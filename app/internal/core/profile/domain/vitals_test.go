package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/profile/domain"
)

var _ = Describe("vitals", func() {
	DescribeTable("EffectiveDefense is the defense that the form changes, cut to a whole number, between 1 and 999",
		func(def, form, want int) { Expect(domain.EffectiveDefense(def, form)).To(Equal(want)) },
		Entry("a slim monster loses a little", 333, -10, 324),
		Entry("a fat monster gains", 400, 10, 410),
		Entry("no form changes nothing", 100, 0, 100),
		Entry("it is never less than 1", 0, 0, 1),
		Entry("and never more than 999", 999, 100, 999),
	)

	DescribeTable("EffectiveSpeed is the speed that the form changes the other way",
		func(spd, form, want int) { Expect(domain.EffectiveSpeed(spd, form)).To(Equal(want)) },
		Entry("a slim monster is quicker", 333, -10, 341),
		Entry("a fat monster is slower", 400, 10, 390),
		Entry("it is never less than 1", 0, 0, 1),
		Entry("and never more than 999", 1000, -100, 999),
	)

	DescribeTable("Loyalty halves spoil and fear and adds them",
		func(spoil, fear, want int) { Expect(domain.Loyalty(spoil, fear)).To(Equal(want)) },
		Entry("both even", 70, 80, 75),
		Entry("both 79 is 78", 79, 79, 78),
		Entry("one 79 and one 80 is 79", 79, 80, 79),
		Entry("nothing", 0, 0, 0),
	)

	DescribeTable("LifeIndexOf adds fatigue and twice the stress, and says how many weeks of life that costs",
		func(fatigue, stress, value, weeks int) {
			Expect(domain.LifeIndexOf(fatigue, stress)).To(Equal(domain.LifeIndex{Value: value, WeeksLost: weeks}))
		},
		Entry("nothing", 0, 0, 0, 0),
		Entry("69 still costs nothing", 69, 0, 69, 0),
		Entry("70 costs a week", 70, 0, 70, 1),
		Entry("104", 104, 0, 104, 1),
		Entry("105 costs two", 105, 0, 105, 2),
		Entry("139", 139, 0, 139, 2),
		Entry("140 costs three", 140, 0, 140, 3),
		Entry("174", 174, 0, 174, 3),
		Entry("175 costs four", 175, 0, 175, 4),
		Entry("209", 209, 0, 209, 4),
		Entry("210 costs five", 210, 0, 210, 5),
		Entry("244", 244, 0, 244, 5),
		Entry("245 costs six", 245, 0, 245, 6),
		Entry("279", 279, 0, 279, 6),
		Entry("280 costs seven, and no more", 280, 0, 280, 7),
		Entry("stress counts twice", 40, 60, 160, 3),
	)

	DescribeTable("AgeText says an age in weeks as years, months and weeks of a 48-week year",
		func(weeks int, want string) { Expect(domain.AgeText(weeks)).To(Equal(want)) },
		Entry("no age", -1, "---"),
		Entry("newborn", 0, "0y,0m,0w"),
		Entry("100 weeks", 100, "2y,1m,0w"),
		Entry("the last week of a year", 191, "3y,11m,3w"),
	)

	DescribeTable("GutsText shows the guts rate, or the guts a second",
		func(guts int, perSecond bool, want string) { Expect(domain.GutsText(guts, perSecond)).To(Equal(want)) },
		Entry("as the game keeps it", 7, false, "7"),
		Entry("a second: 30 over the rate", 7, true, "4.29"),
		Entry("a round one", 12, true, "2.50"),
		Entry("a rate of 50 or more is one guts every 30 seconds", 50, false, "50 (1)"),
		Entry("which is 30 guts a second", 50, true, "30.00"),
		Entry("no rate gives no guts", 0, true, "0.00"),
	)
})
