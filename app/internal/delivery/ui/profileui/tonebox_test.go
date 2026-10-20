package profileui

import (
	"fyne.io/fyne/v2/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ToneBox", func() {
	BeforeEach(func() { test.NewTempApp(GinkgoTB()) })

	It("is small when it is compact, small enough to sit in the corner of an icon of 48 by 32", func() {
		box := Box{Text: "100%", Tone: ToneLimeGreen}
		normal, compact := NewToneBox(), NewToneBox().Compact()
		normal.Set(box)
		compact.Set(box)

		Expect(compact.MinSize().Width).To(BeNumerically("<", normal.MinSize().Width))
		Expect(compact.MinSize().Height).To(BeNumerically("<", normal.MinSize().Height))
		Expect(compact.MinSize().Width).To(BeNumerically("<=", 40))
		Expect(compact.MinSize().Height).To(BeNumerically("<=", 16))
	})

	It("keeps what it was given when it is compact", func() {
		box := Box{Text: "52%", Tone: ToneLightGreen}
		compact := NewToneBox().Compact()
		compact.Set(box)
		Expect(compact.Box()).To(Equal(box))
	})
})
