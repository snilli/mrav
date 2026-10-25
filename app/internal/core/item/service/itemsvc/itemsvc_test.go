package itemsvc_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/apperr"
	"app/internal/core/item"
	"app/internal/core/item/domain"
	"app/internal/core/item/service/itemsvc"
	monster "app/internal/core/monster/domain"
	player "app/internal/core/player/domain"
	mockitem "app/mock/core/item"
	mockport "app/mock/port"
)

var _ = Describe("ItemService", func() {
	var (
		ctx     context.Context
		repo    *mockitem.MockItemRepository
		players *mockport.MockPlayerSource
		svc     item.ItemService
	)
	potato := domain.Item{Code: 0, Name: "Potato"}
	banana := domain.Item{Code: 29, Name: "Magic Banana"}

	BeforeEach(func() {
		ctx = context.Background()
		repo = mockitem.NewMockItemRepository(GinkgoT())
		players = mockport.NewMockPlayerSource(GinkgoT())
		svc = itemsvc.New(repo, players)
	})

	It("fills the twenty slots with the items the player carries", func() {
		p := player.Player{}
		for slot := range p.Items {
			p.Items[slot] = 178
		}
		p.Items[0], p.Items[3] = 29, 0
		players.EXPECT().Player(ctx).Return(p, nil)
		repo.EXPECT().ListItems(ctx, monster.MR2DX).Return([]domain.Item{potato, banana}, nil)

		slots, err := svc.Inventory(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(slots).To(HaveLen(player.ItemSlots))
		Expect(slots[0].Item.Name).To(Equal("Magic Banana"))
		Expect(slots[3].Item.Name).To(Equal("Potato"))
		Expect(slots[1].Item).To(BeNil())
	})

	It("names the items as the game the player plays has them", func() {
		players.EXPECT().Player(ctx).Return(player.Player{Version: monster.MR2}, nil)
		repo.EXPECT().ListItems(ctx, monster.MR2).Return(nil, nil)

		_, err := svc.Inventory(ctx)
		Expect(err).NotTo(HaveOccurred())
	})

	It("locks the places the house has no room for, when the game says how far it is upgraded", func() {
		players.EXPECT().Player(ctx).Return(player.Player{Ranch: player.Ranch{Known: true, House: 1}}, nil)
		repo.EXPECT().ListItems(ctx, monster.MR2DX).Return(nil, nil)

		slots, err := svc.Inventory(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(slots[13].Locked).To(BeFalse())
		Expect(slots[14].Locked).To(BeTrue())
	})

	It("locks nothing when the game does not say", func() {
		players.EXPECT().Player(ctx).Return(player.Player{}, nil)
		repo.EXPECT().ListItems(ctx, monster.MR2DX).Return(nil, nil)

		slots, err := svc.Inventory(ctx)
		Expect(err).NotTo(HaveOccurred())
		for _, s := range slots {
			Expect(s.Locked).To(BeFalse())
		}
	})

	It("returns the source's error without asking the repository", func() {
		players.EXPECT().Player(ctx).Return(player.Player{}, apperr.NotFound("game", "DuckStation is not running"))

		_, err := svc.Inventory(ctx)
		Expect(errors.Is(err, apperr.ErrNotFound)).To(BeTrue())
	})

	It("returns the repository's error when it cannot list the items", func() {
		players.EXPECT().Player(ctx).Return(player.Player{}, nil)
		repo.EXPECT().ListItems(ctx, monster.MR2DX).Return(nil, errors.New("disk gone"))

		_, err := svc.Inventory(ctx)
		Expect(err).To(MatchError(ContainSubstring("disk gone")))
	})
})
