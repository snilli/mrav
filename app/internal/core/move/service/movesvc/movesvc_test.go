package movesvc_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/apperr"
	monster "app/internal/core/monster/domain"
	"app/internal/core/move"
	"app/internal/core/move/domain"
	"app/internal/core/move/service/movesvc"
	mockmove "app/mock/core/move"
	mockport "app/mock/port"
)

var _ = Describe("MoveService", func() {
	var (
		ctx    context.Context
		repo   *mockmove.MockMoveRepository
		source *mockport.MockMonsterSource
		svc    move.MoveService
	)
	golem := domain.Genus{Code: 7, Name: "Golem"}
	pat := domain.Move{Genus: 7, MemorySlot: 0, Col: 1, Row: 1, Name: "Pat", Icon: "Golem_1"}

	BeforeEach(func() {
		ctx = context.Background()
		repo = mockmove.NewMockMoveRepository(GinkgoT())
		source = mockport.NewMockMonsterSource(GinkgoT())
		svc = movesvc.New(repo, source)
	})

	Describe("CurrentGrid", func() {
		It("lays the genus's moves out for the monster the game shows", func() {
			m := monster.Monster{Genus: 7}
			m.Learned[0] = true
			source.EXPECT().Current(ctx).Return(m, nil)
			repo.EXPECT().GetGenus(ctx, monster.Genus(7)).Return(golem, nil)
			repo.EXPECT().ListMoves(ctx, monster.Genus(7)).Return([]domain.Move{pat}, nil)

			grid, err := svc.CurrentGrid(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(grid.Genus).To(Equal(golem))
			Expect(grid.Cells[0][0].Move.Name).To(Equal("Pat"))
			Expect(grid.Cells[0][0].Learned).To(BeTrue())
		})

		It("returns the source's error without asking the repository", func() {
			source.EXPECT().Current(ctx).Return(monster.Monster{}, apperr.NotFound("monster", "the game shows none"))
			// no repository expectation: the mock fails the test if it is called

			_, err := svc.CurrentGrid(ctx)
			Expect(errors.Is(err, apperr.ErrNotFound)).To(BeTrue())
		})

		It("returns not found for a genus the catalog does not have", func() {
			source.EXPECT().Current(ctx).Return(monster.Monster{Genus: 40}, nil)
			repo.EXPECT().GetGenus(ctx, monster.Genus(40)).Return(domain.Genus{}, apperr.NotFound("genus", "40"))

			_, err := svc.CurrentGrid(ctx)
			Expect(errors.Is(err, apperr.ErrNotFound)).To(BeTrue())
		})

		It("returns the repository's error when it cannot list the moves", func() {
			source.EXPECT().Current(ctx).Return(monster.Monster{Genus: 7}, nil)
			repo.EXPECT().GetGenus(ctx, monster.Genus(7)).Return(golem, nil)
			repo.EXPECT().ListMoves(ctx, monster.Genus(7)).Return(nil, errors.New("disk gone"))

			_, err := svc.CurrentGrid(ctx)
			Expect(err).To(MatchError(ContainSubstring("disk gone")))
		})
	})

	Describe("ListGenera", func() {
		It("returns the repository's genera", func() {
			repo.EXPECT().ListGenera(ctx).Return([]domain.Genus{golem}, nil)

			got, err := svc.ListGenera(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal([]domain.Genus{golem}))
		})
	})
})
