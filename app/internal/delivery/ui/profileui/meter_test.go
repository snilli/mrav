package profileui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Meter", func() {
	BeforeEach(func() { test.NewTempApp(GinkgoTB()) })

	// sized is a meter that has been laid out 100 wide, which is where its parts get their sizes.
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
