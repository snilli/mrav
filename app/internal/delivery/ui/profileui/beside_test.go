package profileui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Beside", func() {
	var (
		pad  float32
		a, b *canvas.Rectangle
	)

	BeforeEach(func() {
		test.NewTempApp(GinkgoTB())
		pad = theme.Padding()
		a, b = canvas.NewRectangle(color.Black), canvas.NewRectangle(color.Black)
		a.SetMinSize(fyne.NewSize(100, 20))
		b.SetMinSize(fyne.NewSize(60, 30))
	})

	row := func(l Beside, width float32) *fyne.Container {
		c := container.New(&l, a, b)
		c.Resize(fyne.NewSize(width, c.MinSize().Height))
		return c
	}

	It("puts the first part at the left edge and the second at the right edge, on one line, when it has the room", func() {
		c := row(Beside{}, 400)

		Expect(a.Position()).To(Equal(fyne.NewPos(0, 0)))
		Expect(b.Position()).To(Equal(fyne.NewPos(340, 0)))
		Expect(c.MinSize()).To(Equal(fyne.NewSize(100+60+pad, 30)))
	})

	It("gives each part half of the line, when they are to share it evenly", func() {
		c := row(Beside{Even: true}, 400)

		Expect(a.Size().Width).To(Equal((400 - pad) / 2))
		Expect(b.Size().Width).To(Equal((400 - pad) / 2))
		Expect(b.Position().X).To(Equal((400-pad)/2 + pad))
		Expect(c.MinSize().Width).To(Equal(2*100+pad), "twice the wider part")
	})

	It("puts the second part under the first when it is stacked, as wide as the wider part and as high as both", func() {
		c := row(Beside{Stacked: true}, 150)

		Expect(a.Position()).To(Equal(fyne.NewPos(0, 0)))
		Expect(b.Position()).To(Equal(fyne.NewPos(0, 20+pad)))
		Expect(c.MinSize()).To(Equal(fyne.NewSize(100, 20+30+pad)))
	})
})
