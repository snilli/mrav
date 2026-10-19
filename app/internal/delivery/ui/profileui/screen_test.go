package profileui

import (
	"context"
	"errors"
	"image/png"
	"os"
	"path/filepath"
	"strconv"

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
		Expect(d.goldPeach.On()).To(BeTrue())
		Expect(d.silverPeach.On()).To(BeFalse())
		Expect(d.errantry.On()).To(BeTrue())
	})

	It("says how tired the monster is in the game's words, and sums up its fights and its ailments", func() {
		p := pixieProfile()
		p.Monster.Record = monster.Record{Wins: 12, Losses: 3, KOs: 5, Injuries: 4, Runaways: 6, Sicknesses: 7, Resistance: 8}
		d.apply(p, nil)
		Expect(d.fatigueWord.Text).To(Equal("seems tired"))
		Expect(d.record.Text).To(Equal("12 W · 3 L · 5 KO"))
		Expect(d.history.Text).To(Equal("sick 7 · injured 4 · ran away 6 · resistance 8"))
	})

	Describe("the ranch", func() {
		// ranchProfile is the Pixie with a ranch of a game that keeps one, its gold and grade one short of a step.
		ranchProfile := func() domain.Profile {
			p := pixieProfile()
			p.Player.Ranch = player.Ranch{Known: true, Date: 191, Grade: 6, House: 2, Barn: 1, Altar: 3}
			return p
		}

		It("shows the date in the header with the game's count of weeks, and the panel of upgrades", func() {
			d.apply(ranchProfile(), nil)
			Expect(d.date.Visible()).To(BeTrue())
			Expect(d.date.Text).To(Equal("Y1004 Mar wk 4 (week 191)"))
			Expect(d.ranchBody.Visible()).To(BeTrue())
			Expect(d.ranchItem.Title).To(Equal("Ranch · House & stable: May wk 4, in 8 weeks"))
		})

		It("shows the tournament's countdown among the vitals, and hides it where the year of the first was not read", func() {
			d.apply(ranchProfile(), nil)
			Expect(d.tournamentRow.Visible()).To(BeFalse())
			p := ranchProfile()
			p.Player.Ranch.Date, p.Player.Ranch.Tournament = 860, 1003
			d.apply(p, nil)
			Expect(d.tournamentRow.Visible()).To(BeTrue())
			Expect(d.tournament.Text).To(Equal("Y1019 Jul wk 1, in 64 weeks"))
		})

		It("shows each facility: how far it is, what its next step costs, and where that step stands", func() {
			d.apply(ranchProfile(), nil)
			house, stable, altar := d.facilities[0], d.facilities[1], d.facilities[2]
			Expect(house.name.Text).To(Equal("House"))
			Expect(house.level.Text).To(Equal("2/4"))
			Expect(house.bar.Ratio()).To(BeNumerically("~", 0.5, 1e-9))
			Expect(house.next.Text).To(Equal("16 places · next 12,000 G · grade 6"))
			Expect(house.state.txt.Text).To(Equal("15,500 G more"))
			Expect(house.state.bg.FillColor).To(Equal(toneColors[ToneOrange]))
			Expect(stable.name.Text).To(Equal("Stable"))
			Expect(stable.next.Text).To(Equal("next 30,000 G · grade 7"))
			Expect(altar.next.Text).To(Equal("fully upgraded"))
			Expect(altar.state.txt.Text).To(Equal("MAX"))
		})

		It("lights up the main breeds that are unlocked, in a list that opens from its title with the count", func() {
			p := ranchProfile()
			for i := range player.BreedCount {
				p.Breeds = append(p.Breeds, domain.BreedUnlock{Name: "B" + strconv.Itoa(i), Unlocked: i%2 == 0})
			}
			d.apply(p, nil)
			Expect(d.breeds.Visible()).To(BeTrue())
			Expect(d.breedsItem.Title).To(Equal("Breeds unlocked 19/38"))
			Expect(d.breedFlags[0].txt.Text).To(Equal("B0"))
			Expect(d.breedFlags[0].On()).To(BeTrue())
			Expect(d.breedFlags[1].On()).To(BeFalse())
			Expect(d.breedFlags[37].txt.Text).To(Equal("B37"))
		})

		It("has no list of breeds when the game does not say which are unlocked", func() {
			d.apply(ranchProfile(), nil)
			Expect(d.breeds.Visible()).To(BeFalse())
		})

		It("keeps the four panels closed until they are opened: the ranch, the breeds, and the expeditions", func() {
			d.apply(ranchProfile(), nil)
			for _, panel := range []*widget.Accordion{d.ranchPanel, d.breeds, d.expeditions} {
				for _, item := range panel.Items {
					Expect(item.Open).To(BeFalse(), item.Title)
				}
			}
		})

		It("is not shown for a game that keeps no ranch, and not when there is nothing to show", func() {
			d.apply(pixieProfile(), nil)
			Expect(d.date.Visible()).To(BeFalse())
			Expect(d.ranchBody.Visible()).To(BeFalse())

			d.apply(ranchProfile(), nil)
			Expect(d.ranchBody.Visible()).To(BeTrue())
			d.apply(domain.Profile{}, apperr.NotFound("game", "DuckStation is not running"))
			Expect(d.ranchBody.Visible()).To(BeFalse())
		})
	})

	Describe("the expeditions", func() {
		dated := func() domain.Profile {
			p := pixieProfile()
			p.Player.Ranch = player.Ranch{Known: true, Date: 191, Energy: -5, EnergyKnown: true}
			return p
		}

		It("puts the nearest trip on the title of a list that opens to every trip", func() {
			d.apply(dated(), nil)
			Expect(d.expeditionsBody.Visible()).To(BeTrue())
			Expect(d.expeditionsItem.Title).To(Equal("Next expedition: Phoenix Volcano Adventure, Jun wk 1, in 9 weeks · not fit yet"))
			Expect(d.trips[0].name.Text).To(Equal("Phoenix Volcano Adventure"))
			Expect(d.trips[0].when.Text).To(Equal("Jun wk 1, in 9 weeks"))
			Expect(d.trips[0].note.Text).To(HavePrefix("the first expedition"))
			Expect(d.trips[0].note.Visible()).To(BeTrue())
			Expect(d.trips[1].note.Visible()).To(BeFalse())
		})

		It("marks each condition as met in green and unmet in orange, as the monster stands", func() {
			d.apply(dated(), nil) // the Pixie: life 111, fame 50, grade E; the first trip asks for 140, 40 and D
			chips := d.trips[0].chips
			Expect(chips[0].txt.Text).To(Equal("✗ LIF 140 or more (111)"))
			Expect(chips[0].bg.FillColor).To(Equal(toneColors[ToneOrange]))
			Expect(chips[1].txt.Text).To(Equal("✓ Fame 40 or more (50)"))
			Expect(chips[1].bg.FillColor).To(Equal(toneColors[ToneLimeGreen]))
			Expect(chips[2].txt.Text).To(Equal("✗ Grade D or higher (E)"))
		})

		It("says the energy the monster would start with, and the energy in the game only when it was read", func() {
			d.apply(dated(), nil)
			Expect(d.startEnergy.Text).To(HavePrefix("Starts with 140 energy"))
			Expect(d.gameEnergy.Visible()).To(BeTrue())
			Expect(d.gameEnergy.Text).To(Equal("Energy in the game now: -5 (it stays after an expedition)"))

			d.apply(pixieProfile(), nil)
			Expect(d.gameEnergy.Visible()).To(BeFalse())
		})

		It("is there for a game that keeps no date, in the order of the calendar, since the invitations are the same", func() {
			d.apply(pixieProfile(), nil)
			Expect(d.expeditionsBody.Visible()).To(BeTrue())
			Expect(d.trips[0].name.Text).To(Equal("Torles Mountains (Rovest)"))
			Expect(d.trips[0].when.Text).To(Equal("Feb wk 1"))
		})

		It("is not shown when there is nothing to show", func() {
			d.apply(pixieProfile(), nil)
			d.apply(domain.Profile{}, apperr.NotFound("game", "DuckStation is not running"))
			Expect(d.expeditionsBody.Visible()).To(BeFalse())
		})
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

	It("renders the ranch panel, with its list of breeds open, to a PNG", Label("screenshot"), func() {
		dir := os.Getenv("SCREENSHOT_DIR")
		if dir == "" {
			Skip("set SCREENSHOT_DIR to write the screenshots")
		}
		p := pixieProfile()
		p.Player.Money = 20000
		p.Player.Ranch = player.Ranch{Known: true, Date: 191, Grade: 6, House: 2, Barn: 1, Altar: 1}
		names := []string{"Pixie", "Dragon", "Centaur", "ColorPandora", "Beaclon", "Henger", "Wracky", "Golem", "Zuum", "Durahan"}
		for i := range player.BreedCount {
			p.Breeds = append(p.Breeds, domain.BreedUnlock{Name: names[i%len(names)], Unlocked: i%3 != 0})
		}
		p.Player.Ranch.Energy, p.Player.Ranch.EnergyKnown = 123, true
		d.apply(p, nil)
		d.ranchPanel.OpenAll()
		d.breeds.OpenAll()
		d.expeditions.OpenAll()
		Expect(os.MkdirAll(dir, 0o755)).To(Succeed())
		f, err := os.Create(filepath.Join(dir, "ranch-panel.png"))
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(f.Close)
		Expect(png.Encode(f, software.Render(d.ranchBox, theme.DefaultTheme()))).To(Succeed())
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
