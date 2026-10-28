package drillsvc_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDrillsvc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Drillsvc Suite")
}
