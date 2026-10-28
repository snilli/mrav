package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	"app/internal/core/planner/domain"
)

var _ = Describe("the life that is left", func() {
	settings := func(version monster.Version, lifespan int) domain.Settings {
		s := domain.DefaultSettings()
		s.Version, s.Lifespan = version, lifespan
		return s
	}
	build := func(s domain.Settings, steps ...domain.Step) domain.Plan {
		return domain.Build(domain.NewReactions(nil), shippedEffects(), s, steps, nil)
	}
	type life struct{ Total, Life int }
	lives := func(p domain.Plan) []life {
		var out []life
		for _, r := range p.Rows {
			if r.ShowLife {
				out = append(out, life{r.Total, r.Life})
			}
		}
		return out
	}

	It("takes the lifespan of the game a week for each week lived, and the weeks lost of the week as well, and adds up what is lost", func() {
		p := build(settings(monster.MR2DX, 100),
			domain.Step{Action: domain.LightDrill}, domain.Step{Action: domain.BattleLose}, domain.Step{Item: domain.GoldPeach, Action: domain.Rest})

		Expect([]int{p.Rows[0].Lost, p.Rows[1].Lost, p.Rows[2].Lost}).To(Equal([]int{0, -3, 50}))
		Expect(lives(p)).To(Equal([]life{{0, 99}, {-3, 95}, {47, 144}}))
	})

	It("ends on what the sums under the table give: the lifespan, less the weeks lived, with the weeks lost", func() {
		p := build(settings(monster.MR2, 200),
			domain.Step{Action: domain.HardDrill}, domain.Step{Action: domain.Errantry}, domain.Step{Action: domain.Rest}, domain.Step{Action: domain.BattleWin})

		last := lives(p)[len(lives(p))-1]
		Expect(last.Total).To(Equal(p.Totals.Weeks))
		Expect(last.Life).To(Equal(200 - p.Totals.Lived + p.Totals.Weeks))
	})

	It("takes a week for each week away on an expedition, and the expedition's own weeks lost on its first week", func() {
		p := build(settings(monster.MR2DX, 50), domain.Step{Action: domain.Expedition})

		Expect(lives(p)).To(Equal([]life{{2, 51}, {2, 50}, {2, 49}, {2, 48}}))
	})

	It("takes nothing for a frozen week, and has nothing for the food of a month or the empty week to fill", func() {
		p := build(settings(monster.MR2DX, 50), domain.Step{Action: domain.Freeze}, domain.Step{Action: domain.LightDrill})

		Expect(lives(p)).To(Equal([]life{{0, 50}, {0, 49}}))
		Expect(p.Rows[len(p.Rows)-1].ShowLife).To(BeFalse())
	})

	It("takes the lifespan from the monster of the game", func() {
		Expect(domain.SettingsOf(monster.Monster{Lifespan: 123}, [20]int{}).Lifespan).To(Equal(123))
	})
})
