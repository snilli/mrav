package profilerepo_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"orm"
	"orm/ent"
	entbreed "orm/ent/breed"
	entcocoon "orm/ent/cocoon"
	entitem "orm/ent/item"

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
			return client.Breed.Create().SetVersion(entbreed.VersionMr2dx).SetMainGenus(main).SetSubGenus(sub).SetSeq(seq).SetName(name)
		}
		seed(1, 38, 1, "Moo (CD)").SetGuts(18).SetMatchGuts(true).SetGrowth([]int{3, 3, 0, 1, 1, 3}).SaveX(ctx)
		seed(1, 0, 0, "Tiamat").SetGuts(14).SetGrowth([]int{1, 3, 4, 2, 2, 1}).SaveX(ctx)
		seed(1, 38, 0, "Moo (DNA Capsule)").SetGuts(19).SetMatchGuts(true).SetGrowth([]int{2, 4, 4, 2, 1, 2}).SaveX(ctx)
		seed(1, 39, 0, "Kasumi").SaveX(ctx)
		seed(0, 0, 0, "Pixie").SaveX(ctx)

		got, err := repo.ListBreedRules(ctx, 1, monster.MR2DX)
		Expect(err).NotTo(HaveOccurred())
		tiamat := [monster.StatCount]int{1, 3, 4, 2, 2, 1}
		capsule := [monster.StatCount]int{2, 4, 4, 2, 1, 2}
		cd := [monster.StatCount]int{3, 3, 0, 1, 1, 3}
		ints := func(n int) *int { return &n }
		Expect(got).To(Equal([]domain.BreedRule{
			{Main: 1, Sub: 0, Name: "Tiamat", Guts: ints(14), Growth: &tiamat},
			{Main: 1, Sub: 38, Name: "Moo (DNA Capsule)", Guts: ints(19), MatchGuts: true, Growth: &capsule},
			{Main: 1, Sub: 38, Name: "Moo (CD)", Guts: ints(18), MatchGuts: true, Growth: &cd},
			{Main: 1, Sub: 39, Name: "Kasumi"},
		}))
	})

	It("lists the breed rules of the game it is asked for, and none of the other game's", func() {
		for _, v := range []entbreed.Version{entbreed.VersionMr2, entbreed.VersionMr2dx} {
			client.Breed.Create().SetVersion(v).SetMainGenus(1).SetSubGenus(0).SetSeq(0).SetName("Tiamat " + string(v)).SaveX(ctx)
		}

		mr2, err := repo.ListBreedRules(ctx, 1, monster.MR2)
		Expect(err).NotTo(HaveOccurred())
		mr2dx, err := repo.ListBreedRules(ctx, 1, monster.MR2DX)
		Expect(err).NotTo(HaveOccurred())
		Expect(mr2).To(Equal([]domain.BreedRule{{Main: 1, Sub: 0, Name: "Tiamat mr2"}}))
		Expect(mr2dx).To(Equal([]domain.BreedRule{{Main: 1, Sub: 0, Name: "Tiamat mr2dx"}}))
	})

	It("lists none for a main genus it has no rules for, without failing", func() {
		got, err := repo.ListBreedRules(ctx, 9, monster.MR2DX)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(BeEmpty())
	})

	It("fails, saying which breed, for a rule whose growth rates are not six", func() {
		client.Breed.Create().SetVersion(entbreed.VersionMr2dx).SetMainGenus(2).SetSubGenus(7).SetSeq(0).SetName("Bad").SetGrowth([]int{1, 2}).SaveX(ctx)

		_, err := repo.ListBreedRules(ctx, 2, monster.MR2DX)
		Expect(err).To(MatchError(And(ContainSubstring("breed 2/7"), ContainSubstring("2 rates"))))
	})

	It("lists the cocoon rules", func() {
		client.Cocoon.Create().SetVersion(entcocoon.VersionMr2dx).SetGuts(12).SetGrowth([]int{2, 2, 3, 2, 1, 1}).SetName("Cocooned Worm/Pixie").SaveX(ctx)

		got, err := repo.ListCocoonRules(ctx, monster.MR2DX)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal([]domain.CocoonRule{
			{Guts: 12, Growth: [monster.StatCount]int{2, 2, 3, 2, 1, 1}, Name: "Cocooned Worm/Pixie"},
		}))
	})

	It("lists the cocoon rules of the game it is asked for, one for each growth that a cocoon has there", func() {
		cocoon := func(v entcocoon.Version, guts int, growth ...int) {
			client.Cocoon.Create().SetVersion(v).SetGuts(guts).SetGrowth(growth).SetName("Cocooned Worm/Naga").SaveX(ctx)
		}
		cocoon(entcocoon.VersionMr2, 13, 3, 2, 1, 2, 0, 1)
		cocoon(entcocoon.VersionMr2dx, 13, 3, 2, 1, 2, 0, 1)
		cocoon(entcocoon.VersionMr2dx, 16, 3, 2, 2, 2, 2, 1)

		mr2, err := repo.ListCocoonRules(ctx, monster.MR2)
		Expect(err).NotTo(HaveOccurred())
		mr2dx, err := repo.ListCocoonRules(ctx, monster.MR2DX)
		Expect(err).NotTo(HaveOccurred())
		Expect(mr2).To(Equal([]domain.CocoonRule{
			{Guts: 13, Growth: [monster.StatCount]int{3, 2, 1, 2, 0, 1}, Name: "Cocooned Worm/Naga"},
		}))
		Expect(mr2dx).To(Equal([]domain.CocoonRule{
			{Guts: 13, Growth: [monster.StatCount]int{3, 2, 1, 2, 0, 1}, Name: "Cocooned Worm/Naga"},
			{Guts: 16, Growth: [monster.StatCount]int{3, 2, 2, 2, 2, 1}, Name: "Cocooned Worm/Naga"},
		}))
	})

	It("fails, saying which cocoon, for a rule whose growth rates are not six", func() {
		client.Cocoon.Create().SetVersion(entcocoon.VersionMr2dx).SetGuts(12).SetGrowth([]int{1}).SetName("Cocooned Worm/Bad").SaveX(ctx)

		_, err := repo.ListCocoonRules(ctx, monster.MR2DX)
		Expect(err).To(MatchError(ContainSubstring("Cocooned Worm/Bad")))
	})

	It("lists what each item id is called when a monster wants it, in the game version asked for", func() {
		client.Item.Create().SetVersion(entitem.VersionMr2).SetCode(42).SetName("Tiger Chips").SetDesireName("Disc Chips").SetKind("Chips").
			SetDescription("d").SetEffect("e").SetSalePrice(500).SaveX(ctx)
		client.Item.Create().SetVersion(entitem.VersionMr2dx).SetCode(42).SetName("Tiger Chips").SetDesireName("DX Chips").SetKind("Chips").
			SetDescription("d").SetEffect("e").SetSalePrice(500).SaveX(ctx)

		ps1, err := repo.ListDesireNames(ctx, monster.MR2)
		Expect(err).NotTo(HaveOccurred())
		steam, err := repo.ListDesireNames(ctx, monster.MR2DX)
		Expect(err).NotTo(HaveOccurred())
		Expect(ps1).To(Equal(map[int]string{42: "Disc Chips"}))
		Expect(steam).To(Equal(map[int]string{42: "DX Chips"}))
	})

	It("reads the shipped database, and names a standard Pixie and a cocooned Worm from it", func() {
		shipped, err := testutil.OpenShipped(ctx)
		if errors.Is(err, testutil.ErrNoShippedDB) {
			Skip(err.Error())
		}
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(shipped.Close)
		repo = profilerepo.New(shipped)

		for version, want := range map[monster.Version]struct{ breeds, cocoons int }{
			monster.MR2:   {428, 12},
			monster.MR2DX: {498, 14},
		} {
			total := 0
			for main := range 50 {
				rules, err := repo.ListBreedRules(ctx, main, version)
				Expect(err).NotTo(HaveOccurred())
				total += len(rules)
			}
			Expect(total).To(Equal(want.breeds), string(version))

			cocoons, err := repo.ListCocoonRules(ctx, version)
			Expect(err).NotTo(HaveOccurred())
			Expect(cocoons).To(HaveLen(want.cocoons), string(version))
		}

		cocoons, err := repo.ListCocoonRules(ctx, monster.MR2DX)
		Expect(err).NotTo(HaveOccurred())

		genera, err := repo.ListGenusNames(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(genera).To(HaveLen(38))
		Expect(genera[0]).To(Equal("Pixie"))

		for _, version := range []monster.Version{monster.MR2, monster.MR2DX} {
			desires, err := repo.ListDesireNames(ctx, version)
			Expect(err).NotTo(HaveOccurred())
			Expect(desires).To(HaveLen(179), string(version))
			Expect(desires[131]).To(Equal("Battle"), string(version))
		}

		rules, err := repo.ListBreedRules(ctx, 0, monster.MR2DX)
		Expect(err).NotTo(HaveOccurred())
		pixie := monster.Monster{GutsRate: 7, Growth: [monster.StatCount]int{0, 1, 4, 3, 3, 0}}
		Expect(domain.ResolveBreed(rules, cocoons, pixie)).To(Equal(domain.Breed{Name: "Pixie", GutsOK: true}))
		pixie.Growth[monster.Life] = 1
		Expect(domain.ResolveBreed(rules, cocoons, pixie).Name).To(Equal("Pixie (N/S)"))

		worm := monster.Monster{SubGenus: 36, GutsRate: 12, Growth: [monster.StatCount]int{2, 2, 3, 2, 1, 1}}
		Expect(domain.ResolveBreed(rules, cocoons, worm).Name).To(Equal("Cocooned Worm/Pixie"))

		durahan := monster.Monster{Genus: 9, SubGenus: 9, GutsRate: 12, Growth: [monster.StatCount]int{2, 3, 2, 2, 0, 4}}
		for version, gutsOK := range map[monster.Version]bool{monster.MR2: false, monster.MR2DX: true} {
			rules, err := repo.ListBreedRules(ctx, 9, version)
			Expect(err).NotTo(HaveOccurred())
			Expect(domain.ResolveBreed(rules, nil, durahan)).To(Equal(domain.Breed{Name: "Durahan", GutsOK: gutsOK}), string(version))
		}

		poison := monster.Monster{Genus: 0, SubGenus: 38, GutsRate: 9, Growth: [monster.StatCount]int{4, 4, 4, 4, 4, 4}}
		for _, version := range []monster.Version{monster.MR2, monster.MR2DX} {
			rules, err := repo.ListBreedRules(ctx, 0, version)
			Expect(err).NotTo(HaveOccurred())
			Expect(domain.ResolveBreed(rules, nil, poison).Name).To(Equal("Poison"), string(version))
		}

		for _, c := range []struct {
			version   monster.Version
			main, sub int
			name      string
		}{
			{monster.MR2, 13, 38, "Kung Fu Hare"},
			{monster.MR2DX, 13, 38, "Tornado"},
			{monster.MR2DX, 13, 39, "Kung Fu Hare"},
			{monster.MR2, 18, 38, "Boxer Bajarl"},
			{monster.MR2DX, 18, 38, "Gym Bajarl"},
			{monster.MR2DX, 18, 41, "Boxer Bajarl"},
			{monster.MR2, 1, 39, "[E] Magma Heart"},
			{monster.MR2DX, 1, 39, "Magma Heart"},
			{monster.MR2DX, 5, 15, "Proto"},
			{monster.MR2, 0, 18, "Jinnee"},
			{monster.MR2DX, 7, 33, "Ecologuardia"},
			{monster.MR2, 8, 33, "AlohaSaurian"},
			{monster.MR2DX, 3, 0, "PeachTreeBug"},
			{monster.MR2DX, 7, 40, "ForwardGolem"},
			{monster.MR2, 25, 38, "GentleMocchi"},
			{monster.MR2, 27, 38, "GabooSoldier"},
		} {
			rules, err := repo.ListBreedRules(ctx, c.main, c.version)
			Expect(err).NotTo(HaveOccurred())
			m := monster.Monster{Genus: monster.Genus(c.main), SubGenus: monster.Genus(c.sub), GutsRate: 9}
			Expect(domain.ResolveBreed(rules, nil, m).Name).To(HavePrefix(c.name), string(c.version))
		}

		for main := range 38 {
			rules, err := repo.ListBreedRules(ctx, main, monster.MR2DX)
			Expect(err).NotTo(HaveOccurred())
			for _, r := range rules {
				Expect(r.Name).NotTo(ContainSubstring("[E]"), "DX raises the enemy monsters like any other")
			}
		}

		for _, c := range []struct {
			version monster.Version
			sub     int
			guts    int
			growth  [monster.StatCount]int
			want    domain.Breed
		}{
			{monster.MR2, 0, 12, [monster.StatCount]int{2, 2, 3, 2, 1, 1}, domain.Breed{Name: "Red Worm", GutsOK: true}},
			{monster.MR2DX, 0, 12, [monster.StatCount]int{2, 2, 3, 2, 1, 1}, domain.Breed{Name: "Red Worm", GutsOK: true}},
			{monster.MR2, 36, 15, [monster.StatCount]int{4, 2, 2, 2, 0, 1}, domain.Breed{Name: "Worm", GutsOK: true}},
			{monster.MR2DX, 37, 16, [monster.StatCount]int{3, 2, 2, 2, 2, 1}, domain.Breed{Name: "Purple Worm", GutsOK: true}},
			{monster.MR2, 37, 16, [monster.StatCount]int{3, 2, 2, 2, 2, 1}, domain.Breed{Name: "Purple Worm (N/S)", GutsOK: false}},
			{monster.MR2, 38, 9, [monster.StatCount]int{}, domain.Breed{Name: "Express Worm", GutsOK: true}},
			{monster.MR2DX, 38, 9, [monster.StatCount]int{}, domain.Breed{Name: "Express Worm", GutsOK: true}},
		} {
			rules, err := repo.ListBreedRules(ctx, 36, c.version)
			Expect(err).NotTo(HaveOccurred())
			m := monster.Monster{Genus: 36, SubGenus: monster.Genus(c.sub), GutsRate: c.guts, Growth: c.growth}
			Expect(domain.ResolveBreed(rules, cocoons, m)).To(Equal(c.want), string(c.version))
		}

		hare := monster.Monster{Genus: 13, SubGenus: 0, GutsRate: 12, Growth: [monster.StatCount]int{1, 3, 2, 2, 4, 0}}
		for _, version := range []monster.Version{monster.MR2, monster.MR2DX} {
			rules, err := repo.ListBreedRules(ctx, 13, version)
			Expect(err).NotTo(HaveOccurred())
			Expect(domain.ResolveBreed(rules, nil, hare)).To(Equal(domain.Breed{Name: "Fairy Hare", GutsOK: true}), string(version))
		}
	})
})
