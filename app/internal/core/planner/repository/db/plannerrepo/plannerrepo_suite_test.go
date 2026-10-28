package plannerrepo_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPlannerrepo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Plannerrepo Suite")
}
