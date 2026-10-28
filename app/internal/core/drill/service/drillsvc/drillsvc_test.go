package drillsvc_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/apperr"
	"app/internal/core/drill"
	"app/internal/core/drill/domain"
	"app/internal/core/drill/service/drillsvc"
	monster "app/internal/core/monster/domain"
	mockdrill "app/mock/core/drill"
	mockport "app/mock/port"
)

var _ = Describe("DrillService", func() {
	var (
		ctx    context.Context
		repo   *mockdrill.MockDrillRepository
		source *mockport.MockMonsterSource
		svc    drill.DrillService
	)

	BeforeEach(func() {
		ctx = context.Background()
		repo = mockdrill.NewMockDrillRepository(GinkgoT())
		source = mockport.NewMockMonsterSource(GinkgoT())
		svc = drillsvc.New(repo, source)
	})

	It("builds the page from the monster the game shows and the table of gains, with the bonus for a drill it specialises in", func() {
		m := monster.Monster{LifeStage: 2, TrainBoost: 1}
		m.Growth[monster.Pow] = 3
		source.EXPECT().Current(ctx).Return(m, nil)
		repo.EXPECT().ListGains(ctx).Return([]domain.Cell{
			{Result: domain.Success, Role: domain.Light, Bonus: true, LifeStage: 2, GrowthRate: 3, Slots: [][]int{{2}, {3}, {4}, {5}}},
		}, nil)

		got, err := svc.Current(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Drills[0].Drill.Name).To(Equal("Domino"))
		Expect(got.Drills[0].Outcomes[1].Effects).To(Equal([]domain.Effect{{Stat: monster.Pow, Low: 2, High: 5}}))
		Expect(got.Specialties[0]).To(BeTrue())
	})

	It("returns the source's error without asking the repository", func() {
		source.EXPECT().Current(ctx).Return(monster.Monster{}, apperr.NotFound("monster", "No monster"))

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
