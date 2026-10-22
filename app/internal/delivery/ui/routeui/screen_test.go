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

	It("draws each map as a table of tiles, the route marked in it", func() {
		s.apply(maps(), monsterOf(140, 350), nil)
		Expect(s.Content().Visible()).To(BeTrue())
		Expect(s.boxes).To(HaveLen(1))
		cells := s.boxes[0].cells
		Expect(cells).To(HaveLen(4))
		Expect(cells[0].text.Text).To(Equal("S"))
		Expect(cells[0].bg.FillColor).To(Equal(startFill))
		Expect(cells[1].bg.FillColor).To(Equal(routeFill))
		Expect(cells[2].text.Text).To(Equal("A"))
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

	It("lists the places under the map: the letter, the building, where it is in the route and what a search is worth, and what it may turn up", func() {
		m := aisle()
		m.Spots = []domain.Spot{{Name: "A", At: domain.Point{X: 2}, Building: "Cave", Value: 1000, Loot: []domain.Drop{
			{Item: "Mint Leaf", Pct: 25, Price: 100}, {Item: "Old Sheath", Pct: 18.8, Price: 50},
		}}}
		s.apply([]domain.ExpeditionMap{m}, monsterOf(140, 350), nil)
		Expect(s.boxes[0].places).To(HaveLen(1))
		a := s.boxes[0].places[0]
		Expect(a.letter.Box()).To(Equal(profileui.Box{Text: "A", Tone: profileui.ToneLimeGreen}))
		Expect(a.building.Text).To(Equal("Cave"))
		Expect(a.note.Text).To(Equal("stop 1 · about 1,000 G"))
		Expect(a.empty.Visible()).To(BeFalse())
		Expect(a.loot).To(HaveLen(2))
		Expect(a.loot[0].Text).To(Equal("Mint Leaf 25%"))
		Expect(a.loot[0].TextStyle.Bold).To(BeFalse())
		Expect(a.loot[1].Text).To(Equal("Old Sheath 18.8%"))
		Expect(a.loot[1].TextStyle.Bold).To(BeTrue(), "an item of the story")
	})

	It("lists a place that the route leaves out as left out, and says when nothing turns up there for the monster", func() {
		m := aisle()
		m.Width, m.Tiles = 400, append([]domain.Terrain{domain.Floor, domain.Floor}, make([]domain.Terrain, 398)...)
		m.Spots = []domain.Spot{{Name: "A", At: domain.Point{X: 399}, Building: "Cave", Value: 1000}}
		s.apply([]domain.ExpeditionMap{m}, monsterOf(140, 350), nil)
		a := s.boxes[0].places[0]
		Expect(a.letter.Box()).To(Equal(profileui.Box{Text: "A", Tone: profileui.ToneGray}))
		Expect(a.note.Text).To(Equal("not on the route · about 1,000 G"))
		Expect(a.loot).To(BeEmpty())
		Expect(a.empty.Visible()).To(BeTrue())
		Expect(a.empty.Text).To(Equal("Nothing turns up here for this monster yet"))
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
		Expect(s.boxes[0].cells[399].bg.FillColor).To(Equal(spotFill))
	})

	It("titles the map with its name and what the route is worth, as a heading and not a panel that opens, and says the route", func() {
		s.apply(maps(), monsterOf(140, 350), nil)
		Expect(s.boxes[0].title.Text).To(Equal("Route · Parepare Jungle: 1 place, about 1,000 G"))
		Expect(s.boxes[0].summary.Text).To(Equal("Starts with 140 energy · rocks break 50% · the route costs about 14 energy"))
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

		It("keeps the colour of the route on a place to land that the route crosses", func() {
			s.apply([]domain.ExpeditionMap{teleporting()}, monsterOf(140, 350), nil)
			c := s.boxes[0].cells[1]
			Expect(c.text.Text).To(Equal("E1"))
			Expect(c.bg.FillColor).To(Equal(routeFill))
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
			Expect(words(b)).To(Equal([]string{"Start", "The route", "A place the route searches"}))
			Expect(swatch(b, "Start").text.Text).To(Equal("S"))
			Expect(swatch(b, "Start").bg.FillColor).To(Equal(startFill))
			Expect(swatch(b, "The route").bg.FillColor).To(Equal(routeFill))
			Expect(swatch(b, "A place the route searches").text.Text).To(Equal("A"), "the letter of a place of the map")
			Expect(swatch(b, "A place the route searches").bg.FillColor).To(Equal(stopFill))
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
			Expect(words(b)).To(Equal([]string{"Start", "A place it leaves out"}))
			Expect(swatch(b, "A place it leaves out").bg.FillColor).To(Equal(spotFill))
		})

		It("shows a rock, and a rock that the route breaks, each as it is drawn", func() {
			m := aisle()
			m.Width = 6
			m.Tiles = []domain.Terrain{domain.Floor, domain.Rock, domain.Floor, domain.Wall, domain.Rock, domain.Wall}
			s.apply([]domain.ExpeditionMap{m}, monsterOf(700, 500), nil)
			b := s.boxes[0]
			Expect(words(b)).To(ContainElements("Rock", "Rock the route breaks"))
			Expect(swatch(b, "Rock").bg.FillColor).To(Equal(rockFill))
			Expect(swatch(b, "Rock").text.Text).To(BeEmpty())
			Expect(swatch(b, "Rock the route breaks").bg.FillColor).To(Equal(rockFill))
			Expect(swatch(b, "Rock the route breaks").text.Text).To(Equal("•"))
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
			Expect(s.boxes[0].places).NotTo(BeEmpty())

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
