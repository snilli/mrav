package drillui

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/software"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/apperr"
	"app/internal/core/drill/domain"
	monster "app/internal/core/monster/domain"
	"app/internal/delivery/ui/profileui"
	mockdrill "app/mock/core/drill"
)

func cell(result domain.Result, role domain.Role, bonus bool, slots ...[]int) domain.Cell {
	return domain.Cell{Result: result, Role: role, Bonus: bonus, LifeStage: 4, GrowthRate: 2, Slots: slots}
}

func trainingOf(boost int) domain.Training {
	cells := []domain.Cell{
		cell(domain.Cheat, domain.Light, false, []int{1}, []int{2}),
		cell(domain.Success, domain.Light, false, []int{1}, []int{3}),
		cell(domain.Success, domain.Light, true, []int{2}, []int{4}),
		cell(domain.Great, domain.Light, false, []int{2}, []int{5}),
		cell(domain.Success, domain.Primary, false, []int{3}, []int{4}),
		cell(domain.Success, domain.Secondary, false, []int{2}),
		cell(domain.Success, domain.Drop, false, []int{-2}, []int{-3}),
		cell(domain.Great, domain.Primary, false, []int{4, 8}, []int{6, 10}),
	}
	m := monster.Monster{
		LifeStage: 4, TrainBoost: boost, Growth: [monster.StatCount]int{2, 2, 2, 2, 2, 2},
		Motivation: [monster.DrillCount]int{85, 20, 30, 40, 50, 60, 70, 80, 90, 100},
	}
	return domain.Build(m, domain.NewGains(cells))
}

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
		svc *mockdrill.MockDrillService
		s   *Screen
	)

	BeforeEach(func() {
		test.NewTempApp(GinkgoTB())
		ctx = context.Background()
		svc = mockdrill.NewMockDrillService(GinkgoT())
		s = NewScreen(svc)
	})

	shown := func(drill int, stat monster.Stat) string { return s.rows[drill].cells[stat].Text }
	marked := func(r *row) string {
		if r.bonus.Visible() {
			return r.name.Text + " +1"
		}
		return r.name.Text
	}

	It("lists the ten drills in the order of the page, the swim once, one row each", func() {
		Expect(s.rows).To(HaveLen(10))
		Expect(marked(s.rows[0])).To(Equal("Domino"))
		Expect(marked(s.rows[6])).To(Equal("Pull"))
		Expect(marked(s.rows[9])).To(Equal("Swim"))
	})

	It("heads a column for each stat, named in the colour of the stat, in the order of the vitals page", func() {
		for stat := range monster.StatCount {
			Expect(s.heads[stat].Text).To(Equal(profileui.StatNames[stat]))
			Expect(s.heads[stat].Color).To(Equal(profileui.StatColors[stat]))
		}
	})

	It("lets the three results of the page be chosen, and starts with Success", func() {
		Expect(s.result.Options).To(Equal([]string{"Cheat", "Success", "Great"}))
		Expect(s.result.Selected).To(Equal("Success"))
		Expect(s.result.Horizontal).To(BeTrue())
		Expect(s.result.Required).To(BeTrue(), "a result is always shown")
	})

	It("shows what the chosen result gives, a light drill under its one stat", func() {
		s.apply(trainingOf(0), nil)
		Expect(shown(0, monster.Pow)).To(Equal("1~3"))

		s.result.SetSelected("Cheat")
		Expect(shown(0, monster.Pow)).To(Equal("1~2"))

		s.result.SetSelected("Great")
		Expect(shown(0, monster.Pow)).To(Equal("2~5"))
	})

	It("shows a heavy drill under each of the stats it moves, the one it drops below zero in the colour of an error, and nothing under the rest", func() {
		s.apply(trainingOf(0), nil)

		Expect(shown(6, monster.Pow)).To(Equal("3~4"))
		Expect(shown(6, monster.Life)).To(Equal("2"))
		Expect(shown(6, monster.Speed)).To(Equal("-3~-2"))
		for _, stat := range []monster.Stat{monster.Int, monster.Skill, monster.Defense} {
			Expect(shown(6, stat)).To(BeEmpty())
		}
		Expect(s.rows[6].cells[monster.Life].Color).To(Equal(theme.Color(theme.ColorNameForeground)))
		Expect(s.rows[6].cells[monster.Speed].Color).To(Equal(theme.Color(theme.ColorNameError)))
	})

	It("shows the number under the main stat of a drill in a soft green, so that it is told from the second stat, and keeps the drop red", func() {
		s.apply(trainingOf(0), nil)

		Expect(s.rows[0].cells[monster.Pow].Color).To(Equal(mainColor), "a light drill has the one stat")
		Expect(s.rows[6].cells[monster.Pow].Color).To(Equal(mainColor), "Pull's main stat")
		Expect(s.rows[6].cells[monster.Life].Color).To(Equal(theme.Color(theme.ColorNameForeground)), "Pull's second stat")
		Expect(s.rows[6].cells[monster.Speed].Color).To(Equal(theme.Color(theme.ColorNameError)))
		Expect(s.rows[9].cells[monster.Defense].Color).To(Equal(mainColor), "Swim's main stat is not the first column")
	})

	It("keeps the soft green easy on the eyes, a muted green that is not much brighter than the colours of the stats", func() {
		r, g, b, _ := mainColor.RGBA()
		Expect(g).To(BeNumerically(">", r))
		Expect(g).To(BeNumerically(">", b))
		Expect(g >> 8).To(BeNumerically("<=", 0xd5))
		Expect(r>>8).To(BeNumerically(">=", 0x60), "no hard, pure green")
	})

	It("reads a range over the choices of the game, from the lowest to the highest", func() {
		s.apply(trainingOf(0), nil)
		s.result.SetSelected("Great")
		Expect(shown(6, monster.Pow)).To(Equal("4~10"))
	})

	It("says N/A under the main stat of a drill that has nothing to show for the chosen result", func() {
		s.apply(trainingOf(0), nil)
		s.result.SetSelected("Cheat")
		Expect(shown(6, monster.Pow)).To(Equal("N/A"))
		Expect(shown(6, monster.Life)).To(BeEmpty())

		s.apply(trainingOf(1<<0), nil)
		Expect(shown(0, monster.Pow)).To(Equal("N/A"))
		Expect(shown(1, monster.Int)).To(Equal("1~2"))
	})

	It("keeps the result that was chosen when the page loads again", func() {
		s.apply(trainingOf(0), nil)
		s.result.SetSelected("Great")
		s.apply(trainingOf(1<<1), nil)
		Expect(s.result.Selected).To(Equal("Great"))
		Expect(shown(0, monster.Pow)).To(Equal("2~5"))
	})

	It("clears what a drill showed before when another result has less to show", func() {
		s.apply(trainingOf(0), nil)
		Expect(shown(6, monster.Speed)).NotTo(BeEmpty())
		s.result.SetSelected("Cheat")
		Expect(shown(6, monster.Speed)).To(BeEmpty())
		Expect(s.rows[6].cells[monster.Speed].Color).To(Equal(theme.Color(theme.ColorNameForeground)))
	})

	It("shows the numbers with the bonus for a drill the monster specialises in, and says so under the table", func() {
		s.apply(trainingOf(1<<0), nil)
		Expect(shown(0, monster.Pow)).To(Equal("2~4"))
		s.apply(trainingOf(0), nil)
		Expect(shown(0, monster.Pow)).To(Equal("1~3"))
		Expect(s.legend.text.Text).To(ContainSubstring("bonus"))
	})

	It("explains the +1 under the table with the same white tag that the rows have, beside what it means", func() {
		Expect(s.legend.tag.FillColor).To(Equal(color.White))
		Expect(s.legend.mark.Text).To(Equal("+1"))
		Expect(s.legend.mark.Color).To(Equal(color.Black))
		Expect(s.legend.text.Text).To(Equal("specialty: its numbers are the ones with the bonus"))
		Expect(s.note.Text).NotTo(ContainSubstring("+1"))
	})

	It("says what Mot and Chance are under the table, by their names, in two short lines that the label does not have to wrap", func() {
		Expect(s.note.Text).To(ContainSubstring("motivation"))
		Expect(s.note.Text).To(ContainSubstring("Mot is"))
		Expect(s.note.Text).To(ContainSubstring("Chance is"))
		Expect(strings.Split(s.note.Text, "\n")).To(HaveLen(2))
		Expect(s.note.Wrapping).To(Equal(fyne.TextWrap(fyne.TextWrapOff)))
	})

	It("shows how keen the monster is for each drill as a plain percent against the right edge, with no fill behind it, and in red only where the monster is not keen", func() {
		s.apply(trainingOf(0), nil)
		Expect(s.rows[0].keen.Text).To(Equal("85%"))
		Expect(s.rows[9].keen.Text).To(Equal("100%"))
		Expect(s.rows[0].keen.Alignment).To(Equal(fyne.TextAlignTrailing))
		Expect(s.rows[9].keen.Alignment).To(Equal(fyne.TextAlignTrailing))
		Expect(s.rows[0].keen.Importance).To(Equal(widget.MediumImportance))
		Expect(s.rows[2].keen.Importance).To(Equal(widget.MediumImportance), "30%")
		Expect(s.rows[1].keen.Importance).To(Equal(widget.DangerImportance), "20%")
	})

	It("stars the drills the monster specialises in, and highlights the ones it is also keen on", func() {
		s.apply(trainingOf(1<<0|1<<1|1<<9), nil)
		Expect(marked(s.rows[0])).To(Equal("Domino +1"))
		Expect(marked(s.rows[1])).To(Equal("Study +1"))
		Expect(marked(s.rows[2])).To(Equal("Run"))
		Expect(marked(s.rows[9])).To(Equal("Swim +1"))

		Expect(s.rows[0].bg.FillColor).To(Equal(recommendedFill))
		Expect(s.rows[1].bg.FillColor).To(Equal(color.Transparent))
		Expect(s.rows[2].bg.FillColor).To(Equal(color.Transparent))
		Expect(s.rows[9].bg.FillColor).To(Equal(recommendedFill))
	})

	It("marks the bonus of a drill as the game does, +1 in dark text on a white tag with slightly round corners, beside the name", func() {
		s.apply(trainingOf(1<<0), nil)
		Expect(s.rows[0].name.Text).To(Equal("Domino"))
		Expect(s.rows[0].bonus.Visible()).To(BeTrue())
		Expect(s.rows[0].mark.Text).To(Equal("+1"))
		Expect(s.rows[0].mark.Color).To(Equal(color.Black))
		Expect(s.rows[0].tag.FillColor).To(Equal(color.White))
		Expect(s.rows[0].tag.CornerRadius).To(BeNumerically(">", 0))
		Expect(s.rows[0].tag.CornerRadius).To(BeNumerically("<=", 4))
		Expect(s.rows[1].bonus.Visible()).To(BeFalse())
	})

	It("clears a highlight and a +1 when the page loads again without them", func() {
		s.apply(trainingOf(1<<0), nil)
		Expect(s.rows[0].bg.FillColor).To(Equal(recommendedFill))
		s.apply(trainingOf(0), nil)
		Expect(s.rows[0].bg.FillColor).To(Equal(color.Transparent))
		Expect(marked(s.rows[0])).To(Equal("Domino"))
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

	It("shows the chance of the chosen result for each drill in a column of its own, after the motivation, as the Drill Calculator gives it", func() {
		t := trainingOf(0)
		t.Drills[0].Chances = domain.Chances{Success: 54, Failure: 36.3, Great: 3.4, Cheat: 6.3}
		s.apply(t, nil)
		Expect(s.rows[0].chance.Text).To(Equal("54%"))
		s.result.SetSelected("Cheat")
		Expect(s.rows[0].chance.Text).To(Equal("6.3%"))
		s.result.SetSelected("Great")
		Expect(s.rows[0].chance.Text).To(Equal("3.4%"))
		Expect(s.rows[0].chance.Alignment).To(Equal(fyne.TextAlignTrailing))
	})

	It("is no wider than a column of the window, so that it never makes the window grow", func() {
		s.apply(trainingOf(1<<6), nil)
		s.result.SetSelected("Great")
		Expect(s.Content().MinSize().Width).To(BeNumerically("<=", 440))
	})

	It("draws the grid in every row of the table, the head too, and a frame around the table", func() {
		Expect(s.table.Objects).To(HaveLen(1 + 10))
		for _, o := range s.table.Objects {
			Expect(slices.ContainsFunc(o.(*fyne.Container).Objects, func(c fyne.CanvasObject) bool {
				g, ok := c.(*fyne.Container)
				if !ok {
					return false
				}
				_, ok = g.Layout.(profileui.Grid)
				return ok
			})).To(BeTrue())
		}
		Expect(s.frame.StrokeWidth).To(Equal(profileui.LineWidth))
		Expect(s.frame.StrokeColor).To(Equal(profileui.LineColor))
		Expect(s.frame.FillColor).To(Equal(color.Transparent))
	})

	It("gives the percent of every drill the same width, whatever the number, so that the column reads straight", func() {
		s.apply(trainingOf(0), nil)
		s.table.Resize(fyne.NewSize(466, s.table.MinSize().Height))
		Expect(s.rows[0].keen.Size().Width).To(Equal(s.rows[9].keen.Size().Width), "85% and 100%")
		Expect(s.rows[1].keen.Size().Width).To(Equal(s.rows[9].keen.Size().Width), "20% and 100%")
	})

	It("stacks the rows with no gap, so that the lines between them run unbroken", func() {
		s.table.Resize(fyne.NewSize(466, s.table.MinSize().Height))
		for i := 1; i < len(s.table.Objects); i++ {
			prev, next := s.table.Objects[i-1], s.table.Objects[i]
			Expect(next.Position().Y).To(Equal(prev.Position().Y + prev.Size().Height))
		}
	})

	Describe("sorting by the percent", func() {
		names := func() []string {
			var out []string
			for _, r := range s.rows {
				out = append(out, strings.TrimSuffix(r.name.Text, " +1"))
			}
			return out
		}

		It("heads the motivation column with a button named Mot, short for motivation, and shows the drills in the order of the page until it is used", func() {
			s.apply(trainingOf(0), nil)
			Expect(s.sort.Text).To(Equal("Mot"))
			Expect(s.sort.Icon).To(BeNil())
			Expect(names()).To(Equal([]string{"Domino", "Study", "Run", "Shoot", "Dodge", "Endure", "Pull", "Meditate", "Leap", "Swim"}))
		})

		It("puts the most keen drill first with one tap, the least keen first with a second, and the order of the page with a third, the arrow saying which", func() {
			s.apply(trainingOf(0), nil)

			test.Tap(s.sort)
			Expect(names()).To(Equal([]string{"Swim", "Leap", "Domino", "Meditate", "Pull", "Endure", "Dodge", "Shoot", "Run", "Study"}))
			Expect(s.sort.Icon).To(Equal(theme.MenuDropDownIcon()))

			test.Tap(s.sort)
			Expect(names()).To(Equal([]string{"Study", "Run", "Shoot", "Dodge", "Endure", "Pull", "Meditate", "Domino", "Leap", "Swim"}))
			Expect(s.sort.Icon).To(Equal(theme.MenuDropUpIcon()))

			test.Tap(s.sort)
			Expect(names()[0]).To(Equal("Domino"))
			Expect(s.sort.Icon).To(BeNil())
		})

		It("moves the numbers, the +1 and the highlight of a drill with it", func() {
			s.apply(trainingOf(1<<9), nil)
			test.Tap(s.sort)
			Expect(marked(s.rows[0])).To(Equal("Swim +1"))
			Expect(s.rows[0].keen.Text).To(Equal("100%"))
			Expect(shown(0, monster.Int)).To(Equal("-3~-2"))
			Expect(s.rows[0].bg.FillColor).To(Equal(recommendedFill))
			Expect(s.rows[9].bg.FillColor).To(Equal(color.Transparent))
		})

		It("sorts by the chance of the chosen result with the Chance button, high first, then low first, and one column at a time", func() {
			s.apply(trainingOf(0), nil)
			chances := func() []float64 {
				var out []float64
				for _, r := range s.rows {
					v, err := strconv.ParseFloat(strings.TrimSuffix(r.chance.Text, "%"), 64)
					Expect(err).NotTo(HaveOccurred())
					out = append(out, v)
				}
				return out
			}
			Expect(s.byChance.Text).To(Equal("Chance"))

			test.Tap(s.sort)
			test.Tap(s.byChance)
			Expect(s.byChance.Icon).To(Equal(theme.MenuDropDownIcon()))
			Expect(s.sort.Icon).To(BeNil())
			Expect(slices.IsSortedFunc(chances(), func(a, b float64) int { return cmp.Compare(b, a) })).To(BeTrue(), fmt.Sprint(chances()))
			Expect(slices.Compact(chances())).NotTo(HaveLen(1))

			test.Tap(s.byChance)
			Expect(slices.IsSorted(chances())).To(BeTrue(), fmt.Sprint(chances()))

			s.result.SetSelected("Great")
			Expect(slices.IsSorted(chances())).To(BeTrue(), fmt.Sprint(chances()))

			test.Tap(s.sort)
			Expect(s.byChance.Icon).To(BeNil())
			Expect(s.sort.Icon).To(Equal(theme.MenuDropDownIcon()))
		})

		It("keeps the order when the page loads again", func() {
			s.apply(trainingOf(0), nil)
			test.Tap(s.sort)
			s.apply(trainingOf(1<<1), nil)
			Expect(names()[0]).To(Equal("Swim"))
		})
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
		Eventually(func() string { return marked(s.rows[0]) }).Should(Equal("Domino +1"))
	})

	It("renders the page to a PNG", Label("screenshot"), func() {
		dir := os.Getenv("SCREENSHOT_DIR")
		if dir == "" {
			Skip("set SCREENSHOT_DIR to write the screenshots")
		}
		s.apply(trainingOf(1<<0|1<<6|1<<10), nil)
		Expect(os.MkdirAll(dir, 0o755)).To(Succeed())
		f, err := os.Create(filepath.Join(dir, "drills.png"))
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
