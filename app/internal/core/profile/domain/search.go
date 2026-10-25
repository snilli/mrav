package domain

import (
	"cmp"
	"slices"
)

type Chance struct {
	Good    bool
	MinStat int
	Slot    int
	Code    int
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
			loot = append(loot, Drop{Code: c.Code, Item: c.Item, Pct: c.Pct, Price: c.Price})
		}
	}
	return value, loot
}

func (s Spot) NextTier(by Searcher) (stat string, need int, ok bool) {
	have := by.Intelligence
	stat = "INT"
	if s.BySkill {
		have, stat = by.Skill, "SKI"
	}
	for _, c := range s.Chances {
		if c.Good == by.GoodNature && c.MinStat > have && (!ok || c.MinStat < need) {
			need, ok = c.MinStat, true
		}
	}
	if !ok {
		return "", 0, false
	}
	return stat, need, true
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
