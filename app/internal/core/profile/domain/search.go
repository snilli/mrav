package domain

import (
	"cmp"
	"slices"
)

type Chance struct {
	Good    bool
	MinStat int
	Slot    int
	Item    string
	Pct     float64
	Price   int
}

func (s Spot) Searched(by Searcher) (value float64, loot []Drop) {
	stat := by.Intelligence
	if s.BySkill {
		stat = by.Skill
	}
	tier, found := 0, false
	for _, c := range s.Chances {
		if c.Good == by.GoodNature && c.MinStat <= stat && (!found || c.MinStat > tier) {
			tier, found = c.MinStat, true
		}
	}
	if !found {
		return 0, nil
	}

	var rows []Chance
	for _, c := range s.Chances {
		if c.Good == by.GoodNature && c.MinStat == tier {
			rows = append(rows, c)
			value += c.Pct / 100 * float64(c.Price)
		}
	}
	slices.SortStableFunc(rows, func(a, b Chance) int {
		return cmp.Or(cmp.Compare(b.Pct, a.Pct), cmp.Compare(a.Slot, b.Slot))
	})
	for _, c := range rows {
		if c.Pct > 0 {
			loot = append(loot, Drop{Item: c.Item, Pct: c.Pct, Price: c.Price})
		}
	}
	return value, loot
}

func (m ExpeditionMap) For(by Searcher) ExpeditionMap {
	spots := make([]Spot, len(m.Spots))
	for i, s := range m.Spots {
		s.Value, s.Loot = s.Searched(by)
		spots[i] = s
	}
	m.Spots = spots
	return m
}
