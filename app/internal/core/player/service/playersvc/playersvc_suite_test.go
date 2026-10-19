package playersvc_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPlayersvc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Playersvc Suite")
}
