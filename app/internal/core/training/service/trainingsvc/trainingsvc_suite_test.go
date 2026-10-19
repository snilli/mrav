package trainingsvc_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTrainingsvc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Trainingsvc Suite")
}
