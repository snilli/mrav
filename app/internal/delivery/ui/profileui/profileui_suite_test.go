package profileui_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestProfileui(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Profileui Suite")
}
