package profileui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Columns", func() {
	var pad float32

	BeforeEach(func() {
		test.NewTempApp(GinkgoTB())
		pad = theme.Padding()
	})

	cells := func(n int) []fyne.CanvasObject {
		objs := make([]fyne.CanvasObject, n)
		for i := range objs {
			r := canvas.NewRectangle(color.Black)
			r.SetMinSize(fyne.NewSize(20, 30))
			objs[i] = r
		}
		return objs
	}

	It("gives a column of positive width that width, and the column of zero what is left", func() {
		objs := cells(3)
		Columns{50, 0, 30}.Layout(objs, fyne.NewSize(200, 30))

		Expect(objs[0].Position().X).To(BeZero())
		Expect(objs[0].Size().Width).To(Equal(float32(50)))
		Expect(objs[1].Position().X).To(Equal(50 + pad))
		Expect(objs[1].Size().Width).To(Equal(200 - 50 - 30 - 2*pad))
		Expect(objs[2].Position().X).To(Equal(float32(200 - 30)))
		Expect(objs[2].Size().Width).To(Equal(float32(30)))
	})

	It("gives every object the height of the row", func() {
		objs := cells(2)
		Columns{50, 0}.Layout(objs, fyne.NewSize(200, 44))
		Expect(objs[0].Size().Height).To(Equal(float32(44)))
		Expect(objs[1].Size().Height).To(Equal(float32(44)))
	})

	It("shares what is left between the columns of zero", func() {
		objs := cells(2)
		Columns{0, 0}.Layout(objs, fyne.NewSize(200, 30))
		Expect(objs[0].Size().Width).To(Equal((200 - pad) / 2))
		Expect(objs[1].Size().Width).To(Equal((200 - pad) / 2))
		Expect(objs[1].Position().X).To(Equal((200-pad)/2 + pad))
	})

	It("keeps the columns of a row with fewer objects where a full row has them, so that rows line up", func() {
		full, short := cells(3), cells(2)
		Columns{50, 0, 30}.Layout(full, fyne.NewSize(200, 30))
		Columns{50, 0, 30}.Layout(short, fyne.NewSize(200, 30))
		Expect(short[1].Position()).To(Equal(full[1].Position()))
		Expect(short[1].Size()).To(Equal(full[1].Size()))
	})

	It("never gives a column a negative width, when there is less room than the fixed columns need", func() {
		objs := cells(3)
		Columns{50, 0, 30}.Layout(objs, fyne.NewSize(40, 30))
		Expect(objs[1].Size().Width).To(BeZero())
	})

	It("needs the fixed widths, the least the flexible columns need, the gaps, and the tallest row", func() {
		objs := cells(3)
		objs[1].(*canvas.Rectangle).SetMinSize(fyne.NewSize(70, 45))
		Expect(Columns{50, 0, 30}.MinSize(objs)).To(Equal(fyne.NewSize(50+70+30+2*pad, 45)))
	})
})
