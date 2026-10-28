package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	"app/internal/core/planner/domain"
	player "app/internal/core/player/domain"
)

var _ = Describe("SettingsOf", func() {
	const (
		sculpture  = 6
		geminisPot = 7
		empty      = 178
	)

	inventory := func(codes ...int) [player.ItemSlots]int {
		var items [player.ItemSlots]int
		for i := range items {
			items[i] = empty
		}
		copy(items[:], codes)
		return items
	}

	It("starts the plan from the monster of the game: its game, breed, form, fatigue and stress", func() {
		m := monster.Monster{Version: monster.MR2, Genus: 7, SubGenus: 9, Form: -35, Fatigue: 15, Stress: 25, LifeStage: 3, ItemLike: 5, ItemDislike: 2}

		Expect(domain.SettingsOf(m, inventory())).To(Equal(domain.Settings{
			Version: monster.MR2, Main: 7, Sub: 9, Like: domain.Tablet, Dislike: domain.Meat,
			LifeStage: 4, Form: -35, Fatigue: 15, Stress: 25,
		}))
	})

	It("counts the life stage from one, as the page does, where the game counts from nought, and keeps it from one to ten", func() {
		stage := func(n int) int { return domain.SettingsOf(monster.Monster{LifeStage: n}, inventory()).LifeStage }
		Expect(stage(0)).To(Equal(1))
		Expect(stage(9)).To(Equal(10))
		Expect(stage(-1)).To(Equal(1))
		Expect(stage(12)).To(Equal(10))
	})

	It("gives a baby the milk of a baby, so that the first stage of the game is the first of the page", func() {
		reactions := domain.NewReactions([]domain.FoodReaction{
			{Genus: 14, Food: domain.Milk, Stage: domain.Baby, Reaction: domain.Like},
			{Genus: 14, Food: domain.Milk, Stage: domain.Grown, Reaction: domain.Neutral},
		})
		s := domain.SettingsOf(monster.Monster{Genus: 14, LifeStage: 0}, inventory())
		Expect(reactions.Of(s.Main, domain.Milk, s.LifeStage)).To(Equal(domain.Like))
	})

	DescribeTable("takes the item that the monster likes or dislikes to a food of the month by its code in the game",
		func(code int, want domain.Food) {
			s := domain.SettingsOf(monster.Monster{ItemLike: code, ItemDislike: code}, inventory())
			Expect([]domain.Food{s.Like, s.Dislike}).To(Equal([]domain.Food{want, want}))
		},
		Entry("potato", 0, domain.Potato),
		Entry("fish", 1, domain.Fish),
		Entry("meat", 2, domain.Meat),
		Entry("milk", 3, domain.Milk),
		Entry("cup jelly", 4, domain.CupJelly),
		Entry("tablet", 5, domain.Tablet),
		Entry("a sculpture is not a food", 6, domain.NoFood),
		Entry("a mint leaf is not a food", 30, domain.NoFood),
		Entry("an empty slot is not a food", empty, domain.NoFood),
	)

	It("takes a monster read without a game to be one of DX, as the rest of the app does", func() {
		Expect(domain.SettingsOf(monster.Monster{}, inventory()).Version).To(Equal(monster.MR2DX))
	})

	It("knows which peach the monster in the game has eaten already", func() {
		s := domain.SettingsOf(monster.Monster{SilverPeach: true}, inventory())
		Expect([]bool{s.GoldPeach, s.SilverPeach}).To(Equal([]bool{false, true}))
	})

	It("counts the Gemini's Pots and the sculptures that are kept", func() {
		s := domain.SettingsOf(monster.Monster{}, inventory(geminisPot, 0, sculpture, geminisPot, 30, sculpture, geminisPot))
		Expect([]int{s.Pots, s.Sculptures}).To(Equal([]int{3, 2}))
		Expect(domain.SettingsOf(monster.Monster{}, inventory()).Pots).To(BeZero())
	})
})
