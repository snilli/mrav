package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	"app/internal/core/profile/domain"
)

var _ = Describe("ErrantryChance", func() {
	DescribeTable("is the chance that the calculator of legendcup.com gives",
		func(statGain, lifeGain, fatigue, fear int, byStage map[int][4]int) {
			for stage, rolls := range byStage {
				for week := 1; week <= 4; week++ {
					got := domain.ErrantryChance(stage, week, statGain, lifeGain, fatigue, fear)
					Expect(got).To(BeNumerically("~", float64(rolls[week-1])/29*100, 1e-9), "stage %d, week %d", stage, week)
				}
			}
		},
		Entry("with all in the middle", 3, 3, 0, 0, map[int][4]int{
			1: {29, 20, 6, 0}, 3: {29, 28, 14, 1}, 5: {29, 29, 22, 9}, 10: {29, 29, 29, 29}}),
		Entry("with the lowest stat gain", 1, 3, 0, 0, map[int][4]int{
			1: {29, 16, 2, 0}, 3: {29, 24, 10, 0}, 5: {29, 29, 18, 5}, 10: {29, 29, 29, 25}}),
		Entry("with a stat gain of 2", 2, 3, 0, 0, map[int][4]int{1: {29, 18, 4, 0}, 3: {29, 26, 12, 0}}),
		Entry("with a stat gain of 4, which is no better than 3", 4, 3, 0, 0, map[int][4]int{1: {29, 20, 6, 0}, 3: {29, 28, 14, 1}}),
		Entry("with the highest stat gain", 5, 3, 0, 0, map[int][4]int{
			1: {29, 22, 8, 0}, 3: {29, 29, 16, 3}, 5: {29, 29, 24, 11}}),
		Entry("with the lowest life gain", 3, 1, 0, 0, map[int][4]int{1: {29, 18, 4, 0}, 3: {29, 26, 12, 0}, 10: {29, 29, 29, 27}}),
		Entry("with the highest life gain, which is no better than 3", 3, 5, 0, 0, map[int][4]int{1: {29, 20, 6, 0}, 5: {29, 29, 22, 9}}),
		Entry("with 20 fatigue", 3, 3, 20, 0, map[int][4]int{1: {29, 16, 2, 0}, 3: {29, 24, 10, 0}, 10: {29, 29, 29, 25}}),
		Entry("with 50 fatigue", 3, 3, 50, 0, map[int][4]int{1: {23, 10, 0, 0}, 3: {29, 18, 4, 0}, 10: {29, 29, 29, 19}}),
		Entry("with 20 fear", 3, 3, 0, 20, map[int][4]int{1: {29, 24, 10, 0}, 3: {29, 29, 18, 5}, 5: {29, 29, 26, 13}}),
		Entry("with 50 fear", 3, 3, 0, 50, map[int][4]int{1: {29, 29, 16, 3}, 3: {29, 29, 24, 11}, 5: {29, 29, 29, 19}}),
		Entry("with all the fear", 3, 3, 0, 100, map[int][4]int{1: {29, 29, 26, 13}, 3: {29, 29, 29, 21}}),
	)

	It("is never less than none of the rolls or more than all of them", func() {
		for stage := 1; stage <= 10; stage++ {
			for week := 1; week <= 4; week++ {
				for _, fatigue := range []int{0, 100} {
					for _, fear := range []int{0, 100} {
						got := domain.ErrantryChance(stage, week, 1, 1, fatigue, fear)
						Expect(got).To(BeNumerically(">=", 0))
						Expect(got).To(BeNumerically("<=", 100))
					}
				}
			}
		}
	})
})

var _ = Describe("ErrantryGainOf", func() {
	DescribeTable("is what the page's table says an errantry gives a stat, and the Life that goes with it",
		func(stage, level, min, max, life int) {
			Expect(domain.ErrantryGainOf(stage, level)).To(Equal(domain.ErrantryGain{Min: min, Max: max, Life: life}))
		},
		Entry("the least, at the first stage", 1, 1, 2, 4, 3),
		Entry("the most, at the first stage", 1, 5, 3, 7, 3),
		Entry("in the middle, at the fifth stage", 5, 3, 8, 12, 10),
		Entry("the most, at the fifth stage, where the game stops at 15", 5, 5, 13, 15, 10),
		Entry("at the sixth stage", 6, 4, 9, 13, 9),
		Entry("at the seventh stage", 7, 1, 3, 7, 8),
		Entry("the least, at the last stage", 10, 1, 2, 4, 3),
		Entry("at the last stage", 10, 2, 2, 5, 3),
		Entry("at the third stage, like the fourth", 3, 4, 6, 10, 7),
		Entry("at the fourth stage, like the third", 4, 4, 6, 10, 7),
	)
})

var _ = Describe("ErrantryLocales", func() {
	It("lists the five places in the order of the game, as a copy that is safe to change", func() {
		places := domain.ErrantryLocales()
		Expect(places).To(HaveLen(5))
		Expect(places[0]).To(Equal(domain.ErrantryLocale{Name: "Papas Mountain", Stat: monster.Speed}))

		places[0].Name = "Changed"
		Expect(domain.ErrantryLocales()[0].Name).To(Equal("Papas Mountain"))
	})
})

var _ = Describe("PlanErrantry", func() {
	m := monster.Monster{LifeStage: 4, Fatigue: 30, Fear: 40, Growth: [monster.StatCount]int{1, 2, 3, 4, 0, 2}}

	It("has the five places of an errantry, in the order of the game, each with the stat that it trains", func() {
		var names []string
		var stats []monster.Stat
		for _, o := range domain.PlanErrantry(m) {
			names = append(names, o.Locale.Name)
			stats = append(stats, o.Locale.Stat)
		}
		Expect(names).To(Equal([]string{"Papas Mountain", "Mandy Desert", "Parepare Jungle", "Torble Sea", "Kawrea Volcano"}))
		Expect(stats).To(Equal([]monster.Stat{monster.Speed, monster.Pow, monster.Int, monster.Skill, monster.Defense}))
	})

	It("takes the stat gain of each place from the growth rate of its stat, a rate of 0 being the gain 1", func() {
		var gains []int
		for _, o := range domain.PlanErrantry(m) {
			gains = append(gains, o.Level)
		}
		Expect(gains).To(Equal([]int{1, 3, 4, 5, 3}))
	})

	It("says what each place trains the monster in, for its life stage and the gain of the place's stat", func() {
		for _, o := range domain.PlanErrantry(m) {
			Expect(o.Gains).To(Equal(domain.ErrantryGainOf(5, o.Level)), o.Locale.Name)
		}
	})

	It("works out the chance of each week for the life stage, the fatigue and the fear that the monster has now", func() {
		for _, o := range domain.PlanErrantry(m) {
			for week := 1; week <= domain.ErrantryWeeks; week++ {
				want := domain.ErrantryChance(5, week, o.Level, 2, 30, 40)
				Expect(o.Chances[week-1]).To(Equal(want), "%s, week %d", o.Locale.Name, week)
			}
		}
	})

	It("keeps what a monster has to what the calculator takes: a rate past A is the gain 5, a stage past the tenth the tenth, a fatigue and a fear between 0 and 100", func() {
		extreme := monster.Monster{LifeStage: 14, Fatigue: 130, Fear: -5, Growth: [monster.StatCount]int{9, 9, 9, 9, 9, 9}}
		got := domain.PlanErrantry(extreme)

		for _, o := range got {
			Expect(o.Level).To(Equal(5))
			Expect(o.Chances[0]).To(Equal(domain.ErrantryChance(10, 1, 5, 5, 100, 0)))
		}
	})
})
