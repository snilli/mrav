package testutil_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"orm"

	"app/internal/testutil"
)

var _ = Describe("Unindexed", func() {
	var (
		ctx    context.Context
		client *orm.Client
	)
	exec := func(stmt string) {
		_, err := client.DB.ExecContext(ctx, stmt)
		Expect(err).NotTo(HaveOccurred())
	}
	unindexed := func(query string, args ...any) []string {
		plan, err := testutil.Plan(ctx, client.DB, query, args...)
		Expect(err).NotTo(HaveOccurred())
		return testutil.Unindexed(query, plan)
	}

	BeforeEach(func() {
		ctx = context.Background()
		var err error
		client, err = testutil.NewClient(ctx)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)
		exec("CREATE TABLE t (a INTEGER, b INTEGER)")
	})

	It("names the steps that read a whole table to find rows, and that sort rows an index could give in order", func() {
		got := unindexed("SELECT a FROM t WHERE b = ? ORDER BY a", 1)
		Expect(got).To(ConsistOf(ContainSubstring("SCAN t"), ContainSubstring("TEMP B-TREE")))
	})

	It("passes a statement that an index serves, in the order it asks for", func() {
		exec("CREATE INDEX t_b_a ON t (b, a)")
		Expect(unindexed("SELECT a FROM t WHERE b = ? ORDER BY a", 1)).To(BeEmpty())
	})

	It("lets a listing of a whole table through, which no index could speed up, but not a sort of it", func() {
		Expect(unindexed("SELECT a, b FROM t")).To(BeEmpty())
		Expect(unindexed("SELECT a, b FROM t ORDER BY a")).To(ConsistOf(ContainSubstring("TEMP B-TREE")))
		exec("CREATE INDEX t_a ON t (a)")
		Expect(unindexed("SELECT a, b FROM t ORDER BY a")).To(BeEmpty())
	})

	It("does not count an index that is read from end to end as serving a statement that looks for rows", func() {
		exec("CREATE INDEX t_a ON t (a)")
		got := unindexed("SELECT a, b FROM t WHERE b = ? ORDER BY a", 1)
		Expect(got).To(ConsistOf(ContainSubstring("SCAN t USING INDEX t_a")))
	})

	It("does not mistake a step of a join that is looked up by its key for a scan", func() {
		exec("CREATE TABLE u (id INTEGER PRIMARY KEY, c INTEGER)")
		exec("CREATE INDEX t_b ON t (b)")
		Expect(unindexed("SELECT u.c FROM t JOIN u ON u.id = t.a WHERE t.b = ?", 1)).To(BeEmpty())
	})
})
