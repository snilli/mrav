package gamememory

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("readText", func() {
	read := func(img *image, t Text) string {
		GinkgoHelper()
		r := &reader{mem: img}
		s := readText(r, t)
		Expect(r.err).NotTo(HaveOccurred())
		return s
	}
	ps1Text := Text{Addr: 0x1000, Len: 12}
	steamText := Text{Addr: 0x1000, Len: 24, Steam: true}

	It("reads an English PlayStation name by its glyph codes, up to the byte 255", func() {
		img := newImage()
		img.put(0x1000, 11, 37, 0, 2, 255, 11)
		Expect(read(img, ps1Text)).To(Equal("Aa 1"))
	})

	It("stops at the length when nothing ends the name", func() {
		img := newImage()
		for i := range uint32(16) {
			img.u8(0x1000+i, 11)
		}
		Expect(read(img, ps1Text)).To(Equal("AAAAAAAAAAAA"))
	})

	It("shows a PlayStation code it has no glyph for as an underscore", func() {
		img := newImage()
		img.put(0x1000, 82, 200, 255)
		Expect(read(img, ps1Text)).To(Equal("__"))
	})

	It("reads a Steam name: one byte for a space or a digit, two for a letter, as the bytes B0 to B6 begin a pair", func() {
		img := newImage()
		img.put(0x1000, 0xB0, 0x00, 0xB0, 0x1A, 0x00, 0xA3, 0xFF)
		Expect(read(img, steamText)).To(Equal("Aa 1"))
	})

	It("reads a Steam pair that straddles the end of the name", func() {
		img := newImage()
		for i := range uint32(23) {
			img.u8(0x1000+i, 0x00)
		}
		img.put(0x1000+23, 0xB0, 0x00)
		Expect(read(img, steamText)).To(Equal(strings.Repeat(" ", 23) + "A"))
	})

	It("shows a Steam code it has no glyph for as a question mark", func() {
		img := newImage()
		img.put(0x1000, 0xAF, 0xB6, 0x00, 0xFF)
		Expect(read(img, steamText)).To(Equal("??"))
	})

	It("reads nothing from a text that has no address, such as the breeder's name on the PlayStation", func() {
		Expect(read(newImage(), Text{})).To(BeEmpty())
	})
})
