package playersvc_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/apperr"
	"app/internal/core/player"
	"app/internal/core/player/domain"
	"app/internal/core/player/service/playersvc"
	mockport "app/mock/port"
)

var _ = Describe("PlayerService", func() {
	var (
		ctx    context.Context
		source *mockport.MockPlayerSource
		svc    player.PlayerService
	)

	BeforeEach(func() {
		ctx = context.Background()
		source = mockport.NewMockPlayerSource(GinkgoT())
		svc = playersvc.New(source)
	})

	It("is the player's side of the game that the source shows", func() {
		want := domain.Player{Money: 1500, Breeder: "Bob", Input: 16}
		source.EXPECT().Player(ctx).Return(want, nil)

		got, err := svc.Current(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(want))
	})

	It("returns the source's error", func() {
		source.EXPECT().Player(ctx).Return(domain.Player{}, apperr.NotFound("game", "MR2DX is not running"))

		_, err := svc.Current(ctx)
		Expect(errors.Is(err, apperr.ErrNotFound)).To(BeTrue())
	})
})
