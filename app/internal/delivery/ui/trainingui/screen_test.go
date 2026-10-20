package trainingui

import (
	"context"
	"errors"
	"fmt"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"slices"

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
	"app/internal/core/training/domain"
	"app/internal/delivery/ui/profileui"
	mocktraining "app/mock/core/training"
)

// trainingOf is the Training page of a monster at its life stage 4 whose growth rates are all 2, with a training
// boost of the bits given, in a table whose every text says where it is from, as "k1 (a/b)". Its motivation for
// the first drill is 85 and for the swim 100.
func trainingOf(boost int) domain.Training {
	var gains []domain.Gain
	for _, kind := range []int{domain.KindLight, domain.KindHeavy, domain.KindHeavySecondary} {
		gains = append(gains, domain.Gain{LifeStage: 4, GrowthRate: 2, Kind: kind, Text: fmt.Sprintf("range%d (k%d/b)", kind, kind)})
	}
	m := monster.Monster{
		LifeStage: 4, TrainBoost: boost, Growth: [monster.StatCount]int{2, 2, 2, 2, 2, 2},
		Motivation: [monster.DrillCount]int{85, 20, 30, 40, 50, 60, 70, 80, 90, 100},
	}
	return domain.Build(m, domain.NewGains(gains))
}

// withAccordion is whether there is an accordion anywhere in an object, in the containers and scrolls that hold it.
func withAccordion(o fyne.CanvasObject) bool {
	var children []fyne.CanvasObject
	switch o := o.(type) {
	case *widget.Accordion:
		return true
	case *fyne.Container:
		children = o.Objects
	case *container.Scroll:
		children = []fyne.CanvasObject{o.Content}
	}
	return slices.ContainsFunc(children, withAccordion)
}

var _ = Describe("Screen", func() {
	var (
		ctx context.Context
		svc *mocktraining.MockTrainingService
		s   *Screen
	)

	BeforeEach(func() {
		test.NewTempApp(GinkgoTB())
		ctx = context.Background()
		svc = mocktraining.NewMockTrainingService(GinkgoT())
		s = NewScreen(svc)
	})

	It("lists the ten drills in the order of the page, the swim once, by names that are not buttons, for nothing is shown by tapping", func() {
		Expect(s.rows).To(HaveLen(10))
		Expect(s.rows[0].name.Text).To(Equal("Domino"))
		Expect(s.rows[9].name.Text).To(Equal("Swim"))
		Expect(s.rows[0].name).To(BeAssignableToTypeOf(&widget.Label{}))
	})

	It("shows what each drill gives as what each outcome of it gives, not as a range, a light drill's stat alone and a heavy drill's two", func() {
		s.apply(trainingOf(0), nil)
		Expect(s.rows[0].gain.String()).To(Equal("POW (k0/b)"))
		Expect(s.rows[6].gain.String()).To(Equal("POW (k1/b) / LIF (k3/b)"))
	})

	It("names each stat of a drill in the colour of the stat, and says the rest in the colour of the text", func() {
		s.apply(trainingOf(0), nil)

		colours := map[string]color.Color{}
		for _, seg := range s.rows[6].gain.Segments() {
			colours[seg.Text] = seg.Color
		}
		Expect(colours).To(HaveKeyWithValue("POW", profileui.StatColors[monster.Pow]))
		Expect(colours).To(HaveKeyWithValue("LIF", profileui.StatColors[monster.Life]))
		Expect(colours[" (k1/b)"]).To(Equal(theme.Color(theme.ColorNameForeground)))
		Expect(colours[" / "]).To(Equal(theme.Color(theme.ColorNameForeground)))
	})

	It("shows N/A where the table has no gain", func() {
		t := trainingOf(0)
		t.Drills[0].Primary = "N/A"
		s.apply(t, nil)
		Expect(s.rows[0].gain.String()).To(Equal("N/A"))
	})

	It("gives the column of the drills the room that its longest name needs, and no more", func() {
		var longest float32
		for _, r := range s.rows {
			longest = max(longest, r.name.MinSize().Width)
		}
		Expect(columns[0]).To(BeNumerically(">=", longest))
		Expect(columns[0]).To(BeNumerically("<", longest+10))
	})

	It("shows how keen the monster is for each drill, as a bar and a percent", func() {
		s.apply(trainingOf(0), nil)
		Expect(s.rows[0].percent.Text).To(Equal("85%"))
		Expect(s.rows[0].bar.Ratio()).To(BeNumerically("~", 0.85, 1e-9))
		Expect(s.rows[9].percent.Text).To(Equal("100%"))
	})

	It("stars the drills the monster specialises in, and highlights the ones it is also keen on", func() {
		s.apply(trainingOf(1<<0|1<<1|1<<9), nil)
		Expect(s.rows[0].star.Text).To(Equal("★"))
		Expect(s.rows[1].star.Text).To(Equal("★"))
		Expect(s.rows[2].star.Text).To(BeEmpty())
		Expect(s.rows[9].star.Text).To(Equal("★"))

		Expect(s.rows[0].bg.FillColor).To(Equal(recommendedFill))   // 85% is keen
		Expect(s.rows[1].bg.FillColor).To(Equal(color.Transparent)) // 20% is not
		Expect(s.rows[2].bg.FillColor).To(Equal(color.Transparent)) // keen or not, no specialty
		Expect(s.rows[9].bg.FillColor).To(Equal(recommendedFill))
	})

	It("clears a highlight and a star when the page loads again without them", func() {
		s.apply(trainingOf(1<<0), nil)
		Expect(s.rows[0].bg.FillColor).To(Equal(recommendedFill))
		s.apply(trainingOf(0), nil)
		Expect(s.rows[0].bg.FillColor).To(Equal(color.Transparent))
		Expect(s.rows[0].star.Text).To(BeEmpty())
	})

	It("names the places the monster is good at, the pool among them, or says there are none", func() {
		s.apply(trainingOf(0), nil)
		Expect(s.places.Text).To(Equal("Good places: none"))

		s.apply(trainingOf(1<<0|1<<9|1<<10), nil)
		Expect(s.places.Text).To(Equal("Good places: Pool, Torble Sea"))
	})

	It("shows its table at once, with no panel to open, because the column that holds it already says what it is", func() {
		Expect(withAccordion(s.Content())).To(BeFalse())
	})

	It("shows why there is nothing to show, in place of the page", func() {
		s.apply(domain.Training{}, apperr.NotFound("game", "DuckStation is not running"))
		Expect(s.status.Text).To(Equal("DuckStation is not running"))
		Expect(s.body.Visible()).To(BeFalse())

		s.apply(trainingOf(0), nil)
		Expect(s.status.Visible()).To(BeFalse())
		Expect(s.body.Visible()).To(BeTrue())
	})

	It("shows any other error as an error", func() {
		s.apply(domain.Training{}, errors.New("db gone"))
		Expect(s.status.Text).To(Equal("Error: db gone"))
	})

	It("asks the service for the current page and shows it", func() {
		svc.EXPECT().Current(ctx).Return(trainingOf(1), nil)
		s.Load(ctx)
		Eventually(func() string { return s.rows[0].star.Text }).Should(Equal("★"))
	})

	// A screenshot of the page, for looking at: SCREENSHOT_DIR=/some/dir go test -ginkgo.label-filter=screenshot
	It("renders the page to a PNG", Label("screenshot"), func() {
		dir := os.Getenv("SCREENSHOT_DIR")
		if dir == "" {
			Skip("set SCREENSHOT_DIR to write the screenshots")
		}
		s.apply(trainingOf(1<<0|1<<2|1<<10), nil)
		Expect(os.MkdirAll(dir, 0o755)).To(Succeed())
		f, err := os.Create(filepath.Join(dir, "training-table.png"))
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(f.Close)
		Expect(png.Encode(f, software.Render(container.NewVBox(s.status, s.body), theme.DefaultTheme()))).To(Succeed())
	})
})
