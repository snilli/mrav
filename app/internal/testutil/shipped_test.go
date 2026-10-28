package testutil_test

import (
	"context"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"orm/ent"
	entdrillgain "orm/ent/drillgain"
	"orm/ent/lifestage"
	"orm/ent/migrate"
	entmove "orm/ent/move"
	entsong "orm/ent/song"
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

	It("ships the share of each of the ten life stages for the four life types of the calculator of LegendCup, by the life type byte of the game: Normal is 0", func() {
		ctx := context.Background()
		client, err := testutil.OpenShipped(ctx)
		if errors.Is(err, testutil.ErrNoShippedDB) {
			Skip(err.Error())
		}
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)

		shares := map[int][]int{}
		for _, row := range client.LifeStage.Query().Order(ent.Asc(lifestage.FieldLifeType, lifestage.FieldStage)).AllX(ctx) {
			shares[row.LifeType] = append(shares[row.LifeType], row.Share)
		}
		Expect(shares).To(Equal(map[int][]int{
			0: {10, 10, 15, 15, 10, 5, 10, 5, 5, 15},
			1: {5, 5, 10, 10, 5, 5, 15, 15, 10, 20},
			2: {15, 15, 15, 15, 10, 10, 5, 5, 5, 5},
			3: {5, 5, 10, 10, 10, 5, 15, 15, 10, 15},
		}))
	})

	It("ships the 1218 songs of the MR2DX list of LegendCup, one for each monster id, over 415 breeds of which 17 are the enemies that the Shrine keeps locked", func() {
		ctx := context.Background()
		client, err := testutil.OpenShipped(ctx)
		if errors.Is(err, testutil.ErrNoShippedDB) {
			Skip(err.Error())
		}
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)

		songs, err := client.Song.Query().Count(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(songs).To(Equal(1218))
		var breeds, enemies int
		Expect(client.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM (SELECT DISTINCT main_genus, sub_genus FROM songs)").Scan(&breeds)).To(Succeed())
		Expect(client.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM (SELECT DISTINCT main_genus, sub_genus FROM songs WHERE enemy)").Scan(&enemies)).To(Succeed())
		Expect(breeds).To(Equal(415))
		Expect(enemies).To(Equal(17))
	})

	DescribeTable("ships a song as the script of the page lists it: stats and gains in the order of the page, the life type by the byte of the game, the offset as a flag",
		func(id int, title, artist string, offset bool, stats, gains []int, lifespan, nature, lifeType, guts int, arena, training, techs, specials string, enemy bool) {
			ctx := context.Background()
			client, err := testutil.OpenShipped(ctx)
			if errors.Is(err, testutil.ErrNoShippedDB) {
				Skip(err.Error())
			}
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(client.Close)

			song := client.Song.Query().Where(entsong.MonsterID(id)).OnlyX(ctx)
			Expect([]string{song.Title, song.Artist}).To(Equal([]string{title, artist}))
			Expect(song.HasOffset).To(Equal(offset))
			Expect(song.Stats).To(Equal(stats))
			Expect(song.Gains).To(Equal(gains))
			Expect([]int{song.Lifespan, song.Nature, song.LifeType, song.Guts}).To(Equal([]int{lifespan, nature, lifeType, guts}))
			Expect([]string{song.Arena, song.TrainingBonus, song.ExtraTechs, song.BattleSpecials}).To(Equal([]string{arena, training, techs, specials}))
			Expect(song.Enemy).To(Equal(enemy))
		},
		Entry("Arca by Arca, the baseline of Ape: a Late Bloom with an offset", 309, "Arca", "Arca", true, []int{150, 160, 20, 120, 100, 140}, []int{4, 4, 1, 3, 3, 4}, 400, -45, 2, 18, "E", "Parepare", "", "", false),
		Entry("Sorella by Alfa, a variant of Gold Dust with an extra tech and a special", 944, "Sorella", "Alfa", false, []int{179, 154, 73, 128, 163, 127}, []int{4, 4, 1, 3, 4, 3}, 380, 25, 0, 17, "D", "Parepare", "Boomerang", "Grit", false),
		Entry("Hops by Riley!, King Ape, an enemy", 985, "Hops", "Riley!", false, []int{313, 289, 88, 169, 186, 269}, []int{5, 4, 2, 3, 2, 5}, 360, -95, 2, 13, "D", "Study, Parepare", "Big Bomb", "Fury, Real", true),
		Entry("GP by Gepe, a Sustainable one with an offset", 1221, "GP", "Gepe", true, []int{90, 78, 171, 115, 130, 54}, []int{3, 2, 4, 3, 3, 1}, 420, -90, 3, 11, "B", "Leap", "Heavy Punch", "Hurry, Real", false),
		Entry("Zorn by Zorn, Sueki Suezo, whose lifespan is a week", 482, "Zorn", "Zorn", false, []int{999, 387, 321, 355, 1, 999}, []int{1, 1, 1, 1, 1, 1}, 1, -65, 1, 12, "D", "", "Bite", "Ease", false),
	)

	It("files the baseline of Ape under Ape and Ape", func() {
		ctx := context.Background()
		client, err := testutil.OpenShipped(ctx)
		if errors.Is(err, testutil.ErrNoShippedDB) {
			Skip(err.Error())
		}
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)

		arca := client.Song.Query().Where(entsong.MonsterID(309)).OnlyX(ctx)
		Expect([]int{arca.MainGenus, arca.SubGenus}).To(Equal([]int{35, 35}))
	})

	It("ships the 50 cells (10 life stages by 5 attribute gains) of the Drill Processing page of LegendCup for each result, role and bonus, and the drop of the third stat with no bonus", func() {
		ctx := context.Background()
		client, err := testutil.OpenShipped(ctx)
		if errors.Is(err, testutil.ErrNoShippedDB) {
			Skip(err.Error())
		}
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)

		total, err := client.DrillGain.Query().Count(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(total).To(Equal(1050))
		rows, err := client.DB.QueryContext(ctx, "SELECT result, role, bonus, COUNT(*) FROM drill_gains GROUP BY result, role, bonus")
		Expect(err).NotTo(HaveOccurred())
		defer rows.Close()
		groups := map[string]int{}
		for rows.Next() {
			var result, role string
			var bonus bool
			var n int
			Expect(rows.Scan(&result, &role, &bonus, &n)).To(Succeed())
			groups[fmt.Sprintf("%s %s bonus=%t", result, role, bonus)] = n
		}
		Expect(rows.Err()).NotTo(HaveOccurred())
		Expect(groups).To(HaveLen(21))
		for group, n := range groups {
			Expect(n).To(Equal(50), group)
		}
	})

	DescribeTable("ships a cell as the page lists it, in the order of the page: one list for each choice that the game has, a pair of values where the page gives a pair, and the drop below zero",
		func(result entdrillgain.Result, role entdrillgain.Role, bonus bool, stage, gain int, want [][]int) {
			ctx := context.Background()
			client, err := testutil.OpenShipped(ctx)
			if errors.Is(err, testutil.ErrNoShippedDB) {
				Skip(err.Error())
			}
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(client.Close)

			cell := client.DrillGain.Query().Where(
				entdrillgain.ResultEQ(result), entdrillgain.RoleEQ(role), entdrillgain.Bonus(bonus),
				entdrillgain.LifeStage(stage-1), entdrillgain.GrowthRate(gain-1),
			).OnlyX(ctx)
			Expect(cell.Slots).To(Equal(want))
		},
		Entry("Success, light, no bonus, life stage 1, gain 1", entdrillgain.ResultSuccess, entdrillgain.RoleLight, false, 1, 1, [][]int{{1}, {1}, {2}, {3}}),
		Entry("Success, light, no bonus, life stage 5, gain 5", entdrillgain.ResultSuccess, entdrillgain.RoleLight, false, 5, 5, [][]int{{14}, {15}, {15}, {15}}),
		Entry("Success, light, bonus, life stage 1, gain 1", entdrillgain.ResultSuccess, entdrillgain.RoleLight, true, 1, 1, [][]int{{1}, {2}, {3}, {4}}),
		Entry("Success, heavy primary, no bonus, life stage 1, gain 1", entdrillgain.ResultSuccess, entdrillgain.RolePrimary, false, 1, 1, [][]int{{3}, {3}, {3}, {4}}),
		Entry("Success, heavy primary, no bonus, life stage 5, gain 5", entdrillgain.ResultSuccess, entdrillgain.RolePrimary, false, 5, 5, [][]int{{20}, {20}, {20}, {20}}),
		Entry("Success, heavy primary, bonus, life stage 1, gain 1", entdrillgain.ResultSuccess, entdrillgain.RolePrimary, true, 1, 1, [][]int{{3}, {3}, {4}, {5}}),
		Entry("Success, heavy secondary, no bonus, life stage 1, gain 3, the values not in ascending order", entdrillgain.ResultSuccess, entdrillgain.RoleSecondary, false, 1, 3, [][]int{{2}, {2}, {3}, {2}}),
		Entry("Success, heavy secondary, bonus, life stage 5, gain 5", entdrillgain.ResultSuccess, entdrillgain.RoleSecondary, true, 5, 5, [][]int{{9}, {10}, {10}, {8}}),
		Entry("Success, drop, life stage 1, gain 1", entdrillgain.ResultSuccess, entdrillgain.RoleDrop, false, 1, 1, [][]int{{-2}, {-3}, {-2}, {-3}}),
		Entry("Success, drop, life stage 10, gain 5, the same at every life stage and gain", entdrillgain.ResultSuccess, entdrillgain.RoleDrop, false, 10, 5, [][]int{{-2}, {-3}, {-2}, {-3}}),
		Entry("Cheat, light, no bonus, life stage 1, gain 3", entdrillgain.ResultCheat, entdrillgain.RoleLight, false, 1, 3, [][]int{{1}, {1}, {2}, {1}}),
		Entry("Cheat, light, no bonus, life stage 5, gain 3", entdrillgain.ResultCheat, entdrillgain.RoleLight, false, 5, 3, [][]int{{4}, {5}, {6}, {4}}),
		Entry("Cheat, light, bonus, life stage 1, gain 3", entdrillgain.ResultCheat, entdrillgain.RoleLight, true, 1, 3, [][]int{{1}, {2}, {3}, {1}}),
		Entry("Cheat, heavy primary, no bonus, life stage 1, gain 1", entdrillgain.ResultCheat, entdrillgain.RolePrimary, false, 1, 1, [][]int{{3}, {3}, {3}, {3}}),
		Entry("Cheat, heavy primary, no bonus, life stage 5, gain 3", entdrillgain.ResultCheat, entdrillgain.RolePrimary, false, 5, 3, [][]int{{7}, {8}, {9}, {8}}),
		Entry("Cheat, heavy secondary, no bonus, life stage 5, gain 3", entdrillgain.ResultCheat, entdrillgain.RoleSecondary, false, 5, 3, [][]int{{3}, {4}, {2}, {3}}),
		Entry("Cheat, drop, life stage 1, gain 1", entdrillgain.ResultCheat, entdrillgain.RoleDrop, false, 1, 1, [][]int{{-3}, {-4}, {-1}, {-2}}),
		Entry("Great, light, no bonus, life stage 1, gain 1, five choices", entdrillgain.ResultGreat, entdrillgain.RoleLight, false, 1, 1, [][]int{{2}, {4}, {6}, {8}, {10}}),
		Entry("Great, light, no bonus, life stage 5, gain 5", entdrillgain.ResultGreat, entdrillgain.RoleLight, false, 5, 5, [][]int{{15}, {15}, {15}, {15}, {15}}),
		Entry("Great, light, bonus, life stage 1, gain 1", entdrillgain.ResultGreat, entdrillgain.RoleLight, true, 1, 1, [][]int{{3}, {5}, {7}, {9}, {11}}),
		Entry("Great, heavy primary, no bonus, life stage 1, gain 1, two choices of two values", entdrillgain.ResultGreat, entdrillgain.RolePrimary, false, 1, 1, [][]int{{4, 8}, {6, 10}}),
		Entry("Great, heavy primary, no bonus, life stage 3, gain 5", entdrillgain.ResultGreat, entdrillgain.RolePrimary, false, 3, 5, [][]int{{18, 20}, {20, 20}}),
		Entry("Great, heavy primary, bonus, life stage 1, gain 1", entdrillgain.ResultGreat, entdrillgain.RolePrimary, true, 1, 1, [][]int{{5, 9}, {7, 11}}),
		Entry("Great, heavy secondary, no bonus, life stage 1, gain 5", entdrillgain.ResultGreat, entdrillgain.RoleSecondary, false, 1, 5, [][]int{{2, 4}, {2, 3}}),
		Entry("Great, heavy secondary, bonus, life stage 5, gain 5", entdrillgain.ResultGreat, entdrillgain.RoleSecondary, true, 5, 5, [][]int{{9, 10}, {8, 10}}),
		Entry("Great, drop, life stage 1, gain 1, two choices", entdrillgain.ResultGreat, entdrillgain.RoleDrop, false, 1, 1, [][]int{{-2}, {-3}}),
	)
})
