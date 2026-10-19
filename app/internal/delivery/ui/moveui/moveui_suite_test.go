package moveui_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMoveui(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Moveui Suite")
}
