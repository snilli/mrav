package plannerui

import (
	"context"
	"errors"
	"image/color"
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

	monster "app/internal/core/monster/domain"
	"app/internal/core/planner/domain"
	"app/internal/delivery/ui/profileui"
	"app/internal/testutil"
	mockplanner "app/mock/core/planner"
)

func shippedEffects() domain.Effects {
	effects, err := testutil.ShippedEffects(context.Background())
	if errors.Is(err, testutil.ErrNoShippedDB) {
		Skip(err.Error())
	}
	Expect(err).NotTo(HaveOccurred())
	return effects
}

func currentOf(change func(*domain.Settings)) domain.Current {
	settings := domain.DefaultSettings()
	if change != nil {
		change(&settings)
	}
	return domain.Current{
		Settings: settings,
		Reactions: domain.NewReactions([]domain.FoodReaction{
			{Genus: 35, Food: domain.Tablet, Stage: domain.AnyStage, Reaction: domain.Like},
		}),
		Effects: shippedEffects(),
	}
}

var _ = Describe("Screen", func() {
	var (
		ctx context.Context
		svc *mockplanner.MockPlannerService
		s   *Screen
	)

	BeforeEach(func() {
		test.NewTempApp(GinkgoTB())
		ctx = context.Background()
		svc = mockplanner.NewMockPlannerService(GinkgoT())
		s = NewScreen(svc)
		s.apply(currentOf(nil), nil)
	})

	weeks := func() []*slot {
		var out []*slot
		for _, sl := range s.slots {
			if sl.object.Visible() && !sl.header {
				out = append(out, sl)
			}
		}
		return out
	}
	figures := func(sl *slot) []string { return []string{sl.ftg.Text, sl.str.Text, sl.li.Text, sl.lost.Text} }
	choose := func(i int, item, action string) {
		sl := weeks()[i]
		sl.item.SetSelected(item)
		sl.action.SetSelected(action)
	}
	start := func(change func(*domain.Settings)) { s.apply(currentOf(change), nil) }
	light := func(n int) {
		for range n {
			choose(len(weeks())-1, "No item", "Light Drill")
		}
	}

	Describe("its start", func() {
		It("starts the plan from the monster of the game, and has no form to fill in for it", func() {
			Expect(s.settings).To(Equal(domain.DefaultSettings()))
			var entries []fyne.CanvasObject
			for _, o := range test.LaidOutObjects(s.Content()) {
				if _, ok := o.(*widget.Entry); ok {
					entries = append(entries, o)
				}
			}
			Expect(entries).To(BeEmpty())
		})

		It("works the plan out again when the monster changes in the game", func() {
			choose(0, "No item", "Rest")
			start(func(st *domain.Settings) { st.Fatigue, st.Stress = 100, 60 })
			Expect(figures(weeks()[0])[:2]).To(Equal([]string{"64", "55"}))

			start(func(st *domain.Settings) { st.Fatigue, st.Stress, st.LifeStage = 100, 60, 5 })
			Expect(figures(weeks()[0])[:2]).To(Equal([]string{"56", "51"}))
		})

		It("keeps the weeks and the foods of the plan when the monster changes", func() {
			light(4)
			s.foods[2] = domain.Meat

			start(func(st *domain.Settings) { st.Fatigue = 30 })

			Expect(s.steps).To(HaveLen(4))
			Expect(s.foods).To(Equal(map[int]domain.Food{2: domain.Meat}))
			Expect(s.settings.Fatigue).To(Equal(30))
		})
	})

	Describe("its weeks", func() {
		It("shows one empty week to fill, and nothing lived, to begin with", func() {
			Expect(weeks()).To(HaveLen(1))
			first := weeks()[0]
			Expect(first.week.Text).To(Equal("1wk"))
			Expect(first.item.Selected).To(Equal("No item"))
			Expect(first.action.Selected).To(Equal("No action"))
			Expect(figures(first)).To(Equal([]string{"", "", "", ""}))
			Expect(s.lived.Text).To(Equal("Weeks lived: 0"))
		})

		It("works a week out when an action is chosen for it, and gives an empty week after it", func() {
			choose(0, "No item", "Light Drill")

			Expect(weeks()).To(HaveLen(2))
			Expect(figures(weeks()[0])).To(Equal([]string{"10", "5", "20", "0"}))
			Expect(weeks()[1].week.Text).To(Equal("2wk"))
			Expect(weeks()[1].action.Selected).To(Equal("No action"))
			Expect(s.lived.Text).To(Equal("Weeks lived: 1"))
		})

		It("says under the table what the item of the week that was changed last does, and nothing for no item", func() {
			Expect(s.effect.Visible()).To(BeFalse())
			choose(0, "Candy", "Rest")
			Expect(s.effect.Visible()).To(BeTrue())
			Expect(s.effect.Text).To(Equal("Month 1, week 1, Candy: Stress -2, Form +10"))
			choose(1, "Nuts Oil", "Rest")
			Expect(s.effect.Text).To(Equal("Month 1, week 2, Nuts Oil: Fatigue -28"))
			choose(1, "No item", "Rest")
			Expect(s.effect.Visible()).To(BeFalse())
		})

		It("says nothing under the table for an item that the plan leaves out, as a frozen week takes none", func() {
			choose(0, "Mango", "Freeze")
			Expect(s.effect.Visible()).To(BeFalse())
		})

		It("takes the item of a week with its action", func() {
			choose(0, "Mango", "Hard Drill")
			Expect(figures(weeks()[0])).To(Equal([]string{"15", "12", "39", "0"}))
			Expect(s.steps).To(Equal([]domain.Step{{Item: domain.Mango, Action: domain.HardDrill}}))
		})

		It("heads each month, and puts the food of a new month before its first week", func() {
			light(4)

			var sequence []string
			for _, sl := range s.slots {
				switch {
				case !sl.object.Visible():
				case sl.header:
					sequence = append(sequence, sl.head.Text)
				default:
					sequence = append(sequence, sl.week.Text)
				}
			}
			Expect(sequence).To(Equal([]string{"Month 1", "1wk", "2wk", "3wk", "4wk", "Month 2", "food", "1wk"}))
		})

		It("lets the food of a month be chosen, potato at first, and works the choice in", func() {
			light(4)
			food := weeks()[4]

			Expect(food.food.Visible()).To(BeTrue())
			Expect(food.food.Selected).To(Equal("Potato"))
			Expect(figures(food)).To(Equal([]string{"40", "36", "", ""}))

			food.food.SetSelected("Tablet")
			Expect(figures(weeks()[4])).To(Equal([]string{"40", "5", "", ""}))
			Expect(s.foods).To(Equal(map[int]domain.Food{2: domain.Tablet}))
		})

		It("shows the weeks that follow the first week of an errantry as what they are, which cannot be edited", func() {
			choose(0, "No item", "Errantry")

			rows := weeks()
			Expect(rows).To(HaveLen(6))
			Expect(rows[0].action.Visible()).To(BeTrue())
			for _, i := range []int{1, 2, 3} {
				Expect(rows[i].item.Visible()).To(BeFalse(), "week %d", i)
				Expect(rows[i].action.Visible()).To(BeFalse(), "week %d", i)
				Expect(rows[i].away.Visible()).To(BeTrue(), "week %d", i)
				Expect(rows[i].away.Text).To(Equal("Errantry"), "week %d", i)
			}
		})

		It("offers an expedition to a first week of a month only", func() {
			Expect(weeks()[0].action.Options).To(ContainElement("Expedition"))
			choose(0, "No item", "Light Drill")
			Expect(weeks()[1].action.Options).NotTo(ContainElement("Expedition"))
			Expect(weeks()[1].action.Options).To(ContainElement("Light Drill"))
			Expect(weeks()[1].action.Options[0]).To(Equal("No action"))
		})

		It("takes a week away when both of its choices are made empty, and the weeks after it move up", func() {
			light(1)
			choose(1, "No item", "Rest")
			Expect(s.steps).To(HaveLen(2))

			choose(0, "No item", "No action")

			Expect(s.steps).To(Equal([]domain.Step{{Action: domain.Rest}}))
			Expect(weeks()).To(HaveLen(2))
		})

		It("plays the plan again when an errantry is changed, so that its other weeks go and the steps after it move up", func() {
			choose(0, "No item", "Errantry")
			choose(5, "No item", "Light Drill")
			Expect(s.steps).To(HaveLen(2))
			Expect(weeks()).To(HaveLen(7))

			weeks()[0].action.SetSelected("Rest")

			Expect(s.steps).To(Equal([]domain.Step{{Action: domain.Rest}, {Action: domain.LightDrill}}))
			Expect(weeks()).To(HaveLen(3))
		})

		It("leaves out of what a week may take a peach that another week has, and keeps the one that it has itself", func() {
			choose(0, "Gold Peach", "Rest")

			Expect(weeks()[0].item.Options).To(ContainElement("Gold Peach"))
			Expect(weeks()[1].item.Options).NotTo(ContainElement("Gold Peach"))
			Expect(weeks()[1].item.Options).To(ContainElement("Silver Peach"))
		})

		It("says what the plan left out, in words, and nothing when it left out nothing", func() {
			Expect(s.problems.Text).To(BeEmpty())
			Expect(s.problems.Visible()).To(BeFalse())

			s.steps = []domain.Step{{Action: domain.LightDrill}, {Action: domain.Expedition}}
			s.recompute()

			Expect(s.problems.Visible()).To(BeTrue())
			Expect(s.problems.Text).To(ContainSubstring("Month 1, week 2"))
			Expect(s.problems.Text).To(ContainSubstring("expedition"))
		})
	})

	Describe("its sums", func() {
		It("gives the weeks lived, the weeks lost, and the average a month", func() {
			start(func(st *domain.Settings) { st.Fatigue, st.Stress = 100, 100 })
			choose(0, "No item", "Rest")

			Expect(figures(weeks()[0])).To(Equal([]string{"64", "95", "254", "-6"}))
			Expect(s.lived.Text).To(Equal("Weeks lived: 1"))
			Expect(s.weeks.Text).To(Equal("Weeks lost: 6"))
			Expect(s.average.Text).To(Equal("Average: -24.0 wk/month"))
		})

		It("gives the weeks gained when a peach is given", func() {
			choose(0, "Gold Peach", "Rest")

			Expect(weeks()[0].lost.Text).To(Equal("+50"))
			Expect(s.weeks.Text).To(Equal("Weeks gained: 50"))
			Expect(s.average.Text).To(Equal("Average: +200.0 wk/month"))
		})

		It("adds up the weeks lost, week by week, and gives the life that is left from the lifespan of the game", func() {
			start(func(st *domain.Settings) { st.Fatigue, st.Stress, st.Lifespan = 100, 100, 80 })
			choose(0, "No item", "Rest")
			choose(1, "Gold Peach", "Rest")

			Expect([]string{weeks()[0].total.Text, weeks()[0].life.Text}).To(Equal([]string{"-6", "73"}))
			Expect([]string{weeks()[1].lost.Text, weeks()[1].total.Text, weeks()[1].life.Text}).To(Equal([]string{"+46", "+40", "118"}))
			Expect([]string{weeks()[2].total.Text, weeks()[2].life.Text}).To(Equal([]string{"", ""}), "the week to fill")
			Expect(weeks()[0].total.Color).To(Equal(theme.Color(theme.ColorNameError)))
			Expect(weeks()[1].total.Color).To(Equal(gainColor))
		})

		It("shows the form of each week, from the form of the game", func() {
			start(func(st *domain.Settings) { st.Form = 20 })
			choose(0, "Mango", "Light Drill")
			choose(1, "No item", "Light Drill")
			light(2)
			weeks()[4].food.SetSelected("Meat")

			var shown []string
			for _, sl := range weeks() {
				shown = append(shown, sl.form.Text)
			}
			Expect(shown).To(Equal([]string{"20", "19", "18", "17", "23", ""}))
		})

		It("writes the life that is left in the colour of an error once it runs out", func() {
			start(func(st *domain.Settings) { st.Lifespan = 1 })
			light(2)

			Expect(weeks()[0].life.Color).To(Equal(theme.Color(theme.ColorNameForeground)))
			Expect(weeks()[1].life.Text).To(Equal("-1"))
			Expect(weeks()[1].life.Color).To(Equal(theme.Color(theme.ColorNameError)))
		})

		It("has nothing lost, and an average of nothing, for a plan that has no week", func() {
			Expect(s.weeks.Text).To(Equal("Weeks lost: 0"))
			Expect(s.average.Text).To(Equal("Average: 0.0 wk/month"))
		})

		It("writes the weeks lost in the colour of an error, the weeks gained in a soft green, and none in the colour of the text", func() {
			start(func(st *domain.Settings) { st.Fatigue, st.Stress = 100, 100 })
			choose(0, "No item", "Rest")
			choose(1, "Gold Peach", "Rest")
			choose(2, "No item", "Light Drill")

			Expect(weeks()[0].lost.Color).To(Equal(theme.Color(theme.ColorNameError)))
			Expect(weeks()[1].lost.Color).To(Equal(gainColor))
			start(nil)
			Expect(weeks()[0].lost.Text).To(Equal("0"))
			Expect(weeks()[0].lost.Color).To(Equal(theme.Color(theme.ColorNameForeground)))
		})
	})

	Describe("its table", func() {
		It("draws its lines and a frame the way the table of the drills does", func() {
			Expect(s.frame.StrokeWidth).To(Equal(profileui.LineWidth))
			Expect(s.frame.StrokeColor).To(Equal(profileui.LineColor))
			Expect(s.frame.FillColor).To(Equal(color.Transparent))
		})

		It("is no wider than a column of the window, whatever is in it, so that it never makes the window grow", func() {
			s.steps = []domain.Step{
				{Item: domain.SmokedSnake, Action: domain.BattleLose}, {Item: domain.SilverPeach, Action: domain.Freeze}, {Action: domain.Errantry},
				{Item: domain.MagicBananaMinus, Action: domain.BattleOK}, {Action: domain.Expedition}, {Item: domain.Paradoxine, Action: domain.HardDrill},
			}
			s.recompute()
			Expect(s.Content().MinSize().Width).To(BeNumerically("<=", 440))
		})

		It("clears the plan with its foods and keeps the start, with a button", func() {
			start(func(st *domain.Settings) { st.Main = 7 })
			choose(0, "No item", "Light Drill")
			s.foods[2] = domain.Meat

			test.Tap(s.clear)

			Expect(s.steps).To(BeEmpty())
			Expect(s.foods).To(BeEmpty())
			Expect(weeks()).To(HaveLen(1))
			Expect(s.settings.Main).To(Equal(7))
		})
	})

	Describe("its monster", func() {
		It("asks the service for the monster of the game and starts from it", func() {
			other := NewScreen(svc)
			svc.EXPECT().Current(ctx).Return(currentOf(func(st *domain.Settings) { st.Main = 7 }), nil)

			other.Load(ctx)

			Eventually(func() int { return other.settings.Main }).Should(Equal(7))
		})

		It("keeps the plan when the same monster is loaded again", func() {
			choose(0, "No item", "Light Drill")
			svc.EXPECT().Current(ctx).Return(currentOf(nil), nil)

			s.Load(ctx)

			Consistently(func() int { return len(s.steps) }).Should(Equal(1))
		})

		It("shows why there is nothing to show, in place of the page", func() {
			s.apply(domain.Current{}, errors.New("no game"))
			Expect(s.status.Text).To(Equal("Error: no game"))
			Expect(s.status.Visible()).To(BeTrue())
			Expect(s.body.Visible()).To(BeFalse())

			s.apply(currentOf(nil), nil)
			Expect(s.status.Visible()).To(BeFalse())
			Expect(s.body.Visible()).To(BeTrue())
		})
	})

	It("renders the page to a PNG", Label("screenshot"), func() {
		dir := os.Getenv("SCREENSHOT_DIR")
		if dir == "" {
			Skip("set SCREENSHOT_DIR to write the screenshots")
		}
		start(func(st *domain.Settings) {
			*st = domain.Settings{Version: monster.MR2DX, Main: 7, Sub: 9, Like: domain.Tablet, Dislike: domain.Meat, LifeStage: 4, Form: -35, Pots: 2, Sculptures: 3, Fatigue: 15, Stress: 25, Lifespan: 180}
		})
		s.foods = map[int]domain.Food{2: domain.Tablet, 3: domain.Milk}
		s.steps = []domain.Step{
			{Item: domain.MintLeaf, Action: domain.LightDrill}, {Action: domain.HardDrill}, {Item: domain.Griever, Action: domain.Rest}, {Action: domain.BattleWin},
			{Item: domain.Mango, Action: domain.LightDrill}, {Action: domain.Errantry}, {Item: domain.NutsOil, Action: domain.Rest}, {Action: domain.BattleLose},
			{Action: domain.LightDrill}, {Action: domain.Freeze},
		}
		s.recompute()

		Expect(os.MkdirAll(dir, 0o755)).To(Succeed())
		f, err := os.Create(filepath.Join(dir, "planner.png"))
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(f.Close)
		fyne.CurrentApp().Settings().SetTheme(theme.DefaultTheme())
		c := software.NewCanvas()
		c.SetPadded(false)
		c.SetContent(container.NewVBox(s.status, s.body))
		c.Resize(fyne.NewSize(466, c.Content().MinSize().Height))
		Expect(png.Encode(f, c.Capture())).To(Succeed())
	})
})
