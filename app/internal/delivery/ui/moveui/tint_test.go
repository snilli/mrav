package moveui

import (
	"bytes"
	"image"
	"image/color"
	"image/png"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/move/domain"
)

func icon() []byte {
	img := image.NewNRGBA(image.Rect(0, 0, 3, 1))
	img.Set(0, 0, color.NRGBA{A: 255})
	img.Set(1, 0, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	img.Set(2, 0, color.NRGBA{R: 152, G: 152, B: 152, A: 255})
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		panic(err)
	}
	return b.Bytes()
}

func pixels(b []byte) []color.NRGBA {
	img, err := png.Decode(bytes.NewReader(b))
	Expect(err).NotTo(HaveOccurred())
	var out []color.NRGBA
	for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
		out = append(out, color.NRGBAModel.Convert(img.At(x, 0)).(color.NRGBA))
	}
	return out
}

var _ = Describe("tint", func() {
	black := color.NRGBA{A: 255}
	grey := color.NRGBA{R: 152, G: 152, B: 152, A: 255}

	It("turns the white of a power icon yellow, and leaves the black and the grey as they are", func() {
		out, err := tint(icon(), domain.KindPower)
		Expect(err).NotTo(HaveOccurred())
		Expect(pixels(out)).To(Equal([]color.NRGBA{black, powerFill, grey}))
	})

	It("turns it green for an intelligence icon", func() {
		out, err := tint(icon(), domain.KindIntelligence)
		Expect(err).NotTo(HaveOccurred())
		Expect(pixels(out)).To(Equal([]color.NRGBA{black, intelligenceFill, grey}))
	})

	It("leaves the icon of a move of no known kind as it was", func() {
		out, err := tint(icon(), domain.KindUnknown)
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal(icon()))
	})

	It("says so when it is not given a PNG", func() {
		_, err := tint([]byte("not a picture"), domain.KindPower)
		Expect(err).To(HaveOccurred())
	})

	It("uses a yellow and a green that cannot be taken for each other, or for the white and the grey of the icon", func() {
		Expect(powerFill).NotTo(Equal(intelligenceFill))
		Expect(powerFill.R).To(BeNumerically(">", powerFill.B), "yellow has more red than blue")
		Expect(intelligenceFill.G).To(BeNumerically(">", intelligenceFill.R), "green has more green than red")
	})
})
