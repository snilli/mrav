package errantryui_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestErrantryui(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Errantryui Suite")
}
