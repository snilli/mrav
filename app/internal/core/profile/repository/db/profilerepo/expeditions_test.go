package profilerepo_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"orm"
	"orm/ent"
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

	item := func(code int, name, price string) *ent.Item {
		return client.Item.Create().SetCode(code).SetName(name).SetDesireName(name).SetKind("Item").
			SetDescription("").SetEffect("").SetSalePrice(price).SetPurchasePrice("-----").SaveX(ctx)
	}
	// chance puts a chance of an item into a place, for a tier of the stat that starts at min, in the next row of the
	// table of that tier.
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
	worths := func(by domain.Searcher) map[string]float64 {
		maps, err := repo.ListExpeditionMaps(ctx, by)
		Expect(err).NotTo(HaveOccurred())
		out := map[string]float64{}
		for _, m := range maps {
			for _, s := range m.Spots {
				out[m.Name+"/"+s.Name] = s.Value
			}
		}
		return out
	}

	BeforeEach(func() {
		ctx = context.Background()
		slots = map[[3]int]int{}
		var err error
		client, err = testutil.NewClient(ctx)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)
		repo = profilerepo.New(client)

		gem, nuts := item(900, "Gem", "1000"), item(901, "Nuts", "100")
		item(902, "Nuts", "100") // another item of the same name, which no chance is for: it must not count twice
		glue := item(903, "Glue", "-----")

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
		maps, err := repo.ListExpeditionMaps(ctx, domain.Searcher{})
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
	})

	It("reads a wall, a rock and a teleport as what they are", func() {
		client.ExpeditionMap.Create().SetName("Odd").SetRows("#oT\nS..").SetStartX(0).SetStartY(1).SaveX(ctx)
		maps, err := repo.ListExpeditionMaps(ctx, domain.Searcher{})
		Expect(err).NotTo(HaveOccurred())
		Expect(maps[2].Tiles[:3]).To(Equal([]domain.Terrain{domain.Wall, domain.Rock, domain.Teleport}))
		Expect(maps[2].Start).To(Equal(domain.Point{X: 0, Y: 1}))
	})

	DescribeTable("works out in the database what a search of a place is worth: each item's chance by its price, at the tier of the stat",
		func(by domain.Searcher, place string, want float64) {
			Expect(worths(by)["Cave/"+place]).To(BeNumerically("~", want, 1e-9))
		},
		Entry("the first tier of intelligence", domain.Searcher{Intelligence: 99}, "A", 190.0),
		Entry("the second, from where it starts", domain.Searcher{Intelligence: 100}, "A", 280.0),
		Entry("the second, as high as the stat goes", domain.Searcher{Intelligence: 999}, "A", 280.0),
		Entry("a good nature has a table of its own", domain.Searcher{Intelligence: 150, GoodNature: true}, "A", 600.0),
		Entry("a building that is searched by skill takes no notice of intelligence", domain.Searcher{Intelligence: 999, Skill: 0}, "B", 0.0),
		Entry("and goes by skill", domain.Searcher{Intelligence: 0, Skill: 250}, "B", 1000.0),
		Entry("an item that cannot be sold is worth nothing", domain.Searcher{Skill: 199}, "B", 0.0),
	)

	It("adds up the two rows that a table has for one item", func() {
		gem := client.Item.Query().Where(entitem.Name("Gem")).OnlyX(ctx)
		a := client.ExpeditionPlace.Query().Where(entexpeditionplace.Name("A")).OnlyX(ctx)
		chance(a, false, 0, 0, gem, 10) // a second row of Gem in the first tier of A, which had 10 and has 20 now
		Expect(worths(domain.Searcher{Intelligence: 0})["Cave/A"]).To(BeNumerically("~", 100+100+90, 1e-9))
	})

	It("gives a place that has no chances at all a worth of nothing, and still lists it", func() {
		got := worths(domain.Searcher{Intelligence: 500})
		Expect(got).To(HaveKeyWithValue("Empty/Z", BeNumerically("==", 0)))
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
		got, err := profilerepo.New(client).ListExpeditionMaps(ctx, by)
		Expect(err).NotTo(HaveOccurred())
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

	DescribeTable("has the maps of the three trips that come round every year, with their start, places and gates",
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
	)

	It("is for every trip the name that the game gives the trip, so that its week finds the map", func() {
		for _, m := range maps {
			found := false
			for _, e := range domain.Expeditions {
				found = found || e.Name == m.Name
			}
			Expect(found).To(BeTrue(), m.Name)
		}
	})

	It("works out what a search of a place is worth for a monster, in the database, from the chances and the prices of the items", func() {
		by := domain.Searcher{Intelligence: 463, Skill: 690, GoodNature: true}
		maps = load(by)
		Expect(spot(maps[0], "A").Value).To(BeNumerically("~", 481.8, 1e-6))
		Expect(spot(maps[0], "A").Building).To(Equal("Mamoth Statue"))
		Expect(spot(maps[1], "A").Value).To(BeNumerically("~", 471.78, 1e-6)) // a large tree, searched by skill
		Expect(spot(maps[2], "F").Value).To(BeNumerically("~", 1672.5, 1e-6))

		maps = load(domain.Searcher{Intelligence: 150})
		Expect(spot(maps[0], "A").Value).To(BeNumerically("~", 413.55, 1e-6))
	})

	It("finds a monster that is three steps up from the start of Kawrea in June, as the game showed it at X 0, Y 3", func() {
		ranch := player.Ranch{Known: true, Date: 872, Tile: 27*32 + 15, Steps: 3, WalkKnown: true}
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
			Expect(len(r.Spots)).To(BeNumerically(">=", 3), m.Name)
		}
	})
})
