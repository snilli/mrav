package moverepo

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"orm/stored"

	monster "app/internal/core/monster/domain"
	movedomain "app/internal/core/move/domain"
)

func ptr[T any](v T) *T { return &v }

var _ = Describe("domainConditions", func() {
	DescribeTable("reads each stored shape",
		func(in stored.Condition, want movedomain.Condition) {
			got, err := domainConditions([]stored.Condition{in})
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal([]movedomain.Condition{want}))
		},
		Entry("a stat sum", stored.Condition{Type: stored.StatSum, Stats: []int{1, 2}, Min: ptr(400)},
			movedomain.StatSumAtLeast{Stats: []monster.Stat{monster.Pow, monster.Int}, Min: 400}),
		Entry("a move count", stored.Condition{Type: stored.MoveUsed, Slot: ptr(12), Min: ptr(50)},
			movedomain.MoveUsedAtLeast{Slot: 12, Min: 50}),
		Entry("a good nature", stored.Condition{Type: stored.NatureAtLeast, Min: ptr(20)},
			movedomain.NatureAtLeast{Min: 20}),
		Entry("a bad nature", stored.Condition{Type: stored.NatureAtMost, Max: ptr(-50)},
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
		Entry("a move count without its slot", stored.Condition{Type: stored.MoveUsed, Min: ptr(1)}, "slot"),
		Entry("a nature bound without its value", stored.Condition{Type: stored.NatureAtMost}, "max"),
	)

	It("reads no conditions as nil, the unconditional rule the importer started from", func() {
		got, err := domainConditions([]stored.Condition{})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(BeNil())
	})
})
