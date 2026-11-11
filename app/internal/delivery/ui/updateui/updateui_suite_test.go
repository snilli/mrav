package updateui

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUpdateUI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Update UI Suite")
}
