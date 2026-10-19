package shellui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	movedomain "app/internal/core/move/domain"
	"app/internal/delivery/ui/demoui"
	"app/internal/delivery/ui/itemui"
	"app/internal/delivery/ui/moveui"
	"app/internal/delivery/ui/playerui"
	"app/internal/delivery/ui/profileui"
	"app/internal/delivery/ui/trainingui"
	"app/internal/infrastructure/fakemonster"
	"app/internal/infrastructure/fakeplayer"
	"app/internal/infrastructure/icons"
	mockitem "app/mock/core/item"
	mockmove "app/mock/core/move"
	mockplayer "app/mock/core/player"
	mockprofile "app/mock/core/profile"
	mocktraining "app/mock/core/training"
)

var _ = Describe("Shell", func() {
	var (
		win   fyne.Window
		parts Parts
	)

	// labels are what the Tools menu offers, in order.
	labels := func(s *Shell) []string {
		var out []string
		for _, it := range s.menu().Items {
			if !it.IsSeparator {
				out = append(out, it.Label)
			}
		}
		return out
	}
	// choose does what a tap on an entry of the Tools menu does.
	choose := func(s *Shell, label string) {
		for _, it := range s.menu().Items {
			if it.Label == label {
				it.Action()
				return
			}
		}
		Fail("the Tools menu has no " + label)
	}
	shell := func() *Shell {
		s := New(win, parts)
		win.SetContent(s.Content())
		win.Resize(fyne.NewSize(1500, 900))
		return s
	}

	BeforeEach(func() {
		test.NewTempApp(GinkgoTB())
		win = test.NewWindow(nil)
		parts = Parts{
			Dashboard:  profileui.NewDashboard(mockprofile.NewMockProfileService(GinkgoT())),
			Training:   trainingui.NewScreen(mocktraining.NewMockTrainingService(GinkgoT())),
			Moves:      moveui.NewScreen(mockmove.NewMockMoveService(GinkgoT()), icons.Bytes),
			Items:      itemui.NewScreen(mockitem.NewMockItemService(GinkgoT())),
			Breeder:    playerui.NewBreederScreen(mockplayer.NewMockPlayerService(GinkgoT())),
			Controller: playerui.NewControllerScreen(mockplayer.NewMockPlayerService(GinkgoT())),
		}
	})

	It("keeps the Controller and the Breeder behind a Tools menu", func() {
		Expect(labels(shell())).To(Equal([]string{"Controller", "Breeder"}))
	})

	It("opens the menu when Tools is tapped", func() {
		s := shell()
		Expect(win.Canvas().Overlays().Top()).To(BeNil())
		test.Tap(s.tools)
		Expect(win.Canvas().Overlays().Top()).NotTo(BeNil())
	})

	It("shows the Controller and the Breeder in a dialog each", func() {
		s := shell()
		choose(s, "Controller")
		Expect(win.Canvas().Overlays().Top()).NotTo(BeNil())
		win.Canvas().Overlays().Remove(win.Canvas().Overlays().Top())

		choose(s, "Breeder")
		Expect(win.Canvas().Overlays().Top()).NotTo(BeNil())
	})

	Describe("with a demo panel", func() {
		BeforeEach(func() {
			genera := []movedomain.Genus{{Code: 0, Name: "Pixie"}}
			parts.Demo = demoui.New(fakemonster.New(), fakeplayer.New(), genera)
		})

		It("offers the panel in the menu, checked while it is shown", func() {
			s := shell()
			Expect(labels(s)).To(Equal([]string{"Controller", "Breeder", "Demo panel"}))
			for _, it := range s.menu().Items {
				if it.Label == "Demo panel" {
					Expect(it.Checked).To(BeFalse())
				}
			}
		})

		It("keeps the panel hidden until it is asked for, and hides it again when asked", func() {
			s := shell()
			Expect(s.demo.Visible()).To(BeFalse())

			choose(s, "Demo panel")
			Expect(s.demo.Visible()).To(BeTrue())
			for _, it := range s.menu().Items {
				if it.Label == "Demo panel" {
					Expect(it.Checked).To(BeTrue())
				}
			}

			choose(s, "Demo panel")
			Expect(s.demo.Visible()).To(BeFalse())
		})
	})
})
