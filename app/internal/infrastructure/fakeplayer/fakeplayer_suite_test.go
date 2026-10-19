package fakeplayer_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFakeplayer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fakeplayer Suite")
}
