package movesvc_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMovesvc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Movesvc Suite")
}
