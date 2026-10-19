package domain

import (
	"slices"

	monster "app/internal/core/monster/domain"
)

// Condition is one requirement of an unlock rule.
type Condition interface {
	Holds(m monster.Monster) bool
	// Require says the condition in words, with what the monster has.
	Require(m monster.Monster, n Names) Requirement
}

// StatSumAtLeast needs the named stats to add up to Min.
type StatSumAtLeast struct {
	Stats []monster.Stat
	Min   int
}

func (c StatSumAtLeast) Holds(m monster.Monster) bool { return m.StatSum(c.Stats...) >= c.Min }

// MoveUsedAtLeast needs the move in memory slot Slot to have been used Min times.
type MoveUsedAtLeast struct{ Slot, Min int }

func (c MoveUsedAtLeast) Holds(m monster.Monster) bool {
	return c.Slot >= 0 && c.Slot < monster.SlotCount && m.Used[c.Slot] >= c.Min
}

// NatureAtLeast needs the effective nature to be Min or more.
type NatureAtLeast struct{ Min int }

func (c NatureAtLeast) Holds(m monster.Monster) bool { return m.Nature >= c.Min }

// NatureAtMost needs the effective nature to be Max or less.
type NatureAtMost struct{ Max int }

func (c NatureAtMost) Holds(m monster.Monster) bool { return m.Nature <= c.Max }

// SubGenusIn needs the sub-genus to be one of IDs.
type SubGenusIn struct{ IDs []monster.Genus }

func (c SubGenusIn) Holds(m monster.Monster) bool { return slices.Contains(c.IDs, m.SubGenus) }

// SubGenusNotIn needs the sub-genus to be none of IDs.
type SubGenusNotIn struct{ IDs []monster.Genus }

func (c SubGenusNotIn) Holds(m monster.Monster) bool { return !slices.Contains(c.IDs, m.SubGenus) }

// UnlockRule says when a monster may learn a move. With no conditions and Never unset the move is
// learnable from the start; Never means the game gives no way to learn it.
type UnlockRule struct {
	Never bool
	All   []Condition
}

// Holds reports whether the monster meets every condition.
func (r UnlockRule) Holds(m monster.Monster) bool {
	if r.Never {
		return false
	}
	for _, c := range r.All {
		if !c.Holds(m) {
			return false
		}
	}
	return true
}
