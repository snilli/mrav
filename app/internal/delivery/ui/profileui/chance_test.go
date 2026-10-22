package profileui

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = DescribeTable("ChanceText says the chance of a move as the calculator does, to one decimal and without a needless .0",
	func(percent float64, want string) { Expect(ChanceText(percent)).To(Equal(want)) },
	Entry("all", 100.0, "100%"),
	Entry("all, but for the error of the sums that left it", 99.99999, "100%"),
	Entry("a whole number", 3.0, "3%"),
	Entry("with a decimal", 41.84, "41.8%"),
	Entry("rounded up", 41.86, "41.9%"),
	Entry("under one", 0.5, "0.5%"),
)
