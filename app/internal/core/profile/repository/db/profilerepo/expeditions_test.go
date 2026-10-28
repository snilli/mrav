package profilerepo_test

import (
	"context"
	"errors"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"orm"
	"orm/ent"
	entexpeditionmap "orm/ent/expeditionmap"
	entexpeditionplace "orm/ent/expeditionplace"
	entitem "orm/ent/item"

	player "app/internal/core/player/domain"
	"app/internal/core/profile"
	"app/internal/core/profile/domain"
	"app/internal/core/profile/repository/db/profilerepo"
	"app/internal/testutil"
)

var _ = Describe("profilerepo expedition maps", Label("integration"), func() {
	var (
		ctx    context.Context
		client *orm.Client
		repo   profile.ProfileRepository
	)

	item := func(code int, name string, price int) *ent.Item {
		create := client.Item.Create().SetVersion(entitem.VersionMr2).SetCode(code).SetName(name).SetDesireName(name).SetKind("Item").
			SetDescription("").SetEffect("")
		if price > 0 {
			create.SetSalePrice(price)
		}
		return create.SaveX(ctx)
	}
	slots := map[[3]int]int{}
	chance := func(place *ent.ExpeditionPlace, good bool, tier, min int, it *ent.Item, pct float64) {
		key := [3]int{place.ID, tier, 0}
		if good {
			key[2] = 1
		}
		client.LootChance.Create().SetPlace(place).SetItem(it).SetGood(good).SetTier(tier).SetSlot(slots[key]).
			SetMinStat(min).SetPct(pct).SaveX(ctx)
		slots[key]++
	}
	chancesOf := func(mapName, place string) []domain.Chance {
		maps, err := repo.ListExpeditionMaps(ctx)
		Expect(err).NotTo(HaveOccurred())
		for _, m := range maps {
			for _, s := range m.Spots {
				if m.Name == mapName && s.Name == place {
					return s.Chances
				}
			}
		}
		Fail("no place " + mapName + "/" + place)
		return nil
	}

	BeforeEach(func() {
		ctx = context.Background()
		slots = map[[3]int]int{}
		var err error
		client, err = testutil.NewClient(ctx)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)
		repo = profilerepo.New(client)

		gem, nuts := item(900, "Gem", 1000), item(901, "Nuts", 100)
		item(902, "Nuts", 100)
		glue := item(903, "Glue", 0)

		cave := client.ExpeditionMap.Create().SetName("Cave").SetRows("S.A\n.gB").SetStartX(0).SetStartY(0).SaveX(ctx)
		empty := client.ExpeditionMap.Create().SetName("Empty").SetRows("S").SetStartX(0).SetStartY(0).SaveX(ctx)
		client.ExpeditionLock.Create().SetMap(cave).SetX(1).SetY(1).SetMinInt(300).SaveX(ctx)

		a := client.ExpeditionPlace.Create().SetMap(cave).SetName("A").SetX(2).SetY(0).SetBuilding("Mine").SetStat(entexpeditionplace.StatInt).SaveX(ctx)
		b := client.ExpeditionPlace.Create().SetMap(cave).SetName("B").SetX(2).SetY(1).SetBuilding("Tree").SetStat(entexpeditionplace.StatSkill).SaveX(ctx)
		client.ExpeditionPlace.Create().SetMap(empty).SetName("Z").SetX(0).SetY(0).SetBuilding("Hut").SetStat(entexpeditionplace.StatInt).SaveX(ctx)

		chance(a, false, 0, 0, gem, 10)
		chance(a, false, 0, 0, nuts, 90)
		chance(a, false, 1, 100, gem, 20)
		chance(a, false, 1, 100, nuts, 80)
		chance(a, true, 0, 0, gem, 50)
		chance(a, true, 1, 100, gem, 60)
		chance(b, false, 0, 0, glue, 100)
		chance(b, false, 1, 200, gem, 100)
	})

	It("lists the maps in the order they are kept, with their tiles, their start, their places and their locks", func() {
		maps, err := repo.ListExpeditionMaps(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(maps).To(HaveLen(2))

		cave := maps[0]
		Expect(cave.Name).To(Equal("Cave"))
		Expect([]int{cave.Width, cave.Height}).To(Equal([]int{3, 2}))
		Expect(cave.Start).To(Equal(domain.Point{}))
		Expect(cave.Tiles).To(Equal([]domain.Terrain{
			domain.Floor, domain.Floor, domain.Floor,
			domain.Floor, domain.Gate, domain.Floor,
		}))
		Expect(cave.Locks).To(Equal([]domain.Lock{{At: domain.Point{X: 1, Y: 1}, MinINT: 300}}))
		Expect(cave.Spots).To(HaveLen(2))
		Expect(cave.Spots[0].Name).To(Equal("A"))
		Expect(cave.Spots[0].At).To(Equal(domain.Point{X: 2, Y: 0}))
		Expect(cave.Spots[0].Building).To(Equal("Mine"))
		Expect(cave.Spots[1].Name).To(Equal("B"))
		Expect([]bool{cave.Spots[0].BySkill, cave.Spots[1].BySkill}).To(Equal([]bool{false, true}), "A goes by intelligence, B by skill")
	})

	It("keeps each place with its own map when there are several, and the places of a map by their letter", func() {
		cave := client.ExpeditionMap.Query().Where(entexpeditionmap.Name("Cave")).OnlyX(ctx)
		client.ExpeditionPlace.Create().SetMap(cave).SetName("0").SetX(0).SetY(1).SetBuilding("Door").SetStat(entexpeditionplace.StatInt).SaveX(ctx)
		client.ExpeditionLock.Create().SetMap(cave).SetX(0).SetY(0).SetMinInt(100).SaveX(ctx)
		maps, err := repo.ListExpeditionMaps(ctx)
		Expect(err).NotTo(HaveOccurred())

		names := func(m domain.ExpeditionMap) (out []string) {
			for _, s := range m.Spots {
				out = append(out, s.Name)
			}
			return out
		}
		Expect(names(maps[0])).To(Equal([]string{"0", "A", "B"}))
		Expect(names(maps[1])).To(Equal([]string{"Z"}))
		Expect(maps[0].Locks).To(Equal([]domain.Lock{{At: domain.Point{}, MinINT: 100}, {At: domain.Point{X: 1, Y: 1}, MinINT: 300}}),
			"in the order of the tiles")
		Expect(maps[1].Locks).To(BeEmpty())
	})

	It("gives each place every table of chances that it has, whichever the nature and the tier, with the item, its code and what it sells for", func() {
		Expect(chancesOf("Cave", "A")).To(ConsistOf(
			domain.Chance{MinStat: 0, Slot: 0, Code: 900, Item: "Gem", Pct: 10, Price: 1000},
			domain.Chance{MinStat: 0, Slot: 1, Code: 901, Item: "Nuts", Pct: 90, Price: 100},
			domain.Chance{MinStat: 100, Slot: 0, Code: 900, Item: "Gem", Pct: 20, Price: 1000},
			domain.Chance{MinStat: 100, Slot: 1, Code: 901, Item: "Nuts", Pct: 80, Price: 100},
			domain.Chance{Good: true, MinStat: 0, Slot: 0, Code: 900, Item: "Gem", Pct: 50, Price: 1000},
			domain.Chance{Good: true, MinStat: 100, Slot: 0, Code: 900, Item: "Gem", Pct: 60, Price: 1000},
		))
		Expect(chancesOf("Cave", "B")).To(ConsistOf(
			domain.Chance{MinStat: 0, Slot: 0, Code: 903, Item: "Glue", Pct: 100, Price: 0},
			domain.Chance{MinStat: 200, Slot: 0, Code: 900, Item: "Gem", Pct: 100, Price: 1000},
		))
	})

	It("keeps the price of each item by the item and not by its name, so that two items of one name do not share one", func() {
		a := client.ExpeditionPlace.Query().Where(entexpeditionplace.Name("A")).OnlyX(ctx)
		dear := item(904, "Nuts", 300)
		chance(a, false, 0, 0, dear, 5)
		Expect(chancesOf("Cave", "A")).To(ContainElement(domain.Chance{MinStat: 0, Slot: 2, Code: 904, Item: "Nuts", Pct: 5, Price: 300}))
	})

	It("lists a place that has no chances at all, with none", func() {
		Expect(chancesOf("Empty", "Z")).To(BeEmpty())
	})

	It("reads a wall, a rock and a teleport as what they are", func() {
		client.ExpeditionMap.Create().SetName("Odd").SetRows("#oT\nS..").SetStartX(0).SetStartY(1).SaveX(ctx)
		maps, err := repo.ListExpeditionMaps(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(maps[2].Tiles[:3]).To(Equal([]domain.Terrain{domain.Wall, domain.Rock, domain.Teleport}))
		Expect(maps[2].Start).To(Equal(domain.Point{X: 0, Y: 1}))
	})

})

var _ = Describe("profilerepo shipped expedition maps", Label("integration"), func() {
	var maps []domain.ExpeditionMap

	load := func(by domain.Searcher) []domain.ExpeditionMap {
		ctx := context.Background()
		client, err := testutil.OpenShipped(ctx)
		if errors.Is(err, testutil.ErrNoShippedDB) {
			Skip(err.Error())
		}
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)
		got, err := profilerepo.New(client).ListExpeditionMaps(ctx)
		Expect(err).NotTo(HaveOccurred())
		for i := range got {
			got[i] = got[i].For(by)
		}
		return got
	}
	spot := func(m domain.ExpeditionMap, name string) domain.Spot {
		for _, s := range m.Spots {
			if s.Name == name {
				return s
			}
		}
		Fail("no place " + name + " on " + m.Name)
		return domain.Spot{}
	}

	BeforeEach(func() { maps = load(domain.Searcher{}) })

	DescribeTable("has the maps of the trips, the three that come round every year and then the first expedition, with their start, places and gates",
		func(i int, name string, width, height, places, locks int, start domain.Point) {
			m := maps[i]
			Expect(m.Name).To(Equal(name))
			Expect([]int{m.Width, m.Height, len(m.Tiles), len(m.Spots), len(m.Locks)}).To(Equal([]int{width, height, width * height, places, locks}))
			Expect(m.Start).To(Equal(start))
			Expect(m.Tiles[m.Start.Y*m.Width+m.Start.X]).To(Equal(domain.Floor))
		},
		Entry("Parepare", 0, "Parepare Jungle", 18, 20, 10, 2, domain.Point{X: 8, Y: 19}),
		Entry("Torles", 1, "Torles Mountains", 26, 28, 14, 4, domain.Point{X: 12, Y: 27}),
		Entry("Kawrea", 2, "Kawrea Volcano", 26, 28, 10, 3, domain.Point{X: 12, Y: 27}),
		Entry("the first expedition", 3, "Phoenix Volcano Adventure", 12, 13, 1, 0, domain.Point{X: 6, Y: 12}),
	)

	Describe("the first expedition", func() {
		It("has its cave above the rock in the middle, and two rocks to break on the way round", func() {
			m := maps[3]
			cave := spot(m, "A")
			Expect(cave.Building).To(Equal("Mountain Cave"))
			Expect(cave.At).To(Equal(domain.Point{X: 5, Y: 7}))
			Expect(m.Tiles[(cave.At.Y+1)*m.Width+cave.At.X]).To(Equal(domain.Rock))
			rocks := 0
			for _, t := range m.Tiles {
				if t == domain.Rock {
					rocks++
				}
			}
			Expect(rocks).To(Equal(2))
		})

		It("turns up the Fire Feather at every tier of intelligence, for either nature", func() {
			for _, good := range []bool{true, false} {
				found := map[int]float64{}
				for _, c := range spot(maps[3], "A").Chances {
					if c.Good == good && c.Item == "Fire Feather" {
						found[c.MinStat] = c.Pct
					}
				}
				Expect(found).To(HaveLen(4), "good nature: %v", good)
				for _, pct := range found {
					Expect(pct).To(BeNumerically(">=", 18.8))
				}
			}
		})

		It("has the top row of ten tiles that the game has, where the picture of the guide has eleven", func() {
			m := maps[3]
			top := 0
			for x := 0; x < m.Width; x++ {
				if m.Tiles[x] == domain.Floor {
					top++
				}
			}
			Expect(top).To(Equal(10))
			Expect(m.Tiles[0*m.Width+m.Width-1]).To(Equal(domain.Floor), "the right column of the game is the last")
		})

		It("has a right column of eight tiles, from the top row down, the first rock under it as the corner of the corridor that turns left, and no way down past the rock, as the game has them", func() {
			m := maps[3]
			last := m.Width - 1
			for y := range 8 {
				Expect(m.Tiles[y*m.Width+last]).To(Equal(domain.Floor), "row %d", y)
			}
			Expect(m.Tiles[8*m.Width+last]).To(Equal(domain.Rock))
			Expect(m.Tiles[8*m.Width+last-1]).To(Equal(domain.Floor))
			Expect(m.Tiles[9*m.Width+last]).To(Equal(domain.Wall))
		})

		It("has a route to the cave and back for a monster with the energy for it, which costs 104: 47 out to the tile before the cave over its two rocks at 6 each, the search of 10 made from that tile, and the same 47 back", func() {
			r, err := domain.PlanRoute(maps[3], domain.Walker{Energy: 700, BreakChance: 1, Intelligence: 999})
			Expect(err).NotTo(HaveOccurred())
			Expect(r.Spots).To(HaveLen(1))
			Expect(r.Spots[0].Name).To(Equal("A"))
			Expect(r.Cost).To(BeNumerically("~", 104, 1e-9))
		})

		It("has no route for a monster that starts with 103, one short of the round", func() {
			r, err := domain.PlanRoute(maps[3], domain.Walker{Energy: 103, BreakChance: 1, Intelligence: 999})
			Expect(err).NotTo(HaveOccurred())
			Expect(r.Spots).To(BeEmpty())
		})
	})

	DescribeTable("has the teleports of a map and where each lands, with the labels that the guide gives them",
		func(i int, want []domain.Portal) { Expect(maps[i].Portals).To(ConsistOf(want)) },
		Entry("Parepare, one teleport and the three places where it lands", 0, []domain.Portal{
			{Label: "T1", At: domain.Point{X: 7, Y: 0}, To: []string{"E1", "E2", "E3"}},
			{Label: "E1", At: domain.Point{X: 5, Y: 6}}, {Label: "E2", At: domain.Point{X: 9, Y: 3}}, {Label: "E3", At: domain.Point{X: 9, Y: 14}},
		}),
		Entry("Torles, five teleports, one of which leads to another", 1, []domain.Portal{
			{Label: "T1", At: domain.Point{X: 2, Y: 0}, To: []string{"E1", "E2", "E3"}},
			{Label: "T2", At: domain.Point{X: 9, Y: 9}, To: []string{"E1", "E2", "E3"}},
			{Label: "T3", At: domain.Point{X: 17, Y: 6}, To: []string{"E1", "E2", "E3"}},
			{Label: "T4", At: domain.Point{X: 14, Y: 20}, To: []string{"E3", "E4", "E5"}},
			{Label: "T5", At: domain.Point{X: 25, Y: 4}, To: []string{"E3", "T4"}},
			{Label: "E1", At: domain.Point{X: 0, Y: 25}}, {Label: "E2", At: domain.Point{X: 13, Y: 20}},
			{Label: "E3", At: domain.Point{X: 19, Y: 7}}, {Label: "E4", At: domain.Point{X: 0, Y: 9}},
			{Label: "E5", At: domain.Point{X: 12, Y: 25}},
		}),
		Entry("Kawrea, two teleports, one of which has one place to land", 2, []domain.Portal{
			{Label: "T1", At: domain.Point{X: 7, Y: 4}, To: []string{"E1", "E2", "E3"}},
			{Label: "T2", At: domain.Point{X: 3, Y: 23}, To: []string{"E4"}},
			{Label: "E1", At: domain.Point{X: 7, Y: 0}}, {Label: "E2", At: domain.Point{X: 25, Y: 13}},
			{Label: "E3", At: domain.Point{X: 0, Y: 17}}, {Label: "E4", At: domain.Point{X: 5, Y: 23}},
		}),
	)

	It("puts every teleport on a teleport tile, every place to land on a floor tile, and every teleport tile has a label", func() {
		for _, m := range maps {
			labels := map[string]bool{}
			for _, p := range m.Portals {
				labels[p.Label] = true
			}
			teleports := 0
			for _, p := range m.Portals {
				tile := m.Tiles[p.At.Y*m.Width+p.At.X]
				if strings.HasPrefix(p.Label, "T") {
					teleports++
					Expect(tile).To(Equal(domain.Teleport), m.Name+" "+p.Label)
				} else {
					Expect(tile).To(Equal(domain.Floor), m.Name+" "+p.Label)
				}
				for _, to := range p.To {
					Expect(labels).To(HaveKey(to), m.Name+" "+p.Label+" leads to "+to)
				}
			}
			tiles := 0
			for _, t := range m.Tiles {
				if t == domain.Teleport {
					tiles++
				}
			}
			Expect(tiles).To(Equal(teleports), m.Name)
		}
	})

	It("is for every trip the name that the game gives the trip, so that its week finds the map", func() {
		for _, m := range maps {
			_, found := m.Trip()
			Expect(found).To(BeTrue(), m.Name)
		}
	})

	It("works out what a search of a place is worth for a monster, in the database, from the chances and the prices of the items", func() {
		by := domain.Searcher{Intelligence: 463, Skill: 690, GoodNature: true}
		maps = load(by)
		Expect(spot(maps[0], "A").Value).To(BeNumerically("~", 481.8, 1e-6))
		Expect(spot(maps[0], "A").Building).To(Equal("Mammoth Statue"))
		Expect(spot(maps[1], "A").Value).To(BeNumerically("~", 471.78, 1e-6))
		Expect(spot(maps[2], "F").Value).To(BeNumerically("~", 1672.5, 1e-6))

		maps = load(domain.Searcher{Intelligence: 150})
		Expect(spot(maps[0], "A").Value).To(BeNumerically("~", 413.55, 1e-6))
	})

	It("lists what a search of a place may turn up, with its chance, at the tier of the monster, the likeliest first, and leaves out what it never turns up", func() {
		good := spot(load(domain.Searcher{Intelligence: 150, GoodNature: true})[0], "A")
		Expect(good.Loot).To(Equal([]domain.Drop{
			{Code: 18, Item: "Mint Leaf", Pct: 25, Price: 100},
			{Code: 28, Item: "Magic Banana", Pct: 25, Price: 100},
			{Code: 23, Item: "Nuts Oil", Pct: 18.8, Price: 100},
			{Code: 33, Item: "Griever", Pct: 18.8, Price: 50},
			{Code: 54, Item: "Zuum Chips", Pct: 12.5, Price: 500},
		}))

		bad := spot(load(domain.Searcher{Intelligence: 150})[0], "A")
		Expect(bad.Loot).To(Equal([]domain.Drop{
			{Code: 146, Item: "Pixie Chips", Pct: 18.8, Price: 1000},
			{Code: 153, Item: "Golem Chips", Pct: 18.8, Price: 500},
			{Code: 54, Item: "Zuum Chips", Pct: 18.8, Price: 500},
			{Code: 28, Item: "Magic Banana", Pct: 18.8, Price: 100},
			{Code: 18, Item: "Mint Leaf", Pct: 12.5, Price: 100},
			{Code: 33, Item: "Griever", Pct: 12.5, Price: 50},
		}))
	})

	It("adds up to the worth of the place, since the worth is what the list sells for at its chances", func() {
		for _, m := range load(domain.Searcher{Intelligence: 463, Skill: 690, GoodNature: true}) {
			for _, s := range m.Spots {
				sum := 0.0
				for _, d := range s.Loot {
					sum += d.Pct / 100 * float64(d.Price)
				}
				Expect(sum).To(BeNumerically("~", s.Value, 1e-6), m.Name+"/"+s.Name)
			}
		}
	})

	It("finds a monster that is three steps up from the start of Kawrea in June, as the game showed it at X 0, Y 3", func() {
		ranch := player.Ranch{Known: true, Date: 872, Tile: 27*32 + 15, WalkKnown: true}
		kawrea := maps[2]
		at, ok := kawrea.Stand(ranch)
		Expect(ok).To(BeTrue())
		Expect(at).To(Equal(domain.Point{X: 12, Y: 24}))
		Expect(kawrea.Here(ranch)).To(BeTrue())
		Expect(maps[0].Here(ranch)).To(BeFalse())
	})

	It("has a route to plan on every map, with several places in reach of a monster with the most energy", func() {
		maps = load(domain.Searcher{Intelligence: 999, Skill: 999, GoodNature: true})
		for _, m := range maps {
			r, err := domain.PlanRoute(m, domain.Walker{Energy: 700, BreakChance: 1, Intelligence: 999})
			Expect(err).NotTo(HaveOccurred(), m.Name)
			if m.Name == domain.Phoenix.Name {
				Expect(r.Spots).To(HaveLen(1), m.Name)
				continue
			}
			Expect(len(r.Spots)).To(BeNumerically(">=", 3), m.Name)
		}
	})
})
