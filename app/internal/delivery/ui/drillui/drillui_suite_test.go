package drillui

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDrillui(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Drillui Suite")
}
