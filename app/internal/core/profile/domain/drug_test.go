package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/profile/domain"
)

var _ = Describe("drugs", func() {
	DescribeTable("IsDrug says which item ids are drugs",
		func(id int, want bool) { Expect(domain.IsDrug(id)).To(Equal(want)) },
		Entry("33 is not", 33, false),
		Entry("34 is the first", 34, true),
		Entry("41 is the last", 41, true),
		Entry("42 is not", 42, false),
	)

	DescribeTable("DrugTip explains the drugs that have a tip",
		func(id int, name, effect string) {
			tip := domain.DrugTip(id)
			Expect(tip).To(HavePrefix(name + ":"))
			Expect(tip).To(ContainSubstring(effect))
		},
		Entry("Teromeann", 34, "Teromeann", "Increases SPD by 100%, and POW by 50%"),
		Entry("Kasseitan", 37, "Kasseitan", "Increases SPD by 50%, and POW by 50%"),
		Entry("Troron", 38, "Troron", "-6 Weeks of lifespan on use."),
		Entry("Nageel", 39, "Nageel", "Increases SKI by 100%, and DEF by 50%"),
		Entry("Torokachin FX", 40, "Torokachin FX", "-12 Weeks of lifespan on use."),
		Entry("Paradoxine", 41, "Paradoxine", "decreases SPD and DEF by 10%"),
	)

	It("has no tip for the drugs it has nothing to say about, nor for what is no drug", func() {
		Expect(domain.DrugTip(35)).To(BeEmpty())
		Expect(domain.DrugTip(36)).To(BeEmpty())
		Expect(domain.DrugTip(0)).To(BeEmpty())
	})
})
