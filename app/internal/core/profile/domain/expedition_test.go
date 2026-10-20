package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	"app/internal/core/profile/domain"
)

var _ = Describe("Expeditions", func() {
	It("are three, in the order of the calendar, each in the first week of its month", func() {
		var names []string
		last := 0
		for _, e := range domain.Expeditions {
			names = append(names, e.Name)
			Expect(e.Week).To(Equal(1))
			Expect(e.Month).To(BeNumerically(">=", last))
			last = e.Month
		}
		Expect(names).To(Equal([]string{"Torles Mountains", "Kawrea Volcano", "Parepare Jungle"}))
	})

	It("leave out the first expedition, Phoenix Volcano Adventure, which can be made once only and of which the game does not say whether it was", func() {
		for _, e := range domain.Expeditions {
			Expect(e.Name).NotTo(Equal("Phoenix Volcano Adventure"))
		}
	})

	It("ask a monster for life, fame and a grade, and say how much of each it has, a grade as its rank", func() {
		torles := domain.Expeditions[0] // Life 210, Fame 50, grade C
		m := monster.Monster{Stats: [monster.StatCount]int{monster.Life: 305}, Fame: 40, Rank: 3}
		Expect(torles.Conditions(m)).To(Equal([]domain.Condition{
			{Kind: domain.CondLife, Need: 210, Have: 305},
			{Kind: domain.CondFame, Need: 50, Have: 40},
			{Kind: domain.CondGrade, Need: 2, Have: 3},
		}))
	})

	It("take the exact amount as enough, and a lower grade as not", func() {
		kawrea := domain.Expeditions[1] // Life 280, Fame 50, grade B
		m := monster.Monster{Stats: [monster.StatCount]int{monster.Life: 280}, Fame: 50, Rank: 2}
		conditions := kawrea.Conditions(m)
		Expect(conditions[0].Met()).To(BeTrue())
		Expect(conditions[1].Met()).To(BeTrue())
		Expect(conditions[2].Met()).To(BeFalse()) // grade C, not B
	})

	DescribeTable("RankLetter writes a rank as the game's letter",
		func(rank int, want string) { Expect(domain.RankLetter(rank)).To(Equal(want)) },
		Entry("E", 0, "E"), Entry("D", 1, "D"), Entry("C", 2, "C"),
		Entry("B", 3, "B"), Entry("A", 4, "A"), Entry("S", 5, "S"),
		Entry("none", 6, "?"), Entry("below none", -1, "?"),
	)

	DescribeTable("StartEnergy is 70 for every 70 life, at least 140 and at most 700",
		func(life, want int) { Expect(domain.StartEnergy(life)).To(Equal(want)) },
		Entry("no life", 0, 140),
		Entry("under 140", 139, 140),
		Entry("140", 140, 140),
		Entry("209", 209, 140),
		Entry("210", 210, 210),
		Entry("279", 279, 210),
		Entry("280", 280, 280),
		Entry("349", 349, 280),
		Entry("350", 350, 350),
		Entry("699", 699, 630),
		Entry("700", 700, 700),
		Entry("999", 999, 700),
	)

	DescribeTable("BreakChance is the guide's 50% at 350 power rising to 100% at 500, and carried down to 5% below that",
		func(pow int, want float64) { Expect(domain.BreakChance(pow)).To(BeNumerically("~", want, 1e-9)) },
		Entry("at 350", 350, 0.5),
		Entry("halfway to 500", 425, 0.75),
		Entry("at 500", 500, 1.0),
		Entry("more power does not help", 900, 1.0),
		Entry("under 350 the line goes on", 275, 0.25),
		Entry("but it is never nothing", 100, 0.05),
	)

	DescribeTable("FindChance is the guide's 5% below 100 intelligence, 10% more for each 100, and 95% at 900",
		func(intelligence int, want float64) {
			Expect(domain.FindChance(intelligence)).To(BeNumerically("~", want, 1e-9))
		},
		Entry("none", 0, 0.05),
		Entry("99", 99, 0.05),
		Entry("100", 100, 0.15),
		Entry("300", 300, 0.35),
		Entry("899", 899, 0.85),
		Entry("900", 900, 0.95),
		Entry("more does not help", 1500, 0.95),
	)
})
