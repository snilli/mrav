package status_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/apperr"
	"app/internal/delivery/ui/status"
)

var _ = Describe("Message", func() {
	It("says nothing when nothing is wrong", func() {
		Expect(status.Message(nil)).To(BeEmpty())
	})

	It("says what a not-found error says, such as that the game is not running", func() {
		Expect(status.Message(apperr.NotFound("game", "DuckStation is not running"))).To(Equal("DuckStation is not running"))
	})

	It("finds that message inside an error that wraps it", func() {
		err := fmt.Errorf("loading the page: %w", apperr.NotFound("game", "MR2DX is not running"))
		Expect(status.Message(err)).To(Equal("MR2DX is not running"))
	})

	It("says No monster for a not-found error that says nothing", func() {
		Expect(status.Message(apperr.ErrNotFound)).To(Equal("No monster"))
		Expect(status.Message(apperr.NotFound("monster", ""))).To(Equal("No monster"))
	})

	It("says any other error as an error", func() {
		Expect(status.Message(errors.New("db gone"))).To(Equal("Error: db gone"))
	})
})
