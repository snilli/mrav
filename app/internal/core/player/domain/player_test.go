package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/player/domain"
)

var _ = Describe("Player", func() {
	It("has an errantry sale ready only when the countdown reads 255", func() {
		Expect(domain.Player{ErrantryCD: 255}.ErrantrySaleReady()).To(BeTrue())
		Expect(domain.Player{ErrantryCD: 254}.ErrantrySaleReady()).To(BeFalse())
		Expect(domain.Player{}.ErrantrySaleReady()).To(BeFalse())
	})

	It("says which buttons are held by their bits, in the order of the names", func() {
		held := domain.Player{Input: 16 | 32 | 4096}.Held()
		for b, name := range domain.ButtonNames {
			Expect(held[b]).To(Equal(name == "Circle" || name == "Cross" || name == "Up"), name)
		}
		Expect(domain.Player{}.Held()).To(Equal([domain.ButtonCount]bool{}))
		Expect(domain.Player{Input: 0xFFFF}.Held()[domain.ButtonCount-1]).To(BeTrue())
	})

	It("names the buttons the way MR2AV orders them", func() {
		Expect(domain.ButtonNames[0]).To(Equal("L2"))
		Expect(domain.ButtonNames[4]).To(Equal("Circle"))
		Expect(domain.ButtonNames[11]).To(Equal("Start"))
		Expect(domain.ButtonNames[15]).To(Equal("Left"))
	})
})
