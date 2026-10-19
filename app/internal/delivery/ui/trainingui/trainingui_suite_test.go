package trainingui_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTrainingui(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Trainingui Suite")
}
