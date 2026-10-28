package domain_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	"app/internal/core/planner/domain"
	"app/internal/testutil"
)

func shippedEffects() domain.Effects {
	effects, err := testutil.ShippedEffects(context.Background())
	if errors.Is(err, testutil.ErrNoShippedDB) {
		Skip(err.Error())
	}
	Expect(err).NotTo(HaveOccurred())
	return effects
}

//go:fix inline
func ptr[T any](v T) *T { return new(v) }

var _ = Describe("the effects of the items, foods and actions", func() {
	effects := domain.NewEffects([]domain.Effect{
		{Subject: "item:nuts_oil", Version: monster.MR2, Fatigue: -28},
		{Subject: "item:nuts_oil", Version: monster.MR2, Genus: new(5), Fatigue: -28, StressFactor: new(0.8)},
		{Subject: "item:nuts_oil", Version: monster.MR2, Genus: new(9), Fatigue: -28, StressFactor: new(0.8)},
		{Subject: "action:rest", Version: monster.MR2, Fatigue: -32, Stress: -5},
		{Subject: "action:rest", Version: monster.MR2, LifeStage: new(1), Fatigue: -36, Stress: -5},
		{Subject: "item:candy", Version: monster.MR2, Stress: -2, Form: 10},
		{Subject: "item:candy", Version: monster.MR2DX, Stress: -2, Form: 10},
		{Subject: "item:gold_peach", Version: monster.MR2, Lost: 50},
	})
	settings := func(main, sub, stage int) domain.Settings {
		return domain.Settings{Version: monster.MR2, Main: main, Sub: sub, LifeStage: stage, Fatigue: 50, Stress: 50}
	}
	first := func(s domain.Settings, step domain.Step) domain.Row {
		return domain.Build(domain.NewReactions(nil), effects, s, []domain.Step{step}, nil).Rows[0]
	}

	It("uses the row for the genus of the monster, once, when its main and its sub both have one", func() {
		r := first(settings(5, 9, 3), domain.Step{Item: domain.NutsOil})
		Expect([]int{r.Fatigue, r.Stress}).To(Equal([]int{22, 40}))
	})

	It("uses the row for any genus when the monster has no row of its own", func() {
		r := first(settings(1, 2, 3), domain.Step{Item: domain.NutsOil})
		Expect([]int{r.Fatigue, r.Stress}).To(Equal([]int{22, 50}))
	})

	It("uses the row for the life stage of the monster, and the row for any stage otherwise", func() {
		Expect(first(settings(1, 1, 1), domain.Step{Action: domain.Rest}).Fatigue).To(Equal(14))
		Expect(first(settings(1, 1, 8), domain.Step{Action: domain.Rest}).Fatigue).To(Equal(18))
	})

	It("tells what an item does to the monster in the game, in words", func() {
		Expect(effects.Text(domain.Candy, settings(1, 1, 3))).To(Equal("Stress -2, Form +10"))
		Expect(effects.Text(domain.NutsOil, settings(5, 1, 3))).To(Equal("Fatigue -28, Stress x0.8"))
		Expect(effects.Text(domain.NutsOil, settings(1, 1, 3))).To(Equal("Fatigue -28"))
		Expect(effects.Text(domain.GoldPeach, settings(1, 1, 3))).To(Equal("Life +50 weeks"))
		Expect(effects.Text(domain.MagicBananaZero, settings(1, 1, 3))).To(Equal("no change"))
	})

	It("knows the subjects that an effect can be told for, and no other", func() {
		for _, subject := range []string{"item:mint_leaf", "item:griever", "food:cup_jelly", "action:rest", "action:battle_lose", "action:expedition"} {
			Expect(domain.KnownSubject(subject)).To(BeTrue(), subject)
		}
		for _, subject := range []string{"item:none", "item:pile_of_meat", "food:none", "action:freeze", "action:away", "mango", ""} {
			Expect(domain.KnownSubject(subject)).To(BeFalse(), subject)
		}
	})
})
