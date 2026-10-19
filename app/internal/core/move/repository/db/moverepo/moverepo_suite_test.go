package moverepo_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMoverepo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Moverepo Suite")
}
