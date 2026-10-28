package plannersvc_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	"app/internal/core/planner"
	"app/internal/core/planner/domain"
	"app/internal/core/planner/service/plannersvc"
	player "app/internal/core/player/domain"
	mockplanner "app/mock/core/planner"
	mockport "app/mock/port"
)

var _ = Describe("PlannerService", func() {
	var (
		ctx      context.Context
		repo     *mockplanner.MockPlannerRepository
		monsters *mockport.MockMonsterSource
		players  *mockport.MockPlayerSource
		svc      planner.PlannerService
	)

	BeforeEach(func() {
		ctx = context.Background()
		repo = mockplanner.NewMockPlannerRepository(GinkgoT())
		monsters = mockport.NewMockMonsterSource(GinkgoT())
		players = mockport.NewMockPlayerSource(GinkgoT())
		svc = plannersvc.New(repo, monsters, players)
	})

	It("starts the plan from the monster and the items of the game, with what each genus feels about food", func() {
		monsters.EXPECT().Current(ctx).Return(monster.Monster{Genus: 35, SubGenus: 7, Fatigue: 20, LifeStage: 2}, nil)
		var items [player.ItemSlots]int
		items[0] = 7
		players.EXPECT().Player(ctx).Return(player.Player{Items: items}, nil)
		repo.EXPECT().ListFoodReactions(ctx).Return([]domain.FoodReaction{
			{Genus: 35, Food: domain.Potato, Stage: domain.AnyStage, Reaction: domain.Dislike},
		}, nil)
		repo.EXPECT().ListEffects(ctx).Return([]domain.Effect{{Subject: "item:mango", Version: monster.MR2DX, Fatigue: -10, Form: 1}}, nil)

		got, err := svc.Current(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Settings).To(Equal(domain.SettingsOf(monster.Monster{Genus: 35, SubGenus: 7, Fatigue: 20, LifeStage: 2}, items)))
		Expect(got.Settings.Pots).To(Equal(1))
		Expect(got.Reactions.Of(35, domain.Potato, 3)).To(Equal(domain.Dislike))
		Expect(got.Effects).To(Equal(domain.NewEffects([]domain.Effect{{Subject: "item:mango", Version: monster.MR2DX, Fatigue: -10, Form: 1}})))
	})

	It("returns the error of the monster without asking for the rest", func() {
		monsters.EXPECT().Current(ctx).Return(monster.Monster{}, errors.New("no game"))

		_, err := svc.Current(ctx)
		Expect(err).To(MatchError(ContainSubstring("no game")))
	})

	It("returns the error of the player", func() {
		monsters.EXPECT().Current(ctx).Return(monster.Monster{}, nil)
		players.EXPECT().Player(ctx).Return(player.Player{}, errors.New("no ranch"))

		_, err := svc.Current(ctx)
		Expect(err).To(MatchError(ContainSubstring("no ranch")))
	})

	It("returns the error of the reactions", func() {
		monsters.EXPECT().Current(ctx).Return(monster.Monster{}, nil)
		players.EXPECT().Player(ctx).Return(player.Player{}, nil)
		repo.EXPECT().ListFoodReactions(ctx).Return(nil, errors.New("table gone"))

		_, err := svc.Current(ctx)
		Expect(err).To(MatchError(ContainSubstring("table gone")))
	})

	It("returns the error of the effects", func() {
		monsters.EXPECT().Current(ctx).Return(monster.Monster{}, nil)
		players.EXPECT().Player(ctx).Return(player.Player{}, nil)
		repo.EXPECT().ListFoodReactions(ctx).Return(nil, nil)
		repo.EXPECT().ListEffects(ctx).Return(nil, errors.New("effects gone"))

		_, err := svc.Current(ctx)
		Expect(err).To(MatchError(ContainSubstring("effects gone")))
	})
})
