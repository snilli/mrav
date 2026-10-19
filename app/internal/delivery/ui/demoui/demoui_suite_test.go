package demoui_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDemoui(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Demoui Suite")
}
