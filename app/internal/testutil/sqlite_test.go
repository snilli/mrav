package testutil_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"orm"
	"orm/ent"
	entgenus "orm/ent/genus"
	entmove "orm/ent/move"
	"orm/stored"

	"app/internal/testutil"
)

var _ = Describe("NewClient", func() {
	var (
		ctx    context.Context
		client *orm.Client
		pixie  *ent.Genus
	)

	BeforeEach(func() {
		ctx = context.Background()
		var err error
		client, err = testutil.NewClient(ctx)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)
		pixie = client.Genus.Create().SetCode(0).SetName("Pixie").SaveX(ctx)
	})

	newMove := func(genusID, slot, col, row int) *ent.MoveCreate {
		return client.Move.Create().
			SetGenusID(genusID).SetVersion(entmove.VersionMr2dx).SetMemorySlot(slot).SetCol(col).SetRow(row).
			SetName("Pat").SetKind(entmove.KindPower).
			SetGuts(12).SetForce(6).SetHit(10).SetWither(0).SetCrit(0).
			SetDescription("Starting basic tech.").SetUnlockNever(false).
			SetUnlock([]stored.Condition{}).SetIcon("Pixie_1")
	}

	It("keeps a genus whose code is 0", func() {
		got := client.Genus.Query().Where(entgenus.Code(0)).OnlyX(ctx)
		Expect(got.Name).To(Equal("Pixie"))
	})

	It("rejects a second move in the same memory slot of a genus", func() {
		newMove(pixie.ID, 1, 1, 1).SaveX(ctx)
		_, err := newMove(pixie.ID, 1, 1, 2).Save(ctx)
		Expect(ent.IsConstraintError(err)).To(BeTrue(), "got %v", err)
	})

	It("allows the same memory slot in another genus", func() {
		dragon := client.Genus.Create().SetCode(1).SetName("Dragon").SaveX(ctx)
		newMove(pixie.ID, 1, 1, 1).SaveX(ctx)
		_, err := newMove(dragon.ID, 1, 1, 1).Save(ctx)
		Expect(err).NotTo(HaveOccurred())
	})

	It("rejects a second move in the same cell of a genus", func() {
		newMove(pixie.ID, 1, 1, 1).SaveX(ctx)
		_, err := newMove(pixie.ID, 2, 1, 1).Save(ctx)
		Expect(ent.IsConstraintError(err)).To(BeTrue(), "got %v", err)
	})

	DescribeTable("rejects a cell outside the 4x6 grid",
		func(col, row int) {
			_, err := newMove(pixie.ID, 1, col, row).Save(ctx)
			Expect(ent.IsValidationError(err)).To(BeTrue(), "got %v", err)
		},
		Entry("column 5", 5, 1),
		Entry("column 0", 0, 1),
		Entry("row 7", 1, 7),
		Entry("row 0", 1, 0),
	)

	It("rejects a move of a genus that does not exist (foreign keys are on)", func() {
		_, err := newMove(9999, 1, 1, 1).Save(ctx)
		Expect(ent.IsConstraintError(err)).To(BeTrue(), "got %v", err)
	})

	It("round-trips the JSON columns", func() {
		unlock := []stored.Condition{
			{Type: stored.StatSum, Stats: []int{1, 2}, Min: new(150)},
			{Type: stored.MoveUsed, Slot: new(3), Min: new(30)},
			{Type: stored.SubGenusIn, IDs: []int{2, 5}},
		}
		extra := &stored.Extra{Label: "Recovery:", Raw: 30}
		id := newMove(pixie.ID, 1, 1, 1).SetUnlock(unlock).SetExtra(extra).SaveX(ctx).ID

		got := client.Move.GetX(ctx, id)
		Expect(got.Unlock).To(Equal(unlock))
		Expect(got.Extra).To(Equal(extra))
	})
})
