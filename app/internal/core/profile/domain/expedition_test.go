package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	"app/internal/core/profile/domain"
)

var _ = Describe("Expeditions", func() {
	It("are four, in the order of the calendar, each in the first week of its month", func() {
		var names []string
		last := 0
		for _, e := range domain.Expeditions {
			names = append(names, e.Name)
			Expect(e.Week).To(Equal(1))
			Expect(e.Month).To(BeNumerically(">=", last))
			last = e.Month
		}
		Expect(names).To(Equal([]string{"Torles Mountains", "Phoenix Volcano Adventure", "Kawrea Volcano", "Parepare Jungle"}))
	})

	It("have one that comes first, once, and unlocks the others", func() {
		first := 0
		for _, e := range domain.Expeditions {
			if e.First {
				first++
				Expect(e.Name).To(Equal("Phoenix Volcano Adventure"))
			}
		}
		Expect(first).To(Equal(1))
	})

	It("ask a monster for life, fame and a grade, and say where it stands on each", func() {
		torles := domain.Expeditions[0] // Life 210, Fame 50, grade C
		m := monster.Monster{Stats: [monster.StatCount]int{monster.Life: 305}, Fame: 40, Rank: 3}
		Expect(torles.Conditions(m)).To(Equal([]domain.Condition{
			{Text: "LIF 210 or more", Have: "305", Met: true},
			{Text: "Fame 50 or more", Have: "40", Met: false},
			{Text: "Grade C or higher", Have: "B", Met: true},
		}))
	})

	It("take the exact amount as enough, and a lower grade as not", func() {
		kawrea := domain.Expeditions[2] // Life 280, Fame 50, grade B
		m := monster.Monster{Stats: [monster.StatCount]int{monster.Life: 280}, Fame: 50, Rank: 2}
		conditions := kawrea.Conditions(m)
		Expect(conditions[0].Met).To(BeTrue())
		Expect(conditions[1].Met).To(BeTrue())
		Expect(conditions[2].Met).To(BeFalse()) // grade C, not B
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

	Describe("Loot", func() {
		loot := domain.Loot{
			Building: "Cave", Stat: monster.Int, Mins: []int{0, 100, 200, 300},
			Bad: []domain.Chance{
				{Item: "Gem", Price: 1000, Pct: []float64{10, 20, 30, 40}},
				{Item: "Nuts", Price: 100, Pct: []float64{90, 80, 70, 60}},
			},
			Good: []domain.Chance{{Item: "Gem", Price: 1000, Pct: []float64{50, 50, 50, 50}}},
		}

		DescribeTable("is worth what each item is worth by its chance, at the tier of the stat that the search is by",
			func(stat int, good bool, want float64) {
				Expect(loot.Worth(stat, good)).To(BeNumerically("~", want, 1e-9))
			},
			Entry("the first tier", 99, false, 190.0),
			Entry("the second, where it starts", 100, false, 280.0),
			Entry("the second, to its end", 199, false, 280.0),
			Entry("the last tier, as high as the stat goes", 999, false, 460.0),
			Entry("a good nature has its own table", 500, true, 500.0),
		)

		It("has a stat that it is searched by", func() {
			Expect(loot.Stat).To(Equal(monster.Int))
		})
	})
})
