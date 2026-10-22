package domain

import (
	"cmp"
	"math"
	"slices"

	monster "app/internal/core/monster/domain"
	movedomain "app/internal/core/move/domain"
)

type Learn struct {
	Move   movedomain.Move
	Chance float64
}

func (l Learn) Shown() bool { return math.Round(l.Chance*10) > 0 }

func (l Learn) Sure() bool { return math.Round(l.Chance*10) >= 1000 }

func LearnChances(m monster.Monster, moves []movedomain.Move) [len(errantryLocales)][]Learn {
	type entry struct {
		move    movedomain.Move
		acquire movedomain.Acquire
	}
	var byPlace [len(errantryLocales)][]entry
	for _, mv := range moves {
		for _, a := range mv.Acquire {
			if a.Place >= 1 && a.Place <= len(byPlace) {
				byPlace[a.Place-1] = append(byPlace[a.Place-1], entry{mv, a})
			}
		}
	}

	loyalty := Loyalty(m.Spoil, m.Fear)
	var out [len(errantryLocales)][]Learn
	for place, entries := range byPlace {
		slices.SortStableFunc(entries, func(a, b entry) int { return cmp.Compare(a.acquire.Priority, b.acquire.Priority) })
		left := 100.0
		for _, e := range entries {
			if learned(m, e.move.MemorySlot) || !meets(m, e.acquire) {
				continue
			}
			chance := left
			for _, g := range e.acquire.Gates {
				chance = min(chance, gateChance(float64(loyalty+m.StatSum(g.Stats...)), g, e.acquire.Auto, left))
			}
			left -= chance
			out[place] = append(out[place], Learn{Move: e.move, Chance: chance})
		}
	}
	return out
}

func learned(m monster.Monster, slot int) bool {
	return slot >= 0 && slot < monster.SlotCount && m.Learned[slot]
}

func meets(m monster.Monster, a movedomain.Acquire) bool {
	switch {
	case len(a.SubIn) > 0 && !(movedomain.SubGenusIn{IDs: a.SubIn}).Holds(m):
		return false
	case len(a.SubNotIn) > 0 && !(movedomain.SubGenusNotIn{IDs: a.SubNotIn}).Holds(m):
		return false
	case a.Nature > 0 && !(movedomain.NatureAtLeast{Min: a.Nature}).Holds(m):
		return false
	case a.Nature < 0 && !(movedomain.NatureAtMost{Max: a.Nature}).Holds(m):
		return false
	case a.Chain != nil && !(movedomain.MoveUsedAtLeast{Slot: a.Chain.Slot, Min: a.Chain.Uses}).Holds(m):
		return false
	}
	return true
}

func gateChance(total float64, g movedomain.Gate, auto int, left float64) float64 {
	switch {
	case total < float64(g.Min):
		return left * float64(auto) / 100
	case total < float64(g.Max):
		f := (total - float64(g.Min)) / float64(g.Max-g.Min)
		return f*left + (1-f)*float64(auto)
	}
	return left
}
