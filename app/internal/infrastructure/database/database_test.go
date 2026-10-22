package database_test

import (
	"context"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx/fxtest"

	"orm"

	"app/config"
	"app/internal/infrastructure/database"
)

var _ = Describe("database", func() {
	var (
		ctx  context.Context
		path string
	)

	existing := func() {
		c, err := orm.Open(ctx, path)
		Expect(err).NotTo(HaveOccurred())
		Expect(c.Schema.Create(ctx)).To(Succeed())
		Expect(c.Close()).To(Succeed())
	}

	BeforeEach(func() {
		ctx = context.Background()
		path = filepath.Join(GinkgoT().TempDir(), "mrav.sqlite")
	})

	It("opens a database that exists", func() {
		existing()
		c, err := database.Open(ctx, path)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(c.Close)
		Expect(c.Genus.Query().Count(ctx)).To(Equal(0))
	})

	It("refuses a missing file instead of creating an empty one", func() {
		_, err := database.Open(ctx, path)
		Expect(err).To(MatchError(ContainSubstring(path)))
		Expect(path).NotTo(BeAnExistingFile())
	})

	It("closes with the app", func() {
		existing()
		lc := fxtest.NewLifecycle(GinkgoT())
		c, err := database.ProvideDatabase(lc, &config.Config{DBPath: path})
		Expect(err).NotTo(HaveOccurred())

		lc.RequireStart()
		Expect(c.Genus.Query().Count(ctx)).To(Equal(0))
		lc.RequireStop()
		_, err = c.Genus.Query().Count(ctx)
		Expect(err).To(HaveOccurred())
	})
})
