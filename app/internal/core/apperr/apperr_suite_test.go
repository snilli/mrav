package apperr_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestApperr(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Apperr Suite")
}
