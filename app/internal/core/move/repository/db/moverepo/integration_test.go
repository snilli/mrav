package moverepo_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"orm/ent"
	entmove "orm/ent/move"
	"orm/stored"

	"app/internal/core/apperr"
	monster "app/internal/core/monster/domain"
	"app/internal/core/move"
	movedomain "app/internal/core/move/domain"
	"app/internal/core/move/repository/db/moverepo"
	"app/internal/testutil"
)

var _ = Describe("moverepo", Label("integration"), func() {
	var (
		ctx  context.Context
		repo move.MoveRepository
	)

	pat := movedomain.Move{
		Genus: 0, MemorySlot: 0, Col: 1, Row: 1, Name: "Pat", Kind: movedomain.KindPower,
		Guts: 12, Force: 6, Hit: 10, Description: "Starting basic tech.", Icon: "Pixie_1",
	}
	refreshment := movedomain.Move{
		Genus: 0, MemorySlot: 13, Col: 3, Row: 2, Name: "Refreshment", Kind: movedomain.KindIntelligence,
		Guts: 30, Hit: 1, Description: "d", Icon: "Pixie_12",
		Extra: &movedomain.Extra{Label: "Recovery:", Scale: new(movedomain.ScaleForce), Raw: 30},
		Unlock: movedomain.UnlockRule{All: []movedomain.Condition{
			movedomain.MoveUsedAtLeast{Slot: 12, Min: 50},
			movedomain.StatSumAtLeast{Stats: []monster.Stat{monster.Pow, monster.Int}, Min: 400},
			movedomain.NatureAtMost{Max: -50},
			movedomain.SubGenusNotIn{IDs: []monster.Genus{24, 5}},
		}},
	}
	never := movedomain.Move{
		Genus: 1, MemorySlot: 0, Col: 1, Row: 1, Name: "Never", Kind: movedomain.KindUnknown,
		Description: "d", Icon: "Drago_1", Unlock: movedomain.UnlockRule{Never: true},
	}

	BeforeEach(func() {
		ctx = context.Background()
		client, err := testutil.NewClient(ctx)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)

		// Seed the way the database holds its rows: the unlock rule and the extra row as their stored JSON
		// shapes, written by hand, so what the repository returns is what its mapper makes of them.
		pixie := client.Genus.Create().SetCode(0).SetName("Pixie").SaveX(ctx)
		dragon := client.Genus.Create().SetCode(1).SetName("Dragon").SaveX(ctx)
		seed := func(genusID int, m movedomain.Move) *ent.MoveCreate {
			return client.Move.Create().
				SetGenusID(genusID).SetMemorySlot(m.MemorySlot).SetCol(m.Col).SetRow(m.Row).
				SetName(m.Name).SetKind(entmove.Kind(m.Kind)).
				SetGuts(m.Guts).SetForce(m.Force).SetHit(m.Hit).SetWither(m.Wither).SetCrit(m.Crit).
				SetDescription(m.Description).SetIcon(m.Icon).
				SetUnlockNever(m.Unlock.Never).SetUnlock([]stored.Condition{})
		}
		seed(pixie.ID, refreshment).
			SetUnlock([]stored.Condition{
				{Type: stored.MoveUsed, Slot: new(12), Min: new(50)},
				{Type: stored.StatSum, Stats: []int{1, 2}, Min: new(400)},
				{Type: stored.NatureAtMost, Max: new(-50)},
				{Type: stored.SubGenusNotIn, IDs: []int{24, 5}},
			}).
			SetExtra(&stored.Extra{Label: "Recovery:", Scale: new(0), Raw: 30}).
			SaveX(ctx)
		seed(dragon.ID, never).SaveX(ctx)
		seed(pixie.ID, pat).SaveX(ctx)
		repo = moverepo.New(client)
	})

	It("lists the genera ordered by code", func() {
		got, err := repo.ListGenera(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal([]movedomain.Genus{{Code: 0, Name: "Pixie"}, {Code: 1, Name: "Dragon"}}))
	})

	It("gets one genus, and says not found for another", func() {
		got, err := repo.GetGenus(ctx, 1)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(movedomain.Genus{Code: 1, Name: "Dragon"}))

		_, err = repo.GetGenus(ctx, 9)
		Expect(errors.Is(err, apperr.ErrNotFound)).To(BeTrue())
	})

	It("reads the moves of a genus ordered by memory slot, with their unlock rules and extra rows", func() {
		got, err := repo.ListMoves(ctx, 0)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal([]movedomain.Move{pat, refreshment}))
	})

	It("reads a move the game gives no way to learn", func() {
		got, err := repo.ListMoves(ctx, 1)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal([]movedomain.Move{never}))
	})

	It("lists no moves for a genus it does not have, without failing", func() {
		got, err := repo.ListMoves(ctx, 9)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(BeEmpty())
	})
})
