package testutil_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/testutil"
)

var _ = Describe("OpenShipped", Label("integration"), func() {
	It("opens the database the project ships from any package, or says there is none", func() {
		client, err := testutil.OpenShipped(context.Background())
		if errors.Is(err, testutil.ErrNoShippedDB) {
			Skip(err.Error())
		}
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)

		n, err := client.Genus.Query().Count(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(n).To(Equal(38)) // Pixie to Naga
	})
})
