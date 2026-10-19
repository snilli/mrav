package playerui_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPlayerui(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Playerui Suite")
}
