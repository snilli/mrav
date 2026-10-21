package shellui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Theme", func() {
	BeforeEach(func() { test.NewTempApp(GinkgoTB()) })

	It("pads less than the default theme, so that a screen of numbers fits a window without scrolling", func() {
		def, dense := theme.DefaultTheme(), Theme()
		Expect(dense.Size(theme.SizeNameInnerPadding)).To(BeNumerically("<", def.Size(theme.SizeNameInnerPadding)))
		Expect(dense.Size(theme.SizeNamePadding)).To(BeNumerically("<", def.Size(theme.SizeNamePadding)))
	})

	It("lightens the quiet text of the dark theme, in which the names of the things shown are drawn, but not up to the text", func() {
		brightness := func(c color.Color) uint32 { r, g, b, _ := c.RGBA(); return r + g + b }
		def := theme.DefaultTheme()
		got := Theme().Color(theme.ColorNameDisabled, theme.VariantDark)
		Expect(brightness(got)).To(BeNumerically(">", brightness(def.Color(theme.ColorNameDisabled, theme.VariantDark))))
		Expect(brightness(got)).To(BeNumerically("<", brightness(def.Color(theme.ColorNameForeground, theme.VariantDark))))
		Expect(Theme().Color(theme.ColorNameDisabled, theme.VariantLight)).To(Equal(got))
	})

	It("draws the text smaller than the default theme, so that the window can be small enough to sit beside the game", func() {
		Expect(Theme().Size(theme.SizeNameText)).To(BeNumerically("<", theme.DefaultTheme().Size(theme.SizeNameText)))
	})

	It("is dark even where the system asks for light, as the colours of the pages are picked for a dark background", func() {
		def := theme.DefaultTheme()
		Expect(Theme().Color(theme.ColorNameBackground, theme.VariantLight)).To(Equal(def.Color(theme.ColorNameBackground, theme.VariantDark)))
		Expect(Theme().Color(theme.ColorNameForeground, theme.VariantLight)).To(Equal(def.Color(theme.ColorNameForeground, theme.VariantDark)))
	})

	It("is the default theme in everything else", func() {
		def, dense := theme.DefaultTheme(), Theme()
		Expect(dense.Size(theme.SizeNameScrollBar)).To(Equal(def.Size(theme.SizeNameScrollBar)))
		Expect(dense.Color(theme.ColorNameForeground, theme.VariantDark)).To(Equal(def.Color(theme.ColorNameForeground, theme.VariantDark)))
		Expect(dense.Font(fyne.TextStyle{Bold: true})).To(Equal(def.Font(fyne.TextStyle{Bold: true})))
	})
})
