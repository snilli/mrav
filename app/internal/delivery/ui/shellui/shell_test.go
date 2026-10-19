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
	"app/internal/delivery/ui/profileui"
	"app/internal/delivery/ui/trainingui"
	"app/internal/infrastructure/fakemonster"
	"app/internal/infrastructure/fakeplayer"
	"app/internal/infrastructure/icons"
	mockitem "app/mock/core/item"
	mockmove "app/mock/core/move"
	mockprofile "app/mock/core/profile"
	mocktraining "app/mock/core/training"
)

var _ = Describe("Shell", func() {
	var parts Parts

	BeforeEach(func() {
		test.NewTempApp(GinkgoTB())
		parts = Parts{
			Dashboard: profileui.NewDashboard(mockprofile.NewMockProfileService(GinkgoT())),
			Training:  trainingui.NewScreen(mocktraining.NewMockTrainingService(GinkgoT())),
			Moves:     moveui.NewScreen(mockmove.NewMockMoveService(GinkgoT()), icons.Bytes),
			Items:     itemui.NewScreen(mockitem.NewMockItemService(GinkgoT())),
		}
	})

	It("puts the pages on one screen", func() {
		Expect(New(parts).Content()).NotTo(BeNil())
	})

	It("puts the vitals, then the moves, then the training, from left to right", func() {
		win := test.NewWindow(New(parts).Content())
		win.Resize(fyne.NewSize(1500, 900))
		x := func(o fyne.CanvasObject) float32 { return fyne.CurrentApp().Driver().AbsolutePositionForObject(o).X }

		Expect(x(parts.Dashboard.Vitals())).To(BeNumerically("<", x(parts.Moves.Content())))
		Expect(x(parts.Moves.Content())).To(BeNumerically("<", x(parts.Training.Content())))
	})

	It("has no way to show a demo panel when a game is read, because there is none", func() {
		Expect(New(parts).demoToggle).To(BeNil())
	})

	Describe("with a demo panel", func() {
		BeforeEach(func() {
			genera := []movedomain.Genus{{Code: 0, Name: "Pixie"}}
			parts.Demo = demoui.New(fakemonster.New(), fakeplayer.New(), genera)
		})

		It("keeps the panel hidden until it is asked for, and hides it again when asked", func() {
			s := New(parts)
			Expect(s.demoToggle.Checked).To(BeFalse())
			Expect(s.demo.Visible()).To(BeFalse())

			test.Tap(s.demoToggle)
			Expect(s.demo.Visible()).To(BeTrue())

			test.Tap(s.demoToggle)
			Expect(s.demo.Visible()).To(BeFalse())
		})
	})
})
