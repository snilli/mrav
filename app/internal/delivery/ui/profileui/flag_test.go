package profileui

import (
	"image/color"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Flag", func() {
	BeforeEach(func() { test.NewTempApp(GinkgoTB()) })

	It("starts dim, with no fill, as something that is not so", func() {
		f := NewFlag("Gold Peach", ToneOrange)
		Expect(f.On()).To(BeFalse())
		Expect(f.bg.FillColor).To(Equal(toneNone))
		Expect(f.txt.Color).To(Equal(theme.Color(theme.ColorNameDisabled)))
		Expect(f.txt.Text).To(Equal("Gold Peach"))
	})

	It("lights up in its tone, with text that reads on it, when it is so", func() {
		f := NewFlag("Gold Peach", ToneOrange)
		f.Set(true)
		Expect(f.On()).To(BeTrue())
		Expect(f.bg.FillColor).To(Equal(toneColors[ToneOrange]))
		Expect(f.txt.Color).To(Equal(color.Black))

		dark := NewFlag("Silver Peach", ToneGray)
		dark.Set(true)
		Expect(dark.txt.Color).To(Equal(color.White))
	})

	It("can be given other words, which it keeps whether it is lit or not", func() {
		f := NewFlag("", ToneLimeGreen)
		f.SetText("Dragon")
		f.Set(true)
		Expect(f.txt.Text).To(Equal("Dragon"))
		f.SetText("Pixie")
		Expect(f.txt.Text).To(Equal("Pixie"))
		Expect(f.On()).To(BeTrue())
	})

	It("goes dim again when it stops being so", func() {
		f := NewFlag("Item Used", ToneLimeGreen)
		f.Set(true)
		f.Set(false)
		Expect(f.On()).To(BeFalse())
		Expect(f.bg.FillColor).To(Equal(toneNone))
		Expect(f.txt.Color).To(Equal(theme.Color(theme.ColorNameDisabled)))
	})
})
