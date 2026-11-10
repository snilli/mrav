package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	"app/internal/core/planner/domain"
)

var _ = Describe("following the game while the plan is played", func() {
	var (
		drill    = domain.Step{Action: domain.LightDrill}
		errantry = domain.Step{Action: domain.Errantry}
	)
	build := func(steps ...domain.Step) domain.Plan {
		s := domain.DefaultSettings()
		s.Month, s.Week = 4, 3
		return domain.Build(domain.NewReactions(nil), shippedEffects(), s, steps, nil)
	}

	Describe("the weeks of a plan that have passed", func() {
		It("is nothing to take away when no week has passed", func() {
			passed, ok := build(drill, drill).Passed(0)
			Expect(ok).To(BeTrue())
			Expect(passed.Rows).To(BeEmpty())
			Expect([]int{passed.Step, passed.Month}).To(Equal([]int{0, 1}))
		})

		It("takes the weeks that passed, with the food of a month that they go into, and goes on from the next step", func() {
			p := build(drill, drill, drill, drill)
			passed, ok := p.Passed(3)
			Expect(ok).To(BeTrue())
			Expect(passed.Rows).To(Equal(p.Rows[:4]))
			Expect(passed.Rows[2].Kind).To(Equal(domain.FoodRow))
			Expect([]int{passed.Step, passed.Month}).To(Equal([]int{3, 2}))
		})

		It("takes the food of the month that the game has come to, as the game gives it at the start of that week", func() {
			p := build(drill, drill, drill)
			passed, ok := p.Passed(2)
			Expect(ok).To(BeTrue())
			Expect(passed.Rows).To(Equal(p.Rows[:3]))
			Expect(passed.Rows[2].Kind).To(Equal(domain.FoodRow))
			Expect([]int{passed.Step, passed.Month}).To(Equal([]int{2, 2}))
		})

		It("waits while the game is in the middle of an errantry, as the plan cannot go on from a week away", func() {
			p := build(drill, errantry, drill)
			_, ok := p.Passed(2)
			Expect(ok).To(BeFalse())

			passed, ok := p.Passed(5)
			Expect(ok).To(BeTrue())
			Expect(passed.Step).To(Equal(2))
		})

		It("takes the item out of the week that the game played without one, and leaves the plan as it was", func() {
			p := build(domain.Step{Item: domain.NutsOil, Action: domain.LightDrill}, drill)
			passed, _ := p.Passed(1)
			Expect(passed.Played(false, domain.Settings{}).Rows[0].Item).To(Equal(domain.NoItem))
			Expect(passed.Played(true, domain.Settings{}).Rows[0].Item).To(Equal(domain.NutsOil))
			Expect(p.Rows[0].Item).To(Equal(domain.NutsOil))
		})

		It("ends the last week played where the game is now, as the game is the truth of what the week did", func() {
			p := build(drill, drill, drill)
			passed, _ := p.Passed(2)
			was := p.Rows[1]
			after := domain.Settings{Fatigue: 25, Stress: 19, Form: -86, Lifespan: was.Life - 3}
			last := passed.Played(true, after).Rows[1]
			Expect([]int{last.Fatigue, last.Stress, last.LI, last.Form}).To(Equal([]int{25, 19, 63, -86}))
			Expect([]int{last.Life, last.Lost, last.Total}).To(Equal([]int{was.Life, was.Lost, was.Total}))
			Expect(p.Rows[1].Fatigue).NotTo(Equal(25))
		})

		It("ages a week played by the lifespan that the game takes off a moment after the date turns, once it has", func() {
			r := domain.Row{Life: 99, Lost: 0, Total: 0, ShowLife: true}
			aged, ok := r.Aged(100, 98)
			Expect([]any{aged, ok}).To(Equal([]any{domain.Row{Life: 98, Lost: -1, Total: -1, ShowLife: true}, true}))
			aged, ok = r.Aged(100, 100)
			Expect([]any{aged, ok}).To(Equal([]any{r, false}))
			aged, ok = domain.Row{Life: 99}.Aged(100, 98)
			Expect([]any{aged, ok}).To(Equal([]any{domain.Row{Life: 99}, true}))
		})

		It("takes every week of the plan when more weeks passed than it has, leaving the empty week for the next step", func() {
			passed, ok := build(drill, drill).Passed(9)
			Expect(ok).To(BeTrue())
			Expect(passed.Step).To(Equal(2))
			Expect(passed.Rows).To(HaveLen(3))
		})
	})

	Describe("the weeks since the plan was begun", func() {
		who := domain.WhoOf(monster.Monster{Genus: 34, SubGenus: 34, Name: "Rocky", InitialLifespan: 350})
		begun := domain.Tracking{Started: true, Start: 100, Date: 102, Who: who}
		now := func(date int, w domain.Who) domain.Current { return domain.Current{Date: date, Dated: true, Who: w} }

		It("counts the weeks that the game has gone on since the plan last followed it", func() {
			weeks, ok := begun.Weeks(now(105, who))
			Expect([]any{weeks, ok}).To(Equal([]any{3, true}))
		})

		It("counts back when a save from a week of the plan is loaded", func() {
			weeks, ok := begun.Weeks(now(100, who))
			Expect([]any{weeks, ok}).To(Equal([]any{-2, true}))
		})

		It("does not count for another monster in the game, which the plan is not for", func() {
			other := domain.WhoOf(monster.Monster{Genus: 34, SubGenus: 34, Name: "Rocky", InitialLifespan: 300})
			_, ok := begun.Weeks(now(105, other))
			Expect(ok).To(BeFalse())
			Expect(begun.Another(now(105, other))).To(BeTrue())
			Expect(begun.Another(now(105, who))).To(BeFalse())
			Expect(begun.Another(domain.Current{})).To(BeFalse())
			Expect(domain.Tracking{}.Another(now(105, other))).To(BeFalse())
		})

		It("waits while the game is before the week the plan began", func() {
			_, ok := begun.Weeks(now(99, who))
			Expect(ok).To(BeFalse())
		})

		It("waits while the date of the game cannot be read", func() {
			_, ok := begun.Weeks(domain.Current{Who: who})
			Expect(ok).To(BeFalse())
		})

		It("has nothing to count before a plan is begun", func() {
			_, ok := domain.Tracking{}.Weeks(now(103, who))
			Expect(ok).To(BeFalse())
		})
	})
})
