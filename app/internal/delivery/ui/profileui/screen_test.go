package profileui

import (
	"context"
	"errors"
	"image/png"
	"os"
	"path/filepath"

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
	mockprofile "app/mock/core/profile"
)

func pixieProfile() domain.Profile {
	m := monster.Monster{
		Genus: 12, SubGenus: 3, Name: "Aa 1",
		Stats:  [monster.StatCount]int{111, 222, 666, 444, 555, 333},
		Nature: 130, NatureBase: 30, Age: 100, Lifespan: 250, InitialLifespan: 300,
		Fatigue: 40, Stress: 60, Fame: 50, Form: -10, Spoil: 70, Fear: 80,
		Growth:   [monster.StatCount]int{3, 1, 2, 4, 0, 2},
		PlayType: 2, LifeStage: 5, LifeType: 3, ArenaSpeed: 4, GutsRate: 5, PrizeMoney: 12345, Specials: 1 | 8,
		Motivation: [monster.DrillCount]int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
		Drug:       38, DrugWeeks: 3, GoldPeach: true,
	}
	return domain.Profile{
		Monster: m, Player: player.Player{Money: 1500, NextSale: 3, ErrantryCD: 255},
		Breed: domain.Breed{Name: "Pixie", GutsOK: true}, MainGenus: "Pixie", SubGenus: "Dragon",
		Like: "Cup Jelly", Dislike: "Tablet", Drug: "Troron",
	}
}

var _ = Describe("Dashboard", func() {
	var (
		ctx context.Context
		svc *mockprofile.MockProfileService
		d   *Dashboard
	)

	BeforeEach(func() {
		test.NewTempApp(GinkgoTB())
		ctx = context.Background()
		svc = mockprofile.NewMockProfileService(GinkgoT())
		d = NewDashboard(svc)
	})

	It("shows who the monster is, how old, how much life it has left, and the money, in the header", func() {
		d.apply(pixieProfile(), nil)
		Expect(d.breed.txt.Text).To(Equal("Pixie"))
		Expect(d.givenName.Text).To(Equal("Aa 1"))
		Expect(d.genus.Text).To(Equal("Pixie / Dragon"))
		Expect(d.age.Text).To(Equal(domain.AgeText(100) + " (100w)"))
		Expect(d.lifespan.Text).To(Equal("250w / 300w"))
		Expect(d.lifeIndex.txt.Text).To(Equal("160 (-3 Weeks)"))
		Expect(d.lifeIndex.bg.FillColor).To(Equal(toneColors[ToneOrange]))
		Expect(d.money.Text).To(Equal("1500G"))
	})

	It("shows the stats with their growth, the nature and the rest of the vitals", func() {
		d.apply(pixieProfile(), nil)
		Expect(d.stats[monster.Speed].Text).To(Equal("555 [568]"))
		Expect(d.growth[monster.Skill].txt.Text).To(Equal("A / 5"))
		Expect(d.growth[monster.Skill].bg.FillColor).To(Equal(toneColors[ToneLimeGreen]))
		Expect(d.statTotal.Text).To(Equal("2331"))
		Expect(d.nature.txt.Text).To(Equal("130 (Best)"))
		Expect(d.loyalty.Text).To(Equal("75 (spoil 70 / fear 80)"))
		Expect(d.drug.Text).To(Equal("Troron (3w)"))
		Expect(d.drugTip.Text).To(HavePrefix("Troron:"))
		Expect(d.goldPeach.Checked).To(BeTrue())
		Expect(d.errantry.Checked).To(BeTrue())
	})

	It("draws the stats and the conditions as bars, in the colour of how they stand", func() {
		d.apply(pixieProfile(), nil)
		Expect(d.statBars[monster.Speed].ratio).To(BeNumerically("~", 555.0/999, 1e-9))
		Expect(d.fatigueBar.ratio).To(BeNumerically("~", 0.4, 1e-9))
		Expect(d.fatigueBar.fill.FillColor).To(Equal(toneColors[ToneOrange]))
		Expect(d.lifespanBar.ratio).To(BeNumerically("~", 250.0/300, 1e-9))
		Expect(d.lifespanBar.fill.FillColor).To(Equal(toneColors[ToneLimeGreen]))
		Expect(d.natureGauge.ratio).To(BeNumerically("~", 1, 1e-9))
	})

	It("paints a box with its tone, and clears the paint when the tone goes", func() {
		p := pixieProfile()
		p.Breed.Name = "Pixie (N/S)"
		d.apply(p, nil)
		Expect(d.breed.bg.FillColor).To(Equal(toneColors[ToneLightPink]))

		d.apply(pixieProfile(), nil)
		Expect(d.breed.bg.FillColor).To(Equal(toneNone))
	})

	It("shows why there is nothing to show, in place of the dashboard, and the dashboard again when there is", func() {
		d.apply(pixieProfile(), nil)
		Expect(d.headerBody.Visible()).To(BeTrue())
		Expect(d.vitalsBody.Visible()).To(BeTrue())

		d.apply(domain.Profile{}, apperr.NotFound("game", "DuckStation is not running"))
		Expect(d.status.Text).To(Equal("DuckStation is not running"))
		Expect(d.status.Visible()).To(BeTrue())
		Expect(d.headerBody.Visible()).To(BeFalse())
		Expect(d.vitalsBody.Visible()).To(BeFalse())

		d.apply(pixieProfile(), nil)
		Expect(d.status.Visible()).To(BeFalse())
		Expect(d.headerBody.Visible()).To(BeTrue())
		Expect(d.vitalsBody.Visible()).To(BeTrue())
	})

	It("shows any other error as an error", func() {
		d.apply(domain.Profile{}, errors.New("db gone"))
		Expect(d.status.Text).To(Equal("Error: db gone"))
	})

	It("shows the cocoon facts only for a Worm that can still cocoon", func() {
		d.apply(pixieProfile(), nil)
		Expect(d.cocoon.Visible()).To(BeFalse())

		p := pixieProfile()
		p.Monster.Genus, p.Monster.Rank, p.Monster.CupJellies = 36, 2, 31
		d.apply(p, nil)
		Expect(d.cocoon.Visible()).To(BeTrue())
		Expect(d.cupJellies.txt.Text).To(Equal("31"))
		Expect(d.cupJellies.bg.FillColor).To(Equal(toneColors[ToneLightGreen]))
	})

	It("shows the guts rate as guts a second, and the growth rates from zero, when asked, without waiting for the game", func() {
		d.apply(pixieProfile(), nil)
		Expect(d.guts.txt.Text).To(Equal("5"))
		Expect(d.growth[monster.Life].txt.Text).To(Equal("B / 4"))

		test.Tap(d.gutsPerSecond)
		test.Tap(d.growthFromZero)
		Expect(d.guts.txt.Text).To(Equal("6.00"))
		Expect(d.growth[monster.Life].txt.Text).To(Equal("B / 3"))
	})

	It("asks the service for the current profile and shows it", func() {
		svc.EXPECT().Current(ctx).Return(pixieProfile(), nil)
		d.Load(ctx)
		Eventually(func() string { return d.breed.txt.Text }).Should(Equal("Pixie"))
	})

	// A screenshot of the real thing, for looking at: SCREENSHOT_DIR=/some/dir ginkgo run --label-filter=screenshot
	It("renders the dashboard to a PNG", Label("screenshot"), func() {
		dir := os.Getenv("SCREENSHOT_DIR")
		if dir == "" {
			Skip("set SCREENSHOT_DIR to write the screenshots")
		}
		d.apply(pixieProfile(), nil)
		Expect(os.MkdirAll(dir, 0o755)).To(Succeed())
		f, err := os.Create(filepath.Join(dir, "dashboard.png"))
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(f.Close)
		// The parts are scrolls, whose smallest size is a few pixels, so draw what they hold at its own size.
		Expect(png.Encode(f, software.Render(container.NewVBox(d.status, d.headerBody, d.vitalsBody), theme.DefaultTheme()))).To(Succeed())
	})
})
