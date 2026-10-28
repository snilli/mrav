package domain_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	"app/internal/core/planner/domain"
)

var _ = Describe("a plan", func() {
	var (
		reactions = domain.NewReactions(nil)
		drill     = domain.Step{Action: domain.LightDrill}
		rest      = domain.Step{Action: domain.Rest}
	)

	build := func(steps ...domain.Step) domain.Plan {
		return domain.Build(reactions, shippedEffects(), domain.DefaultSettings(), steps, nil)
	}
	ids := func(p domain.Plan) []string {
		var out []string
		for _, r := range p.Rows {
			out = append(out, fmt.Sprintf("%d-%d", r.Month, r.Week))
		}
		return out
	}
	kinds := func(p domain.Plan) []domain.Kind {
		var out []domain.Kind
		for _, r := range p.Rows {
			out = append(out, r.Kind)
		}
		return out
	}

	It("starts as the settings of the page: DX, an Ape with an Ape, that likes milk and dislikes potato, in its first life stage, nothing in it", func() {
		Expect(domain.DefaultSettings()).To(Equal(domain.Settings{
			Version: monster.MR2DX, Main: 35, Sub: 35, Like: domain.Milk, Dislike: domain.Potato, LifeStage: 1,
		}))
	})

	It("is one empty week to fill, and nothing lived, when it has no step", func() {
		p := build()
		Expect(ids(p)).To(Equal([]string{"1-1"}))
		Expect(p.Rows[0].Kind).To(Equal(domain.NextRow))
		Expect(p.Rows[0].Step).To(Equal(0))
		Expect(p.Rows[0].Shown).To(BeFalse())
		Expect(p.Totals).To(Equal(domain.Totals{}))
		Expect(p.Problems).To(BeEmpty())
	})

	It("gives each step a week, and ends with an empty week for the next one", func() {
		p := build(drill, rest)
		Expect(ids(p)).To(Equal([]string{"1-1", "1-2", "1-3"}))
		Expect(kinds(p)).To(Equal([]domain.Kind{domain.StepRow, domain.StepRow, domain.NextRow}))
		Expect([]int{p.Rows[0].Step, p.Rows[1].Step, p.Rows[2].Step}).To(Equal([]int{0, 1, 2}))
		Expect(p.Totals.Lived).To(Equal(2))
	})

	It("skips a step with neither an item nor an action, as the page does not keep a blank week", func() {
		p := build(drill, domain.Step{}, rest)
		Expect(ids(p)).To(Equal([]string{"1-1", "1-2", "1-3"}))
		Expect([]int{p.Rows[0].Step, p.Rows[1].Step, p.Rows[2].Step}).To(Equal([]int{0, 2, 3}))
	})

	It("puts the food of a month before its first week, Potato unless it is told another", func() {
		steps := []domain.Step{drill, drill, drill, drill}
		p := domain.Build(reactions, shippedEffects(), domain.DefaultSettings(), steps, nil)
		Expect(ids(p)).To(Equal([]string{"1-1", "1-2", "1-3", "1-4", "2-0", "2-1"}))
		Expect(p.Rows[4].Kind).To(Equal(domain.FoodRow))
		Expect(p.Rows[4].Food).To(Equal(domain.Potato))
		Expect(p.Rows[4].Step).To(Equal(-1))

		p = domain.Build(reactions, shippedEffects(), domain.DefaultSettings(), steps, map[int]domain.Food{2: domain.Tablet})
		Expect(p.Rows[4].Food).To(Equal(domain.Tablet))
	})

	It("takes four weeks for an errantry, the month that it enters has no food, and only its first week is a step to edit, the others follow it", func() {
		p := build(drill, drill, domain.Step{Action: domain.Errantry}, drill)
		Expect(ids(p)).To(Equal([]string{"1-1", "1-2", "1-3", "1-4", "2-1", "2-2", "2-3", "2-4"}))
		Expect(kinds(p)).To(Equal([]domain.Kind{
			domain.StepRow, domain.StepRow, domain.StepRow, domain.FollowRow, domain.FollowRow, domain.FollowRow, domain.StepRow, domain.NextRow,
		}))
		Expect([]domain.Action{p.Rows[2].Action, p.Rows[3].Action, p.Rows[4].Action, p.Rows[5].Action}).To(Equal([]domain.Action{
			domain.Errantry, domain.Errantry, domain.Errantry, domain.Errantry,
		}))
		Expect([]int{p.Rows[2].Step, p.Rows[3].Step, p.Rows[5].Step, p.Rows[6].Step}).To(Equal([]int{2, 2, 2, 3}))
	})

	It("keeps the food of a month that an errantry ends with", func() {
		p := build(domain.Step{Action: domain.Errantry})
		Expect(ids(p)).To(Equal([]string{"1-1", "1-2", "1-3", "1-4", "2-0", "2-1"}))
	})

	It("takes the four weeks of its month for an expedition, the three after the first away", func() {
		p := build(domain.Step{Action: domain.Expedition})
		Expect(ids(p)).To(Equal([]string{"1-1", "1-2", "1-3", "1-4", "2-0", "2-1"}))
		Expect([]domain.Action{p.Rows[0].Action, p.Rows[1].Action, p.Rows[2].Action, p.Rows[3].Action}).To(Equal([]domain.Action{
			domain.Expedition, domain.Away, domain.Away, domain.Away,
		}))
		Expect(p.Rows[1].Kind).To(Equal(domain.FollowRow))
	})

	It("leaves the next month without food when the last week of a month is frozen, and not when another week is", func() {
		Expect(ids(build(drill, drill, drill, domain.Step{Action: domain.Freeze}))).To(Equal([]string{"1-1", "1-2", "1-3", "1-4", "2-1"}))
		Expect(ids(build(domain.Step{Action: domain.Freeze}))).To(Equal([]string{"1-1", "1-2"}))
	})

	It("refuses a second peach of a kind but keeps the week, and says so", func() {
		peach := domain.Step{Item: domain.GoldPeach, Action: domain.Rest}
		p := build(peach, peach)
		Expect(p.Problems).To(Equal([]domain.Problem{{Step: 1, Reason: domain.PeachUsed}}))
		Expect(p.Rows[0].Item).To(Equal(domain.GoldPeach))
		Expect(p.Rows[1].Item).To(Equal(domain.NoItem))
		Expect(p.Rows[1].Action).To(Equal(domain.Rest))
		Expect(p.Totals.Lived).To(Equal(2))
	})

	It("refuses a peach that the monster in the game has eaten already, keeps the week, and says so in its own words", func() {
		settings := domain.DefaultSettings()
		settings.GoldPeach = true
		p := domain.Build(reactions, shippedEffects(), settings, []domain.Step{{Item: domain.GoldPeach, Action: domain.Rest}, {Item: domain.SilverPeach, Action: domain.Rest}}, nil)
		Expect(p.Problems).To(Equal([]domain.Problem{{Step: 0, Reason: domain.PeachEaten}}))
		Expect([]domain.Item{p.Rows[0].Item, p.Rows[1].Item}).To(Equal([]domain.Item{domain.NoItem, domain.SilverPeach}))
		Expect(p.Rows[0].Action).To(Equal(domain.Rest))
		Expect(domain.PeachEaten.Text()).To(Equal("the monster has eaten that peach already, so it is left out"))
	})

	It("keeps an expedition that does not begin a month as a week with nothing in it, so that it can be seen and mended, and says so", func() {
		p := build(drill, domain.Step{Action: domain.Expedition}, rest)
		Expect(p.Problems).To(Equal([]domain.Problem{{Step: 1, Reason: domain.ExpeditionWeek}}))
		Expect(ids(p)).To(Equal([]string{"1-1", "1-2", "1-3", "1-4"}))
		Expect(p.Rows[1].Kind).To(Equal(domain.StepRow))
		Expect(p.Rows[1].Action).To(Equal(domain.NoAction))
		Expect(p.Rows[1].Step).To(Equal(1))
		Expect(p.Rows[2].Action).To(Equal(domain.Rest))
	})

	It("drops the item of an expedition that does not begin a month with the rest of the step, as the page does", func() {
		p := build(drill, domain.Step{Item: domain.Mango, Action: domain.Expedition}, rest)
		Expect(p.Rows[1].Item).To(Equal(domain.NoItem))
		Expect(p.Totals).To(Equal(build(drill, domain.Step{Action: domain.Expedition}, rest).Totals))
	})

	It("feeds Potato in a month whose food is not one of the six, the same as a month with no food chosen", func() {
		steps := []domain.Step{drill, drill, drill, drill}
		p := domain.Build(reactions, shippedEffects(), domain.DefaultSettings(), steps, map[int]domain.Food{2: domain.NoFood})
		Expect(p.Rows[4].Food).To(Equal(domain.Potato))
		Expect(p.Totals).To(Equal(domain.Build(reactions, shippedEffects(), domain.DefaultSettings(), steps, nil).Totals))
	})

	It("offers an expedition to the first week of a month only, and every other action to every week", func() {
		for week := 1; week <= 4; week++ {
			row := domain.Row{Month: 1, Week: week}
			Expect(row.Offers(domain.Expedition)).To(Equal(week == 1), fmt.Sprint(week))
			for _, a := range domain.Actions() {
				if a != domain.Expedition {
					Expect(row.Offers(a)).To(BeTrue(), a.Name())
				}
			}
		}
	})

	It("lists the nine actions and the nineteen items of the page, in its order", func() {
		var actions, items []string
		for _, a := range domain.Actions() {
			actions = append(actions, a.Name())
		}
		for _, it := range domain.Items() {
			items = append(items, it.Name())
		}
		Expect(actions).To(Equal([]string{"Rest", "Light Drill", "Hard Drill", "Errantry", "Battle (Win)", "Battle (OK)", "Battle (Lose)", "Expedition", "Freeze"}))
		Expect(items).To(Equal([]string{
			"No item", "Mint Leaf", "Nuts Oil", "MB +10", "MB +0", "MB -10", "Irritater", "Mango", "Paradoxine", "Troron",
			"Candy", "Smoked Snake", "Apple Cake", "Powder", "Sweet Jelly", "Sour Jelly", "Star Prune", "Gold Peach", "Silver Peach", "Griever",
		}))
	})

	Describe("the items that a step may take", func() {
		It("is every item, for a plan that has no peach", func() {
			Expect(domain.ItemsFor(domain.DefaultSettings(), []domain.Step{drill}, 0)).To(Equal(domain.Items()))
			Expect(domain.ItemsFor(domain.DefaultSettings(), nil, 0)).To(Equal(domain.Items()))
		})

		It("leaves out a peach that another step has, as a monster is given each of them once", func() {
			steps := []domain.Step{{Item: domain.GoldPeach, Action: domain.Rest}, drill, {Item: domain.SilverPeach, Action: domain.Rest}}
			names := func(items []domain.Item) []string {
				var out []string
				for _, it := range items {
					out = append(out, it.Name())
				}
				return out
			}

			Expect(names(domain.ItemsFor(domain.DefaultSettings(), steps, 1))).NotTo(ContainElements("Gold Peach", "Silver Peach"))
			Expect(names(domain.ItemsFor(domain.DefaultSettings(), steps, 3))).NotTo(ContainElements("Gold Peach", "Silver Peach"))
		})

		It("offers a peach that a frozen step has, as freezing gives no item", func() {
			steps := []domain.Step{{Item: domain.GoldPeach, Action: domain.Freeze}, drill}
			Expect(domain.ItemsFor(domain.DefaultSettings(), steps, 1)).To(ContainElement(domain.GoldPeach))
		})

		It("leaves out a peach that the monster in the game has eaten already", func() {
			settings := domain.DefaultSettings()
			settings.SilverPeach = true
			got := domain.ItemsFor(settings, []domain.Step{drill}, 1)
			Expect(got).To(ContainElement(domain.GoldPeach))
			Expect(got).NotTo(ContainElement(domain.SilverPeach))
		})

		It("keeps the peach that the step itself has, so that it can be seen and changed", func() {
			steps := []domain.Step{{Item: domain.GoldPeach, Action: domain.Rest}, {Item: domain.SilverPeach, Action: domain.Rest}}
			got := domain.ItemsFor(domain.DefaultSettings(), steps, 0)
			Expect(got).To(ContainElement(domain.GoldPeach))
			Expect(got).NotTo(ContainElement(domain.SilverPeach))
		})
	})

	It("says in words what it refused", func() {
		Expect(domain.PeachUsed.Text()).To(ContainSubstring("peach"))
		Expect(domain.ExpeditionWeek.Text()).To(ContainSubstring("first week"))
	})

	Describe("editing its steps", func() {
		It("changes a step, and does not touch the list that it was given", func() {
			steps := []domain.Step{drill, rest}
			got := domain.SetStep(steps, 1, domain.Step{Item: domain.Mango, Action: domain.HardDrill})
			Expect(got).To(Equal([]domain.Step{drill, {Item: domain.Mango, Action: domain.HardDrill}}))
			Expect(steps).To(Equal([]domain.Step{drill, rest}))
		})

		It("adds a step when it is the one after the last", func() {
			Expect(domain.SetStep([]domain.Step{drill}, 1, rest)).To(Equal([]domain.Step{drill, rest}))
			Expect(domain.SetStep(nil, 0, rest)).To(Equal([]domain.Step{rest}))
		})

		It("takes a step away when it is made empty, and adds nothing when an empty one is added", func() {
			Expect(domain.SetStep([]domain.Step{drill, rest, drill}, 1, domain.Step{})).To(Equal([]domain.Step{drill, drill}))
			Expect(domain.SetStep([]domain.Step{drill}, 1, domain.Step{})).To(Equal([]domain.Step{drill}))
		})

		It("ignores a place that is past the one after the last", func() {
			Expect(domain.SetStep([]domain.Step{drill}, 3, rest)).To(Equal([]domain.Step{drill}))
			Expect(domain.SetStep([]domain.Step{drill}, -1, rest)).To(Equal([]domain.Step{drill}))
		})

		It("plays the plan again from its start after a change, so that making an errantry a rest takes away its other three weeks and the steps after it move up", func() {
			steps := []domain.Step{{Action: domain.Errantry}, drill}
			before := build(steps...)
			Expect(ids(before)).To(Equal([]string{"1-1", "1-2", "1-3", "1-4", "2-0", "2-1", "2-2"}))

			after := build(domain.SetStep(steps, 0, rest)...)
			Expect(ids(after)).To(Equal([]string{"1-1", "1-2", "1-3"}))
			Expect([]domain.Action{after.Rows[0].Action, after.Rows[1].Action}).To(Equal([]domain.Action{domain.Rest, domain.LightDrill}))
		})
	})

	Describe("the totals", func() {
		It("gives the weeks that were lived, the weeks lost (below zero) or gained, and the average for a month of four weeks", func() {
			Expect(domain.TotalsOf(-12, 8)).To(Equal(domain.Totals{Lived: 8, Weeks: -12, Average: -6}))
			Expect(domain.TotalsOf(50, 20)).To(Equal(domain.Totals{Lived: 20, Weeks: 50, Average: 10}))
		})

		It("has an average of nothing when no week was lived", func() {
			Expect(domain.TotalsOf(0, 0).Average).To(BeZero())
			Expect(domain.TotalsOf(-3, 0).Average).To(BeZero())
		})

		It("rounds the average to a tenth the way the page does, a half going up, below zero too", func() {
			Expect(domain.TotalsOf(1, 16).Average).To(BeNumerically("~", 0.3, 1e-9))
			Expect(domain.TotalsOf(-1, 16).Average).To(BeNumerically("~", -0.2, 1e-9))
		})
	})
})
