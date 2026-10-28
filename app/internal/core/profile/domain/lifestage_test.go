package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/profile/domain"
)

var _ = Describe("life stages", func() {
	DescribeTable("StageStarts is the weeks of lifespan left when each of the ten stages begins, from the share of each stage, as the calculator of LegendCup shows them for a lifespan of 500",
		func(shares, want [10]int) {
			Expect(domain.StageStarts(shares, 500)).To(Equal(want))
		},
		Entry("type 1, Normal", [10]int{10, 10, 15, 15, 10, 5, 10, 5, 5, 15}, [10]int{500, 450, 400, 325, 250, 200, 175, 125, 100, 75}),
		Entry("type 2, Precocious", [10]int{5, 5, 10, 10, 5, 5, 15, 15, 10, 20}, [10]int{500, 475, 450, 400, 350, 325, 300, 225, 150, 100}),
		Entry("type 3, Late Bloom", [10]int{15, 15, 15, 15, 10, 10, 5, 5, 5, 5}, [10]int{500, 425, 350, 275, 200, 150, 100, 75, 50, 25}),
		Entry("type 4, Sustainable", [10]int{5, 5, 10, 10, 10, 5, 15, 15, 10, 15}, [10]int{500, 475, 450, 400, 350, 300, 275, 200, 125, 75}),
	)

	It("cuts each start down to a whole week: a Normal monster of 98 weeks leaves its first stage at 88, as the item list says of Prince Hare", func() {
		starts := domain.StageStarts([10]int{10, 10, 15, 15, 10, 5, 10, 5, 5, 15}, 98)
		Expect(starts[0]).To(Equal(98))
		Expect(starts[1]).To(Equal(88))
		Expect(starts[3]).To(Equal(63), "98 * 65 / 100 is 63.7")
	})
})
