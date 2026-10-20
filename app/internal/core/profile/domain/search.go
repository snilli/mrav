package domain

import (
	"cmp"
	"slices"
)

// Chance is a row of the tables of a place: an item that a search may turn up, and the chance in percent that it does, for
// the tier of the stat that starts at MinStat, in the table for a good nature or for the other. A tier is from its MinStat to
// the next one's. Slot is the row of the table that the chance is in, and Price what the item sells for, 0 where it is not sold.
type Chance struct {
	Good    bool
	MinStat int
	Slot    int
	Item    string
	Pct     float64
	Price   int
}

// Searched is what a search of the place is worth on average, in gold, and what it may turn up, for a monster that searches
// as the searcher does: the table is the one for its nature, the tier is the highest that its stat reaches, which is its
// skill where the building goes by skill and its intelligence where it does not, and each item is worth its chance by its
// price. What may turn up is the likeliest first, and by the row of the table where it is as likely, and leaves out what the
// tier never turns up.
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

// For is the map with what a search of each of its places is worth and may turn up for the searcher. The map itself, which many
// may share, is not written to: its places are copied, in their order.
func (m ExpeditionMap) For(by Searcher) ExpeditionMap {
	spots := make([]Spot, len(m.Spots))
	for i, s := range m.Spots {
		s.Value, s.Loot = s.Searched(by)
		spots[i] = s
	}
	m.Spots = spots
	return m
}
