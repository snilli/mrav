package itemsvc_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestItemsvc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Itemsvc Suite")
}
