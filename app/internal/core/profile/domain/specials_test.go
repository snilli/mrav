package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/profile/domain"
)

var _ = Describe("SpecialsOf", func() {
	It("lists nothing when no flag is set", func() {
		s := domain.SpecialsOf(0, 0, 12, 3)
		Expect(s.Names).To(BeEmpty())
		Expect(s.Invalid).To(BeFalse())
		Expect(s.Text()).To(BeEmpty())
	})

	It("lists the ordinary specials in the game's order", func() {
		s := domain.SpecialsOf(1|2|4|8|16|32|64|128|256|512, 0, 12, 3)
		Expect(s.Text()).To(Equal("Power, Anger, Grit, Will, Fight, Fury, Guard, Ease, Hurry, Vigor"))
		Expect(s.Invalid).To(BeFalse())
	})

	DescribeTable("brackets Power for a good nature and Anger for a bad one, which is when they work",
		func(flag, nature int, want string) {
			Expect(domain.SpecialsOf(flag, nature, 12, 3).Text()).To(Equal(want))
		},
		Entry("Power at nature 20", 1, 20, "[Power]"),
		Entry("Power at nature 19", 1, 19, "Power"),
		Entry("Anger at nature -20", 2, -20, "[Anger]"),
		Entry("Anger at nature -19", 2, -19, "Anger"),
	)

	DescribeTable("marks a special that this breed cannot have as invalid",
		func(flag, main, sub int, invalid bool) {
			Expect(domain.SpecialsOf(flag, 0, main, sub).Invalid).To(Equal(invalid))
		},
		Entry("Real on a Joker main", 1024, 26, 0, false),
		Entry("Real on a Joker sub", 1024, 3, 26, false),
		Entry("Real on neither", 1024, 3, 4, true),
		Entry("Drunk on a Kato main", 2048, 16, 0, false),
		Entry("Drunk on a Kato sub only", 2048, 3, 16, true),
		Entry("Unity on a Pandora main", 4096, 3, 0, false),
		Entry("Unity elsewhere", 4096, 4, 3, true),
	)

	It("calls the flags the game does not use an unused special, once, and marks them invalid", func() {
		s := domain.SpecialsOf(8192|16384|32768, 0, 12, 3)
		Expect(s.Text()).To(Equal("Unused Special"))
		Expect(s.Invalid).To(BeTrue())
	})

	It("lists a breed's own special after the ordinary ones", func() {
		s := domain.SpecialsOf(1|1024, 20, 26, 26)
		Expect(s.Text()).To(Equal("[Power], Real"))
		Expect(s.Invalid).To(BeFalse())
	})
})
