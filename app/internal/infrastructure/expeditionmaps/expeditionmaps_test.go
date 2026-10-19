package expeditionmaps_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	"app/internal/core/profile/domain"
	"app/internal/infrastructure/expeditionmaps"
)

var _ = Describe("Load", func() {
	var maps []domain.ExpeditionMap
	byName := func(name string) domain.ExpeditionMap {
		for _, m := range maps {
			if m.Name == name {
				return m
			}
		}
		Fail("no map " + name)
		return domain.ExpeditionMap{}
	}
	spot := func(m domain.ExpeditionMap, name string) domain.Spot {
		for _, s := range m.Spots {
			if s.Name == name {
				return s
			}
		}
		Fail("no place " + name)
		return domain.Spot{}
	}

	BeforeEach(func() {
		var err error
		maps, err = expeditionmaps.Load()
		Expect(err).NotTo(HaveOccurred())
	})

	It("has the three maps that come round every year, in the order of their names in the guide", func() {
		var names []string
		for _, m := range maps {
			names = append(names, m.Name)
		}
		Expect(names).To(Equal([]string{"Parepare Jungle", "Torles Mountains", "Kawrea Volcano"}))
	})

	DescribeTable("is as big as its map, with its start at the bottom, its places and its gates",
		func(name string, width, height, places, locks int) {
			m := byName(name)
			Expect([]int{m.Width, m.Height, len(m.Tiles), len(m.Spots), len(m.Locks)}).To(Equal([]int{width, height, width * height, places, locks}))
			Expect(m.Tiles[m.Start.Y*m.Width+m.Start.X]).To(Equal(domain.Floor))
			Expect(m.Start.Y).To(Equal(height - 1))
		},
		Entry("Parepare", "Parepare Jungle", 18, 20, 10, 2),
		Entry("Torles", "Torles Mountains", 26, 28, 14, 4),
		Entry("Kawrea", "Kawrea Volcano", 26, 28, 10, 3),
	)

	It("puts a place on its letter, with the building there and what it can turn up", func() {
		a := spot(byName("Parepare Jungle"), "A")
		Expect(a.At).To(Equal(domain.Point{X: 0, Y: 0}))
		Expect(a.Loot.Building).To(Equal("Mamoth Statue"))
		Expect(a.Loot.Stat).To(Equal(monster.Int))
		Expect(a.Loot.Mins).To(Equal([]int{0, 100, 200, 300}))
		var pixie domain.Chance
		for _, c := range a.Loot.Bad {
			if c.Item == "Pixie Chips" {
				pixie = c
			}
		}
		Expect(pixie.Price).To(Equal(1000))
		Expect(pixie.Pct).To(Equal([]float64{12.5, 18.8, 18.8, 12.5}))
	})

	It("searches the large trees of Torles by skill, and the rest by intelligence", func() {
		torles := byName("Torles Mountains")
		Expect(spot(torles, "A").Loot.Stat).To(Equal(monster.Skill))
		Expect(spot(torles, "B").Loot.Stat).To(Equal(monster.Int))
	})

	It("gives every place both its tables, a chance for every tier of each item", func() {
		for _, m := range maps {
			for _, s := range m.Spots {
				Expect(s.Loot).NotTo(BeNil(), m.Name+" "+s.Name)
				Expect(s.Loot.Bad).NotTo(BeEmpty(), m.Name+" "+s.Name)
				Expect(s.Loot.Good).NotTo(BeEmpty(), m.Name+" "+s.Name)
				for _, c := range append(append([]domain.Chance{}, s.Loot.Bad...), s.Loot.Good...) {
					Expect(c.Pct).To(HaveLen(len(s.Loot.Mins)), m.Name+" "+s.Name+" "+c.Item)
				}
			}
		}
	})

	It("keeps the intelligence that each gate wants, and every gate tile has one", func() {
		kawrea := byName("Kawrea Volcano")
		wants := map[domain.Point]int{}
		for _, l := range kawrea.Locks {
			wants[l.At] = l.MinINT
		}
		Expect(wants).To(Equal(map[domain.Point]int{{X: 13, Y: 3}: 600, {X: 1, Y: 16}: 600, {X: 14, Y: 19}: 400}))
		for _, m := range maps {
			_, err := domain.PlanRoute(m, domain.Walker{Energy: 700, BreakChance: 1, Intelligence: 999})
			Expect(err).NotTo(HaveOccurred(), m.Name)
		}
	})

	It("has a route to plan on every map, with several places in reach of a monster with the most energy", func() {
		for _, m := range maps {
			for i := range m.Spots {
				m.Spots[i].Value = m.Spots[i].Loot.Worth(999, true)
			}
			r, err := domain.PlanRoute(m, domain.Walker{Energy: 700, BreakChance: 1, Intelligence: 999})
			Expect(err).NotTo(HaveOccurred(), m.Name)
			Expect(len(r.Spots)).To(BeNumerically(">=", 3), m.Name)
		}
	})
})
