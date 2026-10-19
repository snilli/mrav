package itemui_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestItemui(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Itemui Suite")
}
