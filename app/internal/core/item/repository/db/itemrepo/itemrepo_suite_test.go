package itemrepo_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestItemrepo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Itemrepo Suite")
}
