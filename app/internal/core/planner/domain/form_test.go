package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/planner/domain"
)

var _ = Describe("the form of the monster", func() {
	forms := func(start int, foods map[int]domain.Food, steps ...domain.Step) []int {
		settings := domain.DefaultSettings()
		settings.Form = start
		p := domain.Build(domain.NewReactions(nil), shippedEffects(), settings, steps, foods)
		var out []int
		for _, r := range p.Rows {
			if r.Shown {
				out = append(out, r.Form)
			}
		}
		return out
	}

	It("changes with the item of a week and then with its drill, from the form of the game", func() {
		Expect(forms(10, nil,
			domain.Step{Item: domain.Mango, Action: domain.LightDrill},
			domain.Step{Action: domain.HardDrill},
			domain.Step{Item: domain.Candy, Action: domain.Rest},
		)).To(Equal([]int{10, 8, 18}))
	})

	It("changes with the food of a month", func() {
		Expect(forms(0, map[int]domain.Food{2: domain.Meat},
			domain.Step{Action: domain.Rest}, domain.Step{Action: domain.Rest}, domain.Step{Action: domain.Rest}, domain.Step{Action: domain.Rest},
			domain.Step{Action: domain.Rest},
		)).To(Equal([]int{0, 0, 0, 0, 6, 6}))
	})

	It("does not change with a rest, a battle, an errantry, an expedition, or a freeze, nor with what has no change told", func() {
		Expect(forms(-5, nil,
			domain.Step{Item: domain.GoldPeach, Action: domain.BattleWin},
			domain.Step{Action: domain.Errantry},
		)).To(Equal([]int{-5, -5, -5, -5, -5}))
	})

	It("stays from -100 to 100", func() {
		Expect(forms(-90, nil, domain.Step{Item: domain.Powder, Action: domain.HardDrill})).To(Equal([]int{-100}))
		Expect(forms(95, nil, domain.Step{Item: domain.Candy, Action: domain.Rest})).To(Equal([]int{100}))
	})
})
