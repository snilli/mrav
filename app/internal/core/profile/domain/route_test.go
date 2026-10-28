package domain_test

import (
	"errors"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/apperr"
	monster "app/internal/core/monster/domain"
	player "app/internal/core/player/domain"
	"app/internal/core/profile/domain"
)

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

func plan(m domain.ExpeditionMap, energy int, chance float64) (domain.Route, error) {
	return domain.PlanRoute(m, domain.Walker{Energy: energy, BreakChance: chance, Intelligence: 999})
}

func walked(m domain.ExpeditionMap, w domain.Walker, r domain.Route) float64 {
	cost := float64(len(r.Spots)) * domain.SearchCost
	for i, p := range r.Path[1:] {
		q := r.Path[i]
		ExpectWithOffset(1, max(p.X-q.X, q.X-p.X)+max(p.Y-q.Y, q.Y-p.Y)).To(Equal(1), m.Name+": the path goes one tile at a time")
		switch m.Tiles[p.Y*m.Width+p.X] {
		case domain.Rock:
			cost += domain.MoveCost + domain.ObstacleCost/w.BreakChance
		case domain.Gate:
			cost += domain.MoveCost + domain.SearchCost/domain.FindChance(w.Intelligence)
		default:
			cost += domain.MoveCost
		}
	}
	return cost
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
		r, err := plan(open, 24, 1)
		Expect(err).NotTo(HaveOccurred())
		Expect(names(r)).To(ConsistOf("A", "B"))
		Expect(r.Value).To(BeNumerically("==", 7))
		Expect(r.Cost).To(BeNumerically("==", 24))
		Expect(r.Path).To(HaveLen(5))
		Expect(r.Path[0]).To(Equal(open.Start))
		Expect(r.Path[len(r.Path)-1]).To(Equal(open.Start))
	})

	It("walks each next leg from the tile it searched from, so its cost is that of walking its path", func() {
		m := chart(map[rune]float64{'A': 3, 'B': 4},
			"S.A.",
			"....",
			"...B",
		)
		w := domain.Walker{Energy: 100, BreakChance: 1, Intelligence: 999}
		r, err := domain.PlanRoute(m, w)
		Expect(err).NotTo(HaveOccurred())
		Expect(names(r)).To(ConsistOf("A", "B"))
		Expect(r.Cost).To(BeNumerically("~", walked(m, w, r), 1e-9))
	})

	It("goes to the places in the order of the map when either order costs the same", func() {
		r, err := plan(chart(value, "A.S.B"), 100, 1)
		Expect(err).NotTo(HaveOccurred())
		Expect(names(r)).To(Equal([]string{"A", "B"}))
	})

	DescribeTable("takes the places that are worth most for the energy there is",
		func(energy int, want []string, cost float64) {
			r, err := plan(open, energy, 1)
			Expect(err).NotTo(HaveOccurred())
			Expect(names(r)).To(ConsistOf(want))
			Expect(r.Cost).To(BeNumerically("==", cost))
		},
		Entry("both, with the last unit of it", 24, []string{"A", "B"}, 24.0),
		Entry("the dearer one when only one is in reach of it", 20, []string{"B"}, 12.0),
		Entry("nothing when none is", 11, []string(nil), 0.0),
	)

	It("takes the nearer one when it is all that is in reach", func() {
		r, err := plan(chart(value, "S.A..B"), 13, 1)
		Expect(err).NotTo(HaveOccurred())
		Expect(names(r)).To(Equal([]string{"A"}))
		Expect(r.Cost).To(BeNumerically("==", 12))
	})

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
		), 20, 1)
		Expect(err).NotTo(HaveOccurred())
		Expect(names(r)).To(Equal([]string{"A"}))
		Expect(r.Cost).To(BeNumerically("==", 20))
	})

	It("leaves out a place that cannot be reached", func() {
		r, err := plan(chart(value, "S#A"), 100, 1)
		Expect(err).NotTo(HaveOccurred())
		Expect(r.Spots).To(BeEmpty())
		Expect(r.Path).To(BeEmpty())
	})

	DescribeTable("counts a rock at what it costs on average to break it, each time it is crossed",
		func(chance, cost float64) {
			r, err := plan(chart(value, "So.A"), 100, chance)
			Expect(err).NotTo(HaveOccurred())
			Expect(names(r)).To(Equal([]string{"A"}))
			Expect(r.Cost).To(BeNumerically("~", cost, 1e-9))
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
		Expect(r.Cost).To(BeNumerically("==", 14))
	})

	Describe("a gate that wants intelligence", func() {
		shortcut := func() domain.ExpeditionMap {
			m := chart(value, "Sg.A")
			m.Locks = []domain.Lock{{At: domain.Point{X: 1}, MinINT: 300}}
			return m
		}

		It("is crossed by a walker with the intelligence it wants, for what it takes on average to find the way", func() {
			r, err := domain.PlanRoute(shortcut(), domain.Walker{Energy: 100, BreakChance: 1, Intelligence: 300})
			Expect(err).NotTo(HaveOccurred())
			Expect(names(r)).To(Equal([]string{"A"}))
			Expect(r.Cost).To(BeNumerically("~", 2*(2+10/0.35)+10, 1e-9))
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
			Expect(r.Cost).To(BeNumerically("==", 2+2+10))
		})

		It("has to be given its intelligence, or the map is not one", func() {
			m := chart(value, "SgA")
			_, err := domain.PlanRoute(m, domain.Walker{Energy: 100, BreakChance: 1, Intelligence: 999})
			Expect(errors.Is(err, apperr.ErrInvalid)).To(BeTrue())
		})
	})

	Describe("from where the walker stands", func() {
		line := func() domain.ExpeditionMap { return chart(value, "S..A.B") }

		It("plans from the tile that the walker is on, and still ends at the start", func() {
			at := domain.Point{X: 4}
			r, err := domain.PlanRoute(line(), domain.Walker{Energy: 16, BreakChance: 1, At: &at})
			Expect(err).NotTo(HaveOccurred())
			Expect(names(r)).To(Equal([]string{"B"}))
			Expect(r.Cost).To(BeNumerically("==", 14))
			Expect(r.Path[0]).To(Equal(at))
			Expect(r.Path[len(r.Path)-1]).To(Equal(domain.Point{X: 0}))

			r, err = domain.PlanRoute(line(), domain.Walker{Energy: 16, BreakChance: 1})
			Expect(err).NotTo(HaveOccurred())
			Expect(names(r)).To(Equal([]string{"A"}))
		})

		It("is not for a tile that is not on the map", func() {
			at := domain.Point{X: 9}
			_, err := domain.PlanRoute(line(), domain.Walker{Energy: 100, BreakChance: 1, At: &at})
			Expect(errors.Is(err, apperr.ErrInvalid)).To(BeTrue())
		})
	})

	Describe("Here", func() {
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
			at1, ok := up.Stand(at(24, 943))
			Expect(ok).To(BeTrue())
			Expect(at1).To(Equal(domain.Point{X: 0, Y: 1}))
		})

		It("counts from the start tile of its own trip, which for the first expedition is row 14, column 8", func() {
			phoenix := chart(value, ".", ".", "S")
			phoenix.Name = "Phoenix Volcano Adventure"
			one := player.Ranch{Known: true, Date: 8, Tile: 13*32 + 8, WalkKnown: true}
			right, up := phoenix.Walked(one)
			Expect([]int{right, up}).To(Equal([]int{0, 1}))
			Expect(phoenix.Here(one)).To(BeTrue())
			at1, ok := phoenix.Stand(one)
			Expect(ok).To(BeTrue())
			Expect(at1).To(Equal(domain.Point{X: 0, Y: 1}))

			right, up = trip("S.A").Walked(player.Ranch{Tile: 976})
			Expect([]int{right, up}).To(Equal([]int{1, 0}), "the other trips start at row 30, column 15")
		})

		It("counts from the game's own words for how far right and up of the start the monster is, whatever tile the trip started on", func() {
			m := trip("...", "...", "S..")
			begun := player.Ranch{Known: true, Date: 24, Tile: 22*32 + 11, WalkKnown: true, OffsetKnown: true}
			right, up := m.Walked(begun)
			Expect([]int{right, up}).To(Equal([]int{0, 0}), "Parepare starts on row 22, column 11, not where Kawrea does")
			at0, ok := m.Stand(begun)
			Expect(ok).To(BeTrue())
			Expect(at0).To(Equal(domain.Point{X: 0, Y: 2}))

			moved := begun
			moved.Tile, moved.OffsetX, moved.OffsetY = 21*32+13, 2, 1
			right, up = m.Walked(moved)
			Expect([]int{right, up}).To(Equal([]int{2, 1}))
			at1, ok := m.Stand(moved)
			Expect(ok).To(BeTrue())
			Expect(at1).To(Equal(domain.Point{X: 2, Y: 1}))
		})

		It("does not stand the monster on the map with the words that a trip leaves behind", func() {
			left := player.Ranch{Known: true, Date: 24, Tile: 35906, WalkKnown: true, OffsetX: 8, OffsetY: 992, OffsetKnown: true}
			m := trip("...", "...", "S..")
			_, ok := m.Stand(left)
			Expect(ok).To(BeFalse())
			Expect(m.Here(left)).To(BeFalse())
		})
	})

	Describe("Open", func() {
		parepare := func() domain.ExpeditionMap {
			m := chart(value, "S.A")
			m.Name = "Parepare Jungle"
			return m
		}
		fit := func(life, fame, rank int) monster.Monster {
			m := monster.Monster{Fame: fame, Rank: rank}
			m.Stats[monster.Life] = life
			return m
		}
		week := func(date int) player.Ranch { return player.Ranch{Known: true, Date: date} }

		It("is whether the week of the trip has come and the monster is fit for it, however little it has more than it asks", func() {
			Expect(parepare().Open(week(24), fit(140, 40, 2))).To(BeTrue())
		})

		It("does not ask the monster to stand on the map, which it does not before the trip starts", func() {
			Expect(parepare().Open(player.Ranch{Known: true, Date: 24, WalkKnown: false}, fit(300, 60, 4))).To(BeTrue())
		})

		DescribeTable("is not, otherwise",
			func(m domain.ExpeditionMap, r player.Ranch, mon monster.Monster) {
				Expect(m.Open(r, mon)).To(BeFalse())
			},
			Entry("in the week before", parepare(), week(23), fit(300, 60, 4)),
			Entry("in the week after", parepare(), week(25), fit(300, 60, 4)),
			Entry("for a monster short of life", parepare(), week(24), fit(139, 40, 2)),
			Entry("for a monster short of fame", parepare(), week(24), fit(140, 39, 2)),
			Entry("for a monster short of grade", parepare(), week(24), fit(140, 40, 1)),
			Entry("where the date is not known", parepare(), player.Ranch{}, fit(300, 60, 4)),
			Entry("for a map that no trip is of", func() domain.ExpeditionMap { m := parepare(); m.Name = "Nowhere"; return m }(), week(24), fit(300, 60, 4)),
		)

		Describe("with the first expedition", func() {
			named := func(name string) domain.ExpeditionMap {
				m := chart(value, "S.A")
				m.Name = name
				return m
			}
			june := func(known, done bool) player.Ranch {
				return player.Ranch{Known: true, Date: 8, PhoenixKnown: known, PhoenixDone: done}
			}

			It("opens it in its week for a fit monster while the game says it was not made", func() {
				Expect(named("Phoenix Volcano Adventure").Open(june(true, false), fit(140, 40, 1))).To(BeTrue())
				Expect(named("Phoenix Volcano Adventure").Open(june(true, false), fit(139, 40, 1))).To(BeFalse())
			})

			It("does not open it once it was made, or where the game does not say", func() {
				Expect(named("Phoenix Volcano Adventure").Open(june(true, true), fit(300, 60, 4))).To(BeFalse())
				Expect(named("Phoenix Volcano Adventure").Open(june(false, false), fit(300, 60, 4))).To(BeFalse())
			})

			It("shuts the other trips until it was made, and leaves them as they were where the game does not say", func() {
				kawrea := named("Kawrea Volcano")
				Expect(kawrea.Open(june(true, false), fit(300, 60, 4))).To(BeFalse())
				Expect(kawrea.Open(june(true, true), fit(300, 60, 4))).To(BeTrue())
				Expect(kawrea.Open(june(false, false), fit(300, 60, 4))).To(BeTrue())
			})

			Describe("while the monster is still on its trip, which the game marks made as soon as the monster enters the cave", func() {
				onTrip := func(x, y, tile int) player.Ranch {
					r := june(true, true)
					r.Tile, r.WalkKnown = tile, true
					r.OffsetX, r.OffsetY, r.OffsetKnown = x, y, true
					return r
				}

				It("keeps it open and the others shut while the words of the position are those of its trip", func() {
					r := onTrip(-1, 4, 327)
					Expect(domain.PhoenixPending(r)).To(BeTrue())
					Expect(named("Phoenix Volcano Adventure").Open(r, fit(300, 60, 4))).To(BeTrue())
					Expect(named("Kawrea Volcano").Open(r, fit(300, 60, 4))).To(BeFalse())
				})

				It("lets it go when the words are those the last trip left, or of another trip", func() {
					for _, r := range []player.Ranch{onTrip(5, 4, 327), onTrip(-11, 2, 900)} {
						Expect(domain.PhoenixPending(r)).To(BeFalse())
						Expect(named("Phoenix Volcano Adventure").Open(r, fit(300, 60, 4))).To(BeFalse())
						Expect(named("Kawrea Volcano").Open(r, fit(300, 60, 4))).To(BeTrue())
					}
				})

				It("lets it go outside its week", func() {
					r := onTrip(-1, 4, 327)
					r.Date++
					Expect(domain.PhoenixPending(r)).To(BeFalse())
				})
			})
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

var _ = Describe("WaysTo", func() {
	It("costs the walk from the start to each place and its search, leaving out a place it cannot reach", func() {
		m := chart(nil, "S..A#B")
		Expect(m.WaysTo(domain.Walker{Energy: 100})).To(Equal(map[string]domain.Way{"A": {Steps: 2, Cost: 2 + domain.SearchCost}}))
	})

	It("walks from where the monster stands when it is on the map", func() {
		m := chart(nil, "S..A#B")
		at := domain.Point{X: 1}
		Expect(m.WaysTo(domain.Walker{Energy: 100, At: &at})["A"]).To(Equal(domain.Way{Steps: 1, Cost: 1 + domain.SearchCost}))
	})

	It("costs only the search for a place the monster stands next to, as the game searches it from there", func() {
		m := chart(nil, "S..", "..A")
		for _, at := range []domain.Point{{X: 1, Y: 1}, {X: 1}, {X: 2}} {
			Expect(m.WaysTo(domain.Walker{Energy: 100, At: &at})["A"]).To(Equal(domain.Way{Cost: domain.SearchCost}))
		}
	})

	It("counts apart the rocks and the gates on the way, each as one step, while the cost takes what they take on average", func() {
		m := chart(nil, "Sog.A")
		m.Locks = []domain.Lock{{At: domain.Point{X: 2}, MinINT: 300}}
		way := m.WaysTo(domain.Walker{Energy: 100, BreakChance: 0.5, Intelligence: 300})["A"]
		Expect([]int{way.Steps, way.Rocks, way.Gates}).To(Equal([]int{3, 1, 1}))
		Expect(way.Cost).To(BeNumerically("~", 3+5/0.5+10/0.35+domain.SearchCost, 1e-9))
	})
})
