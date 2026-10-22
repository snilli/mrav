package moveui

import (
	"context"
	"errors"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/software"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"orm"

	"app/internal/core/apperr"
	monster "app/internal/core/monster/domain"
	"app/internal/core/move/domain"
	"app/internal/core/move/repository/db/moverepo"
	"app/internal/core/move/service/movesvc"
	"app/internal/delivery/ui/profileui"
	"app/internal/infrastructure/fakemonster"
	"app/internal/infrastructure/icons"
	mockmove "app/mock/core/move"
)

var _ = Describe("Screen", func() {
	var (
		ctx context.Context
		svc *mockmove.MockMoveService
		s   *Screen
	)
	pat := domain.Move{Col: 1, Row: 1, Name: "Pat", Kind: domain.KindPower, Guts: 12, Force: 6, Hit: 10, Description: "Starting basic tech.", Icon: "Pixie_1"}
	slap := domain.Move{Col: 1, Row: 2, Name: "Slap", Kind: domain.KindPower, Guts: 18, Force: 15, Hit: 9, Wither: 5, Description: "Basic tech.", Icon: "Pixie_2"}
	bolt := domain.Move{Col: 1, Row: 3, Name: "Bolt", Kind: domain.KindIntelligence, Guts: 20, Description: "Needs more.", Icon: "Pixie_3"}
	grid := func() domain.Grid {
		g := domain.Grid{Genus: domain.Genus{Code: 0, Name: "Pixie"}}
		g.Cells[0][0] = domain.Cell{Move: &pat, Learned: true, Active: true, Used: 31}
		g.Cells[0][1] = domain.Cell{Move: &slap, CanLearn: true}
		g.Cells[0][2] = domain.Cell{Move: &bolt}
		g.Active[0] = &pat
		return g
	}

	BeforeEach(func() {
		test.NewTempApp(GinkgoTB())
		ctx = context.Background()
		svc = mockmove.NewMockMoveService(GinkgoT())
		s = NewScreen(svc, icons.Bytes)
	})

	It("shows the columns of the grid from the last to the first, with the active moves under them in that order", func() {
		w := test.NewWindow(s.Content())
		DeferCleanup(w.Close)
		w.Resize(fyne.NewSize(900, 900))
		x := func(o fyne.CanvasObject) float32 { return fyne.CurrentApp().Driver().AbsolutePositionForObject(o).X }
		for c := 1; c < domain.Columns; c++ {
			Expect(x(s.cells[c][0])).To(BeNumerically("<", x(s.cells[c-1][0])), "column %d is left of column %d", c+1, c)
			Expect(x(s.active[c].img)).To(BeNumerically("<", x(s.active[c-1].img)), "active move %d is left of %d", c+1, c)
		}
	})

	It("centres the grid of moves and the active moves in the panel, and not at its left", func() {
		w := test.NewWindow(s.Content())
		DeferCleanup(w.Close)
		w.Resize(fyne.NewSize(900, 900))
		x := func(o fyne.CanvasObject) float32 { return fyne.CurrentApp().Driver().AbsolutePositionForObject(o).X }
		for name, o := range map[string]fyne.CanvasObject{"the grid": s.cellGrid, "the active moves": s.activeRow} {
			left := x(o) - x(s.body)
			right := x(s.body) + s.body.Size().Width - (x(o) + o.Size().Width)
			Expect(left).To(BeNumerically(">", 10), name)
			Expect(left).To(BeNumerically("~", right, 1.5), name)
		}
	})

	It("keeps the names of its overlays in step with the icon assets", func() {
		for _, key := range []string{NoMove, Highlight} {
			_, err := icons.Bytes(key)
			Expect(err).NotTo(HaveOccurred(), key)
		}
	})

	It("shows every cell of the grid with its icon, and an empty one as No Move", func() {
		s.apply(grid(), nil)
		Expect(s.cells[0][0].iconKey).To(Equal("Pixie_1"))
		Expect(s.cells[0][1].iconKey).To(Equal("Pixie_2"))
		Expect(s.cells[3][5].iconKey).To(Equal(NoMove))
		Expect(s.genus.Text).To(Equal("Pixie"))
	})

	It("draws no mark in the corner of a learned move or one ready to learn, because its frame says that", func() {
		s.apply(grid(), nil)
		for _, c := range []*cell{s.cells[0][0], s.cells[0][1]} {
			for _, o := range c.body.Objects {
				if o.Visible() {
					Expect(o.Position()).NotTo(Equal(fyne.NewPos(30*zoom, 2*zoom)))
				}
			}
		}
	})

	It("shows the active move of each column under the grid", func() {
		s.apply(grid(), nil)
		Expect(s.active[0].key).To(Equal("Pixie_1"))
		Expect(s.active[1].key).To(Equal(NoMove))
	})

	It("colours the icon of a move by its kind, yellow for power and green for intelligence, and the active move too", func() {
		s.apply(grid(), nil)
		power := func(c color.NRGBA) bool { return c == powerFill }
		green := func(c color.NRGBA) bool { return c == intelligenceFill }
		white := func(c color.NRGBA) bool { return c.R == 255 && c.G == 255 && c.B == 255 }

		pat := s.cells[0][0].icon.Resource.Content()
		Expect(colours(pat, power)).To(BeNumerically(">", 0))
		Expect(colours(pat, white)).To(BeZero())
		Expect(colours(pat, green)).To(BeZero())
		bolt := s.cells[0][2].icon.Resource.Content()
		Expect(colours(bolt, green)).To(BeNumerically(">", 0))
		Expect(colours(bolt, power)).To(BeZero())
		active := s.active[0].img.Resource.Content()
		Expect(colours(active, power)).To(BeNumerically(">", 0))
		empty := s.cells[3][5].icon.Resource.Content()
		Expect(colours(empty, power) + colours(empty, green)).To(BeZero())
	})

	It("keeps the colour of a cell when its move changes, and gives the new move its own", func() {
		s.apply(grid(), nil)
		g := grid()
		g.Cells[0][0] = domain.Cell{Move: &bolt}
		s.apply(g, nil)
		Expect(colours(s.cells[0][0].icon.Resource.Content(), func(c color.NRGBA) bool { return c == intelligenceFill })).To(BeNumerically(">", 0))
	})

	It("marks a move that asks for a good nature or a bad one on its icon, and none that asks for none", func() {
		hug := domain.Move{Col: 2, Row: 1, Name: "Hug", Kind: domain.KindPower, Icon: "Pixie_4",
			Unlock: domain.UnlockRule{All: []domain.Condition{domain.NatureAtLeast{Min: 20}}}}
		bite := domain.Move{Col: 2, Row: 2, Name: "Bite", Kind: domain.KindPower, Icon: "Pixie_5",
			Unlock: domain.UnlockRule{All: []domain.Condition{domain.NatureAtMost{Max: -20}}}}
		g := grid()
		g.Cells[1][0] = domain.Cell{Move: &hug}
		g.Cells[1][1] = domain.Cell{Move: &bite}
		s.apply(g, nil)
		Expect(s.cells[1][0].nature.Visible()).To(BeTrue())
		Expect(s.cells[1][0].nature.Resource.Content()).To(Equal(natureIcon(domain.NatureGood)))
		Expect(s.cells[1][1].nature.Visible()).To(BeTrue())
		Expect(s.cells[1][1].nature.Resource.Content()).To(Equal(natureIcon(domain.NatureBad)))
		Expect(s.cells[0][0].nature.Visible()).To(BeFalse())
		Expect(s.cells[3][5].nature.Visible()).To(BeFalse())

		s.apply(grid(), nil)
		Expect(s.cells[1][0].nature.Visible()).To(BeFalse())
	})

	It("shows the nature that the selected move asks for beside its name", func() {
		hug := domain.Move{Col: 2, Row: 1, Name: "Hug", Kind: domain.KindPower, Icon: "Pixie_4",
			Unlock: domain.UnlockRule{All: []domain.Condition{domain.NatureAtMost{Max: -50}}}}
		g := grid()
		g.Cells[1][0] = domain.Cell{Move: &hug}
		s.apply(g, nil)
		test.Tap(s.cells[1][0])
		Expect(s.detail.nature.Visible()).To(BeTrue())
		Expect(s.detail.nature.Resource.Content()).To(Equal(natureIcon(domain.NatureBad)))
		test.Tap(s.cells[0][0])
		Expect(s.detail.nature.Visible()).To(BeFalse())
	})

	It("shows the details of a tapped cell and highlights only that cell", func() {
		s.apply(grid(), nil)
		test.Tap(s.cells[0][1])
		Expect(s.detail.name.Text).To(Equal("Slap"))
		Expect(s.detail.damage.Text).To(Equal("D (15)"))
		Expect(s.detail.withering.Text).To(Equal("E (5)"))
		Expect(s.detail.description.Text).To(Equal("Type: Basic"))
		Expect(s.detail.status.txt.Text).To(Equal("READY TO LEARN"))
		Expect(s.cells[0][1].highlight.Visible()).To(BeTrue())
		Expect(s.cells[0][0].highlight.Visible()).To(BeFalse())
	})

	It("frames a learned move in green, a move ready to learn in orange with a ribbon, and dims a move that is not yet", func() {
		s.apply(grid(), nil)
		learned, ready, locked, empty := s.cells[0][0], s.cells[0][1], s.cells[0][2], s.cells[3][5]

		Expect(learned.frame.Visible()).To(BeTrue())
		Expect(learned.frame.StrokeColor).To(Equal(metColor))
		Expect(learned.dim.Visible()).To(BeFalse())
		Expect(learned.ribbon.Visible()).To(BeFalse())

		Expect(ready.frame.Visible()).To(BeTrue())
		Expect(ready.frame.StrokeColor).To(Equal(readyColor))
		Expect(ready.ribbon.Visible()).To(BeTrue())
		Expect(ready.dim.Visible()).To(BeFalse())

		Expect(locked.dim.Visible()).To(BeTrue())
		Expect(locked.frame.Visible()).To(BeFalse())
		Expect(locked.ribbon.Visible()).To(BeFalse())

		Expect(empty.frame.Visible()).To(BeFalse())
		Expect(empty.dim.Visible()).To(BeFalse())
		Expect(empty.ribbon.Visible()).To(BeFalse())
	})

	It("takes the frame, the dimming and the ribbon away when a move is learned", func() {
		s.apply(grid(), nil)
		g := grid()
		g.Cells[0][1].CanLearn, g.Cells[0][1].Learned = false, true
		s.apply(g, nil)
		Expect(s.cells[0][1].frame.StrokeColor).To(Equal(metColor))
		Expect(s.cells[0][1].ribbon.Visible()).To(BeFalse())
	})

	It("counts the moves learned, ready to learn, a chance of the errantry and not yet, above the grid", func() {
		s.apply(grid(), nil)
		Expect(s.legend.learned.Text).To(Equal("Learned 1"))
		Expect(s.legend.ready.Text).To(Equal("Ready to learn 1"))
		Expect(s.legend.chance.Text).To(Equal("Errantry chance 0"))
		Expect(s.legend.locked.Text).To(Equal("Not yet 1"))

		g := grid()
		g.Cells[0][2].Chance, g.Cells[0][2].Place = 2.96, "Mandy Desert"
		s.apply(g, nil)
		Expect(s.legend.chance.Text).To(Equal("Errantry chance 1"))
		Expect(s.legend.locked.Text).To(Equal("Not yet 0"))
	})

	Describe("a move that a place of the errantry teaches only by luck", func() {
		luck := func() domain.Grid {
			g := grid()
			g.Cells[0][2].Chance, g.Cells[0][2].Place = 2.96, "Mandy Desert"
			return g
		}

		It("has a ribbon with its chance and a frame, both in a colour of its own, as a move that is learned or ready has, and no dimming", func() {
			s.apply(luck(), nil)
			c, ready := s.cells[0][2], s.cells[0][1]

			Expect(c.ribbon.Visible()).To(BeTrue())
			Expect(c.ribbonTxt.Text).To(Equal("3%"))
			Expect(c.ribbonBg.FillColor).To(Equal(chanceColor))
			Expect(c.frame.Visible()).To(BeTrue())
			Expect(c.frame.StrokeColor).To(Equal(chanceColor))
			Expect(c.dim.Visible()).To(BeFalse())

			Expect(ready.ribbonTxt.Text).To(Equal("READY"), "a move that is ready keeps its word and its colour")
			Expect(ready.ribbonBg.FillColor).To(Equal(readyColor))
		})

		It("goes back to the word and the colour of ready, or to no ribbon, when it changes", func() {
			s.apply(luck(), nil)
			g := luck()
			g.Cells[0][2].CanLearn = true
			s.apply(g, nil)
			Expect(s.cells[0][2].ribbonTxt.Text).To(Equal("READY"))
			Expect(s.cells[0][2].ribbonBg.FillColor).To(Equal(readyColor))
			Expect(s.cells[0][2].frame.StrokeColor).To(Equal(readyColor))

			s.apply(grid(), nil)
			Expect(s.cells[0][2].ribbon.Visible()).To(BeFalse())
			Expect(s.cells[0][2].frame.Visible()).To(BeFalse())
			Expect(s.cells[0][2].dim.Visible()).To(BeTrue())
		})

		It("says its chance alone on the badge of its details, in the colour of its ribbon, and its place on the line of its type", func() {
			s.apply(luck(), nil)
			test.Tap(s.cells[0][2])

			Expect(s.detail.status.txt.Text).To(Equal("3%"))
			Expect(s.detail.status.bg.FillColor).To(Equal(chanceColor))
			Expect(s.detail.description.Text).To(Equal("Type: Needs more · Location: Mandy Desert"))
		})
	})

	It("says where the selected move stands on a badge in the colour of its frame, and none for an empty cell", func() {
		s.apply(grid(), nil)
		for _, tc := range []struct {
			col, row int
			text     string
			fill     color.Color
		}{
			{0, 0, "LEARNED", metColor},
			{0, 1, "READY TO LEARN", readyColor},
			{0, 2, "NOT YET", lockedColor},
		} {
			test.Tap(s.cells[tc.col][tc.row])
			Expect(s.detail.status.object.Visible()).To(BeTrue())
			Expect(s.detail.status.txt.Text).To(Equal(tc.text))
			Expect(s.detail.status.bg.FillColor).To(Equal(tc.fill))
		}
		test.Tap(s.cells[3][5])
		Expect(s.detail.status.object.Visible()).To(BeFalse())
	})

	It("names the stats of a requirement in the colours of the stats, and says the rest in the colour of the text", func() {
		g := grid()
		g.Cells[0][1].Requirements = []domain.Requirement{
			{Text: "POW + INT at least 400", Have: "421", Stats: []monster.Stat{monster.Pow, monster.Int}},
		}
		s.apply(g, nil)
		test.Tap(s.cells[0][1])

		colours := map[string]color.Color{}
		for _, seg := range s.detail.needs[0].colored.Segments() {
			colours[seg.Text] = seg.Color
		}
		Expect(colours).To(HaveKeyWithValue("POW", profileui.StatColors[monster.Pow]))
		Expect(colours).To(HaveKeyWithValue("INT", profileui.StatColors[monster.Int]))
		Expect(colours[" + "]).To(Equal(theme.Color(theme.ColorNameForeground)))
		Expect(colours[" at least 400 (421)"]).To(Equal(theme.Color(theme.ColorNameForeground)))
	})

	It("lists what the selected move still needs, each as met or not, and no list when it needs nothing", func() {
		g := grid()
		g.Cells[0][1].Requirements = []domain.Requirement{
			{Text: "POW + INT at least 400", Have: "421", Met: true, Stats: []monster.Stat{monster.Pow, monster.Int}},
			{Text: "Use Pat at least 50 times", Have: "31 / 50"},
			{Text: "The game gives no way to learn it"},
		}
		s.apply(g, nil)
		test.Tap(s.cells[0][1])
		Expect(s.detail.needsBox.Visible()).To(BeTrue())
		Expect(s.detail.needs).To(HaveLen(3))
		Expect(s.detail.needs[0].mark.Text).To(Equal("✓"))
		Expect(s.detail.needs[0].words()).To(Equal("POW + INT at least 400 (421)"))
		Expect(s.detail.needs[1].mark.Text).To(Equal("✗"))
		Expect(s.detail.needs[1].words()).To(Equal("Use Pat at least 50 times (31 / 50)"))
		Expect(s.detail.needs[2].words()).To(Equal("The game gives no way to learn it"))

		test.Tap(s.cells[0][0])
		Expect(s.detail.needsBox.Visible()).To(BeFalse())
		Expect(s.detail.needs).To(BeEmpty())
	})

	It("keeps the selection and brings its details up to date when the grid changes", func() {
		s.apply(grid(), nil)
		test.Tap(s.cells[0][1])
		Expect(s.detail.status.txt.Text).To(Equal("READY TO LEARN"))

		g := grid()
		g.Cells[0][1].Learned = true
		s.apply(g, nil)
		Expect(s.detail.status.txt.Text).To(Equal("LEARNED"))
		Expect(s.cells[0][1].highlight.Visible()).To(BeTrue())
	})

	It("shows No monster when the game shows none", func() {
		s.apply(grid(), nil)
		s.apply(domain.Grid{}, apperr.NotFound("monster", "No monster"))
		Expect(s.genus.Text).To(Equal("No monster"))
		Expect(s.cells[0][0].iconKey).To(Equal(NoMove))
	})

	It("says what a not-found error says, such as that the game is not running, and No monster when it says nothing", func() {
		s.apply(domain.Grid{}, apperr.NotFound("game", "DuckStation is not running"))
		Expect(s.genus.Text).To(Equal("DuckStation is not running"))

		s.apply(domain.Grid{}, apperr.ErrNotFound)
		Expect(s.genus.Text).To(Equal("No monster"))
	})

	It("shows any other error on the genus line", func() {
		s.apply(domain.Grid{}, errors.New("db gone"))
		Expect(s.genus.Text).To(ContainSubstring("db gone"))
	})

	It("asks the service for the current grid and shows it", func() {
		svc.EXPECT().CurrentGrid(ctx).Return(grid(), nil)
		s.Load(ctx)
		Eventually(func() string { return s.genus.Text }).Should(Equal("Pixie"))
	})

	It("renders the real Dragon grid to a PNG", Label("screenshot"), func() {
		dir := os.Getenv("SCREENSHOT_DIR")
		if dir == "" {
			Skip("set SCREENSHOT_DIR to write the screenshots")
		}
		const dbPath = "../../../../../data/mrav.sqlite"
		if _, err := os.Stat(dbPath); err != nil {
			Skip("no " + dbPath)
		}
		client, err := orm.Open(ctx, dbPath)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)

		src := fakemonster.New()
		real := NewScreen(movesvc.New(moverepo.New(client), src), icons.Bytes)
		src.Update(func(m *monster.Monster) {
			m.Genus = 1
			for stat := range m.Stats {
				m.Stats[stat] = 300
			}
			m.Nature = 20
			for slot := 0; slot < monster.SlotCount; slot += 3 {
				m.Learned[slot], m.Used[slot] = true, 25+slot
			}
			m.Active = [monster.ActiveGroups]int{0, 6, 12, 18}
		})
		g, err := movesvc.New(moverepo.New(client), src).CurrentGrid(ctx)
		Expect(err).NotTo(HaveOccurred())
		real.apply(g, nil)
		test.Tap(real.cells[2][1])

		Expect(os.MkdirAll(dir, 0o755)).To(Succeed())
		f, err := os.Create(filepath.Join(dir, "move-viewer-dragon.png"))
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(f.Close)
		Expect(png.Encode(f, software.Render(real.body, theme.DefaultTheme()))).To(Succeed())
	})
})
