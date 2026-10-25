package routeui

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	player "app/internal/core/player/domain"
	"app/internal/core/profile/domain"
	"app/internal/delivery/ui/profileui"
)

func monsterOf(life, pow int) domain.Profile { return monsterWith(life, pow, 0, 0, 0) }

func monsterWith(life, pow, intelligence, skill, nature int) domain.Profile {
	m := monster.Monster{Fame: 100, Rank: 5}
	m.Stats[monster.Life], m.Stats[monster.Pow] = life, pow
	m.Stats[monster.Int], m.Stats[monster.Skill], m.Nature = intelligence, skill, nature
	return domain.Profile{Monster: m, Player: player.Player{Ranch: player.Ranch{Known: true, Date: 24}}}
}

func aisle() domain.ExpeditionMap {
	return domain.ExpeditionMap{
		Name: "Parepare Jungle", Width: 4, Height: 1,
		Tiles: []domain.Terrain{domain.Floor, domain.Floor, domain.Floor, domain.Wall},
		Spots: []domain.Spot{{Name: "A", At: domain.Point{X: 2}, Value: 1000}},
	}
}

func trip(name string) domain.ExpeditionMap {
	return domain.ExpeditionMap{
		Name: name, Width: 1, Height: 6, Tiles: make([]domain.Terrain, 6), Start: domain.Point{Y: 5},
		Spots: []domain.Spot{{Name: "A", At: domain.Point{}, Value: 1000}},
	}
}

var _ = Describe("NewView", func() {
	Describe("on a trip", func() {
		twoPlaces := func() domain.ExpeditionMap {
			m := aisle()
			m.Width, m.Tiles = 6, []domain.Terrain{domain.Floor, domain.Floor, domain.Floor, domain.Floor, domain.Floor, domain.Wall}
			m.Spots = []domain.Spot{
				{Name: "A", At: domain.Point{X: 2}, Value: 1000, Chances: []domain.Chance{{Code: 46, Item: "Suezo Chips", Pct: 100}}},
				{Name: "B", At: domain.Point{X: 4}, Value: 500},
			}
			return m
		}

		It("keeps a searched place on the route, as it can be searched again, and says what it gave", func() {
			v, err := NewView(monsterOf(140, 350), twoPlaces(), map[string][]int{"A": {46, 46}})
			Expect(err).NotTo(HaveOccurred())
			Expect(v.Route.Spots).To(HaveLen(2))
			Expect(v.Places[0].Name).To(Equal("A"))
			Expect(v.Places[0].Stop).To(Equal(1))
			Expect(v.Places[0].Got).To(Equal([]string{"Suezo Chips", "Suezo Chips"}))
		})

		It("names an item it does not know by its code", func() {
			v, err := NewView(monsterOf(140, 350), twoPlaces(), map[string][]int{"B": {99}})
			Expect(err).NotTo(HaveOccurred())
			Expect(v.Places[1].Got).To(Equal([]string{"item #99"}))
		})
	})

	DescribeTable("says what the stat a place goes by must reach for more",
		func(intelligence int, need string) {
			m := aisle()
			m.Spots[0].Chances = []domain.Chance{{MinStat: 100, Item: "Nuts", Pct: 100}, {MinStat: 300, Item: "Gem", Pct: 100}}
			p := monsterWith(140, 350, intelligence, 0, 0)
			m = m.For(domain.Searcher{Intelligence: intelligence})
			v, err := NewView(p, m, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(v.Places[0].Need).To(Equal(need))
		},
		Entry("too low to find anything", 50, "needs INT 100"),
		Entry("high enough for the first table only", 150, "INT 300 for more"),
		Entry("at the top table", 300, ""),
	)

	It("plans the route with the energy the monster starts with and the chance that its power breaks a rock", func() {
		v, err := NewView(monsterOf(140, 350), aisle(), nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Name).To(Equal("Parepare Jungle"))
		Expect(v.Summary).To(Equal("Starts with 140 energy · rocks break 50% · the route costs about 13 energy"))
		Expect(v.Places).To(Equal([]PlaceView{{Name: "A", Stop: 1, Worth: "about 1,000 G"}}))
		Expect(v.Route.Spots).To(HaveLen(1))
	})

	It("takes more energy from more life, and a better chance from more power", func() {
		v, err := NewView(monsterOf(700, 500), aisle(), nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Summary).To(HavePrefix("Starts with 700 energy · rocks break 100%"))
	})

	It("takes the chance to break a rock from the trip, since the rocks of the first expedition give way at less power", func() {
		first := aisle()
		first.Name = "Phoenix Volcano Adventure"
		v, err := NewView(monsterOf(140, 100), first, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Summary).To(ContainSubstring("rocks break 50%"))

		v, err = NewView(monsterOf(140, 100), aisle(), nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Summary).To(ContainSubstring("rocks break 5%"))
	})

	It("counts X and Y from the start tile of the first expedition, as the game shows them", func() {
		p := monsterOf(700, 350)
		p.Player.Ranch = player.Ranch{Known: true, Date: 8, Tile: 12*32 + 8, WalkKnown: true, Energy: 600, EnergyKnown: true}
		v, err := NewView(p, trip("Phoenix Volcano Adventure"), nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Standing).To(ContainElement(profileui.Box{Text: "X 0 · Y 2", Tone: profileui.ToneHotPink}))
	})

	It("says that no place is in reach when the energy cannot get to one and back", func() {
		m := aisle()
		m.Width, m.Tiles = 400, append([]domain.Terrain{domain.Floor, domain.Floor}, make([]domain.Terrain, 398)...)
		m.Spots = []domain.Spot{{Name: "A", At: domain.Point{X: 399}, Value: 1000}}
		v, err := NewView(monsterOf(140, 350), m, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Summary).To(Equal("Starts with 140 energy · rocks break 50%"), "no cost to give where there is no route")
		Expect(v.Title()).To(Equal("Route · Parepare Jungle: nothing in reach"))
		Expect(v.Places).To(HaveLen(1), "the place is listed, though the route does not go to it")
		Expect(v.Places[0].Stop).To(BeZero())
	})

	It("numbers the stops in the order to go", func() {
		m := aisle()
		m.Width, m.Tiles = 6, make([]domain.Terrain, 6)
		m.Spots = []domain.Spot{{Name: "A", At: domain.Point{X: 2}, Value: 1500}, {Name: "B", At: domain.Point{X: 5}, Value: 2500}}
		v, err := NewView(monsterOf(700, 500), m, nil)
		Expect(err).NotTo(HaveOccurred())
		var got []string
		for _, p := range v.Places {
			got = append(got, fmt.Sprintf("%d. %s · %s", p.Stop, p.Name, p.Worth))
		}
		Expect(strings.Join(got, "|")).To(Equal("1. A · about 1,500 G|2. B · about 2,500 G"))
	})

	It("lists the places that the route leaves out after the ones it takes, so that what they hold can still be read", func() {
		m := aisle()
		m.Width = 6
		m.Tiles = []domain.Terrain{domain.Floor, domain.Floor, domain.Floor, domain.Wall, domain.Floor, domain.Floor}
		m.Spots = []domain.Spot{{Name: "A", At: domain.Point{X: 5}, Value: 2500}, {Name: "B", At: domain.Point{X: 2}, Value: 1500}}
		v, err := NewView(monsterOf(700, 500), m, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Places).To(HaveLen(2))
		Expect(v.Places[0]).To(Equal(PlaceView{Name: "B", Stop: 1, Worth: "about 1,500 G"}))
		Expect(v.Places[1]).To(Equal(PlaceView{Name: "A", Stop: 0, Worth: "about 2,500 G"}))
	})

	It("names the building of a place, where it has one", func() {
		m := aisle()
		m.Spots = []domain.Spot{{Name: "A", At: domain.Point{X: 2}, Building: "Cave", Value: 200}}
		v, err := NewView(monsterOf(140, 350), m, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Places[0].Building).To(Equal("Cave"))
		Expect(v.Title()).To(ContainSubstring("about 200 G"))
	})

	It("gives what a search of a place may turn up, with the chance of each in percent, and marks the items of the story", func() {
		m := aisle()
		m.Spots = []domain.Spot{{Name: "A", At: domain.Point{X: 2}, Value: 200, Loot: []domain.Drop{
			{Item: "Mint Leaf", Pct: 25, Price: 100}, {Item: "Old Sheath", Pct: 18.8, Price: 50}, {Item: "Pixie Chips", Pct: 6.25, Price: 1000},
		}}}
		v, err := NewView(monsterOf(140, 350), m, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Places[0].Loot).To(Equal([]DropView{
			{Item: "Mint Leaf", Chance: "25%"},
			{Item: "Old Sheath", Chance: "18.8%", Story: true},
			{Item: "Pixie Chips", Chance: "6.3%"},
		}))
	})

	It("says nothing of the worth of a place that is worth nothing", func() {
		m := aisle()
		m.Spots = []domain.Spot{{Name: "A", At: domain.Point{X: 2}}}
		v, err := NewView(monsterOf(140, 350), m, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Places[0].Worth).To(BeEmpty())
	})

	It("says nothing of the stats that the map already shows: the INT in its gates and the nature in what a place gives", func() {
		v, err := NewView(monsterWith(140, 350, 944, 0, 1), aisle(), nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Summary).NotTo(ContainSubstring("INT"))
		Expect(v.Summary).NotTo(ContainSubstring("nature"))
	})

	It("takes the way through a gate when the intelligence of the monster finds it", func() {
		m := aisle()
		m.Tiles = []domain.Terrain{domain.Floor, domain.Gate, domain.Floor, domain.Wall}
		m.Locks = []domain.Lock{{At: domain.Point{X: 1}, MinINT: 300}}
		v, err := NewView(monsterWith(700, 350, 300, 0, 0), m, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Route.Spots).To(HaveLen(1))
		Expect(v.Intelligence).To(Equal(300))

		v, err = NewView(monsterWith(700, 350, 299, 0, 0), m, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Route.Spots).To(BeEmpty())
	})

	It("marks where the monster stands when it is on the trip of the map, and nowhere otherwise", func() {
		m := aisle()
		m.Name = "Parepare Jungle"
		on := func(date int) domain.Profile {
			p := monsterOf(140, 350)
			p.Player.Ranch = player.Ranch{Known: true, Date: date, Tile: 976, WalkKnown: true}
			return p
		}
		v, err := NewView(on(24), m, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Marker).To(Equal(&domain.Point{X: 1, Y: 0}))

		v, err = NewView(on(25), m, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Marker).To(BeNil())
	})

	Describe("on the trip", func() {
		parepare := func() domain.ExpeditionMap {
			m := aisle()
			m.Name = "Parepare Jungle"
			return m
		}
		walking := func(energy int, known bool) domain.Profile {
			p := monsterOf(700, 350)
			p.Player.Ranch = player.Ranch{Known: true, Date: 24, Tile: 976, WalkKnown: true, Energy: energy, EnergyKnown: known}
			return p
		}

		It("plans from the tile of the monster with the energy that it has left, once it has used some", func() {
			v, err := NewView(walking(20, true), parepare(), nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(v.Summary).To(Equal("Rocks break 50% · the route costs about 12 energy"), "the energy it has and where it is are in the boxes")
			Expect(v.Route.Path[0]).To(Equal(domain.Point{X: 1}))
		})

		It("says that no place is in reach from where it is, when its energy is too little", func() {
			v, err := NewView(walking(11, true), parepare(), nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(v.Summary).To(Equal("Rocks break 50%"))
		})

		It("puts what the game shows of the monster in boxes: its energy and where it is as the game counts it", func() {
			v, err := NewView(walking(20, true), parepare(), nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(v.Standing).To(Equal([]profileui.Box{
				{Text: "Energy 20", Tone: profileui.ToneOrange},
				{Text: "X 1 · Y 0", Tone: profileui.ToneHotPink},
			}))
		})

		DescribeTable("colours the energy by how near it is to running out",
			func(energy int, tone profileui.Tone) {
				v, err := NewView(walking(energy, true), parepare(), nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(v.Standing[0].Tone).To(Equal(tone))
			},
			Entry("plenty", 100, profileui.ToneLightGreen),
			Entry("getting low", 99, profileui.ToneOrange),
			Entry("low", 20, profileui.ToneOrange),
			Entry("nearly out", 19, profileui.ToneOrangeRed),
			Entry("below nothing", -5, profileui.ToneOrangeRed),
		)

		It("leaves out the energy where the energy is not known", func() {
			v, err := NewView(walking(20, false), parepare(), nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(v.Standing).To(Equal([]profileui.Box{
				{Text: "X 1 · Y 0", Tone: profileui.ToneHotPink},
			}))
		})

		It("has nothing to put in boxes when the monster is not on the trip of the map", func() {
			v, err := NewView(monsterOf(700, 350), parepare(), nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(v.Standing).To(BeEmpty())
		})

		DescribeTable("plans from the start with the energy it begins with, when it has used none or the energy cannot be its own",
			func(p domain.Profile) {
				v, err := NewView(p, parepare(), nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(v.Summary).To(Equal("Starts with 700 energy · rocks break 50% · the route costs about 13 energy"))
				Expect(v.Marker).NotTo(BeNil())
			},
			Entry("no energy used yet", walking(700, true)),
			Entry("more energy than a monster of this life begins with, which is left from another trip", walking(800, true)),
			Entry("the energy is not known", walking(20, false)),
		)
	})

	It("says where each teleport leads, in words, and says nothing of the places where a teleport lands", func() {
		m := aisle()
		m.Portals = []domain.Portal{
			{Label: "T1", At: domain.Point{X: 1}, To: []string{"E1", "E2", "E3"}},
			{Label: "T2", At: domain.Point{X: 2}, To: []string{"E4"}},
			{Label: "T3", At: domain.Point{X: 3}, To: []string{"E3", "T4"}},
			{Label: "E1", At: domain.Point{X: 0}},
		}
		v, err := NewView(monsterOf(140, 350), m, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Teleports).To(Equal([]string{
			"T1 takes you to E1, E2 or E3, at random",
			"T2 takes you to E4",
			"T3 takes you to E3 or T4, at random",
		}))
	})

	It("has no word of teleports for a map that has none", func() {
		v, err := NewView(monsterOf(140, 350), aisle(), nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Teleports).To(BeEmpty())
	})

	It("gives the error of a map that is not one", func() {
		bad := aisle()
		bad.Tiles = bad.Tiles[:2]
		_, err := NewView(monsterOf(140, 350), bad, nil)
		Expect(err).To(HaveOccurred())
	})
})

var _ = DescribeTable("walkText says the steps to a place, and the rocks and gates on the way apart, as one may be broken or open already",
	func(way domain.Way, ok bool, want string) { Expect(walkText(way, ok)).To(Equal(want)) },
	Entry("out of reach", domain.Way{}, false, "out of reach"),
	Entry("next to it", domain.Way{Cost: domain.SearchCost}, true, "right here"),
	Entry("a few steps", domain.Way{Steps: 2}, true, "2 energy to walk"),
	Entry("over a rock", domain.Way{Steps: 3, Rocks: 1}, true, "3 energy to walk · 1 rock"),
	Entry("over rocks and a gate", domain.Way{Steps: 9, Rocks: 2, Gates: 1}, true, "9 energy to walk · 2 rocks · 1 gate"),
)
