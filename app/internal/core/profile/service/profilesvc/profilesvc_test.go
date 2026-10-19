package profilesvc_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/apperr"
	monster "app/internal/core/monster/domain"
	player "app/internal/core/player/domain"
	"app/internal/core/profile"
	"app/internal/core/profile/domain"
	"app/internal/core/profile/service/profilesvc"
	mockprofile "app/mock/core/profile"
	mockport "app/mock/port"
)

var _ = Describe("ProfileService", func() {
	var (
		ctx      context.Context
		repo     *mockprofile.MockProfileRepository
		monsters *mockport.MockMonsterSource
		players  *mockport.MockPlayerSource
		svc      profile.ProfileService
	)
	growth := [monster.StatCount]int{0, 1, 4, 3, 3, 0}
	henger := monster.Monster{Genus: 5, SubGenus: 3, GutsRate: 9, Growth: growth, ItemLike: 4, ItemDislike: 130, Drug: 38}
	rule := domain.BreedRule{Main: 5, Sub: 3, Name: "Hengerdina", Otherwise: "Hengerdina (N/S)", Growth: &growth}
	names := map[monster.Genus]string{5: "Henger", 3: "C. Pandora"}
	desires := map[int]string{4: "Cup Jelly", 130: "Nothing", 38: "Troron"}

	// expectData is every repository call a profile of the monster takes.
	expectData := func() {
		monsters.EXPECT().Current(ctx).Return(henger, nil)
		players.EXPECT().Player(ctx).Return(player.Player{Money: 1500}, nil)
		repo.EXPECT().ListGenusNames(ctx).Return(names, nil)
		repo.EXPECT().ListBreedRules(ctx, 5).Return([]domain.BreedRule{rule}, nil)
		repo.EXPECT().ListCocoonRules(ctx).Return(nil, nil)
		repo.EXPECT().ListDesireNames(ctx).Return(desires, nil)
	}

	BeforeEach(func() {
		ctx = context.Background()
		repo = mockprofile.NewMockProfileRepository(GinkgoT())
		monsters = mockport.NewMockMonsterSource(GinkgoT())
		players = mockport.NewMockPlayerSource(GinkgoT())
		svc = profilesvc.New(repo, monsters, players)
	})

	It("names the monster, its breeds and what it wants, and carries the monster and the player", func() {
		expectData()

		got, err := svc.Current(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Monster).To(Equal(henger))
		Expect(got.Player.Money).To(Equal(1500))
		Expect(got.Breed.Name).To(Equal("Hengerdina"))
		Expect(got.MainGenus).To(Equal("Henger"))
		Expect(got.SubGenus).To(Equal("C. Pandora"))
		Expect(got.Like).To(Equal("Cup Jelly"))
		Expect(got.Dislike).To(Equal("Nothing"))
		Expect(got.Drug).To(Equal("Troron"))
	})

	It("says the monster is on no drug when its drug id is not a drug's, and --- for an id the list lacks", func() {
		m := henger
		m.Drug, m.ItemLike = 0, 99
		monsters.EXPECT().Current(ctx).Return(m, nil)
		players.EXPECT().Player(ctx).Return(player.Player{}, nil)
		repo.EXPECT().ListGenusNames(ctx).Return(names, nil)
		repo.EXPECT().ListBreedRules(ctx, 5).Return([]domain.BreedRule{rule}, nil)
		repo.EXPECT().ListCocoonRules(ctx).Return(nil, nil)
		repo.EXPECT().ListDesireNames(ctx).Return(desires, nil)

		got, err := svc.Current(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Drug).To(Equal("---"))
		Expect(got.Like).To(Equal("---"))
	})

	It("returns the monster source's error without asking anything else", func() {
		monsters.EXPECT().Current(ctx).Return(monster.Monster{}, apperr.NotFound("monster", "No monster"))
		// no other expectation: the mocks fail the test if they are called

		_, err := svc.Current(ctx)
		Expect(errors.Is(err, apperr.ErrNotFound)).To(BeTrue())
	})

	It("returns the player source's error", func() {
		monsters.EXPECT().Current(ctx).Return(henger, nil)
		players.EXPECT().Player(ctx).Return(player.Player{}, errors.New("memory gone"))

		_, err := svc.Current(ctx)
		Expect(err).To(MatchError(ContainSubstring("memory gone")))
	})

	DescribeTable("returns the repository's error, from whichever of its lists fails",
		func(fail func()) {
			monsters.EXPECT().Current(ctx).Return(henger, nil)
			players.EXPECT().Player(ctx).Return(player.Player{}, nil)
			fail()

			_, err := svc.Current(ctx)
			Expect(err).To(MatchError(ContainSubstring("disk gone")))
		},
		Entry("the genus names", func() {
			repo.EXPECT().ListGenusNames(ctx).Return(nil, errors.New("disk gone"))
		}),
		Entry("the breed rules", func() {
			repo.EXPECT().ListGenusNames(ctx).Return(names, nil)
			repo.EXPECT().ListBreedRules(ctx, 5).Return(nil, errors.New("disk gone"))
		}),
		Entry("the cocoon rules", func() {
			repo.EXPECT().ListGenusNames(ctx).Return(names, nil)
			repo.EXPECT().ListBreedRules(ctx, 5).Return(nil, nil)
			repo.EXPECT().ListCocoonRules(ctx).Return(nil, errors.New("disk gone"))
		}),
		Entry("the desire names", func() {
			repo.EXPECT().ListGenusNames(ctx).Return(names, nil)
			repo.EXPECT().ListBreedRules(ctx, 5).Return(nil, nil)
			repo.EXPECT().ListCocoonRules(ctx).Return(nil, nil)
			repo.EXPECT().ListDesireNames(ctx).Return(nil, errors.New("disk gone"))
		}),
	)
})
