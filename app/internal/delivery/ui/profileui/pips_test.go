package profileui

import (
	"image/color"

	"fyne.io/fyne/v2/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pips", func() {
	BeforeEach(func() { test.NewTempApp(GinkgoTB()) })

	It("is one dot for each step, the ones done lit", func() {
		p := NewPips()
		p.Set(2, 4)

		Expect(p.Done()).To(Equal(2))
		Expect(p.Total()).To(Equal(4))
		Expect(p.dots).To(HaveLen(4))
		for i, dot := range p.dots {
			if i < 2 {
				Expect(dot.FillColor).To(Equal(toneColors[ToneAccent]), "dot %d is lit", i)
			} else {
				Expect(dot.FillColor).To(Equal(color.Transparent), "dot %d is not", i)
			}
		}
	})

	It("changes the number of dots with the total, and lights none for nothing done", func() {
		p := NewPips()
		p.Set(1, 4)
		p.Set(0, 2)

		Expect(p.dots).To(HaveLen(2))
		Expect(p.Done()).To(Equal(0))
		for _, dot := range p.dots {
			Expect(dot.FillColor).To(Equal(color.Transparent))
		}
	})

	It("keeps what is done to the total, and to nothing below", func() {
		p := NewPips()
		p.Set(9, 3)
		Expect(p.Done()).To(Equal(3))
		p.Set(-1, 3)
		Expect(p.Done()).To(Equal(0))
	})
})
