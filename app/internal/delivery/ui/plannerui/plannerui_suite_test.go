package plannerui

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPlannerui(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Plannerui Suite")
}
