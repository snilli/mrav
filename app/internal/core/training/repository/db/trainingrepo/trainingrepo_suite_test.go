package trainingrepo_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTrainingrepo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Trainingrepo Suite")
}
