package updateui

import (
	"net/url"

	"fyne.io/fyne/v2/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/release/domain"
)

var _ = Describe("Bar", func() {
	var (
		opened []string
		b      *Bar
	)
	release := domain.Release{Tag: "v0.5.1", URL: "https://github.com/snilli/mrav/releases/tag/v0.5.1"}

	BeforeEach(func() {
		test.NewTempApp(GinkgoTB())
		opened = nil
		b = New(func(u *url.URL) error { opened = append(opened, u.String()); return nil })
	})

	It("stays out of sight until a release is offered", func() {
		Expect(b.row.Visible()).To(BeFalse())
		Expect(b.Object().MinSize().Height).To(BeZero())
	})

	It("says which version is out", func() {
		b.show(release)
		Expect(b.row.Visible()).To(BeTrue())
		Expect(b.message.Text).To(Equal("mrav v0.5.1 is out."))
	})

	It("opens the release page on Download", func() {
		b.show(release)
		test.Tap(b.download)
		Expect(opened).To(Equal([]string{release.URL}))
	})

	It("goes away on Close", func() {
		b.show(release)
		test.Tap(b.close)
		Expect(b.row.Visible()).To(BeFalse())
		Expect(b.Object().MinSize().Height).To(BeZero())
	})
})
