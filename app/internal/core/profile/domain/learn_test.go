package domain_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	movedomain "app/internal/core/move/domain"
	"app/internal/core/profile/domain"
)

func listed(l []domain.Learn) []string {
	out := make([]string, len(l))
	for i, x := range l {
		out[i] = fmt.Sprintf("%d:%.1f", x.Move.MemorySlot, x.Chance)
	}
	return out
}

var _ = DescribeTable("Learn.Sure says whether a place teaches the move for sure: its chance is all to one decimal",
	func(chance float64, want bool) { Expect(domain.Learn{Chance: chance}.Sure()).To(Equal(want)) },
	Entry("all", 100.0, true),
	Entry("all, but for the error of the sums that left it", 99.99999, true),
	Entry("rounds up to all", 99.96, true),
	Entry("a little short", 99.9, false),
	Entry("a chance", 3.0, false),
	Entry("none", 0.0, false),
)

var _ = Describe("LearnChances", func() {
	pow := []monster.Stat{monster.Pow}
	papas := func(priority int) movedomain.Acquire {
		return movedomain.Acquire{Place: 1, Priority: priority, Auto: 3, Gates: []movedomain.Gate{{Stats: pow, Min: 100, Max: 200}}}
	}
	far := func(priority int) movedomain.Acquire {
		return movedomain.Acquire{Place: 1, Priority: priority, Auto: 3, Gates: []movedomain.Gate{{Stats: pow, Min: 400, Max: 500}}}
	}
	mv := func(slot int, a ...movedomain.Acquire) movedomain.Move {
		return movedomain.Move{MemorySlot: slot, Name: fmt.Sprintf("move %d", slot), Acquire: a}
	}
	withPow := func(p int) monster.Monster { return monster.Monster{Stats: [monster.StatCount]int{monster.Pow: p}} }
	papasOf := func(m monster.Monster, moves ...movedomain.Move) []string {
		return listed(domain.LearnChances(m, moves)[0])
	}

	Describe("the chance of one move", func() {
		DescribeTable("rises with the stat total from the auto learn share to all that is left",
			func(pow int, want string) {
				Expect(papasOf(withPow(pow), mv(1, papas(1)))).To(Equal([]string{want}))
			},
			Entry("giving the share of what is left when the total is short", 50, "1:3.0"),
			Entry("a hundred, the bottom, is still the share", 100, "1:3.0"),
			Entry("halfway, half of what is left and half of the share", 150, "1:51.5"),
			Entry("all that is left at the top", 200, "1:100.0"),
			Entry("and above it", 900, "1:100.0"),
		)

		It("counts loyalty, which is half of the spoil and of the fear, with the stats", func() {
			m := withPow(110)
			m.Spoil, m.Fear = 80, 80
			Expect(papasOf(m, mv(1, papas(1)))).To(Equal([]string{"1:90.3"}))
		})

		It("adds up only the stats of the gate", func() {
			m := monster.Monster{Stats: [monster.StatCount]int{monster.Pow: 900, monster.Int: 150}}
			a := movedomain.Acquire{Place: 1, Priority: 1, Auto: 3,
				Gates: []movedomain.Gate{{Stats: []monster.Stat{monster.Int, monster.Skill}, Min: 100, Max: 200}}}
			Expect(papasOf(m, mv(1, a))).To(Equal([]string{"1:51.5"}))
		})

		It("takes all that is left when it asks for no stats", func() {
			Expect(papasOf(withPow(0), mv(1, movedomain.Acquire{Place: 1, Priority: 1}))).To(Equal([]string{"1:100.0"}))
		})

		It("is the lowest of its gates when it has two", func() {
			a := movedomain.Acquire{Place: 1, Priority: 1, Auto: 3, Gates: []movedomain.Gate{
				{Stats: pow, Min: 100, Max: 150}, {Stats: []monster.Stat{monster.Speed}, Min: 150, Max: 250}}}
			m := monster.Monster{Stats: [monster.StatCount]int{monster.Pow: 120, monster.Speed: 190}, Spoil: 20, Fear: 20}
			Expect(papasOf(m, mv(1, a))).To(Equal([]string{"1:61.2"}))
		})
	})

	Describe("the moves of a place, which share what is left", func() {
		It("tries them by priority, whatever order they come in", func() {
			Expect(papasOf(withPow(150), mv(9, far(2)), mv(4, papas(1)))).To(Equal([]string{"4:51.5", "9:1.5"}))
		})

		It("leaves nothing for the next ones once one is sure", func() {
			Expect(papasOf(withPow(300), mv(1, papas(1)), mv(2, papas(2)))).To(Equal([]string{"1:100.0", "2:0.0"}))
		})

		It("gives the next one no more than is left", func() {
			sure := movedomain.Acquire{Place: 1, Priority: 2, Auto: 3, Gates: []movedomain.Gate{{Stats: pow, Min: 10, Max: 20}}}
			Expect(papasOf(withPow(150), mv(1, papas(1)), mv(2, sure))).To(Equal([]string{"1:51.5", "2:48.5"}))
		})

		It("starts each place with all of it", func() {
			second := movedomain.Acquire{Place: 2, Priority: 1, Auto: 3, Gates: []movedomain.Gate{{Stats: pow, Min: 10, Max: 20}}}
			got := domain.LearnChances(withPow(300), []movedomain.Move{mv(1, papas(1)), mv(2, second)})
			Expect(listed(got[0])).To(Equal([]string{"1:100.0"}))
			Expect(listed(got[1])).To(Equal([]string{"2:100.0"}))
			Expect(got[2]).To(BeEmpty())
		})

		It("lists a move at every place that teaches it", func() {
			fifth := far(1)
			fifth.Place = 5
			got := domain.LearnChances(withPow(150), []movedomain.Move{mv(1, papas(1), fifth)})
			Expect(listed(got[0])).To(Equal([]string{"1:51.5"}))
			Expect(listed(got[4])).To(Equal([]string{"1:3.0"}))
		})

		It("ignores a place that is not one of the five", func() {
			Expect(func() {
				domain.LearnChances(withPow(1), []movedomain.Move{mv(1, movedomain.Acquire{Place: 9})})
			}).NotTo(Panic())
		})
	})

	Describe("what keeps a move out of the list, and leaves what is left alone", func() {
		keptOut := func(m monster.Monster, first movedomain.Acquire) {
			Expect(papasOf(m, mv(1, first), mv(2, papas(2)))).To(Equal([]string{"2:51.5"}))
		}

		It("having learned it", func() {
			m := withPow(150)
			m.Learned[1] = true
			keptOut(m, papas(1))
		})

		It("a sub-genus that is not one it is for", func() {
			m := withPow(150)
			m.SubGenus = 3
			a := papas(1)
			a.SubIn = []monster.Genus{16, 26}
			keptOut(m, a)
		})

		It("a sub-genus that it is not for", func() {
			m := withPow(150)
			m.SubGenus = 26
			a := papas(1)
			a.SubNotIn = []monster.Genus{0, 26}
			keptOut(m, a)
		})

		It("a nature that is not good enough", func() {
			m := withPow(150)
			m.Nature = 19
			a := papas(1)
			a.Nature = 20
			keptOut(m, a)
		})

		It("a nature that is not bad enough", func() {
			m := withPow(150)
			m.Nature = -19
			a := papas(1)
			a.Nature = -20
			keptOut(m, a)
		})

		It("a move that has not been used enough", func() {
			m := withPow(150)
			m.Used[7] = 49
			a := papas(1)
			a.Chain = &movedomain.Chain{Slot: 7, Uses: 50}
			keptOut(m, a)
		})
	})

	DescribeTable("lists a move that the monster meets the rest of",
		func(change func(*monster.Monster), a movedomain.Acquire) {
			m := withPow(150)
			change(&m)
			a.Place, a.Priority, a.Auto, a.Gates = 1, 1, 3, papas(1).Gates
			Expect(papasOf(m, mv(1, a))).To(Equal([]string{"1:51.5"}))
		},
		Entry("with a sub-genus it is for", func(m *monster.Monster) { m.SubGenus = 16 },
			movedomain.Acquire{SubIn: []monster.Genus{16}}),
		Entry("with a sub-genus it is not barred from", func(m *monster.Monster) { m.SubGenus = 3 },
			movedomain.Acquire{SubNotIn: []monster.Genus{26}}),
		Entry("with the nature it asks for", func(m *monster.Monster) { m.Nature = 20 },
			movedomain.Acquire{Nature: 20}),
		Entry("with the bad nature it asks for", func(m *monster.Monster) { m.Nature = -20 },
			movedomain.Acquire{Nature: -20}),
		Entry("with a worse nature than a bad one asks for", func(m *monster.Monster) { m.Nature = -90 },
			movedomain.Acquire{Nature: -20}),
		Entry("with the move used as often as it asks", func(m *monster.Monster) { m.Used[7] = 50 },
			movedomain.Acquire{Chain: &movedomain.Chain{Slot: 7, Uses: 50}}),
	)
})
