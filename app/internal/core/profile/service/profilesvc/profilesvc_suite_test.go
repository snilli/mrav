package profilesvc_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestProfilesvc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Profilesvc Suite")
}
