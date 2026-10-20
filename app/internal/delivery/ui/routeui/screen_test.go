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
	player "app/internal/core/player/domain"
	"app/internal/core/profile/domain"
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
		s = NewScreen(svc, true)
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

	It("shows a place that the route leaves out apart from one that it takes", func() {
		m := aisle()
		m.Width, m.Tiles = 400, append([]domain.Terrain{domain.Floor, domain.Floor}, make([]domain.Terrain, 398)...)
		m.Spots = []domain.Spot{{Name: "A", At: domain.Point{X: 399}, Value: 1000}}
		s.apply([]domain.ExpeditionMap{m}, monsterOf(140, 350), nil)
		Expect(s.boxes[0].cells[399].bg.FillColor).To(Equal(spotFill))
	})

	It("titles the panel with the map and what the route is worth, and says the route in it", func() {
		s.apply(maps(), monsterOf(140, 350), nil)
		Expect(s.boxes[0].item.Title).To(Equal("Route · Aisle: 1 place, about 1,000 G"))
		Expect(s.boxes[0].summary.Text).To(Equal("140 energy, rocks break 50%, INT 0, bad nature: 1 place, about 14 energy, worth about 1,000 G"))
		Expect(s.boxes[0].stops.Text).To(Equal("1. A · 1,000 G"))
	})

	It("says that nothing is in reach in the title, when it is not", func() {
		m := aisle()
		m.Width, m.Tiles = 400, append([]domain.Terrain{domain.Floor, domain.Floor}, make([]domain.Terrain, 398)...)
		m.Spots = []domain.Spot{{Name: "A", At: domain.Point{X: 399}, Value: 1000}}
		s.apply([]domain.ExpeditionMap{m}, monsterOf(140, 350), nil)
		Expect(s.boxes[0].item.Title).To(Equal("Route · Aisle: nothing in reach"))
	})

	It("leaves the panels alone when the same route comes again, so that an open one stays open", func() {
		s.apply(maps(), monsterOf(140, 350), nil)
		item := s.boxes[0].item
		s.apply(maps(), monsterOf(140, 350), nil)
		Expect(s.boxes[0].item).To(BeIdenticalTo(item))

		s.apply(maps(), monsterOf(700, 350), nil) // more life: the route is another
		Expect(s.boxes[0].item).NotTo(BeIdenticalTo(item))
	})

	It("keeps a panel open through the next route, so that a step on the trip does not close it, and one that is closed stays closed", func() {
		s.apply(maps(), monsterOf(140, 350), nil)
		s.boxes[0].panel.Open(0)
		s.apply(maps(), monsterOf(700, 350), nil) // more life: the route is another
		Expect(s.boxes[0].item.Open).To(BeTrue())

		s.boxes[0].panel.Close(0)
		s.apply(maps(), monsterOf(140, 350), nil)
		Expect(s.boxes[0].item.Open).To(BeFalse())
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

	It("centres a map in a panel that is wider than it", func() {
		s.apply(maps(), monsterOf(140, 350), nil) // the aisle, 4 tiles of 17 and 1 between: 71 wide
		b := s.boxes[0]
		b.table.Resize(fyne.NewSize(300, 20))
		Expect(b.cells[0].object.Position().X).To(BeNumerically("~", (300-71)/2, 0.5))
		Expect(b.cells[1].object.Position().X).To(BeNumerically("~", (300-71)/2+18, 0.5))

		b.table.Resize(fyne.NewSize(50, 20)) // and not left of its own edge when the panel is narrower
		Expect(b.cells[0].object.Position().X).To(BeNumerically(">=", 0))
	})

	It("keeps nothing that scrolls inside a panel, so that the wheel over a map scrolls the column", func() {
		s.apply(maps(), monsterOf(140, 350), nil)
		for _, o := range s.boxes[0].item.Detail.(*fyne.Container).Objects {
			Expect(o).NotTo(BeAssignableToTypeOf(&container.Scroll{}))
		}
	})

	Describe("in a game, not in the demo", func() {
		// onTrip is a monster that last stood on the tile given, at the date given of the game's count: 24 is October,
		// week 1, when the trip of Parepare comes round, and the tile 975 is the start of every trip.
		onTrip := func(date, tile int) domain.Profile {
			p := monsterWith(420, 400, 600, 300, 0)
			p.Player.Ranch = player.Ranch{Known: true, Date: date, Tile: tile, Steps: 2, WalkKnown: true}
			return p
		}
		real := []domain.ExpeditionMap{trip("Parepare Jungle"), trip("Torles Mountains"), trip("Kawrea Volcano")}

		BeforeEach(func() {
			s = NewScreen(svc, false)
		})

		It("shows only the map of the trip that this week brings, with the monster on it", func() {
			s.apply(real, onTrip(24, 975), nil)
			Expect(s.Content().Visible()).To(BeTrue())
			Expect(s.boxes).To(HaveLen(1))
			Expect(s.boxes[0].item.Title).To(HavePrefix("Route · Parepare Jungle: "))
			Expect(s.boxes[0].cells[5].text.Text).To(Equal("★")) // on the start, at the bottom of the corridor
		})

		It("finds the trip of a monster that is three steps up from the start in June, as the game showed it at X 0, Y 3", func() {
			p := monsterWith(537, 918, 463, 690, 100)
			p.Player.Ranch = player.Ranch{Known: true, Date: 872, Tile: 27*32 + 15, Steps: 3, WalkKnown: true, Energy: 487, EnergyKnown: true}
			s.apply(real, p, nil)
			Expect(s.boxes).To(HaveLen(1))
			Expect(s.boxes[0].item.Title).To(HavePrefix("Route · Kawrea Volcano: "))
			Expect(s.boxes[0].summary.Text).To(HavePrefix("487 energy left, at X 0, Y 3, rocks break 100%, INT 463, good nature: "))
			here := s.boxes[0].cells[2] // three tiles up from the start of the corridor, which is its tile 5
			Expect(here.text.Text).To(Equal("★"))
			Expect(here.bg.FillColor).To(Equal(markerFill))
			Expect(s.boxes[0].standing).To(HaveLen(3)) // the energy, the place and the steps, in boxes above the map
			Expect(s.boxes[0].item.Open).To(BeTrue(), "the panel of the trip that the monster is on is open when it appears")
			Expect(s.boxes[0].stops.Text).NotTo(BeEmpty())

			if dir := os.Getenv("SCREENSHOT_DIR"); dir != "" {
				f, err := os.Create(filepath.Join(dir, "route-kawrea-here.png"))
				Expect(err).NotTo(HaveOccurred())
				defer f.Close()
				Expect(png.Encode(f, software.Render(s.boxes[0].panel, theme.DefaultTheme()))).To(Succeed())
			}
		})

		It("leaves a panel that the breeder has closed closed, when it comes again on the trip", func() {
			p := monsterWith(537, 918, 463, 690, 100)
			p.Player.Ranch = player.Ranch{Known: true, Date: 872, Tile: 27*32 + 15, Steps: 3, WalkKnown: true, Energy: 487, EnergyKnown: true}
			s.apply(real, p, nil)
			s.boxes[0].panel.Close(0)
			p.Player.Ranch.Steps, p.Player.Ranch.Energy, p.Player.Ranch.Tile = 4, 486, 26*32+15
			s.apply(real, p, nil)
			Expect(s.boxes[0].item.Open).To(BeFalse())
		})

		It("has no boxes above a map that the monster is not on, and the panel is closed", func() {
			s = NewScreen(svc, true)
			s.apply(real, monsterOf(420, 400), nil)
			for _, b := range s.boxes {
				Expect(b.standing).To(BeEmpty())
				Expect(b.item.Open).To(BeFalse())
			}
		})

		It("shows nothing in the other weeks", func() {
			s.apply(real, onTrip(24, 975), nil)
			s.apply(real, onTrip(25, 975), nil)
			Expect(s.boxes).To(BeEmpty())
			Expect(s.Content().Visible()).To(BeFalse())
		})

		It("shows nothing where the place of the monster on the map is not known", func() {
			p := onTrip(24, 975)
			p.Player.Ranch.WalkKnown = false
			s.apply(real, p, nil)
			Expect(s.boxes).To(BeEmpty())
		})

		It("shows every map in the demo, and marks the monster only on the one that it is on", func() {
			s = NewScreen(svc, true)
			s.apply(real, onTrip(24, 975), nil)
			Expect(s.boxes).To(HaveLen(3))
			marked := 0
			for _, b := range s.boxes {
				for _, c := range b.cells {
					if c.text.Text == "★" {
						marked++
					}
				}
			}
			Expect(marked).To(Equal(1))
		})
	})
})
