package domain

import (
	"cmp"
	"slices"

	monster "app/internal/core/monster/domain"
	movedomain "app/internal/core/move/domain"
)

// Learn is a move that a place of the errantry may teach a monster as it is now, and the chance in percent that it does, which is
// 0 when only the stats are short.
type Learn struct {
	Move   movedomain.Move
	Chance float64
}

// LearnChances works out, for each place of the errantry in the order of the game, the moves that it may teach the monster as it
// is now, in the order that the place tries them, each with its chance. A move that is learned, or whose sub-genus, nature or
// chain the monster does not meet, is not listed, and leaves what is left to the next ones. It is calcChance of the Errantry
// calculator of legendcup.com, which works so: a place starts with 100 percent, and each move that it may teach takes its chance
// from what is left. The chance of a move rises with the total of the monster's loyalty and the stats of its gate: below the
// bottom it is the auto learn share of what is left, and from there it climbs to all that is left at the top. A move with two
// gates has the lower of the two chances, and one with none takes all that is left.
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

// learned says whether the monster has learned the move in a memory slot.
func learned(m monster.Monster, slot int) bool {
	return slot >= 0 && slot < monster.SlotCount && m.Learned[slot]
}

// meets says whether the monster is of a sub-genus, has a nature and has used the moves that an entry asks for. The stats are not
// asked here: they only change the chance.
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

// gateChance is the chance, in percent, that a stat total gives of what is left. Between the bottom and the top of the gate the
// auto learn share is not cut down to what is left, as the calculator has it, so the chance jumps at the bottom when little is
// left.
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
