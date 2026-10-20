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
		Expect(n).To(Equal(38)) // Pixie to Naga
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
				if column.Unique { // a unique column is an index of its own, named by the column
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

	It("ships what the errantry asks for each move it teaches: 651 entries over 560 moves, from the Errantry calculator", func() {
		ctx := context.Background()
		client, err := testutil.OpenShipped(ctx)
		if errors.Is(err, testutil.ErrNoShippedDB) {
			Skip(err.Error())
		}
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)

		moves, err := client.Move.Query().Where(entmove.AcquireNotNil()).All(ctx)
		Expect(err).NotTo(HaveOccurred())
		entries := 0
		for _, m := range moves {
			entries += len(m.Acquire)
		}
		Expect(moves).To(HaveLen(560))
		Expect(entries).To(Equal(651))
	})
})
