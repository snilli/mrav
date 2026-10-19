package gamememory

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGamememory(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gamememory Suite")
}
