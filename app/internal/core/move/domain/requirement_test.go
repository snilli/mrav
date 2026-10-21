package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	"app/internal/core/move/domain"
)

var _ = Describe("Requirements", func() {
	names := domain.Names{
		Move:  func(slot int) string { return map[int]string{3: "Kick"}[slot] },
		Genus: func(g monster.Genus) string { return map[monster.Genus]string{1: "Dragon", 5: "Henger", 24: "Jill"}[g] },
	}
	require := func(c domain.Condition, m monster.Monster) domain.Requirement { return c.Require(m, names) }

	It("says a stat sum in the names of the stats, with what the monster has", func() {
		m := monster.Monster{Stats: [monster.StatCount]int{0, 300, 154, 0, 0, 0}}
		c := domain.StatSumAtLeast{Stats: []monster.Stat{monster.Pow, monster.Int}, Min: 400}
		stats := []monster.Stat{monster.Pow, monster.Int} // which stats the words are about, to show them in their colours
		Expect(require(c, m)).To(Equal(domain.Requirement{Text: "POW + INT at least 400", Have: "454", Met: true, Stats: stats}))

		m.Stats[monster.Int] = 99
		Expect(require(c, m)).To(Equal(domain.Requirement{Text: "POW + INT at least 400", Have: "399", Met: false, Stats: stats}))
	})

	It("says how often a move has been used out of how often it must be", func() {
		m := monster.Monster{}
		m.Used[3] = 31
		c := domain.MoveUsedAtLeast{Slot: 3, Min: 50}
		Expect(require(c, m)).To(Equal(domain.Requirement{Text: "Use Kick at least 50 times", Have: "31 / 50", Met: false}))

		m.Used[3] = 50
		Expect(require(c, m).Met).To(BeTrue())
	})

	It("falls back to the slot number for a move it has no name for", func() {
		c := domain.MoveUsedAtLeast{Slot: 9, Min: 10}
		Expect(c.Require(monster.Monster{}, domain.Names{}).Text).To(Equal("Use move 9 at least 10 times"))
	})

	It("says a nature bound with the nature the monster has, positive or negative", func() {
		Expect(require(domain.NatureAtLeast{Min: 20}, monster.Monster{Nature: 88})).
			To(Equal(domain.Requirement{Text: "Nature at least 20", Have: "88", Met: true}))
		Expect(require(domain.NatureAtMost{Max: -50}, monster.Monster{Nature: 88})).
			To(Equal(domain.Requirement{Text: "Nature at most -50", Have: "88", Met: false}))
		Expect(require(domain.NatureAtMost{Max: -50}, monster.Monster{Nature: -70}).Met).To(BeTrue())
	})

	It("names the sub-genera that qualify, and the one the monster has", func() {
		c := domain.SubGenusIn{IDs: []monster.Genus{24, 5}}
		Expect(require(c, monster.Monster{SubGenus: 1})).
			To(Equal(domain.Requirement{Text: "Sub-genus is Jill or Henger", Have: "Dragon", Met: false}))
		Expect(require(c, monster.Monster{SubGenus: 5}).Met).To(BeTrue())
	})

	It("names the sub-genera that disqualify", func() {
		c := domain.SubGenusNotIn{IDs: []monster.Genus{24}}
		Expect(require(c, monster.Monster{SubGenus: 24})).
			To(Equal(domain.Requirement{Text: "Sub-genus is not Jill", Have: "Jill", Met: false}))
		Expect(require(c, monster.Monster{SubGenus: 1}).Met).To(BeTrue())
	})

	It("falls back to the number for a genus it has no name for", func() {
		Expect(domain.SubGenusIn{IDs: []monster.Genus{9}}.Require(monster.Monster{SubGenus: 2}, domain.Names{}).Text).
			To(Equal("Sub-genus is #9"))
	})

	Describe("of a rule", func() {
		It("lists every condition in order, met or not", func() {
			rule := domain.UnlockRule{All: []domain.Condition{
				domain.NatureAtLeast{Min: 20},
				domain.MoveUsedAtLeast{Slot: 3, Min: 50},
			}}
			got := rule.Requirements(monster.Monster{Nature: 88}, names)
			Expect(got).To(HaveLen(2))
			Expect(got[0].Met).To(BeTrue())
			Expect(got[1].Met).To(BeFalse())
		})

		It("says there is no way to learn a move the game gives no way to learn", func() {
			got := domain.UnlockRule{Never: true}.Requirements(monster.Monster{}, names)
			Expect(got).To(Equal([]domain.Requirement{{Text: "The game gives no way to learn it"}}))
		})

		It("asks nothing for a move with no condition", func() {
			Expect(domain.UnlockRule{}.Requirements(monster.Monster{}, names)).To(BeEmpty())
		})
	})
})
