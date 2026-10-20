package moverepo

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"orm/stored"

	monster "app/internal/core/monster/domain"
	movedomain "app/internal/core/move/domain"
)

var _ = Describe("domainConditions", func() {
	DescribeTable("reads each stored shape",
		func(in stored.Condition, want movedomain.Condition) {
			got, err := domainConditions([]stored.Condition{in})
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal([]movedomain.Condition{want}))
		},
		Entry("a stat sum", stored.Condition{Type: stored.StatSum, Stats: []int{1, 2}, Min: new(400)},
			movedomain.StatSumAtLeast{Stats: []monster.Stat{monster.Pow, monster.Int}, Min: 400}),
		Entry("a move count", stored.Condition{Type: stored.MoveUsed, Slot: new(12), Min: new(50)},
			movedomain.MoveUsedAtLeast{Slot: 12, Min: 50}),
		Entry("a good nature", stored.Condition{Type: stored.NatureAtLeast, Min: new(20)},
			movedomain.NatureAtLeast{Min: 20}),
		Entry("a bad nature", stored.Condition{Type: stored.NatureAtMost, Max: new(-50)},
			movedomain.NatureAtMost{Max: -50}),
		Entry("sub-genera that qualify", stored.Condition{Type: stored.SubGenusIn, IDs: []int{1, 11}},
			movedomain.SubGenusIn{IDs: []monster.Genus{1, 11}}),
		Entry("sub-genera that do not", stored.Condition{Type: stored.SubGenusNotIn, IDs: []int{24}},
			movedomain.SubGenusNotIn{IDs: []monster.Genus{24}}),
	)

	DescribeTable("rejects what it cannot read, naming it",
		func(in stored.Condition, want string) {
			_, err := domainConditions([]stored.Condition{in})
			Expect(err).To(MatchError(ContainSubstring(want)))
		},
		Entry("an unknown type", stored.Condition{Type: "moon_phase"}, "moon_phase"),
		Entry("a stat sum without its minimum", stored.Condition{Type: stored.StatSum, Stats: []int{1}}, "min"),
		Entry("a move count without its slot", stored.Condition{Type: stored.MoveUsed, Min: new(1)}, "slot"),
		Entry("a nature bound without its value", stored.Condition{Type: stored.NatureAtMost}, "max"),
	)

	It("reads no conditions as nil, the unconditional rule the importer started from", func() {
		got, err := domainConditions([]stored.Condition{})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(BeNil())
	})
})

var _ = Describe("domainAcquire", func() {
	It("reads every field of an entry", func() {
		got, err := domainAcquire([]stored.Acquire{{
			Place: 4, Priority: 2, Auto: 3, Nature: -20,
			Gates:    []stored.Gate{{Stats: []int{1, 3}, Min: 250, Max: 400}},
			Chain:    &stored.Chain{Slot: 7, Uses: 50},
			SubIn:    []int{16},
			SubNotIn: []int{0, 26},
		}})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal([]movedomain.Acquire{{
			Place: 4, Priority: 2, Auto: 3, Nature: -20,
			Gates:    []movedomain.Gate{{Stats: []monster.Stat{monster.Pow, monster.Skill}, Min: 250, Max: 400}},
			Chain:    &movedomain.Chain{Slot: 7, Uses: 50},
			SubIn:    []monster.Genus{16},
			SubNotIn: []monster.Genus{0, 26},
		}}))
	})

	It("reads an entry of nothing but a place as one that asks for nothing", func() {
		got, err := domainAcquire([]stored.Acquire{{Place: 1, Priority: 1}})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal([]movedomain.Acquire{{Place: 1, Priority: 1}}))
	})

	DescribeTable("rejects what it cannot read, naming it",
		func(in stored.Acquire, want string) {
			_, err := domainAcquire([]stored.Acquire{in})
			Expect(err).To(MatchError(ContainSubstring(want)))
		},
		Entry("a place before the first", stored.Acquire{Place: 0}, "place 0"),
		Entry("a place after the last", stored.Acquire{Place: 6}, "place 6"),
		Entry("a gate whose top is not above its bottom",
			stored.Acquire{Place: 1, Gates: []stored.Gate{{Stats: []int{1}, Min: 250, Max: 250}}}, "gate"),
	)

	It("reads no entries as nil", func() {
		got, err := domainAcquire(nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(BeNil())
	})
})
