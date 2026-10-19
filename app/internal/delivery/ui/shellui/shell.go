// Package shellui is the one screen of the viewer: the monster's header on top, its vitals, its training and its
// moves side by side below it, and its items along the bottom. What is looked at seldom, the Controller, the
// Breeder and the demo panel, waits behind a Tools menu.
package shellui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"app/internal/delivery/ui/demoui"
	"app/internal/delivery/ui/itemui"
	"app/internal/delivery/ui/moveui"
	"app/internal/delivery/ui/playerui"
	"app/internal/delivery/ui/profileui"
	"app/internal/delivery/ui/trainingui"
)

const (
	vitalsShare   = 0.34 // of the width; the training and the moves share the rest
	trainingShare = 0.57 // of what the vitals leave
	drawerWidth   = 340  // the demo panel's, which scrolls and so would otherwise ask for nothing
)

// Parts are the pages the shell puts on its screen.
type Parts struct {
	Dashboard  *profileui.Dashboard
	Training   *trainingui.Screen
	Moves      *moveui.Screen
	Items      *itemui.Screen
	Breeder    *playerui.BreederScreen
	Controller *playerui.ControllerScreen
	Demo       *demoui.Panel // nil when a game is read, because nobody sets that monster by hand
}

// Shell is the content of the window.
type Shell struct {
	win     fyne.Window
	parts   Parts
	tools   *widget.Button
	demo    fyne.CanvasObject // the drawer for the demo panel, hidden until asked for; nil without a panel
	content fyne.CanvasObject
}

func New(win fyne.Window, p Parts) *Shell {
	s := &Shell{win: win, parts: p}
	s.tools = widget.NewButtonWithIcon("Tools", theme.MenuDropDownIcon(), s.showMenu)
	s.tools.IconPlacement = widget.ButtonIconTrailingText

	header := container.NewBorder(nil, widget.NewSeparator(), nil, container.NewCenter(s.tools), p.Dashboard.Header())
	trainingAndMoves := container.NewHSplit(p.Training.Content(), p.Moves.Content())
	trainingAndMoves.SetOffset(trainingShare)
	columns := container.NewHSplit(p.Dashboard.Vitals(), trainingAndMoves)
	columns.SetOffset(vitalsShare)
	items := container.NewVBox(widget.NewSeparator(), p.Items.Content())

	var drawer fyne.CanvasObject
	if p.Demo != nil {
		width := canvas.NewRectangle(color.Transparent)
		width.SetMinSize(fyne.NewSize(drawerWidth, 0))
		s.demo = container.NewStack(width, p.Demo.Content())
		s.demo.Hide()
		drawer = s.demo
	}
	s.content = container.NewBorder(header, items, nil, drawer, columns)
	return s
}

// Content is what to put in the window.
func (s *Shell) Content() fyne.CanvasObject { return s.content }

// menu is what Tools offers, made anew each time so that it shows how things stand.
func (s *Shell) menu() *fyne.Menu {
	items := []*fyne.MenuItem{
		fyne.NewMenuItem("Controller", func() { s.show("Controller", s.parts.Controller.Content()) }),
		fyne.NewMenuItem("Breeder", func() { s.show("Breeder", s.parts.Breeder.Content()) }),
	}
	if s.demo != nil {
		panel := fyne.NewMenuItem("Demo panel", s.toggleDemo)
		panel.Checked = s.demo.Visible()
		items = append(items, fyne.NewMenuItemSeparator(), panel)
	}
	return fyne.NewMenu("", items...)
}

// showMenu drops the menu down from the Tools button.
func (s *Shell) showMenu() {
	below := fyne.CurrentApp().Driver().AbsolutePositionForObject(s.tools).AddXY(0, s.tools.Size().Height)
	widget.NewPopUpMenu(s.menu(), s.win.Canvas()).ShowAtPosition(below)
}

func (s *Shell) show(title string, content fyne.CanvasObject) {
	dialog.ShowCustom(title, "Close", content, s.win)
}

func (s *Shell) toggleDemo() {
	if s.demo.Visible() {
		s.demo.Hide()
	} else {
		s.demo.Show()
	}
	s.content.Refresh()
}
