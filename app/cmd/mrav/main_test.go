package main

import (
	"fyne.io/fyne/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
)

var _ = Describe("the dependency graph", func() {
	It("can build the window from everything it needs", func() {
		Expect(fx.ValidateApp(options(), fx.Populate(new(fyne.Window)))).To(Succeed())
	})
})
