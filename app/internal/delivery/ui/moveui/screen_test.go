package moveui

import (
	"context"
	"errors"
	"image/png"
	"os"
	"path/filepath"

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
	grid := func() domain.Grid {
		g := domain.Grid{Genus: domain.Genus{Code: 0, Name: "Pixie"}}
		g.Cells[0][0] = domain.Cell{Move: &pat, Learned: true, Active: true, Used: 31}
		g.Cells[0][1] = domain.Cell{Move: &slap, CanLearn: true}
		g.Active[0] = &pat
		return g
	}

	BeforeEach(func() {
		test.NewTempApp(GinkgoTB())
		ctx = context.Background()
		svc = mockmove.NewMockMoveService(GinkgoT())
		s = NewScreen(svc, icons.Bytes)
	})

	It("keeps the names of its overlays in step with the icon assets", func() {
		for _, key := range []string{NoMove, Checked, CanGet, Highlight} {
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

	It("marks a learned move with the check and a learnable one with the marker", func() {
		s.apply(grid(), nil)
		Expect(s.cells[0][0].check.Visible()).To(BeTrue())
		Expect(s.cells[0][0].canGet.Visible()).To(BeFalse())
		Expect(s.cells[0][1].check.Visible()).To(BeFalse())
		Expect(s.cells[0][1].canGet.Visible()).To(BeTrue())
		Expect(s.cells[2][0].check.Visible()).To(BeFalse())
	})

	It("shows the active move of each column under the grid", func() {
		s.apply(grid(), nil)
		Expect(s.active[0].key).To(Equal("Pixie_1"))
		Expect(s.active[1].key).To(Equal(NoMove))
	})

	It("shows the details of a tapped cell and highlights only that cell", func() {
		s.apply(grid(), nil)
		test.Tap(s.cells[0][1])
		Expect(s.detail.name.Text).To(Equal("Slap"))
		Expect(s.detail.damage.Text).To(Equal("D (15)"))
		Expect(s.detail.withering.Text).To(Equal("E (5)"))
		Expect(s.detail.description.Text).To(Equal("Basic tech."))
		Expect(s.detail.canLearn.Checked).To(BeTrue())
		Expect(s.cells[0][1].highlight.Visible()).To(BeTrue())
		Expect(s.cells[0][0].highlight.Visible()).To(BeFalse())
	})

	It("lists what the selected move still needs, each as met or not, and no list when it needs nothing", func() {
		g := grid()
		g.Cells[0][1].Requirements = []domain.Requirement{
			{Text: "Pow + Int at least 400", Have: "421", Met: true},
			{Text: "Use Pat at least 50 times", Have: "31 / 50"},
			{Text: "The game gives no way to learn it"},
		}
		s.apply(g, nil)
		test.Tap(s.cells[0][1])
		Expect(s.detail.needsBox.Visible()).To(BeTrue())
		Expect(s.detail.needs).To(HaveLen(3))
		Expect(s.detail.needs[0].mark.Text).To(Equal("✓"))
		Expect(s.detail.needs[0].label.Text).To(Equal("Pow + Int at least 400 (421)"))
		Expect(s.detail.needs[1].mark.Text).To(Equal("✗"))
		Expect(s.detail.needs[1].label.Text).To(Equal("Use Pat at least 50 times (31 / 50)"))
		Expect(s.detail.needs[2].label.Text).To(Equal("The game gives no way to learn it"))

		test.Tap(s.cells[0][0])
		Expect(s.detail.needsBox.Visible()).To(BeFalse())
		Expect(s.detail.needs).To(BeEmpty())
	})

	It("keeps the selection and brings its details up to date when the grid changes", func() {
		s.apply(grid(), nil)
		test.Tap(s.cells[0][1])
		Expect(s.detail.learned.Checked).To(BeFalse())

		g := grid()
		g.Cells[0][1].Learned = true
		s.apply(g, nil)
		Expect(s.detail.learned.Checked).To(BeTrue())
		Expect(s.cells[0][1].highlight.Visible()).To(BeTrue())
	})

	It("shows No monster when the game shows none", func() {
		s.apply(grid(), nil)
		s.apply(domain.Grid{}, apperr.NotFound("monster", "No monster"))
		Expect(s.genus.Text).To(Equal("No monster"))
		Expect(s.cells[0][0].iconKey).To(Equal(NoMove))
		Expect(s.cells[0][0].check.Visible()).To(BeFalse())
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

	// A screenshot of the real thing, for looking at: SCREENSHOT_DIR=/some/dir ginkgo run --label-filter=screenshot
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
		// The screen is a scroll, whose smallest size is a few pixels, so draw what it holds at its own size.
		Expect(png.Encode(f, software.Render(real.body, theme.DefaultTheme()))).To(Succeed())
	})
})
