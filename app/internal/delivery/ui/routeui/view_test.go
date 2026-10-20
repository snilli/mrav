package routeui

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	player "app/internal/core/player/domain"
	"app/internal/core/profile/domain"
	"app/internal/delivery/ui/profileui"
)

// monsterOf is a profile whose monster has the life and the power given, and no intelligence, skill or nature.
func monsterOf(life, pow int) domain.Profile { return monsterWith(life, pow, 0, 0, 0) }

// monsterWith is a profile whose monster has the stats and the nature given.
func monsterWith(life, pow, intelligence, skill, nature int) domain.Profile {
	var m monster.Monster
	m.Stats[monster.Life], m.Stats[monster.Pow] = life, pow
	m.Stats[monster.Int], m.Stats[monster.Skill], m.Nature = intelligence, skill, nature
	return domain.Profile{Monster: m}
}

// aisle is a made-up map of one row: the start, a floor, a place worth 1,000 G, and a wall.
func aisle() domain.ExpeditionMap {
	return domain.ExpeditionMap{
		Name: "Aisle", Width: 4, Height: 1,
		Tiles: []domain.Terrain{domain.Floor, domain.Floor, domain.Floor, domain.Wall},
		Spots: []domain.Spot{{Name: "A", At: domain.Point{X: 2}, Value: 1000}},
	}
}

// trip is a made-up map with the name of a trip, so that the week of that trip finds it: a corridor of 6 tiles that
// goes up from the start, with a place worth 1,000 G at the top.
func trip(name string) domain.ExpeditionMap {
	return domain.ExpeditionMap{
		Name: name, Width: 1, Height: 6, Tiles: make([]domain.Terrain, 6), Start: domain.Point{Y: 5},
		Spots: []domain.Spot{{Name: "A", At: domain.Point{}, Value: 1000}},
	}
}

var _ = Describe("NewView", func() {
	It("plans the route with the energy the monster starts with and the chance that its power breaks a rock", func() {
		v, err := NewView(monsterOf(140, 350), aisle())
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Name).To(Equal("Aisle"))
		Expect(v.Summary).To(Equal("140 energy, rocks break 50%, INT 0, bad nature: 1 place, about 14 energy, worth about 1,000 G"))
		Expect(v.Stops).To(Equal([]string{"1. A · 1,000 G"}))
		Expect(v.Route.Spots).To(HaveLen(1))
	})

	It("takes more energy from more life, and a better chance from more power", func() {
		v, err := NewView(monsterOf(700, 500), aisle())
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Summary).To(HavePrefix("700 energy, rocks break 100%, INT 0, bad nature: "))
	})

	It("says that no place is in reach when the energy cannot get to one and back", func() {
		m := aisle()
		m.Width, m.Tiles = 400, append([]domain.Terrain{domain.Floor, domain.Floor}, make([]domain.Terrain, 398)...)
		m.Spots = []domain.Spot{{Name: "A", At: domain.Point{X: 399}, Value: 1000}}
		v, err := NewView(monsterOf(140, 350), m)
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Summary).To(Equal("140 energy, rocks break 50%, INT 0, bad nature: no place is in reach"))
		Expect(v.Stops).To(BeEmpty())
	})

	It("numbers the stops in the order to go", func() {
		m := aisle()
		m.Width, m.Tiles = 6, make([]domain.Terrain, 6)
		m.Spots = []domain.Spot{{Name: "A", At: domain.Point{X: 2}, Value: 1500}, {Name: "B", At: domain.Point{X: 5}, Value: 2500}}
		v, err := NewView(monsterOf(700, 500), m)
		Expect(err).NotTo(HaveOccurred())
		Expect(strings.Join(v.Stops, "|")).To(Equal("1. A · 1,500 G|2. B · 2,500 G"))
	})

	It("names the building of a place beside its letter in the stops, where it has one", func() {
		m := aisle()
		m.Spots = []domain.Spot{{Name: "A", At: domain.Point{X: 2}, Building: "Cave", Value: 200}}
		v, err := NewView(monsterOf(140, 350), m)
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Stops).To(Equal([]string{"1. A · Cave · 200 G"}))
		Expect(v.Summary).To(ContainSubstring("worth about 200 G"))
	})

	It("says that the nature is good above 0, which the game takes it to be", func() {
		v, err := NewView(monsterWith(140, 350, 0, 0, 1), aisle())
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Summary).To(ContainSubstring("good nature"))
		v, err = NewView(monsterWith(140, 350, 0, 0, 0), aisle())
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Summary).To(ContainSubstring("bad nature"))
	})

	It("takes the way through a gate when the intelligence of the monster finds it", func() {
		m := aisle()
		m.Tiles = []domain.Terrain{domain.Floor, domain.Gate, domain.Floor, domain.Wall}
		m.Locks = []domain.Lock{{At: domain.Point{X: 1}, MinINT: 300}}
		v, err := NewView(monsterWith(700, 350, 300, 0, 0), m)
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Route.Spots).To(HaveLen(1))
		Expect(v.Intelligence).To(Equal(300))

		v, err = NewView(monsterWith(700, 350, 299, 0, 0), m)
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Route.Spots).To(BeEmpty())
	})

	It("marks where the monster stands when it is on the trip of the map, and nowhere otherwise", func() {
		m := aisle()
		m.Name = "Parepare Jungle" // its trip comes round in October, week 1: the date 24 of the game's count
		on := func(date int) domain.Profile {
			p := monsterOf(140, 350)
			p.Player.Ranch = player.Ranch{Known: true, Date: date, Tile: 976, WalkKnown: true}
			return p
		}
		v, err := NewView(on(24), m)
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Marker).To(Equal(&domain.Point{X: 1, Y: 0}))

		v, err = NewView(on(25), m)
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Marker).To(BeNil())
	})

	Describe("on the trip", func() {
		// The aisle as the map of Parepare, whose trip comes round at the date 24 of the game's count, with the monster
		// on the tile 976, one step right of the start, which is the one between the start and the place.
		parepare := func() domain.ExpeditionMap {
			m := aisle()
			m.Name = "Parepare Jungle"
			return m
		}
		walking := func(steps, energy int, known bool) domain.Profile {
			p := monsterOf(700, 350)
			p.Player.Ranch = player.Ranch{Known: true, Date: 24, Tile: 976, WalkKnown: true, Steps: steps, Energy: energy, EnergyKnown: known}
			return p
		}

		It("plans from the tile of the monster with the energy that it has left, once it has made steps", func() {
			v, err := NewView(walking(3, 20, true), parepare())
			Expect(err).NotTo(HaveOccurred())
			Expect(v.Summary).To(Equal("20 energy left, at X 1, Y 0, rocks break 50%, INT 0, bad nature: 1 place, about 13 energy, worth about 1,000 G"))
			Expect(v.Route.Path[0]).To(Equal(domain.Point{X: 1}))
		})

		It("says that no place is in reach from where it is, when its energy is too little", func() {
			v, err := NewView(walking(3, 12, true), parepare())
			Expect(err).NotTo(HaveOccurred())
			Expect(v.Summary).To(Equal("12 energy left, at X 1, Y 0, rocks break 50%, INT 0, bad nature: no place is in reach"))
		})

		It("puts what the panel says of the monster in boxes: its energy, where it is as the game counts it, and its steps", func() {
			v, err := NewView(walking(3, 20, true), parepare())
			Expect(err).NotTo(HaveOccurred())
			Expect(v.Standing).To(Equal([]profileui.Box{
				{Text: "Energy 20", Tone: profileui.ToneOrange},
				{Text: "X 1 · Y 0", Tone: profileui.ToneHotPink},
				{Text: "3 steps", Tone: profileui.ToneLightYellow},
			}))
		})

		DescribeTable("colours the energy by how near it is to running out",
			func(energy int, tone profileui.Tone) {
				v, err := NewView(walking(1, energy, true), parepare())
				Expect(err).NotTo(HaveOccurred())
				Expect(v.Standing[0].Tone).To(Equal(tone))
			},
			Entry("plenty", 100, profileui.ToneLightGreen),
			Entry("getting low", 99, profileui.ToneOrange),
			Entry("low", 20, profileui.ToneOrange),
			Entry("nearly out", 19, profileui.ToneOrangeRed),
			Entry("below nothing", -5, profileui.ToneOrangeRed),
		)

		It("leaves out the energy where it is not known, and says one step in the singular", func() {
			v, err := NewView(walking(1, 20, false), parepare())
			Expect(err).NotTo(HaveOccurred())
			Expect(v.Standing).To(Equal([]profileui.Box{
				{Text: "X 1 · Y 0", Tone: profileui.ToneHotPink},
				{Text: "1 step", Tone: profileui.ToneLightYellow},
			}))
		})

		It("has nothing to put in boxes when the monster is not on the trip of the map", func() {
			v, err := NewView(monsterOf(700, 350), parepare())
			Expect(err).NotTo(HaveOccurred())
			Expect(v.Standing).To(BeEmpty())
		})

		DescribeTable("plans from the start with the energy it begins with, when it has made no steps or the energy cannot be its own",
			func(p domain.Profile) {
				v, err := NewView(p, parepare())
				Expect(err).NotTo(HaveOccurred())
				Expect(v.Summary).To(Equal("700 energy, rocks break 50%, INT 0, bad nature: 1 place, about 14 energy, worth about 1,000 G"))
				Expect(v.Marker).NotTo(BeNil())
			},
			Entry("no steps yet", walking(0, 20, true)),
			Entry("more energy than a monster of this life begins with, which is left from another trip", walking(3, 800, true)),
			Entry("the energy is not known", walking(3, 20, false)),
		)
	})

	It("gives the error of a map that is not one", func() {
		bad := aisle()
		bad.Tiles = bad.Tiles[:2]
		_, err := NewView(monsterOf(140, 350), bad)
		Expect(err).To(HaveOccurred())
	})
})
