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

type Tier struct {
	From, To int
}

func (t Tier) Holds(stat int) bool {
	return stat >= t.From && (t.To == 0 || stat <= t.To)
}

func (s Spot) Tiers(item string, good bool) []Tier {
	var mins []int
	pct := map[int]float64{}
	for _, c := range s.Chances {
		if c.Good != good {
			continue
		}
		if !slices.Contains(mins, c.MinStat) {
			mins = append(mins, c.MinStat)
		}
		if c.Item == item {
			pct[c.MinStat] += c.Pct
		}
	}
	slices.Sort(mins)
	var out []Tier
	for i, m := range mins {
		if pct[m] <= 0 {
			continue
		}
		to := 0
		if i+1 < len(mins) {
			to = mins[i+1] - 1
		}
		if n := len(out); n > 0 && out[n-1].To == m-1 {
			out[n-1].To = to
			continue
		}
		out = append(out, Tier{From: m, To: to})
	}
	return out
}
