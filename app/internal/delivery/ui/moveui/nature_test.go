package moveui

import (
	"bytes"
	"image/color"
	"image/png"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/move/domain"
)

func colours(b []byte, is func(color.NRGBA) bool) int {
	img, err := png.Decode(bytes.NewReader(b))
	Expect(err).NotTo(HaveOccurred())
	n := 0
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			if is(color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)) {
				n++
			}
		}
	}
	return n
}

var _ = Describe("natureIcon", func() {
	yellow := func(c color.NRGBA) bool { return c.A > 0 && c.R > 200 && c.G > 200 && c.B < 100 }
	dark := func(c color.NRGBA) bool { return c.A > 0 && c.R < 100 && c.G < 100 && c.B < 100 }
	light := func(c color.NRGBA) bool { return c.A > 0 && c.R > 230 && c.G > 230 && c.B > 230 }

	It("has none for a move that asks for no nature", func() {
		Expect(natureIcon(domain.NatureNone)).To(BeNil())
	})

	It("is the size of one of the game's marks, 24 by 16, for a good nature and for a bad one", func() {
		for _, n := range []domain.Nature{domain.NatureGood, domain.NatureBad} {
			img, err := png.Decode(bytes.NewReader(natureIcon(n)))
			Expect(err).NotTo(HaveOccurred())
			Expect([]int{img.Bounds().Dx(), img.Bounds().Dy()}).To(Equal([]int{24, 16}))
		}
	})

	It("is a light tile with dark brows for a good nature", func() {
		good := natureIcon(domain.NatureGood)
		Expect(colours(good, light)).To(BeNumerically(">", 150), "most of the tile is light")
		Expect(colours(good, dark)).To(BeNumerically(">", 10), "and has brows drawn on it")
		Expect(colours(good, yellow)).To(BeZero())
	})

	It("is a dark tile with yellow eyes for a bad nature", func() {
		bad := natureIcon(domain.NatureBad)
		Expect(colours(bad, dark)).To(BeNumerically(">", 150), "most of the tile is dark")
		Expect(colours(bad, yellow)).To(BeNumerically(">", 10), "and has eyes drawn on it")
		Expect(colours(bad, light)).To(BeZero())
	})

	It("has the eyes in a pair, one the other's reflection", func() {
		img, err := png.Decode(bytes.NewReader(natureIcon(domain.NatureBad)))
		Expect(err).NotTo(HaveOccurred())
		for y := range 16 {
			for x := range 12 {
				left := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
				right := color.NRGBAModel.Convert(img.At(23-x, y)).(color.NRGBA)
				Expect(right).To(Equal(left), "at %d, %d", x, y)
			}
		}
	})
})
