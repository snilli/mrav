package plannerrepo_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	enteffect "orm/ent/effect"
	entfood "orm/ent/foodreaction"

	monster "app/internal/core/monster/domain"
	"app/internal/core/planner"
	"app/internal/core/planner/domain"
	"app/internal/core/planner/repository/db/plannerrepo"
	"app/internal/testutil"
)

var _ = Describe("plannerrepo", Label("integration"), func() {
	var ctx context.Context

	BeforeEach(func() { ctx = context.Background() })

	It("lists the reactions ordered by genus, food and age, as the domain has them", func() {
		client, err := testutil.NewClient(ctx)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)
		client.FoodReaction.Create().SetGenus(7).SetFood(entfood.FoodTablet).SetStage(entfood.StageAny).SetReaction(entfood.ReactionDislike).SaveX(ctx)
		client.FoodReaction.Create().SetGenus(3).SetFood(entfood.FoodMilk).SetStage(entfood.StageGrown).SetReaction(entfood.ReactionLike).SaveX(ctx)
		client.FoodReaction.Create().SetGenus(3).SetFood(entfood.FoodMilk).SetStage(entfood.StageBaby).SetReaction(entfood.ReactionNeutral).SaveX(ctx)
		client.FoodReaction.Create().SetGenus(3).SetFood(entfood.FoodCupJelly).SetStage(entfood.StageAny).SetReaction(entfood.ReactionLike).SaveX(ctx)
		var repo planner.PlannerRepository = plannerrepo.New(client)

		got, err := repo.ListFoodReactions(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal([]domain.FoodReaction{
			{Genus: 3, Food: domain.CupJelly, Stage: domain.AnyStage, Reaction: domain.Like},
			{Genus: 3, Food: domain.Milk, Stage: domain.Baby, Reaction: domain.Neutral},
			{Genus: 3, Food: domain.Milk, Stage: domain.Grown, Reaction: domain.Like},
			{Genus: 7, Food: domain.Tablet, Stage: domain.AnyStage, Reaction: domain.Dislike},
		}))
	})

	It("lists the effects ordered by subject, game, genus and life stage, and fails on a subject that the planner does not know", func() {
		client, err := testutil.NewClient(ctx)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)
		client.Effect.Create().SetSubject("item:nuts_oil").SetVersion(enteffect.VersionMr2).SetGenus(5).SetStressFactor(0.8).SetFatigue(-28).SetStress(0).SetLost(0).SetForm(0).SaveX(ctx)
		client.Effect.Create().SetSubject("item:nuts_oil").SetVersion(enteffect.VersionMr2).SetFatigue(-28).SetStress(0).SetLost(0).SetForm(0).SaveX(ctx)
		client.Effect.Create().SetSubject("action:battle_win").SetVersion(enteffect.VersionMr2dx).SetFatigue(30).SetStress(-45).SetFormDivisor(13).SetLost(0).SetForm(0).SaveX(ctx)
		client.Effect.Create().SetSubject("action:rest").SetVersion(enteffect.VersionMr2).SetLifeStage(1).SetFatigue(-36).SetStress(-5).SetLost(0).SetForm(0).SaveX(ctx)
		repo := plannerrepo.New(client)

		got, err := repo.ListEffects(ctx)
		Expect(err).NotTo(HaveOccurred())
		five, one, thirteen, factor := 5, 1, 13, 0.8
		Expect(got).To(Equal([]domain.Effect{
			{Subject: "action:battle_win", Version: monster.MR2DX, Fatigue: 30, Stress: -45, FormDivisor: &thirteen},
			{Subject: "action:rest", Version: monster.MR2, LifeStage: &one, Fatigue: -36, Stress: -5},
			{Subject: "item:nuts_oil", Version: monster.MR2, Fatigue: -28},
			{Subject: "item:nuts_oil", Version: monster.MR2, Genus: &five, StressFactor: &factor, Fatigue: -28},
		}))

		client.Effect.Create().SetSubject("item:pile_of_meat").SetVersion(enteffect.VersionMr2).SetFatigue(0).SetStress(0).SetLost(0).SetForm(12).SaveX(ctx)
		_, err = plannerrepo.New(client).ListEffects(ctx)
		Expect(err).To(MatchError(ContainSubstring("item:pile_of_meat")))
	})

	It("fails on two effects for the same subject, game, genus and life stage, which the unique index lets in when the genus or the stage is empty", func() {
		client, err := testutil.NewClient(ctx)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)
		client.Effect.Create().SetSubject("item:mango").SetVersion(enteffect.VersionMr2).SetFatigue(-10).SetStress(0).SetLost(0).SetForm(1).SaveX(ctx)
		client.Effect.Create().SetSubject("item:mango").SetVersion(enteffect.VersionMr2).SetFatigue(-12).SetStress(0).SetLost(0).SetForm(1).SaveX(ctx)

		_, err = plannerrepo.New(client).ListEffects(ctx)
		Expect(err).To(MatchError(ContainSubstring("item:mango")))
	})

	It("lists none from empty tables, without failing", func() {
		client, err := testutil.NewClient(ctx)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)
		repo := plannerrepo.New(client)

		reactions, err := repo.ListFoodReactions(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(reactions).To(BeEmpty())
		effects, err := repo.ListEffects(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(effects).To(BeEmpty())
	})

	Describe("the shipped database", func() {
		var repo planner.PlannerRepository

		BeforeEach(func() {
			client, err := testutil.OpenShipped(ctx)
			if errors.Is(err, testutil.ErrNoShippedDB) {
				Skip(err.Error())
			}
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(client.Close)
			repo = plannerrepo.New(client)
		})

		It("has what each of the 38 genera feels about the six foods: five foods for any age and the milk for a baby and for a grown monster", func() {
			got, err := repo.ListFoodReactions(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(HaveLen(38 * 7))
		})

		DescribeTable("has the effects that the Training Planner and the items and drills pages of LegendCup give",
			func(subject string, version monster.Version, main, stage int, want string) {
				got, err := repo.ListEffects(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(got).To(HaveLen(85))
				s := domain.Settings{Version: version, Main: main, Sub: main, LifeStage: stage}
				Expect(domain.NewEffects(got).Text(itemNamed(subject), s)).To(Equal(want))
			},
			Entry("a mint leaf", "Mint Leaf", monster.MR2, 1, 3, "Stress x0.5, Form -5"),
			Entry("nuts oil for any genus", "Nuts Oil", monster.MR2, 1, 3, "Fatigue -28"),
			Entry("nuts oil for a Naga", "Nuts Oil", monster.MR2DX, 9, 3, "Fatigue -28, Stress x0.8"),
			Entry("the banana of +10 loyalty on the PlayStation", "MB +10", monster.MR2, 1, 3, "Life -1 week, Form -1"),
			Entry("the banana of +10 loyalty on Steam", "MB +10", monster.MR2DX, 1, 3, "Stress -10, Form -1"),
			Entry("the banana of +0 loyalty on the PlayStation", "MB +0", monster.MR2, 1, 3, "no change"),
			Entry("powder", "Powder", monster.MR2DX, 1, 3, "Fatigue +5, Form -24"),
			Entry("a griever", "Griever", monster.MR2, 1, 3, "Fatigue +10, Stress x0.6, Form -1"),
			Entry("a gold peach", "Gold Peach", monster.MR2DX, 1, 3, "Life +50 weeks"),
		)

		DescribeTable("has what the lists of the page say, for the foods of the month",
			func(genus int, food domain.Food, lifeStage int, want domain.Reaction) {
				got, err := repo.ListFoodReactions(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(domain.NewReactions(got).Of(genus, food, lifeStage)).To(Equal(want))
			},
			Entry("Baku dislikes a tablet", 14, domain.Tablet, 3, domain.Dislike),
			Entry("Pixie is neutral to a tablet", 0, domain.Tablet, 3, domain.Neutral),
			Entry("Centaur likes a tablet", 2, domain.Tablet, 3, domain.Like),
			Entry("Baku likes meat", 14, domain.Meat, 3, domain.Like),
			Entry("Kato is neutral to meat", 16, domain.Meat, 3, domain.Neutral),
			Entry("Pixie is not for meat, which the page does not give it", 0, domain.Meat, 3, domain.Dislike),
			Entry("Golem dislikes a cup jelly", 7, domain.CupJelly, 3, domain.Dislike),
			Entry("ColorPandora is neutral to a cup jelly", 3, domain.CupJelly, 3, domain.Neutral),
			Entry("Pixie likes a cup jelly", 0, domain.CupJelly, 3, domain.Like),
			Entry("Ape dislikes fish", 35, domain.Fish, 3, domain.Dislike),
			Entry("Monol is neutral to fish", 34, domain.Fish, 3, domain.Neutral),
			Entry("Pixie likes fish", 0, domain.Fish, 3, domain.Like),
			Entry("Kato is neutral to the milk of a baby", 16, domain.Milk, 1, domain.Neutral),
			Entry("Baku likes the milk of a baby", 14, domain.Milk, 1, domain.Like),
			Entry("Kato dislikes the milk of a grown monster", 16, domain.Milk, 2, domain.Dislike),
			Entry("Baku is neutral to the milk of a grown monster", 14, domain.Milk, 2, domain.Neutral),
			Entry("ColorPandora likes the milk of a grown monster", 3, domain.Milk, 2, domain.Like),
			Entry("Ape dislikes potato", 35, domain.Potato, 3, domain.Dislike),
			Entry("Plant is neutral to potato", 33, domain.Potato, 3, domain.Neutral),
			Entry("ColorPandora likes potato", 3, domain.Potato, 3, domain.Like),
		)
	})
})

func itemNamed(name string) domain.Item {
	for _, it := range domain.Items() {
		if it.Name() == name {
			return it
		}
	}
	Fail("no item called " + name)
	return domain.NoItem
}
