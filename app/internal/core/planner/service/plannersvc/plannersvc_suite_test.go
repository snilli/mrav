package plannersvc_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPlannersvc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Plannersvc Suite")
}
