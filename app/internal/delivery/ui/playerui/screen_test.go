package playerui

import (
	"context"
	"errors"

	"fyne.io/fyne/v2/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/apperr"
	player "app/internal/core/player/domain"
	mockplayer "app/mock/core/player"
)

var _ = Describe("BreederScreen", func() {
	var (
		ctx context.Context
		svc *mockplayer.MockPlayerService
		s   *BreederScreen
	)

	BeforeEach(func() {
		test.NewTempApp(GinkgoTB())
		ctx = context.Background()
		svc = mockplayer.NewMockPlayerService(GinkgoT())
		s = NewBreederScreen(svc)
	})

	It("shows the breeder's name", func() {
		s.apply(player.Player{Breeder: "Bob"}, nil)
		Expect(s.name.Text).To(Equal("Bob"))
		Expect(s.status.Visible()).To(BeFalse())
	})

	It("shows why there is nothing to show, in place of the name", func() {
		s.apply(player.Player{}, apperr.NotFound("game", "MR2DX is not running"))
		Expect(s.status.Text).To(Equal("MR2DX is not running"))
		Expect(s.name.Visible()).To(BeFalse())

		s.apply(player.Player{Breeder: "Bob"}, nil)
		Expect(s.name.Visible()).To(BeTrue())
	})

	It("asks the service for the player and shows the breeder", func() {
		svc.EXPECT().Current(ctx).Return(player.Player{Breeder: "Ann"}, nil)
		s.Load(ctx)
		Eventually(func() string { return s.name.Text }).Should(Equal("Ann"))
	})
})

var _ = Describe("ControllerScreen", func() {
	var (
		ctx context.Context
		svc *mockplayer.MockPlayerService
		s   *ControllerScreen
	)

	BeforeEach(func() {
		test.NewTempApp(GinkgoTB())
		ctx = context.Background()
		svc = mockplayer.NewMockPlayerService(GinkgoT())
		s = NewControllerScreen(svc)
	})

	It("shows the buttons held, group by group", func() {
		s.apply(player.Player{Input: 1<<0 | 1<<5 | 1<<8 | 1<<13}, nil) // L2, Cross, Select, Right
		Expect(s.shoulder.Text).To(Equal("L2"))
		Expect(s.face.Text).To(Equal("Cross"))
		Expect(s.misc.Text).To(Equal("Select"))
		Expect(s.dpad.Text).To(Equal("Right"))
	})

	It("shows a button let go of as gone", func() {
		s.apply(player.Player{Input: 1 << 4}, nil)
		s.apply(player.Player{}, nil)
		Expect(s.face.Text).To(BeEmpty())
	})

	It("shows why there is nothing to show", func() {
		s.apply(player.Player{}, errors.New("memory gone"))
		Expect(s.status.Text).To(Equal("Error: memory gone"))
		Expect(s.body.Visible()).To(BeFalse())
	})

	It("asks the service for the player and shows the buttons", func() {
		svc.EXPECT().Current(ctx).Return(player.Player{Input: 1 << 11}, nil)
		s.Load(ctx)
		Eventually(func() string { return s.misc.Text }).Should(Equal("Start"))
	})
})
