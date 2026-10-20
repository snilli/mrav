package profileui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
)

var _ = Describe("StatColors", func() {
	It("has a colour for each stat, none the same as another, in the colours of the breeder's sheet", func() {
		Expect(StatColors[monster.Life]).To(Equal(color.NRGBA{R: 254, G: 254, B: 145, A: 255}), "yellow")
		Expect(StatColors[monster.Pow]).To(Equal(color.NRGBA{R: 238, G: 134, B: 131, A: 255}), "red")
		Expect(StatColors[monster.Int]).To(Equal(color.NRGBA{R: 130, G: 209, B: 114, A: 255}), "green")
		Expect(StatColors[monster.Skill]).To(Equal(color.NRGBA{R: 238, G: 134, B: 248, A: 255}), "magenta")
		Expect(StatColors[monster.Speed]).To(Equal(color.NRGBA{R: 140, G: 230, B: 232, A: 255}), "cyan")
		Expect(StatColors[monster.Defense]).To(Equal(color.NRGBA{R: 127, G: 127, B: 246, A: 255}), "blue")
	})
})

var _ = Describe("ColorText", func() {
	BeforeEach(func() { test.NewTempApp(GinkgoTB()) })

	It("is its segments one after the other, each in its colour, and in the colour of the text where it has none", func() {
		c := NewColorText(Segment{Text: "POW", Color: StatColors[monster.Pow]}, Segment{Text: " (2/3/4/5)"})

		Expect(c.String()).To(Equal("POW (2/3/4/5)"))
		Expect(c.texts[0].Color).To(Equal(StatColors[monster.Pow]))
		Expect(c.texts[1].Color).To(Equal(theme.Color(theme.ColorNameForeground)))
	})

	It("lays the segments side by side, in the room that its padding leaves", func() {
		c := NewColorText(Segment{Text: "POW"}, Segment{Text: " (2/3/4/5)"})
		c.Resize(c.MinSize())

		pad := theme.InnerPadding()
		first := c.texts[0].MinSize().Width
		Expect(c.texts[0].Position()).To(Equal(fyne.NewPos(pad, pad)))
		Expect(c.texts[1].Position()).To(Equal(fyne.NewPos(pad+first, pad)))
		Expect(c.MinSize().Width).To(Equal(2*pad + first + c.texts[1].MinSize().Width))
	})

	It("takes new segments, and the width follows them", func() {
		c := NewColorText(Segment{Text: "POW"})
		narrow := c.MinSize().Width

		c.Set(Segment{Text: "POW"}, Segment{Text: " / "}, Segment{Text: "LIF"})

		Expect(c.String()).To(Equal("POW / LIF"))
		Expect(c.MinSize().Width).To(BeNumerically(">", narrow))
	})
})
