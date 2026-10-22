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
	profile "app/internal/core/profile/domain"
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
			repo.EXPECT().ListGenera(ctx).Return([]domain.Genus{golem}, nil)

			grid, err := svc.CurrentGrid(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(grid.Genus).To(Equal(golem))
			Expect(grid.Cells[0][0].Move.Name).To(Equal("Pat"))
			Expect(grid.Cells[0][0].Learned).To(BeTrue())
		})

		It("names the genera that a move's unlock rule asks for", func() {
			jill := domain.Genus{Code: 24, Name: "Jill"}
			odd := domain.Move{Genus: 7, MemorySlot: 1, Col: 1, Row: 2, Name: "Odd", Icon: "Golem_2",
				Unlock: domain.UnlockRule{All: []domain.Condition{domain.SubGenusIn{IDs: []monster.Genus{24}}}}}
			source.EXPECT().Current(ctx).Return(monster.Monster{Genus: 7, SubGenus: 7}, nil)
			repo.EXPECT().GetGenus(ctx, monster.Genus(7)).Return(golem, nil)
			repo.EXPECT().ListMoves(ctx, monster.Genus(7)).Return([]domain.Move{odd}, nil)
			repo.EXPECT().ListGenera(ctx).Return([]domain.Genus{golem, jill}, nil)

			grid, err := svc.CurrentGrid(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(grid.Cells[0][1].Requirements).To(Equal([]domain.Requirement{
				{Text: "Sub-genus is Jill", Have: "Golem", Met: false},
			}))
		})

		It("marks ready only what the errantry teaches now, as the Errantry page lists it", func() {
			press := domain.Move{Genus: 7, MemorySlot: 1, Col: 1, Row: 2, Name: "Press", Icon: "Golem_2",
				Acquire: []domain.Acquire{{Place: 2, Priority: 1, Auto: 3, Gates: []domain.Gate{{Stats: []monster.Stat{monster.Pow}, Min: 250, Max: 400}}}}}
			roll := domain.Move{Genus: 7, MemorySlot: 2, Col: 1, Row: 3, Name: "Roll", Icon: "Golem_3",
				Acquire: []domain.Acquire{{Place: 2, Priority: 2}}}
			m := monster.Monster{Genus: 7}
			m.Stats[monster.Pow] = 500
			source.EXPECT().Current(ctx).Return(m, nil)
			repo.EXPECT().GetGenus(ctx, monster.Genus(7)).Return(golem, nil)
			repo.EXPECT().ListMoves(ctx, monster.Genus(7)).Return([]domain.Move{pat, press, roll}, nil)
			repo.EXPECT().ListGenera(ctx).Return([]domain.Genus{golem}, nil)

			grid, err := svc.CurrentGrid(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(grid.Cells[0][0].CanLearn).To(BeFalse(), "Pat: no place teaches it")
			Expect(grid.Cells[0][1].CanLearn).To(BeTrue(), "Press: Mandy teaches it for sure")
			Expect(grid.Cells[0][2].CanLearn).To(BeFalse(), "Roll: its rule is met, but Press takes all of Mandy's chance")
			Expect(grid.Cells[0][2].Requirements).To(Equal([]domain.Requirement{{Text: "Mandy Desert teaches Press first"}}))
			Expect(grid.Cells[0][1].Chance).To(BeNumerically("~", 100), "Press: all of it")
			Expect(grid.Cells[0][1].Place).To(Equal("Mandy Desert"))
			Expect(grid.Cells[0][0].Chance).To(BeZero(), "Pat: no place teaches it")
			Expect(grid.Cells[0][2].Chance).To(BeZero(), "Roll: Press takes all of what is left")
			Expect(grid.Cells[0][2].Place).To(BeEmpty())
		})

		It("gives a move that a place teaches only by luck its chance and its place, and does not mark it ready", func() {
			luck := domain.Move{Genus: 7, MemorySlot: 1, Col: 1, Row: 2, Name: "Luck", Icon: "Golem_2",
				Acquire: []domain.Acquire{{Place: 2, Priority: 1, Auto: 3, Gates: []domain.Gate{{Stats: []monster.Stat{monster.Pow}, Min: 250, Max: 400}}}}}
			m := monster.Monster{Genus: 7}
			m.Stats[monster.Pow] = 120
			source.EXPECT().Current(ctx).Return(m, nil)
			repo.EXPECT().GetGenus(ctx, monster.Genus(7)).Return(golem, nil)
			repo.EXPECT().ListMoves(ctx, monster.Genus(7)).Return([]domain.Move{pat, luck}, nil)
			repo.EXPECT().ListGenera(ctx).Return([]domain.Genus{golem}, nil)

			grid, err := svc.CurrentGrid(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(grid.Cells[0][1].CanLearn).To(BeFalse(), "a chance of 3 percent is not for sure")
			Expect(grid.Cells[0][1].Chance).To(BeNumerically("~", 3))
			Expect(grid.Cells[0][1].Place).To(Equal("Mandy Desert"))
		})

		It("takes the better of the places that teach a move, and marks it ready when that is all", func() {
			locales := profile.ErrantryLocales()
			both := domain.Move{Genus: 7, MemorySlot: 1, Col: 1, Row: 2, Name: "Both", Icon: "Golem_2", Acquire: []domain.Acquire{
				{Place: 1, Priority: 1, Auto: 3, Gates: []domain.Gate{{Stats: []monster.Stat{monster.Pow}, Min: 400, Max: 500}}},
				{Place: 2, Priority: 1},
			}}
			m := monster.Monster{Genus: 7}
			m.Stats[monster.Pow] = 100
			source.EXPECT().Current(ctx).Return(m, nil)
			repo.EXPECT().GetGenus(ctx, monster.Genus(7)).Return(golem, nil)
			repo.EXPECT().ListMoves(ctx, monster.Genus(7)).Return([]domain.Move{pat, both}, nil)
			repo.EXPECT().ListGenera(ctx).Return([]domain.Genus{golem}, nil)

			grid, err := svc.CurrentGrid(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(grid.Cells[0][1].CanLearn).To(BeTrue(), "the second place teaches it for sure")
			Expect(grid.Cells[0][1].Chance).To(BeNumerically("~", 100))
			Expect(grid.Cells[0][1].Place).To(Equal(locales[1].Name))
		})

		It("returns the repository's error when it cannot list the genera", func() {
			source.EXPECT().Current(ctx).Return(monster.Monster{Genus: 7}, nil)
			repo.EXPECT().GetGenus(ctx, monster.Genus(7)).Return(golem, nil)
			repo.EXPECT().ListMoves(ctx, monster.Genus(7)).Return([]domain.Move{pat}, nil)
			repo.EXPECT().ListGenera(ctx).Return(nil, errors.New("disk gone"))

			_, err := svc.CurrentGrid(ctx)
			Expect(err).To(MatchError(ContainSubstring("disk gone")))
		})

		It("returns the source's error without asking the repository", func() {
			source.EXPECT().Current(ctx).Return(monster.Monster{}, apperr.NotFound("monster", "the game shows none"))

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
