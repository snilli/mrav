package apperr_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/apperr"
)

var _ = Describe("apperr", func() {
	It("names the field and keeps its kind", func() {
		err := apperr.Invalid("icon", "is empty")
		Expect(err).To(MatchError("icon: is empty"))
		Expect(errors.Is(err, apperr.ErrInvalid)).To(BeTrue())
		Expect(errors.Is(err, apperr.ErrNotFound)).To(BeFalse())
	})

	It("leaves out an empty field", func() {
		err := apperr.NotFound("", "no such genus")
		Expect(err).To(MatchError("no such genus"))
		Expect(errors.Is(err, apperr.ErrNotFound)).To(BeTrue())
	})
})
