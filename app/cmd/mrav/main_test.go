package main

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"

	"github.com/stretchr/testify/mock"

	"app/config"
	monster "app/internal/core/monster/domain"
	movedomain "app/internal/core/move/domain"
	"app/internal/infrastructure/fakemonster"
	"app/internal/infrastructure/fakeplayer"
	"app/internal/infrastructure/gamememory"
	mockitem "app/mock/core/item"
	mockmove "app/mock/core/move"
	mockprofile "app/mock/core/profile"
	mocktraining "app/mock/core/training"
)

var _ = Describe("newSource", func() {
	It("gives the demo panel its fakes only in the demo, and reads a game otherwise", func() {
		monsters, players, fakeMonster, fakePlayer := newSource(&config.Config{Game: "demo"})
		Expect(fakeMonster).NotTo(BeNil())
		Expect(fakePlayer).NotTo(BeNil())
		Expect(monsters).To(BeIdenticalTo(fakeMonster))
		Expect(players).To(BeIdenticalTo(fakePlayer))

		for _, game := range []string{"duckstation", "mr2dx"} {
			monsters, players, fakeMonster, fakePlayer := newSource(&config.Config{Game: game})
			Expect(monsters).To(BeAssignableToTypeOf(&gamememory.Source{}), game)
			Expect(players).To(BeIdenticalTo(monsters), game) // one game, read for both
			Expect(fakeMonster).To(BeNil(), game)
			Expect(fakePlayer).To(BeNil(), game)
		}
	})

	It("reads whichever game is running with auto, where a game can be read, and is the demo where it cannot", func() {
		DeferCleanup(func(was bool) { canReadGames = was }, canReadGames)

		canReadGames = true
		monsters, players, fakeMonster, fakePlayer := newSource(&config.Config{Game: "auto"})
		Expect(monsters).To(BeAssignableToTypeOf(&gamememory.Auto{}))
		Expect(players).To(BeIdenticalTo(monsters))
		Expect(fakeMonster).To(BeNil())
		Expect(fakePlayer).To(BeNil())

		canReadGames = false
		monsters, players, fakeMonster, fakePlayer = newSource(&config.Config{Game: "auto"})
		Expect(fakeMonster).NotTo(BeNil())
		Expect(fakePlayer).NotTo(BeNil())
		Expect(monsters).To(BeIdenticalTo(fakeMonster))
		Expect(players).To(BeIdenticalTo(fakePlayer))
	})
})

var _ = Describe("newDemo", func() {
	It("starts the demo on a random monster, so that the pages have something to show", func() {
		test.NewTempApp(GinkgoTB())
		svc := mockmove.NewMockMoveService(GinkgoT())
		svc.EXPECT().ListGenera(mock.Anything).Return([]movedomain.Genus{{Code: 0, Name: "Pixie"}}, nil)
		monsters, players := fakemonster.New(), fakeplayer.New()

		panel, err := newDemo(nil, svc, monsters, players)
		Expect(err).NotTo(HaveOccurred())
		Expect(panel).NotTo(BeNil())
		m, _ := monsters.Current(context.Background())
		Expect(m.Stats[monster.Pow]).To(BeNumerically(">", 0))
		p, _ := players.Player(context.Background())
		Expect(p.Breeder).To(Equal("Demo Breeder"))
	})

	It("has no demo when a game is read", func() {
		panel, err := newDemo(nil, mockmove.NewMockMoveService(GinkgoT()), nil, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(panel).To(BeNil())
	})
})

var _ = Describe("newWindow", func() {
	It("puts every page on one screen, not in tabs", func() {
		a := test.NewTempApp(GinkgoTB())
		p := newParts(a,
			mockprofile.NewMockProfileService(GinkgoT()), mockmove.NewMockMoveService(GinkgoT()),
			mocktraining.NewMockTrainingService(GinkgoT()), mockitem.NewMockItemService(GinkgoT()), nil)

		w := newWindow(a, p, nil)
		Expect(w.Content()).NotTo(BeNil())
		Expect(w.Content()).NotTo(BeAssignableToTypeOf(&container.AppTabs{}))
	})

	It("opens at startSize, which is smaller than the screen of the game, to sit beside it and not cover it", func() {
		a := test.NewTempApp(GinkgoTB())
		p := newParts(a,
			mockprofile.NewMockProfileService(GinkgoT()), mockmove.NewMockMoveService(GinkgoT()),
			mocktraining.NewMockTrainingService(GinkgoT()), mockitem.NewMockItemService(GinkgoT()), nil)

		w := newWindow(a, p, nil)
		Expect(w.Canvas().Size()).To(Equal(startSize))
	})

	It("asks to be no less high than minHeight, which is less than it opens at, so that it can be made smaller", func() {
		a := test.NewTempApp(GinkgoTB())
		p := newParts(a,
			mockprofile.NewMockProfileService(GinkgoT()), mockmove.NewMockMoveService(GinkgoT()),
			mocktraining.NewMockTrainingService(GinkgoT()), mockitem.NewMockItemService(GinkgoT()), nil)

		w := newWindow(a, p, nil)
		Expect(w.Content().MinSize().Height).To(BeNumerically(">=", minHeight))
		Expect(minHeight).To(BeNumerically("<", startSize.Height))
	})
})

var _ = Describe("loaders", func() {
	It("lists every page, which are all asked at the rate in the config", func() {
		test.NewTempApp(GinkgoTB())
		p := newParts(nil,
			mockprofile.NewMockProfileService(GinkgoT()), mockmove.NewMockMoveService(GinkgoT()),
			mocktraining.NewMockTrainingService(GinkgoT()), mockitem.NewMockItemService(GinkgoT()), nil)

		Expect(loaders(p)).To(ConsistOf(p.Dashboard, p.Moves, p.Training, p.Items, p.Route, p.Errantry))
	})
})

var _ = Describe("the dependency graph", func() {
	It("can build the window from everything it needs", func() {
		Expect(fx.ValidateApp(options(), fx.Populate(new(fyne.Window)))).To(Succeed())
	})
})
