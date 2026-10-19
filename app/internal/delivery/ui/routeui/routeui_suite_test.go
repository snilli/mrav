package routeui

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRouteui(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Routeui Suite")
}
