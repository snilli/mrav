package profileui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Grid", func() {
	BeforeEach(func() { test.NewTempApp(GinkgoTB()) })

	It("draws a line between each two columns of a row where the columns end, and one under the row", func() {
		columns := Columns{104, 60, 0, 0, 0}
		g := NewGrid(columns)
		g.Resize(fyne.NewSize(466, 40))

		edges := columns.Edges(466)
		Expect(g.Objects).To(HaveLen(len(edges) + 1))
		for i, x := range edges {
			Expect(g.Objects[i].Position().X).To(BeNumerically("~", x-LineWidth/2, 0.01))
			Expect(g.Objects[i].Size()).To(Equal(fyne.NewSize(LineWidth, 40)))
		}
		bottom := g.Objects[len(edges)]
		Expect(bottom.Position()).To(Equal(fyne.NewPos(0, 40-LineWidth)))
		Expect(bottom.Size()).To(Equal(fyne.NewSize(466, LineWidth)))
	})

	It("draws only the line under the row when the row is one column, as for a heading that runs across", func() {
		g := NewGrid(Columns{0})
		g.Resize(fyne.NewSize(300, 30))
		Expect(g.Objects).To(HaveLen(1))
		Expect(g.Objects[0].Size()).To(Equal(fyne.NewSize(300, LineWidth)))
	})

	It("asks for no room of its own, so that it never makes a row taller or wider", func() {
		Expect(NewGrid(Columns{50, 0, 30}).MinSize()).To(Equal(fyne.Size{}))
	})

	It("draws its lines in a colour that shows against the background of the dark theme", func() {
		light := func(c color.Color) int {
			r, g, b, _ := c.RGBA()
			return int((r + g + b) / 3 >> 8)
		}
		background := theme.DefaultTheme().Color(theme.ColorNameBackground, theme.VariantDark)
		Expect(light(LineColor) - light(background)).To(BeNumerically(">=", 80))
	})
})
