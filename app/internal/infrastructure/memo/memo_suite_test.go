package memo_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMemo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Memo Suite")
}
