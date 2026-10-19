package profilerepo_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"orm"
	"orm/ent"

	monster "app/internal/core/monster/domain"
	"app/internal/core/profile"
	"app/internal/core/profile/domain"
	"app/internal/core/profile/repository/db/profilerepo"
	"app/internal/testutil"
)

var _ = Describe("profilerepo", Label("integration"), func() {
	var (
		ctx    context.Context
		client *orm.Client
		repo   profile.ProfileRepository
	)

	BeforeEach(func() {
		ctx = context.Background()
		var err error
		client, err = testutil.NewClient(ctx)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)
		repo = profilerepo.New(client)
	})

	It("lists the genus names by code", func() {
		client.Genus.Create().SetCode(0).SetName("Pixie").SaveX(ctx)
		client.Genus.Create().SetCode(5).SetName("Henger").SaveX(ctx)

		got, err := repo.ListGenusNames(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(map[monster.Genus]string{0: "Pixie", 5: "Henger"}))
	})

	It("lists the breed rules of one main genus, ordered by sub genus and then by their order", func() {
		seed := func(main, sub, seq int, name string) *ent.BreedCreate {
			return client.Breed.Create().SetMainGenus(main).SetSubGenus(sub).SetSeq(seq).SetName(name)
		}
		seed(1, 38, 1, "Moo (CD)").SetOtherwise("Moo (N/S)").SetGuts(18).SetMatchGuts(true).SetGrowth([]int{3, 3, 0, 1, 1, 3}).SaveX(ctx)
		seed(1, 0, 0, "Tiamat").SetOtherwise("Tiamat (N/S)").SetGuts(14).SetGrowth([]int{1, 3, 4, 2, 2, 1}).SaveX(ctx)
		seed(1, 38, 0, "Moo (DNA Capsule)").SetOtherwise("Moo (N/S)").SetGuts(19).SetMatchGuts(true).SetGrowth([]int{2, 4, 4, 2, 1, 2}).SaveX(ctx)
		seed(1, 39, 0, "Kasumi").SaveX(ctx) // a fixed name: no guts rate, no growth rates
		seed(0, 0, 0, "Pixie").SaveX(ctx)   // another main genus

		got, err := repo.ListBreedRules(ctx, 1)
		Expect(err).NotTo(HaveOccurred())
		tiamat := [monster.StatCount]int{1, 3, 4, 2, 2, 1}
		capsule := [monster.StatCount]int{2, 4, 4, 2, 1, 2}
		cd := [monster.StatCount]int{3, 3, 0, 1, 1, 3}
		ints := func(n int) *int { return &n }
		Expect(got).To(Equal([]domain.BreedRule{
			{Main: 1, Sub: 0, Name: "Tiamat", Otherwise: "Tiamat (N/S)", Guts: ints(14), Growth: &tiamat},
			{Main: 1, Sub: 38, Name: "Moo (DNA Capsule)", Otherwise: "Moo (N/S)", Guts: ints(19), MatchGuts: true, Growth: &capsule},
			{Main: 1, Sub: 38, Name: "Moo (CD)", Otherwise: "Moo (N/S)", Guts: ints(18), MatchGuts: true, Growth: &cd},
			{Main: 1, Sub: 39, Name: "Kasumi"},
		}))
	})

	It("lists none for a main genus it has no rules for, without failing", func() {
		got, err := repo.ListBreedRules(ctx, 9)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(BeEmpty())
	})

	It("fails, saying which breed, for a rule whose growth rates are not six", func() {
		client.Breed.Create().SetMainGenus(2).SetSubGenus(7).SetSeq(0).SetName("Bad").SetGrowth([]int{1, 2}).SaveX(ctx)

		_, err := repo.ListBreedRules(ctx, 2)
		Expect(err).To(MatchError(And(ContainSubstring("breed 2/7"), ContainSubstring("2 rates"))))
	})

	It("lists the cocoon rules", func() {
		client.Cocoon.Create().SetGuts(12).SetGrowth([]int{2, 2, 3, 2, 1, 1}).SetName("Cocooned Worm/Pixie").SaveX(ctx)

		got, err := repo.ListCocoonRules(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal([]domain.CocoonRule{
			{Guts: 12, Growth: [monster.StatCount]int{2, 2, 3, 2, 1, 1}, Name: "Cocooned Worm/Pixie"},
		}))
	})

	It("fails, saying which cocoon, for a rule whose growth rates are not six", func() {
		client.Cocoon.Create().SetGuts(12).SetGrowth([]int{1}).SetName("Cocooned Worm/Bad").SaveX(ctx)

		_, err := repo.ListCocoonRules(ctx)
		Expect(err).To(MatchError(ContainSubstring("Cocooned Worm/Bad")))
	})

	It("lists what each item id is called when a monster wants it", func() {
		client.Item.Create().SetCode(42).SetName("Tiger Chips").SetDesireName("Disc Chips").SetKind("Chips").
			SetDescription("d").SetEffect("e").SetSalePrice("500").SetPurchasePrice("-----").SaveX(ctx)

		got, err := repo.ListDesireNames(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(map[int]string{42: "Disc Chips"}))
	})

	It("reads the shipped database, and names a standard Pixie and a cocooned Worm from it", func() {
		shipped, err := testutil.OpenShipped(ctx)
		if errors.Is(err, testutil.ErrNoShippedDB) {
			Skip(err.Error())
		}
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(shipped.Close)
		repo = profilerepo.New(shipped)

		total := 0
		for main := range 50 {
			rules, err := repo.ListBreedRules(ctx, main)
			Expect(err).NotTo(HaveOccurred())
			total += len(rules)
		}
		Expect(total).To(Equal(399))

		cocoons, err := repo.ListCocoonRules(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(cocoons).To(HaveLen(12))

		genera, err := repo.ListGenusNames(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(genera).To(HaveLen(38))
		Expect(genera[0]).To(Equal("Pixie"))

		desires, err := repo.ListDesireNames(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(desires).To(HaveLen(179))
		Expect(desires[131]).To(Equal("Battle"))

		rules, err := repo.ListBreedRules(ctx, 0)
		Expect(err).NotTo(HaveOccurred())
		pixie := monster.Monster{GutsRate: 7, Growth: [monster.StatCount]int{0, 1, 4, 3, 3, 0}}
		Expect(domain.ResolveBreed(rules, cocoons, pixie)).To(Equal(domain.Breed{Name: "Pixie", GutsOK: true}))
		pixie.Growth[monster.Life] = 1
		Expect(domain.ResolveBreed(rules, cocoons, pixie).Name).To(Equal("Pixie (N/S)"))

		worm := monster.Monster{SubGenus: 36, GutsRate: 12, Growth: [monster.StatCount]int{2, 2, 3, 2, 1, 1}}
		Expect(domain.ResolveBreed(rules, cocoons, worm).Name).To(Equal("Cocooned Worm/Pixie"))
	})
})
