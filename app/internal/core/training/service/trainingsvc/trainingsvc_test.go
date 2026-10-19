package trainingsvc_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/apperr"
	monster "app/internal/core/monster/domain"
	"app/internal/core/training"
	"app/internal/core/training/domain"
	"app/internal/core/training/service/trainingsvc"
	mocktraining "app/mock/core/training"
	mockport "app/mock/port"
)

var _ = Describe("TrainingService", func() {
	var (
		ctx    context.Context
		repo   *mocktraining.MockTrainingRepository
		source *mockport.MockMonsterSource
		svc    training.TrainingService
	)

	BeforeEach(func() {
		ctx = context.Background()
		repo = mocktraining.NewMockTrainingRepository(GinkgoT())
		source = mockport.NewMockMonsterSource(GinkgoT())
		svc = trainingsvc.New(repo, source)
	})

	It("builds the page from the monster the game shows and the table of gains", func() {
		m := monster.Monster{LifeStage: 2, TrainBoost: 1}
		m.Growth[monster.Pow] = 3
		source.EXPECT().Current(ctx).Return(m, nil)
		repo.EXPECT().ListGains(ctx).Return([]domain.Gain{{LifeStage: 2, GrowthRate: 3, Kind: domain.KindLight, Text: "2-5 (2/3/4/5)"}}, nil)

		got, err := svc.Current(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Drills[0].Drill.Name).To(Equal("Domino"))
		Expect(got.Drills[0].Primary).To(Equal("2-5 (2/3/4/5)"))
		Expect(got.Specialties[0]).To(BeTrue())
	})

	It("returns the source's error without asking the repository", func() {
		source.EXPECT().Current(ctx).Return(monster.Monster{}, apperr.NotFound("monster", "No monster"))
		// no repository expectation: the mock fails the test if it is called

		_, err := svc.Current(ctx)
		Expect(errors.Is(err, apperr.ErrNotFound)).To(BeTrue())
	})

	It("returns the repository's error when it cannot list the gains", func() {
		source.EXPECT().Current(ctx).Return(monster.Monster{}, nil)
		repo.EXPECT().ListGains(ctx).Return(nil, errors.New("disk gone"))

		_, err := svc.Current(ctx)
		Expect(err).To(MatchError(ContainSubstring("disk gone")))
	})
})
