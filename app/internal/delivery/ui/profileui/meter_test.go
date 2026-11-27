package profileui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Meter", func() {
	BeforeEach(func() { test.NewTempApp(GinkgoTB()) })

	sized := func(m *Meter) *Meter {
		test.WidgetRenderer(m)
		m.Resize(fyne.NewSize(100, 10))
		return m
	}

	It("fills in proportion to the value, in the colour of the tone", func() {
		m := sized(NewMeter())
		m.Set(0.5, ToneLimeGreen)
		Expect(m.fill.Size().Width).To(BeNumerically("~", 50, 0.01))
		Expect(m.fill.Position().X).To(BeZero())
		Expect(m.fill.FillColor).To(Equal(toneColors[ToneLimeGreen]))
	})

	It("never fills past either end", func() {
		m := sized(NewMeter())
		m.Set(1.7, ToneRed)
		Expect(m.fill.Size().Width).To(BeNumerically("~", 100, 0.01))
		m.Set(-0.3, ToneRed)
		Expect(m.fill.Size().Width).To(BeZero())
	})

	It("tells how full it is, which is the ratio it was given, kept between 0 and 1", func() {
		m := sized(NewMeter())
		m.Set(0.3, ToneNone)
		Expect(m.Ratio()).To(BeNumerically("~", 0.3, 1e-9))
		m.Set(4, ToneNone)
		Expect(m.Ratio()).To(Equal(1.0))
	})

	It("fills in the accent colour when it is given no tone", func() {
		m := sized(NewMeter())
		m.Set(0.5, ToneNone)
		Expect(m.fill.FillColor).To(Equal(toneColors[ToneAccent]))
	})

	It("fills in the colour it was given, whatever the tone, until it is a gauge, which marks in white", func() {
		blue := color.NRGBA{B: 255, A: 255}
		m := sized(NewMeter().Colored(blue))
		m.Set(0.5, ToneRed)
		Expect(m.fill.FillColor).To(Equal(blue))
		m.Set(0.5, ToneNone)
		Expect(m.fill.FillColor).To(Equal(blue))

		g := sized(NewGauge().Colored(blue))
		g.Set(0.5, ToneRed)
		Expect(g.fill.FillColor).To(Equal(color.White))
	})

	It("grows with the width it is given", func() {
		m := sized(NewMeter())
		m.Set(0.25, ToneOrange)
		m.Resize(fyne.NewSize(200, 10))
		Expect(m.fill.Size().Width).To(BeNumerically("~", 50, 0.01))
	})

	It("shows a gauge as a marker at the value, which stays on the track at both ends", func() {
		g := sized(NewGauge())
		g.Set(0.5, ToneNone)
		Expect(g.fill.Size().Width).To(Equal(float32(markerWidth)))
		Expect(g.fill.Position().X).To(BeNumerically("~", 48, 0.01))

		g.Set(1, ToneNone)
		Expect(g.fill.Position().X).To(BeNumerically("~", 100-markerWidth, 0.01))
		g.Set(0, ToneNone)
		Expect(g.fill.Position().X).To(BeZero())
	})
})

var _ = Describe("Gauge span", func() {
	BeforeEach(func() { test.NewTempApp(GinkgoTB()) })

	It("shades the reach from low to high and marks the base", func() {
		g := NewGauge()
		test.WidgetRenderer(g)
		g.Resize(fyne.NewSize(100, 30))
		g.SetSpan(Span{Low: 0.25, Base: 0.5, High: 1})
		Expect(g.band.Visible()).To(BeTrue())
		Expect(g.band.Position().X).To(BeNumerically("~", 25, 0.01))
		Expect(g.band.Size().Width).To(BeNumerically("~", 75, 0.01))
		Expect(g.base.Position().X).To(BeNumerically("~", 0.5*(100-baseWidth), 0.01))
		Expect(g.base.Size().Width).To(BeNumerically("==", baseWidth))
		Expect(g.base.Size().Height).To(BeNumerically(">", barHeight))
	})

	It("shows no span until given one", func() {
		g := NewGauge()
		test.WidgetRenderer(g)
		Expect(g.band.Visible()).To(BeFalse())
		Expect(g.base.Visible()).To(BeFalse())
	})
})

var _ = Describe("Meter hint", func() {
	BeforeEach(func() { test.NewTempApp(GinkgoTB()) })

	hovered := func(hint string) *Meter {
		g := NewGauge()
		g.Hint(hint)
		w := test.NewWindow(g)
		DeferCleanup(w.Close)
		w.Resize(fyne.NewSize(200, 40))
		g.MouseIn(&desktop.MouseEvent{})
		return g
	}

	It("shows its hint while the mouse is over it", func() {
		g := hovered("Min -70 · Base 30 · Max 100")
		Expect(g.popup).NotTo(BeNil())
		Expect(g.popup.Visible()).To(BeTrue())
		Expect(g.hintText.Text).To(Equal("Min -70 · Base 30 · Max 100"))
		g.MouseOut()
		Expect(g.popup.Visible()).To(BeFalse())
	})

	It("shows nothing on hover without a hint", func() {
		g := hovered("")
		Expect(g.popup).To(BeNil())
	})
})

var _ = Describe("Gauge with a span", func() {
	BeforeEach(func() { test.NewTempApp(GinkgoTB()) })

	spanned := func(base float64) *Meter {
		g := NewGauge()
		test.WidgetRenderer(g)
		g.Resize(fyne.NewSize(100, 30))
		g.SetSpan(Span{Low: 0, Base: base, High: 1, Zero: 0.5})
		return g
	}

	DescribeTable("fills from the base to the value, red below zero and green above",
		func(base, value, redX, redW, greenX, greenW float64) {
			g := spanned(base)
			g.Set(value, ToneNone)
			Expect(g.fill.FillColor).To(Equal(toneColors[ToneRed]))
			Expect(g.plus.FillColor).To(Equal(toneColors[ToneLimeGreen]))
			Expect(g.fill.Size().Width).To(BeNumerically("~", redW, 0.01))
			Expect(g.plus.Size().Width).To(BeNumerically("~", greenW, 0.01))
			if redW > 0 {
				Expect(g.fill.Position().X).To(BeNumerically("~", redX, 0.01))
			}
			if greenW > 0 {
				Expect(g.plus.Position().X).To(BeNumerically("~", greenX, 0.01))
			}
		},
		Entry("base -30, now 9: red up to 0, green past it", 0.35, 0.545, 35.0, 15.0, 50.0, 4.5),
		Entry("base -30, now -50: all red", 0.35, 0.25, 25.0, 10.0, 0.0, 0.0),
		Entry("base 20, now 40: all green", 0.6, 0.7, 0.0, 0.0, 60.0, 10.0),
		Entry("base 20, now -10: green down to 0, red past it", 0.6, 0.45, 45.0, 5.0, 50.0, 10.0),
	)

	It("marks the base in white, over the fill", func() {
		g := spanned(0.5)
		Expect(g.base.FillColor).To(Equal(color.White))
		objs := test.WidgetRenderer(g).Objects()
		Expect(objs[len(objs)-1]).To(BeIdenticalTo(g.base))
	})
})
