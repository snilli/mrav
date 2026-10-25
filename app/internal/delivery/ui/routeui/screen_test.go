package routeui

import (
	"context"
	"image/png"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/software"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/apperr"
	monster "app/internal/core/monster/domain"
	player "app/internal/core/player/domain"
	"app/internal/core/profile/domain"
	"app/internal/delivery/ui/profileui"
	mockprofile "app/mock/core/profile"
)

var _ = Describe("Screen", func() {
	var (
		ctx context.Context
		svc *mockprofile.MockProfileService
		s   *Screen
	)
	maps := func() []domain.ExpeditionMap { return []domain.ExpeditionMap{aisle()} }

	BeforeEach(func() {
		test.NewTempApp(GinkgoTB())
		ctx = context.Background()
		svc = mockprofile.NewMockProfileService(GinkgoT())
		s = NewScreen(svc)
	})

	Describe("the item finder", func() {
		findable := func() []domain.ExpeditionMap {
			m := aisle()
			m.Spots[0].Chances = []domain.Chance{{Item: "Gem", Pct: 20}, {Item: "Nuts", Pct: 80}}
			m.Spots[0].Loot = []domain.Drop{{Item: "Gem", Pct: 20}, {Item: "Nuts", Pct: 80}}
			return []domain.ExpeditionMap{m}
		}
		standing := func() domain.Profile {
			p := monsterOf(140, 350)
			p.Player.Ranch.Tile, p.Player.Ranch.WalkKnown = 30*32+16, true
			return p
		}
		checks := func() []string {
			var out []string
			for _, o := range s.checks.Objects {
				out = append(out, o.(*widget.Check).Text)
			}
			return out
		}

		It("shares one map with the route, and switches only what is under it", func() {
			s.apply(findable(), monsterOf(140, 350), nil)
			b := s.boxes[0]
			Expect(s.modeNames()).To(Equal([]string{"Route", "Find items"}))
			Expect(b.routePart.Visible()).To(BeTrue())
			Expect(b.findPart.Visible()).To(BeFalse())
			Expect(s.picker.Visible()).To(BeFalse())

			s.setMode(modeFind)
			Expect(s.boxes[0]).To(BeIdenticalTo(b))
			Expect(b.routePart.Visible()).To(BeFalse())
			Expect(b.findPart.Visible()).To(BeTrue())
			Expect(s.picker.Visible()).To(BeTrue())
		})

		It("puts the switch and the item picker under the map, above what they change", func() {
			s.apply(findable(), monsterOf(140, 350), nil)
			b := s.boxes[0]
			Expect(b.controls.Objects).To(Equal([]fyne.CanvasObject{s.bar, s.picker}))
			order := b.object.(*fyne.Container).Objects
			at := func(o fyne.CanvasObject) int {
				for i, x := range order {
					if x == o {
						return i
					}
				}
				return -1
			}
			Expect(at(b.holder)).To(BeNumerically("<", at(b.controls)))
			Expect(at(b.controls)).To(BeNumerically("<", at(b.routePart)))
		})

		It("offers every item of the open maps to pick, as many as wanted, folded away until asked for or searched", func() {
			s.apply(findable(), monsterOf(140, 350), nil)
			Expect(s.checks.Visible()).To(BeFalse())
			Expect(s.search.Visible()).To(BeFalse())
			Expect(s.showItems.Text).To(Equal("+ Add item"))
			test.Tap(s.showItems)
			Expect(s.checks.Visible()).To(BeTrue())
			Expect(s.showItems.Text).To(Equal("Done"))
			Expect(checks()).To(Equal([]string{"Gem", "Nuts"}))
			test.Tap(s.showItems)
			s.filter.SetText("nu")
			Expect(s.checks.Visible()).To(BeTrue())
			Expect(checks()).To(Equal([]string{"Nuts"}))
		})

		It("shows each picked item as a chip that takes it off when tapped, with the list folded", func() {
			s.apply(findable(), monsterOf(140, 350), nil)
			chips := func() []string {
				var out []string
				for _, o := range s.chips.Objects {
					if b, ok := o.(*widget.Button); ok {
						out = append(out, b.Text)
					}
				}
				return out
			}
			Expect(chips()).To(BeEmpty())
			s.pick("Nuts", true)
			s.pick("Gem", true)
			Expect(chips()).To(Equal([]string{"Gem ×", "Nuts ×"}))
			test.Tap(s.chips.Objects[0].(*widget.Button))
			Expect(chips()).To(Equal([]string{"Nuts ×"}))
			Expect(s.picked["Gem"]).To(BeFalse())
		})

		twoPlaces := func() []domain.ExpeditionMap {
			m := aisle()
			m.Width, m.Tiles = 6, make([]domain.Terrain, 6)
			m.Spots = []domain.Spot{
				{Name: "A", At: domain.Point{X: 2}, Building: "Cave",
					Chances: []domain.Chance{{Item: "Gem", Pct: 20}, {Item: "Nuts", Pct: 80}, {Item: "Mint", Pct: 10}, {MinStat: 500, Item: "Gold", Pct: 5}},
					Loot:    []domain.Drop{{Item: "Nuts", Pct: 80}, {Item: "Gem", Pct: 20}, {Item: "Mint", Pct: 10}}},
				{Name: "B", At: domain.Point{X: 4}, Building: "Pond",
					Chances: []domain.Chance{{Item: "Nuts", Pct: 90}}, Loot: []domain.Drop{{Item: "Nuts", Pct: 90}}},
			}
			return []domain.ExpeditionMap{m}
		}
		finding := func(items ...string) *box {
			s.apply(twoPlaces(), standing(), nil)
			s.setMode(modeFind)
			for _, it := range items {
				s.pick(it, true)
			}
			return s.boxes[0]
		}
		rows := func(b *box) []string {
			var out []string
			for _, r := range b.rows {
				out = append(out, r.Text)
			}
			return out
		}

		It("numbers the places of the picked items on the map, best first, dims the others, and still shows where the monster stands", func() {
			b := finding("Nuts")
			Expect(b.cells[1].bg.FillColor).To(Equal(markerFill))
			Expect([]any{b.cells[4].text.Text, b.cells[4].bg.FillColor}).To(Equal([]any{"1", pickFill}))
			Expect([]any{b.cells[2].text.Text, b.cells[2].bg.FillColor}).To(Equal([]any{"2", pickFill}))

			s.pick("Nuts", false)
			s.pick("Gem", true)
			b = s.boxes[0]
			Expect(b.cells[2].text.Text).To(Equal("1"))
			Expect([]any{b.cells[4].text.Text, b.cells[4].bg.FillColor}).To(Equal([]any{"B", dimFill}))

			s.setMode(modeRoute)
			Expect(s.boxes[0].cells[1].bg.FillColor).To(Equal(markerFill))
			Expect(s.boxes[0].cells[2].text.Text).To(Equal("A"))
		})

		It("outlines on the map the place it tells about", func() {
			b := finding("Nuts")
			Expect(b.cells[4].bg.StrokeWidth).To(BeNumerically(">", 0))
			test.Tap(b.rows[1])
			b = s.boxes[0]
			Expect(b.cells[2].bg.StrokeWidth).To(BeNumerically(">", 0))
			Expect(b.cells[4].bg.StrokeWidth).To(BeZero())
		})

		It("lists each place in one line, and tells the first under the map until another row or tile is tapped", func() {
			b := finding("Nuts")
			Expect(rows(b)).To(Equal([]string{"1. B · Pond · 2 energy to walk", "2. A · Cave · right here"}))
			Expect(b.detailTitle.Text).To(Equal("1. B · Pond · 2 energy to walk"))
			Expect(b.detailItems.Text).To(Equal("Nuts 90%"))

			test.Tap(b.rows[1])
			b = s.boxes[0]
			Expect(b.detailTitle.Text).To(Equal("2. A · Cave · right here"))
			Expect(b.detailItems.Text).To(Equal("Nuts 80%"))

			test.Tap(b.taps["B"])
			Expect(s.boxes[0].detailTitle.Text).To(HavePrefix("1. B"))
		})

		It("with every picked item at one place, numbers that place alone and tells all its items", func() {
			b := finding("Gem", "Mint")
			Expect(rows(b)).To(Equal([]string{"1. A · Cave · right here"}))
			Expect(b.cells[2].text.Text).To(Equal("1"))
			Expect(b.cells[4].bg.FillColor).To(Equal(dimFill))
			Expect(b.detailItems.Text).To(Equal("Gem 20% · Mint 10%"))
		})

		It("folds what the monster cannot find yet into one line", func() {
			b := finding("Gold")
			Expect(b.rows).To(BeEmpty())
			Expect(b.notYetToggle.Text).To(Equal("▸ 1 more needs a higher stat"))
			Expect(b.notYetBox.Visible()).To(BeFalse())
			test.Tap(b.notYetToggle)
			b = s.boxes[0]
			Expect(b.notYetBox.Visible()).To(BeTrue())
			Expect(b.notYetToggle.Text).To(Equal("▾ 1 more needs a higher stat"))
		})

		It("folds the key of the map away until asked for, keeps it so, and tells the tiles of the mode shown", func() {
			s.apply(findable(), monsterOf(140, 350), nil)
			b := s.boxes[0]
			Expect(b.keyBox.Visible()).To(BeFalse())
			Expect(b.keyToggle.Text).To(Equal("▸ Map key"))
			test.Tap(b.keyToggle)
			Expect(b.keyBox.Visible()).To(BeTrue())
			Expect(b.keyToggle.Text).To(Equal("▾ Map key"))

			s.apply(findable(), monsterOf(141, 350), nil)
			Expect(s.boxes[0].keyBox.Visible()).To(BeTrue())

			s.setMode(modeFind)
			var words []string
			for _, k := range s.boxes[0].key {
				words = append(words, k.label.Text)
			}
			Expect(words).To(ContainElement("A place with a picked item, best first"))
			Expect(words).NotTo(ContainElement("A stop, in the order to go"))
		})

		It("keeps the picks and the mode when the game is read again, and drops the lights when nothing is picked", func() {
			s.apply(findable(), monsterOf(140, 350), nil)
			s.setMode(modeFind)
			s.pick("Gem", true)
			s.apply(findable(), monsterOf(141, 350), nil)
			Expect(s.boxes[0].cells[2].bg.FillColor).To(Equal(pickFill))
			Expect(s.boxes[0].findPart.Visible()).To(BeTrue())
			s.pick("Gem", false)
			Expect(s.boxes[0].cells[2].bg.FillColor).To(Equal(dimFill))
			Expect(s.boxes[0].rows).To(BeEmpty())
		})
	})

	It("loads the monster, and shows the maps that its profile has", func() {
		p := monsterOf(140, 350)
		p.Maps = maps()
		svc.EXPECT().Current(ctx).Return(p, nil)
		s.Load(ctx)
		Expect(s.Content().Visible()).To(BeTrue())
		Expect(s.boxes).To(HaveLen(1))
	})

	It("is not there where the profile has no maps", func() {
		svc.EXPECT().Current(ctx).Return(monsterOf(140, 350), nil)
		s.Load(ctx)
		Expect(s.Content().Visible()).To(BeFalse())
	})

	It("draws each map as a table of tiles, the places of the route marked in it and no line between them", func() {
		s.apply(maps(), monsterOf(140, 350), nil)
		Expect(s.Content().Visible()).To(BeTrue())
		Expect(s.boxes).To(HaveLen(1))
		cells := s.boxes[0].cells
		Expect(cells).To(HaveLen(4))
		Expect(cells[0].text.Text).To(Equal("S"))
		Expect(cells[0].bg.FillColor).To(Equal(startFill))
		Expect(cells[1].bg.FillColor).To(Equal(floorFill))
		Expect(cells[2].text.Text).To(Equal("1"))
		Expect(cells[2].bg.FillColor).To(Equal(stopFill))
		Expect(cells[3].bg.FillColor).To(Equal(wallFill))
	})

	It("draws a map on a table as big as the biggest of all the maps, the smaller in the middle of it, with tiles of one size", func() {
		big := trip("Torles Mountains")
		big.Width, big.Height, big.Tiles = 8, 3, make([]domain.Terrain, 24)
		s.apply([]domain.ExpeditionMap{aisle(), big}, monsterOf(140, 350), nil)
		Expect(s.boxes).To(HaveLen(1))
		b := s.boxes[0]
		b.table.Resize(b.table.MinSize())

		Expect(b.table.MinSize()).To(Equal(fyne.NewSize(8*(cellSize+cellGap)-cellGap, 3*(cellSize+cellGap)-cellGap)))
		Expect(b.cells[0].object.Position()).To(Equal(fyne.NewPos(2*(cellSize+cellGap), 1*(cellSize+cellGap))),
			"the smaller map two tiles in and one down, which is the middle")
		Expect(b.cells[0].object.Size()).To(Equal(fyne.NewSize(cellSize, cellSize)), "tiles of one size")
		Expect(b.padding).To(HaveLen(8*3 - 4))
		for _, c := range b.padding {
			Expect(c.bg.FillColor).To(Equal(wallFill), "walls round the smaller map")
		}
	})

	Describe("the route under the map", func() {
		trail := func() []domain.ExpeditionMap {
			m := aisle()
			m.Width, m.Tiles = 7, make([]domain.Terrain, 7)
			m.Spots = []domain.Spot{
				{Name: "A", At: domain.Point{X: 2}, Building: "Cave", Value: 1000,
					Chances: []domain.Chance{{Code: 46, Item: "Suezo Chips", Pct: 25}, {Item: "Mint Leaf", Pct: 75}},
					Loot:    []domain.Drop{{Code: 46, Item: "Suezo Chips", Pct: 25}, {Item: "Mint Leaf", Pct: 75}}},
				{Name: "B", At: domain.Point{X: 4}, Building: "Pond", Value: 500, Loot: []domain.Drop{{Item: "Nuts", Pct: 100}}},
				{Name: "C", At: domain.Point{X: 6}, Building: "Hut", Value: 300, Loot: []domain.Drop{{Item: "Candy", Pct: 100}}},
			}
			return []domain.ExpeditionMap{m}
		}
		searchedA := func() *box {
			s.haul = domain.Haul{Map: "Parepare Jungle", Got: map[string][]int{"A": {46}}}
			s.apply(trail(), monsterOf(140, 350), nil)
			return s.boxes[0]
		}
		chips := func(b *box) []string {
			var out []string
			for _, c := range b.drops {
				out = append(out, c.Box().Text)
			}
			return out
		}

		It("keeps a searched place on the route with its number, as it can be searched again", func() {
			b := searchedA()
			Expect([]any{b.cells[2].text.Text, b.cells[2].bg.FillColor}).To(Equal([]any{"1", stopFill}))
			Expect([]any{b.cells[4].text.Text, b.cells[4].bg.FillColor}).To(Equal([]any{"2", stopFill}))
		})

		It("tells the first stop under the map, greening what it gave with how many times, and leaving the rest as they may still drop", func() {
			s.haul = domain.Haul{Map: "Parepare Jungle", Got: map[string][]int{"A": {46, 46}}}
			s.apply(trail(), monsterOf(140, 350), nil)
			b := s.boxes[0]
			Expect(b.routeTitle.Text).To(Equal("1. A · Cave · 1 energy to walk"))
			Expect(chips(b)).To(Equal([]string{"✓ Suezo Chips 25% ×2", "Mint Leaf 75%"}))
			Expect(b.drops[0].Box().Tone).To(Equal(profileui.ToneLimeGreen))
			Expect(b.drops[1].Box().Tone).To(Equal(profileui.ToneNone))
			Expect(b.routeNote.Text).To(Equal("Got Suezo Chips ×2 · about 1,000 G"))
		})

		It("lists each place in one line, with how many finds it gave, and shows the one tapped", func() {
			b := searchedA()
			var rows []string
			for _, r := range b.routeRows {
				rows = append(rows, r.Text)
			}
			Expect(rows).To(Equal([]string{"1. A · Cave · 1 energy to walk · got 1", "2. B · Pond · 3 energy to walk", "3. C · Hut · 5 energy to walk"}))
			test.Tap(b.routeRows[1])
			Expect(s.boxes[0].routeTitle.Text).To(HavePrefix("2. B"))
		})

		It("outlines on the map the place it tells about", func() {
			b := searchedA()
			Expect(b.cells[2].bg.StrokeWidth).To(BeNumerically(">", 0))
			Expect(b.cells[4].bg.StrokeWidth).To(BeZero())
			test.Tap(b.routeTaps["B"])
			b = s.boxes[0]
			Expect(b.cells[4].bg.StrokeWidth).To(BeNumerically(">", 0))
			Expect(b.cells[2].bg.StrokeWidth).To(BeZero())
		})

		It("tells about a place off the route when tapped, with why the route leaves it", func() {
			m := trail()[0]
			m.Spots[2].Loot, m.Spots[2].Value, m.Spots[2].Chances = nil, 0, []domain.Chance{{MinStat: 300, Item: "Gem", Pct: 100}}
			s.apply([]domain.ExpeditionMap{m}, monsterOf(140, 350), nil)
			b := s.boxes[0]
			Expect(b.cells[6].text.Text).To(Equal("C"))
			test.Tap(b.routeTaps["C"])
			b = s.boxes[0]
			Expect(b.routeTitle.Text).To(Equal("C · Hut · off the route"))
			Expect(b.routeNote.Text).To(Equal("needs INT 300"))
			Expect(b.cells[6].bg.StrokeWidth).To(BeNumerically(">", 0))
		})

		It("says a place off the route is out of reach, or not worth the energy, and lets its line be tapped", func() {
			m := aisle()
			m.Width, m.Tiles = 400, append([]domain.Terrain{domain.Floor, domain.Floor}, make([]domain.Terrain, 398)...)
			m.Tiles[3] = domain.Wall
			m.Spots = []domain.Spot{
				{Name: "A", At: domain.Point{X: 399}, Building: "Cave", Value: 1000},
				{Name: "B", At: domain.Point{X: 2}, Building: "Pond", Value: 0, Loot: []domain.Drop{{Item: "Nuts", Pct: 100}}},
			}
			s.apply([]domain.ExpeditionMap{m}, monsterOf(140, 350), nil)
			test.Tap(s.boxes[0].offToggle)
			test.Tap(s.boxes[0].offRows[0])
			Expect(s.boxes[0].routeNote.Text).To(Equal("out of reach"))
			test.Tap(s.boxes[0].offRows[1])
			Expect(s.boxes[0].routeNote.Text).To(Equal("left out: the energy is worth more elsewhere"))
		})

		It("folds the places off the route into one line", func() {
			m := aisle()
			m.Width, m.Tiles = 400, append([]domain.Terrain{domain.Floor, domain.Floor}, make([]domain.Terrain, 398)...)
			m.Spots = []domain.Spot{{Name: "A", At: domain.Point{X: 399}, Building: "Cave", Value: 1000}}
			s.apply([]domain.ExpeditionMap{m}, monsterOf(140, 350), nil)
			b := s.boxes[0]
			Expect(b.routeRows).To(BeEmpty())
			Expect(b.offToggle.Text).To(Equal("▸ 1 place off the route"))
			Expect(b.offBox.Visible()).To(BeFalse())
			test.Tap(b.offToggle)
			Expect(s.boxes[0].offBox.Visible()).To(BeTrue())
		})
	})

	It("draws a map that has no tiles without failing", func() {
		m := aisle()
		m.Width, m.Height, m.Tiles = 0, 0, nil
		Expect(func() { s.apply([]domain.ExpeditionMap{m}, monsterOf(140, 350), nil) }).NotTo(Panic())
	})

	It("shows a place that the route leaves out apart from one that it takes", func() {
		m := aisle()
		m.Width, m.Tiles = 400, append([]domain.Terrain{domain.Floor, domain.Floor}, make([]domain.Terrain, 398)...)
		m.Spots = []domain.Spot{{Name: "A", At: domain.Point{X: 399}, Value: 1000}}
		s.apply([]domain.ExpeditionMap{m}, monsterOf(140, 350), nil)
		Expect(s.boxes[0].cells[399].bg.FillColor).To(Equal(dimFill))
	})

	It("titles the map with its name and what the route is worth, as a heading and not a panel that opens, and says the route", func() {
		s.apply(maps(), monsterOf(140, 350), nil)
		Expect(s.boxes[0].title.Text).To(Equal("Route · Parepare Jungle: 1 place, about 1,000 G"))
		Expect(s.boxes[0].summary.Text).To(Equal("Starts with 140 energy · rocks break 50% · the route costs about 13 energy"))
	})

	It("says that nothing is in reach in the title, when it is not", func() {
		m := aisle()
		m.Width, m.Tiles = 400, append([]domain.Terrain{domain.Floor, domain.Floor}, make([]domain.Terrain, 398)...)
		m.Spots = []domain.Spot{{Name: "A", At: domain.Point{X: 399}, Value: 1000}}
		s.apply([]domain.ExpeditionMap{m}, monsterOf(140, 350), nil)
		Expect(s.boxes[0].title.Text).To(Equal("Route · Parepare Jungle: nothing in reach"))
	})

	It("leaves what is drawn alone when the same route comes again, and draws it again when the route is another", func() {
		s.apply(maps(), monsterOf(140, 350), nil)
		object := s.boxes[0].object
		s.apply(maps(), monsterOf(140, 350), nil)
		Expect(s.boxes[0].object).To(BeIdenticalTo(object))

		s.apply(maps(), monsterOf(700, 350), nil)
		Expect(s.boxes[0].object).NotTo(BeIdenticalTo(object))
	})

	It("draws again when the monster comes onto the trip or moves on it, though the route is the same", func() {
		at := func(tile int) domain.Profile {
			p := monsterOf(140, 350)
			p.Player.Ranch = player.Ranch{Known: true, Date: 24, Tile: tile, WalkKnown: true}
			return p
		}
		s.apply(maps(), monsterOf(140, 350), nil)
		off := s.boxes[0].object

		s.apply(maps(), at(976), nil)
		Expect(s.boxes[0].object).NotTo(BeIdenticalTo(off))
		Expect(s.boxes[0].standing).NotTo(BeEmpty())
		here := s.boxes[0].object

		s.apply(maps(), at(976), nil)
		Expect(s.boxes[0].object).To(BeIdenticalTo(here), "the same place: left alone")

		s.apply(maps(), at(977), nil)
		Expect(s.boxes[0].object).NotTo(BeIdenticalTo(here))
	})

	It("says why there is no route when the monster cannot be had, and hides the maps", func() {
		s.apply(maps(), monsterOf(140, 350), nil)
		s.apply(maps(), domain.Profile{}, apperr.NotFound("", "no monster"))
		Expect(s.status.Visible()).To(BeTrue())
		Expect(s.status.Text).NotTo(BeEmpty())
		Expect(s.boxes).To(BeEmpty())
	})

	It("shows a gate that the monster can find the way through apart from one that is shut to it, and a teleport", func() {
		m := aisle()
		m.Width = 5
		m.Tiles = []domain.Terrain{domain.Floor, domain.Gate, domain.Floor, domain.Teleport, domain.Wall}
		m.Locks = []domain.Lock{{At: domain.Point{X: 1}, MinINT: 300}}
		s.apply([]domain.ExpeditionMap{m}, monsterWith(700, 350, 300, 0, 0), nil)
		cells := s.boxes[0].cells
		Expect(cells[1].text.Text).To(Equal("3"))
		Expect(cells[1].bg.FillColor).To(Equal(gateFill))
		Expect(cells[3].text.Text).To(Equal("T"))
		Expect(cells[3].bg.FillColor).To(Equal(teleportFill))

		s.apply([]domain.ExpeditionMap{m}, monsterWith(700, 350, 200, 0, 0), nil)
		Expect(s.boxes[0].cells[1].bg.FillColor).To(Equal(shutFill))
	})

	Describe("with the teleports of a map", func() {
		teleporting := func() domain.ExpeditionMap {
			m := aisle()
			m.Width = 6
			m.Tiles = []domain.Terrain{domain.Floor, domain.Floor, domain.Floor, domain.Teleport, domain.Floor, domain.Wall}
			m.Portals = []domain.Portal{
				{Label: "T1", At: domain.Point{X: 3}, To: []string{"E1", "E2"}},
				{Label: "E1", At: domain.Point{X: 1}},
				{Label: "E2", At: domain.Point{X: 4}},
			}
			return m
		}

		It("writes the label of a teleport on its tile, and of a place where one lands on that tile, as the guide does", func() {
			s.apply([]domain.ExpeditionMap{teleporting()}, monsterOf(140, 350), nil)
			cells := s.boxes[0].cells
			Expect(cells[3].text.Text).To(Equal("T1"))
			Expect(cells[3].bg.FillColor).To(Equal(teleportFill))
			Expect(cells[4].text.Text).To(Equal("E2"))
			Expect(cells[4].bg.FillColor).To(Equal(teleportFill))
		})

		It("draws a place to land as one, even where the route crosses it", func() {
			s.apply([]domain.ExpeditionMap{teleporting()}, monsterOf(140, 350), nil)
			c := s.boxes[0].cells[1]
			Expect(c.text.Text).To(Equal("E1"))
			Expect(c.bg.FillColor).To(Equal(teleportFill))
		})

		It("says under the map where each teleport leads", func() {
			s.apply([]domain.ExpeditionMap{teleporting()}, monsterOf(140, 350), nil)
			Expect(s.boxes[0].teleports.Visible()).To(BeTrue())
			Expect(s.boxes[0].teleports.Text).To(Equal("T1 takes you to E1 or E2, at random"))
		})

		It("has no line about teleports where the map has none", func() {
			s.apply(maps(), monsterOf(140, 350), nil)
			Expect(s.boxes[0].teleports.Visible()).To(BeFalse())
		})
	})

	Describe("the key to the map", func() {
		words := func(b *box) []string {
			var out []string
			for _, k := range b.key {
				out = append(out, k.label.Text)
			}
			return out
		}
		swatch := func(b *box, word string) *cell {
			for _, k := range b.key {
				if k.label.Text == word {
					return k.swatch
				}
			}
			Fail("no tile in the key for " + word)
			return nil
		}

		It("shows the tiles themselves with a few words each, only the kinds that the map has", func() {
			s.apply(maps(), monsterOf(140, 350), nil)
			b := s.boxes[0]
			Expect(words(b)).To(Equal([]string{"Start", "A stop, in the order to go"}))
			Expect(swatch(b, "Start").text.Text).To(Equal("S"))
			Expect(swatch(b, "Start").bg.FillColor).To(Equal(startFill))
			Expect(swatch(b, "A stop, in the order to go").text.Text).To(Equal("1"))
			Expect(swatch(b, "A stop, in the order to go").bg.FillColor).To(Equal(stopFill))
		})

		It("marks a place searched and left off the route in its own colour, and says so in the key", func() {
			m := aisle()
			m.Spots[0].Value = 0
			v, err := NewView(monsterOf(140, 350), m, map[string][]int{"A": {46}})
			Expect(err).NotTo(HaveOccurred())
			b := newBox(v, tableSize{w: 4, h: 1})
			Expect(b.cells[2].text.Text).To(Equal("A"))
			Expect(b.cells[2].bg.FillColor).To(Equal(doneFill))
			Expect(words(b)).To(ContainElement("A place searched"))
			Expect(swatch(b, "A place searched").bg.FillColor).To(Equal(doneFill))
			Expect(swatch(b, "A place searched").text.Text).To(Equal("A"))
		})

		It("has no long text of its own, as the map is told by its tiles", func() {
			s.apply(maps(), monsterOf(140, 350), nil)
			for _, w := range words(s.boxes[0]) {
				Expect(len(w)).To(BeNumerically("<=", 30), w)
			}
		})

		It("shows a place that the route leaves out, and says nothing of a route where there is none", func() {
			m := aisle()
			m.Width, m.Tiles = 400, append([]domain.Terrain{domain.Floor, domain.Floor}, make([]domain.Terrain, 398)...)
			m.Spots = []domain.Spot{{Name: "A", At: domain.Point{X: 399}, Value: 1000}}
			s.apply([]domain.ExpeditionMap{m}, monsterOf(140, 350), nil)
			b := s.boxes[0]
			Expect(words(b)).To(Equal([]string{"Start", "A place off the route"}))
			Expect(swatch(b, "A place off the route").bg.FillColor).To(Equal(dimFill))
		})

		It("shows a rock as it is drawn, whether or not the route breaks it", func() {
			m := aisle()
			m.Width = 6
			m.Tiles = []domain.Terrain{domain.Floor, domain.Rock, domain.Floor, domain.Wall, domain.Rock, domain.Wall}
			s.apply([]domain.ExpeditionMap{m}, monsterOf(700, 500), nil)
			b := s.boxes[0]
			Expect(words(b)).To(ContainElement("Rock"))
			Expect(words(b)).NotTo(ContainElement("Rock the route breaks"))
			Expect(swatch(b, "Rock").bg.FillColor).To(Equal(rockFill))
			Expect(swatch(b, "Rock").text.Text).To(BeEmpty())
			for _, c := range b.cells {
				Expect(c.text.Text).NotTo(Equal("•"))
			}
		})

		It("shows a hidden way that the monster can open, with the number that it has on the map, and one that it cannot", func() {
			m := aisle()
			m.Width = 6
			m.Tiles = []domain.Terrain{domain.Floor, domain.Gate, domain.Floor, domain.Wall, domain.Gate, domain.Wall}
			m.Locks = []domain.Lock{{At: domain.Point{X: 1}, MinINT: 300}, {At: domain.Point{X: 4}, MinINT: 500}}
			s.apply([]domain.ExpeditionMap{m}, monsterWith(700, 350, 300, 0, 0), nil)
			b := s.boxes[0]
			Expect(words(b)).To(ContainElements("Hidden way, needs 300 INT", "Hidden way, INT too low"))
			Expect(swatch(b, "Hidden way, needs 300 INT").text.Text).To(Equal("3"))
			Expect(swatch(b, "Hidden way, needs 300 INT").bg.FillColor).To(Equal(gateFill))
			Expect(swatch(b, "Hidden way, INT too low").text.Text).To(Equal("5"))
			Expect(swatch(b, "Hidden way, INT too low").bg.FillColor).To(Equal(shutFill))
		})

		It("shows a teleport and a place where one lands as the map labels them", func() {
			m := aisle()
			m.Width = 6
			m.Tiles = []domain.Terrain{domain.Floor, domain.Floor, domain.Floor, domain.Teleport, domain.Floor, domain.Wall}
			m.Portals = []domain.Portal{
				{Label: "T1", At: domain.Point{X: 3}, To: []string{"E1"}},
				{Label: "E1", At: domain.Point{X: 4}},
			}
			s.apply([]domain.ExpeditionMap{m}, monsterOf(140, 350), nil)
			b := s.boxes[0]
			Expect(words(b)).To(ContainElements("Teleport, not on the route", "Where a teleport lands"))
			Expect(swatch(b, "Teleport, not on the route").text.Text).To(Equal("T1"))
			Expect(swatch(b, "Teleport, not on the route").bg.FillColor).To(Equal(teleportFill))
			Expect(swatch(b, "Where a teleport lands").text.Text).To(Equal("E1"))
			Expect(swatch(b, "Where a teleport lands").bg.FillColor).To(Equal(teleportFill))
		})

		It("shows where the monster is only when it is on the trip", func() {
			p := monsterOf(140, 350)
			s.apply(maps(), p, nil)
			Expect(words(s.boxes[0])).NotTo(ContainElement("Where the monster is"))

			p.Player.Ranch = player.Ranch{Known: true, Date: 24, Tile: 976, WalkKnown: true}
			s.apply(maps(), p, nil)
			b := s.boxes[0]
			Expect(words(b)).To(ContainElement("Where the monster is"))
			Expect(swatch(b, "Where the monster is").text.Text).To(Equal("★"))
			Expect(swatch(b, "Where the monster is").bg.FillColor).To(Equal(markerFill))
		})
	})

	It("centres a map in a panel that is wider than it", func() {
		s.apply(maps(), monsterOf(140, 350), nil)
		b := s.boxes[0]
		b.table.Resize(fyne.NewSize(300, 20))
		Expect(b.cells[0].object.Position().X).To(BeNumerically("~", (300-71)/2, 0.5))
		Expect(b.cells[1].object.Position().X).To(BeNumerically("~", (300-71)/2+18, 0.5))

		b.table.Resize(fyne.NewSize(50, 20))
		Expect(b.cells[0].object.Position().X).To(BeNumerically(">=", 0))
	})

	It("keeps nothing that scrolls inside a map, so that the wheel over a map scrolls the column", func() {
		s.apply(maps(), monsterOf(140, 350), nil)
		for _, o := range s.boxes[0].object.(*fyne.Container).Objects {
			Expect(o).NotTo(BeAssignableToTypeOf(&container.Scroll{}))
		}
	})

	Describe("when the monster is on a trip", func() {
		onTrip := func(date, tile int) domain.Profile {
			p := monsterWith(420, 400, 600, 300, 0)
			p.Player.Ranch = player.Ranch{Known: true, Date: date, Tile: tile, WalkKnown: true}
			return p
		}
		real := []domain.ExpeditionMap{trip("Parepare Jungle"), trip("Torles Mountains"), trip("Kawrea Volcano")}

		It("shows only the map of the trip that this week brings, with the monster on it", func() {
			s.apply(real, onTrip(40, 975), nil)
			Expect(s.Content().Visible()).To(BeTrue())
			Expect(s.boxes).To(HaveLen(1))
			Expect(s.boxes[0].title.Text).To(HavePrefix("Route · Torles Mountains: "))
			Expect(s.boxes[0].cells[5].text.Text).To(Equal("★"))
		})

		It("draws the map of the trip as big as the biggest of all the maps, as it does when every map is shown", func() {
			wide := trip("Parepare Jungle")
			wide.Width, wide.Height, wide.Tiles = 3, 6, make([]domain.Terrain, 18)
			s.apply([]domain.ExpeditionMap{trip("Torles Mountains"), wide, trip("Kawrea Volcano")}, onTrip(40, 975), nil)
			Expect(s.boxes).To(HaveLen(1))
			b := s.boxes[0]
			b.table.Resize(b.table.MinSize())

			Expect(b.table.MinSize().Width).To(BeNumerically("~", 3*(cellSize+cellGap)-cellGap, 0.5))
			Expect(b.cells[0].object.Position().X).To(BeNumerically("~", cellSize+cellGap, 0.5), "a tile in from the edge")
		})

		It("finds the trip of a monster that is three steps up from the start in June, as the game showed it at X 0, Y 3", func() {
			p := monsterWith(537, 918, 463, 690, 100)
			p.Player.Ranch = player.Ranch{Known: true, Date: 872, Tile: 27*32 + 15, WalkKnown: true, Energy: 487, EnergyKnown: true}
			s.apply(real, p, nil)
			Expect(s.boxes).To(HaveLen(1))
			Expect(s.boxes[0].title.Text).To(HavePrefix("Route · Kawrea Volcano: "))
			Expect(s.boxes[0].summary.Text).To(HavePrefix("Rocks break 100%"), "its energy and where it is are in the boxes above")
			here := s.boxes[0].cells[2]
			Expect(here.text.Text).To(Equal("★"))
			Expect(here.bg.FillColor).To(Equal(markerFill))
			Expect(s.boxes[0].standing).To(HaveLen(2))
			Expect(s.boxes[0].routeRows).NotTo(BeEmpty())

			if dir := os.Getenv("SCREENSHOT_DIR"); dir != "" {
				f, err := os.Create(filepath.Join(dir, "route-kawrea-here.png"))
				Expect(err).NotTo(HaveOccurred())
				defer f.Close()
				Expect(png.Encode(f, software.Render(s.boxes[0].object, theme.DefaultTheme()))).To(Succeed())
			}
		})

		It("has no boxes above a map that the monster is not on", func() {
			s.apply(real, monsterOf(420, 400), nil)
			Expect(s.boxes).NotTo(BeEmpty())
			for _, b := range s.boxes {
				Expect(b.standing).To(BeEmpty())
			}
		})

		It("shows nothing in the other weeks", func() {
			s.apply(real, onTrip(40, 975), nil)
			s.apply(real, onTrip(41, 975), nil)
			Expect(s.boxes).To(BeEmpty())
			Expect(s.Content().Visible()).To(BeFalse())
		})

		It("shows the map of the week where the place of the monster on it is not known, with no mark on it", func() {
			p := onTrip(40, 975)
			p.Player.Ranch.WalkKnown = false
			s.apply(real, p, nil)
			Expect(s.boxes).To(HaveLen(1))
			Expect(s.boxes[0].title.Text).To(HavePrefix("Route · Torles Mountains: "))
			for _, c := range s.boxes[0].cells {
				Expect(c.text.Text).NotTo(Equal("★"))
			}
		})

		It("shows nothing of the trip of the week to a monster that is not fit for it", func() {
			weak := onTrip(40, 975)
			weak.Monster.Stats[monster.Life] = 209
			s.apply(real, weak, nil)
			Expect(s.boxes).To(BeEmpty())
			Expect(s.Content().Visible()).To(BeFalse())
		})
	})
})
