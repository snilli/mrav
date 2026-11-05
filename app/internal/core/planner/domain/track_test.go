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

		It("leaves the food of a month to come while the game is on the first week of that month, as it is given in that week", func() {
			p := build(drill, drill, drill)
			passed, ok := p.Passed(2)
			Expect(ok).To(BeTrue())
			Expect(passed.Rows).To(Equal(p.Rows[:2]))
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

		It("takes every week of the plan when more weeks passed than it has, leaving the empty week for the next step", func() {
			passed, ok := build(drill, drill).Passed(9)
			Expect(ok).To(BeTrue())
			Expect(passed.Step).To(Equal(2))
			Expect(passed.Rows).To(HaveLen(2))
		})
	})

	Describe("the weeks since the plan was begun", func() {
		who := domain.WhoOf(monster.Monster{Genus: 34, SubGenus: 34, Name: "Rocky", InitialLifespan: 350})
		begun := domain.Tracking{Started: true, Date: 100, Who: who}
		now := func(date int, w domain.Who) domain.Current { return domain.Current{Date: date, Dated: true, Who: w} }

		It("counts the weeks that the game has gone on since then", func() {
			weeks, reset := begun.Weeks(now(103, who))
			Expect([]any{weeks, reset}).To(Equal([]any{3, false}))
		})

		It("starts again when the monster in the game is another one", func() {
			other := domain.WhoOf(monster.Monster{Genus: 34, SubGenus: 34, Name: "Rocky", InitialLifespan: 300})
			_, reset := begun.Weeks(now(103, other))
			Expect(reset).To(BeTrue())
		})

		It("starts again when the game goes back in time, as a save that was loaded", func() {
			_, reset := begun.Weeks(now(99, who))
			Expect(reset).To(BeTrue())
		})

		It("starts again when the date of the game cannot be read", func() {
			_, reset := begun.Weeks(domain.Current{Who: who})
			Expect(reset).To(BeTrue())
		})

		It("has nothing to count before a plan is begun", func() {
			weeks, reset := domain.Tracking{}.Weeks(now(103, who))
			Expect([]any{weeks, reset}).To(Equal([]any{0, false}))
		})
	})
})
