package testutil_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"orm/ent/migrate"
	entmove "orm/ent/move"

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
		Expect(n).To(Equal(38))
	})

	It("ships with the indexes that the schema declares and no others, because the viewer never migrates and the tests run on a schema that Ent made", func() {
		ctx := context.Background()
		client, err := testutil.OpenShipped(ctx)
		if errors.Is(err, testutil.ErrNoShippedDB) {
			Skip(err.Error())
		}
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)

		declared := map[string]bool{}
		for _, table := range migrate.Tables {
			for _, index := range table.Indexes {
				declared[table.Name+"."+index.Name] = true
			}
			for _, column := range table.Columns {
				if column.Unique {
					declared[table.Name+"."+table.Name+"_"+column.Name+"_key"] = true
				}
			}
		}

		rows, err := client.DB.QueryContext(ctx, "SELECT tbl_name, name FROM sqlite_master WHERE type = 'index' AND name NOT LIKE 'sqlite_autoindex%'")
		Expect(err).NotTo(HaveOccurred())
		defer rows.Close()
		shipped := map[string]bool{}
		for rows.Next() {
			var table, name string
			Expect(rows.Scan(&table, &name)).To(Succeed())
			shipped[table+"."+name] = true
		}
		Expect(rows.Err()).NotTo(HaveOccurred())

		var missing, extra []string
		for name := range declared {
			if !shipped[name] {
				missing = append(missing, name)
			}
		}
		for name := range shipped {
			if !declared[name] {
				extra = append(extra, name)
			}
		}
		Expect(missing).To(BeEmpty(), "declared by the schema, not in data/mrav.sqlite")
		Expect(extra).To(BeEmpty(), "in data/mrav.sqlite, not declared by the schema")
	})

	It("ships no extra row of a move with a scale of its own, because the game rates the numbers of those rows on the scale of force", func() {
		ctx := context.Background()
		client, err := testutil.OpenShipped(ctx)
		if errors.Is(err, testutil.ErrNoShippedDB) {
			Skip(err.Error())
		}
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)

		var n int
		Expect(client.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM moves WHERE json_extract(extra, '$.scale') IS NOT NULL").Scan(&n)).To(Succeed())
		Expect(n).To(BeZero())
	})

	DescribeTable("ships the Mocchi's Petal Storm as a drain of 50%, as the LegendCup page has it",
		func(version entmove.Version) {
			ctx := context.Background()
			client, err := testutil.OpenShipped(ctx)
			if errors.Is(err, testutil.ErrNoShippedDB) {
				Skip(err.Error())
			}
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(client.Close)

			storm := client.Move.Query().Where(entmove.VersionEQ(version), entmove.GenusID(26), entmove.Name("Petal Storm")).OnlyX(ctx)
			Expect(storm.Extra).NotTo(BeNil())
			Expect(storm.Extra.Text).To(Equal("50%"))
		},
		Entry("the PlayStation game, mr2", entmove.VersionMr2),
		Entry("the Steam game, mr2dx", entmove.VersionMr2dx),
	)

	DescribeTable("leaves the Golem sub breed 38 out of the fourth place's lesson of Heavy Kick: Forward Golem on the PlayStation, Dream Golem on Steam",
		func(version entmove.Version) {
			ctx := context.Background()
			client, err := testutil.OpenShipped(ctx)
			if errors.Is(err, testutil.ErrNoShippedDB) {
				Skip(err.Error())
			}
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(client.Close)

			kick := client.Move.Query().Where(entmove.VersionEQ(version), entmove.GenusID(8), entmove.Name("Heavy Kick")).OnlyX(ctx)
			var fourth []int
			for _, a := range kick.Acquire {
				if a.Place == 4 {
					fourth = a.SubNotIn
				}
			}
			Expect(fourth).To(Equal([]int{4, 5, 6, 10, 14, 17, 22, 31, 38}))
		},
		Entry("the PlayStation game, mr2", entmove.VersionMr2),
		Entry("the Steam game, mr2dx", entmove.VersionMr2dx),
	)

	DescribeTable("ships every move once for each version of the game, with what the Errantry asks for each move it teaches, from the Errantry calculator",
		func(version entmove.Version, withAcquire, entries int) {
			ctx := context.Background()
			client, err := testutil.OpenShipped(ctx)
			if errors.Is(err, testutil.ErrNoShippedDB) {
				Skip(err.Error())
			}
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(client.Close)

			all, err := client.Move.Query().Where(entmove.VersionEQ(version)).Count(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(all).To(Equal(637))

			moves, err := client.Move.Query().Where(entmove.VersionEQ(version), entmove.AcquireNotNil()).All(ctx)
			Expect(err).NotTo(HaveOccurred())
			got := 0
			for _, m := range moves {
				got += len(m.Acquire)
			}
			Expect(moves).To(HaveLen(withAcquire))
			Expect(got).To(Equal(entries))
		},
		Entry("the Steam game, mr2dx: 651 entries over 560 moves", entmove.VersionMr2dx, 560, 651),
		Entry("the PlayStation game, mr2: 650 entries over 559 moves", entmove.VersionMr2, 559, 650),
	)
})
