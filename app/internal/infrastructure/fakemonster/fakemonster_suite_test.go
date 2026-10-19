package fakemonster_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFakemonster(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fakemonster Suite")
}
