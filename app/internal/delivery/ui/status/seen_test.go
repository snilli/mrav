package status_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/delivery/ui/status"
)

var _ = Describe("Seen", func() {
	type page struct{ Names []string }

	It("is changed the first time, and then only when the value or the error is not what was last shown", func() {
		var s status.Seen[page]

		Expect(s.Changed(page{}, nil)).To(BeTrue(), "nothing was shown yet")
		Expect(s.Changed(page{}, nil)).To(BeFalse())
		Expect(s.Changed(page{Names: []string{"Ray"}}, nil)).To(BeTrue())
		Expect(s.Changed(page{Names: []string{"Ray"}}, nil)).To(BeFalse(), "an equal value, though not the same slice")
		Expect(s.Changed(page{Names: []string{"Ray"}}, errors.New("gone"))).To(BeTrue())
		Expect(s.Changed(page{Names: []string{"Ray"}}, errors.New("gone"))).To(BeFalse(), "an error that says the same")
		Expect(s.Changed(page{Names: []string{"Ray"}}, nil)).To(BeTrue())
	})
})
