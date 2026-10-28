package drillrepo_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDrillrepo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Drillrepo Suite")
}
