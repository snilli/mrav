package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	"app/internal/core/profile/domain"
)

var _ = Describe("labels", func() {
	DescribeTable("GrowthGrade is the letter of a growth rate",
		func(rate int, want string) { Expect(domain.GrowthGrade(rate)).To(Equal(want)) },
		Entry("0 is E", 0, "E"),
		Entry("1 is D", 1, "D"),
		Entry("2 is C", 2, "C"),
		Entry("3 is B", 3, "B"),
		Entry("4 is A", 4, "A"),
		Entry("5 is S", 5, "S"),
		Entry("6 is S2", 6, "S2"),
		Entry("9 is S5", 9, "S5"),
		Entry("10 is F", 10, "F"),
		Entry("11 is not a grade", 11, "---"),
		Entry("nor is a negative", -1, "---"),
	)

	DescribeTable("ArenaGrade is the letter of the arena speed",
		func(speed int, want string) { Expect(domain.ArenaGrade(speed)).To(Equal(want)) },
		Entry("0 is E", 0, "E"),
		Entry("4 is A", 4, "A"),
		Entry("5 is not a grade", 5, "---"),
		Entry("nor is a negative", -1, "---"),
	)

	DescribeTable("NatureOf names the effective nature",
		func(nature int, want domain.NatureLabel) { Expect(domain.NatureOf(nature)).To(Equal(want)) },
		Entry("the lowest", -100, domain.Worst),
		Entry("-60 is the worst there is", -60, domain.Worst),
		Entry("-59 is bad", -59, domain.Bad),
		Entry("-20 is bad", -20, domain.Bad),
		Entry("-19 is neutral", -19, domain.Neutral),
		Entry("0 is neutral", 0, domain.Neutral),
		Entry("19 is neutral", 19, domain.Neutral),
		Entry("20 is good", 20, domain.Good),
		Entry("59 is good", 59, domain.Good),
		Entry("60 is the best there is", 60, domain.Best),
		Entry("the highest", 130, domain.Best),
	)

	DescribeTable("FormLabel describes the form",
		func(form int, want string) { Expect(domain.FormLabel(form)).To(Equal(want)) },
		Entry("-100", -100, "Skinny"),
		Entry("-60", -60, "Skinny"),
		Entry("-59", -59, "Slim"),
		Entry("-20", -20, "Slim"),
		Entry("-19", -19, "Normal"),
		Entry("19", 19, "Normal"),
		Entry("20", 20, "Fat"),
		Entry("59", 59, "Fat"),
		Entry("60", 60, "Plump"),
		Entry("100", 100, "Plump"),
		Entry("past the ends there is nothing to say", 101, "---"),
		Entry("either end", -101, "---"),
	)

	DescribeTable("MotivationOf says how eager a monster is for a drill",
		func(percent int, want domain.Level) { Expect(domain.MotivationOf(percent)).To(Equal(want)) },
		Entry("0", 0, domain.LevelLow),
		Entry("29", 29, domain.LevelLow),
		Entry("30", 30, domain.LevelMid),
		Entry("80", 80, domain.LevelMid),
		Entry("81", 81, domain.LevelHigh),
		Entry("100", 100, domain.LevelHigh),
		Entry("past 100 there is nothing to say", 101, domain.LevelNone),
	)

	DescribeTable("LifeStageText names the stage, and says when the monster is dead",
		func(stage, lifespan int, want string) { Expect(domain.LifeStageText(stage, lifespan)).To(Equal(want)) },
		Entry("the first", 0, 100, "1 - Baby"),
		Entry("the fifth", 4, 100, "5 - Prime"),
		Entry("the last", 9, 1, "10 - Twilight"),
		Entry("a stage there is not", 10, 100, "---"),
		Entry("a monster with no lifespan left is dead whatever its stage", 4, 0, "Dead :("),
	)

	DescribeTable("LifeTypeName names the aging type",
		func(lifeType int, want string) { Expect(domain.LifeTypeName(lifeType)).To(Equal(want)) },
		Entry("0", 0, "Type 1: Normal"),
		Entry("1", 1, "Type 2: Precocious"),
		Entry("2", 2, "Type 3: Late Bloom"),
		Entry("3", 3, "Type 4: Sustainable"),
		Entry("4", 4, "Type 5: Prodigy [X]"),
		Entry("5", 5, "Unrecognised"),
	)

	DescribeTable("PlayTypeName names the play",
		func(play int, want string) { Expect(domain.PlayTypeName(play)).To(Equal(want)) },
		Entry("0", 0, "Mud Fight"),
		Entry("1", 1, "Sumo Battle"),
		Entry("2", 2, "Sparring"),
		Entry("3", 3, "---"),
	)

	Describe("GenusLabel", func() {
		names := map[monster.Genus]string{0: "Pixie", 5: "Henger", 37: "Naga"}

		It("is the name of a genus, as a main breed or a sub breed", func() {
			Expect(domain.GenusLabel(5, true, names)).To(Equal("Henger"))
			Expect(domain.GenusLabel(37, false, names)).To(Equal("Naga"))
		})

		It("numbers the unknown sub breeds 38 to 43, and shows nothing for them as a main breed", func() {
			Expect(domain.GenusLabel(38, false, names)).To(Equal("??? #1"))
			Expect(domain.GenusLabel(43, false, names)).To(Equal("??? #6"))
			Expect(domain.GenusLabel(38, true, names)).To(Equal("---"))
		})

		It("shows nothing for a code that is no genus", func() {
			Expect(domain.GenusLabel(44, false, names)).To(Equal("---"))
			Expect(domain.GenusLabel(12, true, names)).To(Equal("---")) // a genus the table does not have
		})
	})
})
