package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	"app/internal/core/move/domain"
)

func rule(cs ...domain.Condition) domain.UnlockRule { return domain.UnlockRule{All: cs} }

var _ = Describe("UnlockRule.Holds", func() {
	// Pow 200, Int 300, Speed 500; nature +20; sub-genus 5; move slot 3 used 30 times.
	m := monster.Monster{SubGenus: 5, Nature: 20}
	m.Stats[monster.Pow], m.Stats[monster.Int], m.Stats[monster.Speed] = 200, 300, 500
	m.Used[3] = 30

	DescribeTable("answers for the monster",
		func(r domain.UnlockRule, want bool) { Expect(r.Holds(m)).To(Equal(want)) },
		Entry("no condition: learnable from the start", domain.UnlockRule{}, true),
		Entry("never", domain.UnlockRule{Never: true}, false),
		Entry("stat sum met exactly", rule(domain.StatSumAtLeast{Stats: []monster.Stat{monster.Pow, monster.Int}, Min: 500}), true),
		Entry("stat sum one short", rule(domain.StatSumAtLeast{Stats: []monster.Stat{monster.Pow, monster.Int}, Min: 501}), false),
		Entry("move used exactly enough", rule(domain.MoveUsedAtLeast{Slot: 3, Min: 30}), true),
		Entry("move used one short", rule(domain.MoveUsedAtLeast{Slot: 3, Min: 31}), false),
		Entry("move used in a slot outside 0-23", rule(domain.MoveUsedAtLeast{Slot: 24, Min: 0}), false),
		Entry("nature at least, met", rule(domain.NatureAtLeast{Min: 20}), true),
		Entry("nature at least, short", rule(domain.NatureAtLeast{Min: 21}), false),
		Entry("nature at most, met", rule(domain.NatureAtMost{Max: 20}), true),
		Entry("nature at most, exceeded", rule(domain.NatureAtMost{Max: 19}), false),
		Entry("sub-genus in the list", rule(domain.SubGenusIn{IDs: []monster.Genus{1, 5}}), true),
		Entry("sub-genus not in the list", rule(domain.SubGenusIn{IDs: []monster.Genus{1, 2}}), false),
		Entry("sub-genus outside the exclusions", rule(domain.SubGenusNotIn{IDs: []monster.Genus{1, 2}}), true),
		Entry("sub-genus among the exclusions", rule(domain.SubGenusNotIn{IDs: []monster.Genus{5}}), false),
		Entry("every condition must hold: all true", rule(domain.NatureAtLeast{Min: 20}, domain.MoveUsedAtLeast{Slot: 3, Min: 30}), true),
		Entry("every condition must hold: one false", rule(domain.NatureAtLeast{Min: 20}, domain.MoveUsedAtLeast{Slot: 3, Min: 31}), false),
		Entry("never wins over conditions", domain.UnlockRule{Never: true, All: []domain.Condition{domain.NatureAtLeast{Min: 0}}}, false),
	)
})
