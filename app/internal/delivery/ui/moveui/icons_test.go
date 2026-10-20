package moveui

import (
	"errors"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/move/domain"
)

var _ = Describe("Icons", func() {
	var icons *Icons

	BeforeEach(func() {
		test.NewTempApp(GinkgoTB())
		icons = NewIcons(func(key string) ([]byte, error) {
			if key == "Pixie_1" {
				return icon(), nil
			}
			return nil, errors.New("no such icon")
		})
	})

	image := func(kind domain.Kind) *canvas.Image {
		return icons.Image(domain.Move{Icon: "Pixie_1", Kind: kind}, 48, 32)
	}

	It("draws the icon of a move at the size asked, its pixels not smoothed", func() {
		img := image(domain.KindUnknown)

		Expect(img.MinSize()).To(Equal(fyne.NewSize(48, 32)))
		Expect(img.ScaleMode).To(Equal(canvas.ImageScalePixels))
		Expect(img.Resource).NotTo(BeNil())
	})

	It("colours it as the Moves page does, yellow for a move of power and green for one of intelligence", func() {
		plain, power, smart := image(domain.KindUnknown).Resource, image(domain.KindPower).Resource, image(domain.KindIntelligence).Resource

		Expect(power.Content()).NotTo(Equal(plain.Content()))
		Expect(smart.Content()).NotTo(Equal(plain.Content()))
		Expect(smart.Content()).NotTo(Equal(power.Content()))
	})

	It("draws a move whose icon is missing as an empty cell, which is nothing at all when there is none of that either", func() {
		img := icons.Image(domain.Move{Icon: "Missing", Kind: domain.KindPower}, 48, 32)

		Expect(img).NotTo(BeNil())
		Expect(img.Resource).To(BeNil())
	})
})
