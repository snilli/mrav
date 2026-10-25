package routeui

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/profile/domain"
)

var _ = Describe("NewFindView", func() {
	by := domain.Searcher{Intelligence: 100}
	findable := func() View {
		m := aisle()
		m.Width, m.Tiles = 7, make([]domain.Terrain, 7)
		m.Spots = []domain.Spot{
			{Name: "A", At: domain.Point{X: 2}, Building: "Cave", Chances: []domain.Chance{{Item: "Gem", Pct: 20}, {Item: "Nuts", Pct: 80}}},
			{Name: "B", At: domain.Point{X: 4}, Building: "Tree", Chances: []domain.Chance{{MinStat: 300, Item: "Gem", Pct: 100}}},
			{Name: "C", At: domain.Point{X: 5}},
			{Name: "D", At: domain.Point{X: 6}, Building: "Pond", Chances: []domain.Chance{{Item: "Nuts", Pct: 90}}},
		}
		return View{Map: m.For(by), Ways: map[string]domain.Way{"A": {Steps: 2, Cost: 12}, "B": {Steps: 4, Cost: 14}, "D": {Steps: 6, Cost: 16.4}}}
	}
	names := func(v FindView) []string {
		var out []string
		for _, p := range v.Places {
			out = append(out, p.Name)
		}
		return out
	}

	It("lists every item that a search on the map may turn up, once each, by name", func() {
		Expect(NewFindView(findable(), nil, by).Items).To(Equal([]string{"Gem", "Nuts"}))
	})

	It("lists nothing more when nothing is picked", func() {
		v := NewFindView(findable(), nil, by)
		Expect(v.Places).To(BeEmpty())
		Expect(v.NotYet).To(BeEmpty())
		Expect(v.Counts).To(BeEmpty())
	})

	It("says where each picked item turns up for this monster, with the place's name, the energy to get there and search, and the chance", func() {
		v := NewFindView(findable(), map[string]bool{"Gem": true}, by)
		Expect(v.Places).To(Equal([]FindPlace{
			{Name: "A", Building: "Cave", At: domain.Point{X: 2}, Energy: "2 energy to walk", Items: []FindItem{{Item: "Gem", Chance: "20%"}}},
		}))
	})

	It("keeps apart what the monster cannot find there yet, with the stat it needs", func() {
		v := NewFindView(findable(), map[string]bool{"Gem": true}, by)
		Expect(v.NotYet).To(Equal([]string{"Gem at B (Tree) · needs INT 300"}))
		Expect(v.Counts).To(Equal(map[string]int{"A": 1}))
	})

	It("puts first the place that gives the most picked items, then the likeliest, then the nearest", func() {
		v := NewFindView(findable(), map[string]bool{"Gem": true, "Nuts": true}, by)
		Expect(names(v)).To(Equal([]string{"A", "D"}))
		Expect(v.Places[0].Items).To(Equal([]FindItem{{Item: "Gem", Chance: "20%"}, {Item: "Nuts", Chance: "80%"}}))
		Expect(v.Places[1].Energy).To(Equal("6 energy to walk"))

		nuts := NewFindView(findable(), map[string]bool{"Nuts": true}, by)
		Expect(names(nuts)).To(Equal([]string{"D", "A"}), "90% before 80%")

		same := findable()
		same.Map.Spots[3].Loot[0].Pct = 80
		Expect(names(NewFindView(same, map[string]bool{"Nuts": true}, by))).To(Equal([]string{"A", "D"}), "the nearer of two alike")
	})

	It("says when a place is out of reach, and puts it last", func() {
		far := findable()
		delete(far.Ways, "A")
		v := NewFindView(far, map[string]bool{"Gem": true, "Nuts": true}, by)
		Expect(names(v)).To(Equal([]string{"D", "A"}))
		Expect(v.Places[1].Energy).To(Equal("out of reach"))
	})
})
