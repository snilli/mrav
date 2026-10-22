package domain

import (
	"slices"

	monster "app/internal/core/monster/domain"
)

type Condition interface {
	Holds(m monster.Monster) bool
	Require(m monster.Monster, n Names) Requirement
}

type StatSumAtLeast struct {
	Stats []monster.Stat
	Min   int
}

func (c StatSumAtLeast) Holds(m monster.Monster) bool { return m.StatSum(c.Stats...) >= c.Min }

type MoveUsedAtLeast struct{ Slot, Min int }

func (c MoveUsedAtLeast) Holds(m monster.Monster) bool {
	return c.Slot >= 0 && c.Slot < monster.SlotCount && m.Used[c.Slot] >= c.Min
}

type NatureAtLeast struct{ Min int }

func (c NatureAtLeast) Holds(m monster.Monster) bool { return m.Nature >= c.Min }

type NatureAtMost struct{ Max int }

func (c NatureAtMost) Holds(m monster.Monster) bool { return m.Nature <= c.Max }

type SubGenusIn struct{ IDs []monster.Genus }

func (c SubGenusIn) Holds(m monster.Monster) bool { return slices.Contains(c.IDs, m.SubGenus) }

type SubGenusNotIn struct{ IDs []monster.Genus }

func (c SubGenusNotIn) Holds(m monster.Monster) bool { return !slices.Contains(c.IDs, m.SubGenus) }

type UnlockRule struct {
	Never bool
	All   []Condition
}

type Nature int

const (
	NatureNone Nature = iota
	NatureGood
	NatureBad
)

func (r UnlockRule) Nature() Nature {
	for _, c := range r.All {
		switch c := c.(type) {
		case NatureAtLeast:
			if c.Min > 0 {
				return NatureGood
			}
		case NatureAtMost:
			if c.Max < 0 {
				return NatureBad
			}
		}
	}
	return NatureNone
}

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
