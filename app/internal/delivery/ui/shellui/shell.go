// Package shellui is the one screen of the viewer: the monster's header on top, its vitals, its moves and its
// training side by side below it, and its items along the bottom. In the demo, a check box in the header shows the
// panel that sets the monster by hand.
package shellui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"app/internal/delivery/ui/demoui"
	"app/internal/delivery/ui/itemui"
	"app/internal/delivery/ui/moveui"
	"app/internal/delivery/ui/profileui"
	"app/internal/delivery/ui/trainingui"
)

const (
	vitalsShare = 0.34 // of the width; the moves and the training share the rest
	movesShare  = 0.43 // of what the vitals leave, in the middle; the training has the rest, on the right
	drawerWidth = 340  // the demo panel's, which scrolls and so would otherwise ask for nothing
)

// Parts are the pages the shell puts on its screen.
type Parts struct {
	Dashboard *profileui.Dashboard
	Training  *trainingui.Screen
	Moves     *moveui.Screen
	Items     *itemui.Screen
	Demo      *demoui.Panel // nil when a game is read, because nobody sets that monster by hand
}

// Shell is the content of the window.
type Shell struct {
	demoToggle *widget.Check     // shows and hides the demo drawer; nil without a panel
	demo       fyne.CanvasObject // the drawer for the demo panel, hidden until asked for; nil without a panel
	content    fyne.CanvasObject
}

func New(p Parts) *Shell {
	s := &Shell{}

	var drawer, toggle fyne.CanvasObject
	if p.Demo != nil {
		width := canvas.NewRectangle(color.Transparent)
		width.SetMinSize(fyne.NewSize(drawerWidth, 0))
		s.demo = container.NewStack(width, p.Demo.Content())
		s.demo.Hide()
		s.demoToggle = widget.NewCheck("Demo panel", s.showDemo)
		drawer, toggle = s.demo, container.NewCenter(s.demoToggle)
	}

	header := container.NewBorder(nil, widget.NewSeparator(), nil, toggle, p.Dashboard.Header())
	// The right column is panels, each closed until it is opened: the training, the ranch's upgrades, the breeds that
	// are unlocked, and the expeditions. They scroll together, for when several are open.
	right := container.NewVScroll(container.NewVBox(p.Training.Content(), p.Dashboard.Ranch()))
	movesAndTraining := container.NewHSplit(p.Moves.Content(), right)
	movesAndTraining.SetOffset(movesShare)
	columns := container.NewHSplit(p.Dashboard.Vitals(), movesAndTraining)
	columns.SetOffset(vitalsShare)
	items := container.NewVBox(widget.NewSeparator(), p.Items.Content())

	s.content = container.NewBorder(header, items, nil, drawer, columns)
	return s
}

// Content is what to put in the window.
func (s *Shell) Content() fyne.CanvasObject { return s.content }

func (s *Shell) showDemo(on bool) {
	if on {
		s.demo.Show()
	} else {
		s.demo.Hide()
	}
	s.content.Refresh()
}
