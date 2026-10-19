package domain_test

import (
	"errors"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/apperr"
	player "app/internal/core/player/domain"
	"app/internal/core/profile/domain"
)

// chart reads a map from rows of text: # is a wall, o a rock, . a floor, S the start, and a letter a place, worth
// what values says. Every map here is made up for the test.
func chart(values map[rune]float64, rows ...string) domain.ExpeditionMap {
	m := domain.ExpeditionMap{Width: len(rows[0]), Height: len(rows)}
	for y, row := range rows {
		for x, c := range row {
			kind := domain.Floor
			switch c {
			case '#':
				kind = domain.Wall
			case 'o':
				kind = domain.Rock
			case 'g':
				kind = domain.Gate
			case 'T':
				kind = domain.Teleport
			case 'S':
				m.Start = domain.Point{X: x, Y: y}
			case '.':
			default:
				m.Spots = append(m.Spots, domain.Spot{Name: string(c), At: domain.Point{X: x, Y: y}, Value: values[c]})
			}
			m.Tiles = append(m.Tiles, kind)
		}
	}
	return m
}

// plan is PlanRoute for a walker with the energy and the chance of breaking a rock given, and so much intelligence that
// no gate is shut to it.
func plan(m domain.ExpeditionMap, energy int, chance float64) (domain.Route, error) {
	return domain.PlanRoute(m, domain.Walker{Energy: energy, BreakChance: chance, Intelligence: 999})
}

func names(r domain.Route) []string {
	var out []string
	for _, s := range r.Spots {
		out = append(out, s.Name)
	}
	return out
}

var _ = Describe("PlanRoute", func() {
	value := map[rune]float64{'A': 3, 'B': 4}
	open := chart(value,
		"S.A",
		"...",
		".B.",
	)

	It("visits every place when the energy is enough, and comes back to the start", func() {
		r, err := plan(open, 28, 1)
		Expect(err).NotTo(HaveOccurred())
		Expect(names(r)).To(ConsistOf("A", "B"))
		Expect(r.Value).To(BeNumerically("==", 7))
		Expect(r.Cost).To(BeNumerically("==", 28)) // 8 steps and 2 searches of 10
		Expect(r.Path).To(HaveLen(9))
		Expect(r.Path[0]).To(Equal(open.Start))
		Expect(r.Path[len(r.Path)-1]).To(Equal(open.Start))
	})

	It("goes to the places in the order of the map when either order costs the same", func() {
		r, err := plan(open, 28, 1)
		Expect(err).NotTo(HaveOccurred())
		Expect(names(r)).To(Equal([]string{"A", "B"}))
	})

	DescribeTable("takes the places that are worth most for the energy there is",
		func(energy int, want []string, cost float64) {
			r, err := plan(open, energy, 1)
			Expect(err).NotTo(HaveOccurred())
			Expect(names(r)).To(ConsistOf(want)) // either order of the two costs the same
			Expect(r.Cost).To(BeNumerically("==", cost))
		},
		Entry("both, with the last unit of it", 28, []string{"A", "B"}, 28.0),
		Entry("the dearer one when only one is in reach of it", 20, []string{"B"}, 16.0),
		Entry("the nearer one when it is all that is in reach", 14, []string{"A"}, 14.0),
		Entry("nothing when none is", 13, []string(nil), 0.0),
	)

	It("takes the nearer of two places that are worth the same, when only one can be had", func() {
		r, err := plan(chart(map[rune]float64{'A': 5, 'B': 5}, "S.A", "...", ".B."), 20, 1)
		Expect(err).NotTo(HaveOccurred())
		Expect(names(r)).To(Equal([]string{"A"}))
	})

	It("goes round a wall", func() {
		r, err := plan(chart(value,
			"S#A",
			".#.",
			"...",
		), 22, 1)
		Expect(err).NotTo(HaveOccurred())
		Expect(names(r)).To(Equal([]string{"A"}))
		Expect(r.Cost).To(BeNumerically("==", 22)) // 6 steps each way and a search
	})

	It("leaves out a place that cannot be reached", func() {
		r, err := plan(chart(value, "S#A"), 100, 1)
		Expect(err).NotTo(HaveOccurred())
		Expect(r.Spots).To(BeEmpty())
		Expect(r.Path).To(BeEmpty())
	})

	DescribeTable("counts a rock at what it costs on average to break it, each time it is crossed",
		func(chance, cost float64) {
			r, err := plan(chart(value, "SoA"), 100, chance)
			Expect(err).NotTo(HaveOccurred())
			Expect(names(r)).To(Equal([]string{"A"}))
			Expect(r.Cost).To(BeNumerically("~", cost, 1e-9)) // 4 steps, a search, and the rock twice
		},
		Entry("when it always breaks: 5 a time", 1.0, 24.0),
		Entry("when it breaks half the time: 10 a time", 0.5, 34.0),
	)

	It("goes round a rock when that is cheaper than breaking it", func() {
		r, err := plan(chart(value,
			"SoA",
			"...",
			"...",
		), 100, 1)
		Expect(err).NotTo(HaveOccurred())
		Expect(r.Cost).To(BeNumerically("==", 18)) // 4 steps each way, round it, and a search
	})

	Describe("a gate that wants intelligence", func() {
		shortcut := func() domain.ExpeditionMap {
			m := chart(value, "SgA")
			m.Locks = []domain.Lock{{At: domain.Point{X: 1}, MinINT: 300}}
			return m
		}

		It("is crossed by a walker with the intelligence it wants, for what it takes on average to find the way", func() {
			r, err := domain.PlanRoute(shortcut(), domain.Walker{Energy: 100, BreakChance: 1, Intelligence: 300})
			Expect(err).NotTo(HaveOccurred())
			Expect(names(r)).To(Equal([]string{"A"}))
			// 35% to find it in a search of 10: 1 + 10/0.35 to cross, twice, then the 2 steps onto the place and
			// the start, and a search
			Expect(r.Cost).To(BeNumerically("~", 2*(1+10/0.35)+2+10, 1e-9))
		})

		It("is shut to a walker with less", func() {
			r, err := domain.PlanRoute(shortcut(), domain.Walker{Energy: 1000, BreakChance: 1, Intelligence: 299})
			Expect(err).NotTo(HaveOccurred())
			Expect(r.Spots).To(BeEmpty())
		})

		It("is gone round when the way round is cheaper", func() {
			m := chart(value,
				"SgA",
				"...",
			)
			m.Locks = []domain.Lock{{At: domain.Point{X: 1}, MinINT: 300}}
			r, err := domain.PlanRoute(m, domain.Walker{Energy: 100, BreakChance: 1, Intelligence: 300})
			Expect(err).NotTo(HaveOccurred())
			Expect(r.Cost).To(BeNumerically("==", 4+4+10)) // 4 steps each way, round it
		})

		It("has to be given its intelligence, or the map is not one", func() {
			m := chart(value, "SgA")
			_, err := domain.PlanRoute(m, domain.Walker{Energy: 100, BreakChance: 1, Intelligence: 999})
			Expect(errors.Is(err, apperr.ErrInvalid)).To(BeTrue())
		})
	})

	Describe("from where the walker stands", func() {
		// S . . A . B: the start, then the place A three tiles on, and the place B five.
		line := func() domain.ExpeditionMap { return chart(value, "S..A.B") }

		It("plans from the tile that the walker is on, and still ends at the start", func() {
			at := domain.Point{X: 4}
			r, err := domain.PlanRoute(line(), domain.Walker{Energy: 16, BreakChance: 1, At: &at})
			Expect(err).NotTo(HaveOccurred())
			Expect(names(r)).To(Equal([]string{"B"})) // 1 step to B, a search, and 5 steps back to the start
			Expect(r.Cost).To(BeNumerically("==", 16))
			Expect(r.Path[0]).To(Equal(at))
			Expect(r.Path[len(r.Path)-1]).To(Equal(domain.Point{X: 0}))

			r, err = domain.PlanRoute(line(), domain.Walker{Energy: 16, BreakChance: 1})
			Expect(err).NotTo(HaveOccurred())
			Expect(names(r)).To(Equal([]string{"A"})) // from the start, B is 5 steps and a search and 5 back: 20
		})

		It("is not for a tile that is not on the map", func() {
			at := domain.Point{X: 9}
			_, err := domain.PlanRoute(line(), domain.Walker{Energy: 100, BreakChance: 1, At: &at})
			Expect(errors.Is(err, apperr.ErrInvalid)).To(BeTrue())
		})
	})

	Describe("Here", func() {
		// The trip of Parepare comes round in October, week 1: the date 24 of the game's count. Every trip starts on the
		// tile 975 of the game's count, and one step right of it is the tile 976, which is one step right of the start
		// of the map.
		trip := func(rows ...string) domain.ExpeditionMap {
			m := chart(value, rows...)
			m.Name = "Parepare Jungle"
			return m
		}
		at := func(date, tile int) player.Ranch {
			return player.Ranch{Known: true, Date: date, Tile: tile, WalkKnown: true}
		}

		It("is whether the week of the trip has come and the monster last stood on a tile of the map that it can walk", func() {
			Expect(trip("S.A").Here(at(24, 976))).To(BeTrue())
		})

		DescribeTable("is not, otherwise",
			func(m domain.ExpeditionMap, r player.Ranch) { Expect(m.Here(r)).To(BeFalse()) },
			Entry("in the week before", trip("S.A"), at(23, 976)),
			Entry("in the week after", trip("S.A"), at(25, 976)),
			Entry("with the tile on a wall", trip("S#A"), at(24, 976)),
			Entry("with the tile outside the map", trip("S.A"), at(24, 25*32+15)),
			Entry("where the place on the map is not known", trip("S.A"), player.Ranch{Known: true, Date: 24, Tile: 976}),
			Entry("where the date is not known", trip("S.A"), player.Ranch{Tile: 976, WalkKnown: true}),
			Entry("for a map that no trip is of", func() domain.ExpeditionMap { m := trip("S.A"); m.Name = "Nowhere"; return m }(), at(24, 976)),
		)

		It("stands the monster on the tile of the map that is as far from its start as the monster is from the start of the trip", func() {
			m := trip("S.A")
			at0, ok := m.Stand(at(24, 975))
			Expect(ok).To(BeTrue())
			Expect(at0).To(Equal(domain.Point{X: 0, Y: 0}))
			at2, ok := m.Stand(at(24, 977))
			Expect(ok).To(BeTrue())
			Expect(at2).To(Equal(domain.Point{X: 2, Y: 0}))

			up := chart(value, ".", ".", "S")
			up.Name = "Parepare Jungle"
			at1, ok := up.Stand(at(24, 943)) // a step up of the start
			Expect(ok).To(BeTrue())
			Expect(at1).To(Equal(domain.Point{X: 0, Y: 1}))
		})
	})

	It("does not walk through a teleport, which leads somewhere at random", func() {
		r, err := plan(chart(value, "STA"), 1000, 1)
		Expect(err).NotTo(HaveOccurred())
		Expect(r.Spots).To(BeEmpty())
	})

	It("cannot cross a rock that never breaks", func() {
		r, err := plan(chart(value, "SoA"), 100, 0)
		Expect(err).NotTo(HaveOccurred())
		Expect(r.Spots).To(BeEmpty())
	})

	It("refuses a map that is not what it says it is", func() {
		bad := open
		bad.Tiles = bad.Tiles[:4]
		_, err := plan(bad, 100, 1)
		Expect(errors.Is(err, apperr.ErrInvalid)).To(BeTrue())

		outside := open
		outside.Start = domain.Point{X: 9, Y: 9}
		_, err = plan(outside, 100, 1)
		Expect(errors.Is(err, apperr.ErrInvalid)).To(BeTrue())
	})

	It("refuses a map with more places than it can weigh", func() {
		row := "S" + strings.Repeat("A", 17)
		_, err := plan(chart(map[rune]float64{'A': 1}, row), 1000, 1)
		Expect(errors.Is(err, apperr.ErrInvalid)).To(BeTrue())
	})
})
