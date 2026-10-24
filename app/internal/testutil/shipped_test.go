package testutil_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"orm/ent"
	"orm/ent/migrate"
	entmove "orm/ent/move"
	"orm/stored"

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

	DescribeTable("ships the stat sums that each game asks for five moves: POW on the PlayStation and INT on Steam for four of them, and the other way round for Whirlwind",
		func(version entmove.Version, genus int, name string, stats []int, text string) {
			ctx := context.Background()
			client, err := testutil.OpenShipped(ctx)
			if errors.Is(err, testutil.ErrNoShippedDB) {
				Skip(err.Error())
			}
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(client.Close)

			m := client.Move.Query().Where(entmove.VersionEQ(version), entmove.GenusID(genus), entmove.Name(name)).OnlyX(ctx)
			var got []int
			for _, c := range m.Unlock {
				if c.Type == stored.StatSum {
					got = c.Stats
				}
			}
			Expect(got).To(Equal(stats))
			Expect(m.Description).To(ContainSubstring(text))
		},
		Entry("mr2 Hidden Sting", entmove.VersionMr2, 11, "Hidden Sting", []int{1, 3}, "POW + SKI should total over 650"),
		Entry("mr2 Javelin", entmove.VersionMr2, 11, "Javelin", []int{1, 4}, "POW + SPD should total over 650"),
		Entry("mr2 Flying Mask", entmove.VersionMr2, 16, "Flying Mask", []int{1, 4}, "POW + SPD should total over 400"),
		Entry("mr2 Fist Shot", entmove.VersionMr2, 8, "Fist Shot", []int{1, 4}, "POW + SPD should total over 800"),
		Entry("mr2 Whirlwind", entmove.VersionMr2, 16, "Whirlwind", []int{2}, "INT should be over 450"),
		Entry("mr2dx Hidden Sting", entmove.VersionMr2dx, 11, "Hidden Sting", []int{2, 3}, "INT + SKI"),
		Entry("mr2dx Javelin", entmove.VersionMr2dx, 11, "Javelin", []int{2, 4}, "INT + SPD should total over 650"),
		Entry("mr2dx Flying Mask", entmove.VersionMr2dx, 16, "Flying Mask", []int{2, 4}, "INT + SPD should total over 400"),
		Entry("mr2dx Fist Shot", entmove.VersionMr2dx, 8, "Fist Shot", []int{2, 4}, "INT + SPD"),
		Entry("mr2dx Whirlwind", entmove.VersionMr2dx, 16, "Whirlwind", []int{1}, "POW (not a typo) should be over 450"),
	)

	DescribeTable("ships the seven healing moves with a force of nothing, the amount they heal being the extra row, as the game's own table has it",
		func(version entmove.Version) {
			ctx := context.Background()
			client, err := testutil.OpenShipped(ctx)
			if errors.Is(err, testutil.ErrNoShippedDB) {
				Skip(err.Error())
			}
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(client.Close)

			moves, err := client.Move.Query().Where(entmove.VersionEQ(version)).All(ctx)
			Expect(err).NotTo(HaveOccurred())
			healing := 0
			for _, m := range moves {
				if m.Extra != nil && m.Extra.Label == "Recovery:" {
					healing++
					Expect(m.Extra.Raw).To(BeNumerically(">", 0), m.Name)
					Expect(m.Force).To(BeZero(), m.Name)
				}
			}
			Expect(healing).To(Equal(7))
		},
		Entry("the PlayStation game, mr2", entmove.VersionMr2),
		Entry("the Steam game, mr2dx", entmove.VersionMr2dx),
	)

	It("ships Kato's Oil Fire as a move with no way to learn on the PlayStation, and with the second place on Steam", func() {
		ctx := context.Background()
		client, err := testutil.OpenShipped(ctx)
		if errors.Is(err, testutil.ErrNoShippedDB) {
			Skip(err.Error())
		}
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)

		fire := func(version entmove.Version) *ent.Move {
			return client.Move.Query().Where(entmove.VersionEQ(version), entmove.GenusID(17), entmove.Name("Oil Fire")).OnlyX(ctx)
		}
		mr2, mr2dx := fire(entmove.VersionMr2), fire(entmove.VersionMr2dx)
		Expect(mr2.UnlockNever).To(BeTrue())
		Expect(mr2.Unlock).To(BeEmpty())
		Expect(mr2.Acquire).To(BeEmpty())
		Expect(mr2.Description).NotTo(ContainSubstring("should total over"))
		Expect(mr2dx.UnlockNever).To(BeFalse())
		Expect(mr2dx.Acquire).NotTo(BeEmpty())
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
