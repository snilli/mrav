package shellui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	movedomain "app/internal/core/move/domain"
	"app/internal/delivery/ui/demoui"
	"app/internal/delivery/ui/errantryui"
	"app/internal/delivery/ui/itemui"
	"app/internal/delivery/ui/moveui"
	"app/internal/delivery/ui/profileui"
	"app/internal/delivery/ui/routeui"
	"app/internal/delivery/ui/trainingui"
	"app/internal/infrastructure/fakemonster"
	"app/internal/infrastructure/fakeplayer"
	"app/internal/infrastructure/icons"
	mockitem "app/mock/core/item"
	mockmove "app/mock/core/move"
	mockprofile "app/mock/core/profile"
	mocktraining "app/mock/core/training"
)

type floored struct {
	fyne.Window
	width float32
}

func (w floored) Resize(size fyne.Size) {
	w.Window.Resize(fyne.NewSize(max(size.Width, w.width), size.Height))
}

var _ = Describe("Shell", func() {
	var (
		parts Parts
		win   fyne.Window
	)

	BeforeEach(func() {
		test.NewTempApp(GinkgoTB())
		win = test.NewWindow(nil)
		parts = Parts{
			Dashboard: profileui.NewDashboard(mockprofile.NewMockProfileService(GinkgoT())),
			Training:  trainingui.NewScreen(mocktraining.NewMockTrainingService(GinkgoT())),
			Moves:     moveui.NewScreen(mockmove.NewMockMoveService(GinkgoT()), icons.Bytes),
			Items:     itemui.NewScreen(mockitem.NewMockItemService(GinkgoT())),
			Route:     routeui.NewScreen(mockprofile.NewMockProfileService(GinkgoT())),
			Errantry:  errantryui.NewScreen(mockprofile.NewMockProfileService(GinkgoT()), mockmove.NewMockMoveService(GinkgoT()), moveui.NewIcons(icons.Bytes)),
		}
	})

	It("puts the pages on one screen", func() {
		Expect(New(win, parts).Content()).NotTo(BeNil())
	})

	It("has no way to show a demo panel when a game is read, because there is none", func() {
		Expect(New(win, parts).demoToggle).To(BeNil())
	})

	Describe("with a demo panel", func() {
		BeforeEach(func() {
			genera := []movedomain.Genus{{Code: 0, Name: "Pixie"}}
			parts.Demo = demoui.New(fakemonster.New(), fakeplayer.New(), genera)
		})

		It("keeps the panel hidden until it is asked for, and hides it again when asked", func() {
			s := New(win, parts)
			Expect(s.demoToggle.Checked).To(BeFalse())
			Expect(s.demo.Visible()).To(BeFalse())

			test.Tap(s.demoToggle)
			Expect(s.demo.Visible()).To(BeTrue())

			test.Tap(s.demoToggle)
			Expect(s.demo.Visible()).To(BeFalse())
		})
	})

	Describe("its columns", func() {
		var s *Shell

		BeforeEach(func() {
			s = New(win, parts)
			win.SetContent(s.Content())
			win.Resize(fyne.NewSize(500, 700))
		})

		width := func() float32 { return win.Canvas().Size().Width }
		last := func() *column { return s.columns[len(s.columns)-1] }
		shown := func() []string {
			var names []string
			for _, c := range s.columns {
				names = append(names, c.shown.name)
			}
			return names
		}

		It("opens with one column, showing the vitals, with a button to add another and none to take it away", func() {
			Expect(shown()).To(Equal([]string{"Vitals"}))
			Expect(s.columns[0].add.Visible()).To(BeTrue())
			Expect(s.columns[0].remove.Visible()).To(BeFalse())
		})

		It("says on the add button what it does, so that it is seen to add a column", func() {
			Expect(s.columns[0].add.Text).To(Equal("Col"))
			Expect(s.columns[0].add.Importance).To(Equal(widget.LowImportance))
		})

		It("adds a column that shows the first page that is not shown, and the window gets a column wider", func() {
			before, room := width(), s.row.Size().Width

			test.Tap(s.columns[0].add)

			Expect(shown()).To(Equal([]string{"Vitals", "Moves"}))
			Expect(width()).To(BeNumerically("~", before+room+theme.Padding(), 1), "a column and the gap beside it")
			Expect(s.columns[0].add.Visible()).To(BeFalse(), "only the last column has the button")
			Expect(last().add.Visible()).To(BeTrue())
			Expect(s.columns[0].remove.Visible()).To(BeTrue())
			Expect(last().remove.Visible()).To(BeTrue())
		})

		It("puts the columns side by side, in the order that they were added, and stops at three", func() {
			test.Tap(last().add)
			test.Tap(last().add)

			Expect(shown()).To(Equal([]string{"Vitals", "Moves", "Training"}))
			Expect(last().add.Visible()).To(BeFalse())
			x := func(c *column) float32 { return fyne.CurrentApp().Driver().AbsolutePositionForObject(c.box).X }
			Expect(x(s.columns[0])).To(BeNumerically("<", x(s.columns[1])))
			Expect(x(s.columns[1])).To(BeNumerically("<", x(s.columns[2])))
		})

		It("offers each column the pages that no other column shows, and its own", func() {
			test.Tap(s.columns[0].add)

			Expect(s.columns[0].choice.Options).To(ConsistOf("Vitals", "Training", "Errantry", "Ranch", "Calendar", "Route", "Items"))
			Expect(s.columns[1].choice.Options).To(ConsistOf("Moves", "Training", "Errantry", "Ranch", "Calendar", "Route", "Items"))
		})

		It("shows the page that a column is told to choose, and offers the one that it showed to the others", func() {
			test.Tap(s.columns[0].add)

			s.columns[0].choice.SetSelected("Items")

			Expect(shown()).To(Equal([]string{"Items", "Moves"}))
			Expect(s.columns[1].choice.Options).To(ContainElement("Vitals"))
			Expect(s.columns[1].choice.Options).NotTo(ContainElement("Items"))
		})

		It("takes a column away with its button, and the window gives up the width that it had", func() {
			before := width()
			test.Tap(s.columns[0].add)

			test.Tap(last().remove)

			Expect(shown()).To(Equal([]string{"Vitals"}))
			Expect(width()).To(BeNumerically("~", before, 1))
			Expect(s.columns[0].remove.Visible()).To(BeFalse())
			Expect(s.columns[0].add.Visible()).To(BeTrue())
		})

		It("takes back no more than it gave up, when the window would not get as narrow as a column", func() {
			s = New(floored{win, 900}, parts)
			win.SetContent(s.Content())
			win.Resize(fyne.NewSize(500, 700))
			test.Tap(last().add)
			test.Tap(last().add)
			win.Resize(fyne.NewSize(1500, 700))

			for range 3 {
				test.Tap(last().remove)
				test.Tap(last().remove)
				test.Tap(last().add)
				test.Tap(last().add)
			}

			Expect(width()).To(BeNumerically("~", 1500, 1))
		})

		It("puts the header on more lines while there is one column, and on fewer when there are more", func() {
			one := parts.Dashboard.Header().MinSize().Height

			test.Tap(s.columns[0].add)

			Expect(parts.Dashboard.Header().MinSize().Height).To(BeNumerically("<", one))
		})
	})
})
