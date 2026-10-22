package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/move/domain"
)

var _ = Describe("Rate", func() {
	DescribeTable("says a raw value in the game's words, as GenerateStatValue does",
		func(scale domain.Scale, raw int, want string) {
			Expect(domain.Rate(scale, raw)).To(Equal(want))
		},
		Entry("force: zero is no value", domain.ScaleForce, 0, "---"),
		Entry("force: top of E", domain.ScaleForce, 9, "E (9)"),
		Entry("force: bottom of D", domain.ScaleForce, 10, "D (10)"),
		Entry("force: top of D", domain.ScaleForce, 19, "D (19)"),
		Entry("force: C", domain.ScaleForce, 20, "C (20)"),
		Entry("force: B", domain.ScaleForce, 30, "B (30)"),
		Entry("force: top of A", domain.ScaleForce, 49, "A (49)"),
		Entry("force: S", domain.ScaleForce, 50, "S (50)"),
		Entry("force: a negative value falls to S, as the game does", domain.ScaleForce, -3, "S (-3)"),

		Entry("hit: the middle of C is zero", domain.ScaleHit, 0, "C (0)"),
		Entry("hit: bottom of C", domain.ScaleHit, -4, "C (-4)"),
		Entry("hit: top of D", domain.ScaleHit, -5, "D (-5)"),
		Entry("hit: bottom of D", domain.ScaleHit, -14, "D (-14)"),
		Entry("hit: E", domain.ScaleHit, -15, "E (-15)"),
		Entry("hit: bottom of B", domain.ScaleHit, 1, "B (1)"),
		Entry("hit: top of B", domain.ScaleHit, 4, "B (4)"),
		Entry("hit: bottom of A", domain.ScaleHit, 5, "A (5)"),
		Entry("hit: top of A", domain.ScaleHit, 14, "A (14)"),
		Entry("hit: S", domain.ScaleHit, 15, "S (15)"),
		Entry("hit: the highest value the game keeps is S", domain.ScaleHit, 156, "S (156)"),
		Entry("hit: a byte read from memory wraps to a negative", domain.ScaleHit, 200, "E (200)"),

		Entry("withering: zero is no value", domain.ScaleWither, 0, "---"),
		Entry("withering: B", domain.ScaleWither, 39, "B (39)"),

		Entry("sharpness: zero is no value", domain.ScaleCrit, 0, "---"),
		Entry("sharpness: top of E", domain.ScaleCrit, 9, "E (9)"),
		Entry("sharpness: D", domain.ScaleCrit, 10, "D (10)"),
		Entry("sharpness: top of D", domain.ScaleCrit, 14, "D (14)"),
		Entry("sharpness: C", domain.ScaleCrit, 15, "C (15)"),
		Entry("sharpness: B", domain.ScaleCrit, 24, "B (24)"),
		Entry("sharpness: A", domain.ScaleCrit, 29, "A (29)"),
		Entry("sharpness: S", domain.ScaleCrit, 30, "S (30)"),

		Entry("an unknown scale is read as force, as the game's default is", domain.Scale(9), 27, "C (27)"),
	)

	DescribeTable("agrees with the grades the source's author wrote by hand",
		func(scale domain.Scale, raw int, want string) {
			Expect(domain.Rate(scale, raw)).To(Equal(want))
		},
		Entry("damage 27", domain.ScaleForce, 27, "C (27)"),
		Entry("hit -8", domain.ScaleHit, -8, "D (-8)"),
		Entry("withering 39", domain.ScaleWither, 39, "B (39)"),
		Entry("sharpness 5", domain.ScaleCrit, 5, "E (5)"),
	)
})
