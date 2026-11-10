package profileui

import (
	"image/color"

	"fyne.io/fyne/v2/driver/desktop"
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

	It("can be tapped once it is given something to do, showing a hand and an outline while it is dim", func() {
		f := NewFlag("Ready", ToneLimeGreen)
		Expect(f.Cursor()).To(Equal(desktop.DefaultCursor))
		tapped := 0
		f.OnTapped = func() { tapped++ }
		f.Set(false)
		test.Tap(f)
		Expect(tapped).To(Equal(1))
		Expect(f.Cursor()).To(Equal(desktop.PointerCursor))
		Expect(f.bg.StrokeWidth).To(BeNumerically(">", 0))
		f.Set(true)
		Expect(f.bg.StrokeWidth).To(BeZero())
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
