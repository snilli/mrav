package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/planner/domain"
)

var _ = Describe("food", func() {
	It("lists the six foods in the order of the page, which is the order of its choices", func() {
		Expect(domain.Foods()).To(Equal([]domain.Food{domain.Potato, domain.Milk, domain.Fish, domain.CupJelly, domain.Meat, domain.Tablet}))
	})

	It("keeps the code that the game gives each food, which a monster's like and dislike are told in", func() {
		Expect([]int{int(domain.Potato), int(domain.Fish), int(domain.Meat), int(domain.Milk), int(domain.CupJelly), int(domain.Tablet)}).To(Equal([]int{0, 1, 2, 3, 4, 5}))
	})

	It("names each food, and finds it again from its key, which is what the database keeps", func() {
		for _, food := range domain.Foods() {
			parsed, ok := domain.ParseFood(food.Key())
			Expect(ok).To(BeTrue(), food.Name())
			Expect(parsed).To(Equal(food))
		}
		Expect(domain.CupJelly.Name()).To(Equal("Cup Jelly"))
		Expect(domain.CupJelly.Key()).To(Equal("cup_jelly"))
		_, ok := domain.ParseFood("battle")
		Expect(ok).To(BeFalse())
		Expect(domain.NoFood.Name()).To(Equal("Other"))
	})

	It("finds a reaction and an age again from its key", func() {
		for _, r := range []domain.Reaction{domain.Like, domain.Neutral, domain.Dislike} {
			parsed, ok := domain.ParseReaction(r.Key())
			Expect(ok).To(BeTrue())
			Expect(parsed).To(Equal(r))
		}
		for _, s := range []domain.Stage{domain.AnyStage, domain.Baby, domain.Grown} {
			parsed, ok := domain.ParseStage(s.Key())
			Expect(ok).To(BeTrue())
			Expect(parsed).To(Equal(s))
		}
		_, ok := domain.ParseReaction("love")
		Expect(ok).To(BeFalse())
		_, ok = domain.ParseStage("old")
		Expect(ok).To(BeFalse())
	})

	It("changes the stress by what the page says for each food and reaction", func() {
		table := map[domain.Food][3]int{
			domain.Potato: {0, 4, 16}, domain.Fish: {-6, -3, 2}, domain.Meat: {-8, -6, 0},
			domain.Milk: {-3, -2, 4}, domain.CupJelly: {-7, -5, 1}, domain.Tablet: {-15, -13, -10},
		}
		for food, want := range table {
			Expect([3]int{domain.FoodStress(food, domain.Like), domain.FoodStress(food, domain.Neutral), domain.FoodStress(food, domain.Dislike)}).To(Equal(want), food.Name())
		}
	})

	Describe("what a genus feels about a food", func() {
		reactions := domain.NewReactions([]domain.FoodReaction{
			{Genus: 7, Food: domain.Tablet, Stage: domain.AnyStage, Reaction: domain.Dislike},
			{Genus: 7, Food: domain.Milk, Stage: domain.Baby, Reaction: domain.Neutral},
			{Genus: 7, Food: domain.Milk, Stage: domain.Grown, Reaction: domain.Dislike},
		})

		It("tells it for any age when the food does not depend on age", func() {
			Expect(reactions.Of(7, domain.Tablet, 1)).To(Equal(domain.Dislike))
			Expect(reactions.Of(7, domain.Tablet, 6)).To(Equal(domain.Dislike))
		})

		It("tells the milk of a baby, the first life stage, from that of a grown monster", func() {
			Expect(reactions.Of(7, domain.Milk, 1)).To(Equal(domain.Neutral))
			Expect(reactions.Of(7, domain.Milk, 2)).To(Equal(domain.Dislike))
			Expect(reactions.Of(7, domain.Milk, 10)).To(Equal(domain.Dislike))
		})

		It("falls back, for a genus that it has nothing for, to what the page does for the rest: like, and dislike for meat", func() {
			Expect(reactions.Of(38, domain.Potato, 1)).To(Equal(domain.Like))
			Expect(reactions.Of(38, domain.Tablet, 1)).To(Equal(domain.Like))
			Expect(reactions.Of(38, domain.Meat, 1)).To(Equal(domain.Dislike))
		})
	})
})
