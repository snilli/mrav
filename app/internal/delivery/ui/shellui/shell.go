// Package shellui is the one screen of the viewer: the monster's header on top, and under it from one to three
// columns. Each column shows one page, which it is told to by its choice, from the vitals, the moves, the training, the
// errantry, the ranch, the route and the items; a button adds a column, and another takes one away, and the window grows and shrinks
// by the width of the column, to be small enough to sit beside the game. In the demo, a check box in the header shows
// the panel that sets the monster by hand.
package shellui

import (
	"image/color"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"app/internal/delivery/ui/demoui"
	"app/internal/delivery/ui/errantryui"
	"app/internal/delivery/ui/itemui"
	"app/internal/delivery/ui/moveui"
	"app/internal/delivery/ui/profileui"
	"app/internal/delivery/ui/routeui"
	"app/internal/delivery/ui/trainingui"
)

const (
	maxColumns  = 3
	drawerWidth = 340 // the demo panel's, which scrolls and so would otherwise ask for nothing
)

// Parts are the pages the shell puts on its screen.
type Parts struct {
	Dashboard *profileui.Dashboard
	Training  *trainingui.Screen
	Moves     *moveui.Screen
	Items     *itemui.Screen
	Route     *routeui.Screen
	Errantry  *errantryui.Screen
	Demo      *demoui.Panel // nil when a game is read, because nobody sets that monster by hand
}

// An option is a page that a column can show, by the name that it is chosen by.
type option struct {
	name   string
	object fyne.CanvasObject
}

// A column shows one option, which its choice picks; its buttons add a column after it, when it is the last, and take
// it away, when it is not the only one.
type column struct {
	choice *widget.Select
	remove *widget.Button
	add    *widget.Button
	body   *fyne.Container // holds what is shown
	box    fyne.CanvasObject
	shown  *option
}

func (c *column) show(o *option) {
	c.shown = o
	c.body.Objects = []fyne.CanvasObject{o.object}
	c.body.Refresh()
}

// Shell is the content of the window.
type Shell struct {
	win     fyne.Window
	dash    *profileui.Dashboard
	options []*option
	columns []*column
	row     *fyne.Container // the columns, side by side and as wide as each other
	// taken is how much narrower the window got for each column that was taken away, last first to be given back. It is
	// less than the width of the column when the window would not get any narrower, and giving back the width of the
	// column would make the window wider each time.
	taken      []float32
	demoToggle *widget.Check     // shows and hides the demo drawer; nil without a panel
	demo       fyne.CanvasObject // the drawer for the demo panel, hidden until asked for; nil without a panel
	content    fyne.CanvasObject
}

// New puts the pages on a screen for the window to show, which is the one that grows and shrinks with its columns.
func New(win fyne.Window, p Parts) *Shell {
	s := &Shell{win: win, dash: p.Dashboard, row: container.New(layout.NewGridLayoutWithColumns(1))}
	scrolled := func(o fyne.CanvasObject) fyne.CanvasObject { return container.NewVScroll(o) }
	s.options = []*option{
		{"Vitals", p.Dashboard.Vitals()}, // these two scroll already
		{"Moves", p.Moves.Content()},
		{"Training", scrolled(p.Training.Content())},
		{"Errantry", scrolled(p.Errantry.Content())},
		{"Ranch", scrolled(p.Dashboard.Ranch())},
		{"Route", scrolled(p.Route.Content())},
		{"Items", scrolled(p.Items.Content())},
	}
	s.columns = []*column{s.newColumn(s.options[0])}
	s.sync()

	top := p.Dashboard.Header()
	var drawer fyne.CanvasObject
	if p.Demo != nil {
		width := canvas.NewRectangle(color.Transparent)
		width.SetMinSize(fyne.NewSize(drawerWidth, 0))
		s.demo = container.NewStack(width, p.Demo.Content())
		s.demo.Hide()
		s.demoToggle = widget.NewCheck("Demo panel", s.showDemo)
		drawer = s.demo
		top = container.NewVBox(container.NewHBox(layout.NewSpacer(), s.demoToggle), top)
	}
	s.content = container.NewBorder(container.NewVBox(top, widget.NewSeparator()), nil, nil, drawer, s.row)
	return s
}

func (s *Shell) newColumn(o *option) *column {
	c := &column{body: container.NewStack()}
	c.choice = widget.NewSelect(nil, func(name string) { s.choose(c, name) })
	c.remove = widget.NewButtonWithIcon("", theme.CancelIcon(), func() { s.removeColumn(c) })
	c.add = widget.NewButtonWithIcon("", theme.ContentAddIcon(), s.addColumn)
	bar := container.NewBorder(nil, nil, nil, container.NewHBox(c.add, c.remove), c.choice)
	c.box = container.NewBorder(bar, nil, nil, nil, c.body)
	c.show(o)
	return c
}

// offered are the names that a column may choose from: the pages that no other column shows, and its own.
func (s *Shell) offered(c *column) []string {
	var names []string
	for _, o := range s.options {
		if !slices.ContainsFunc(s.columns, func(other *column) bool { return other != c && other.shown == o }) {
			names = append(names, o.name)
		}
	}
	return names
}

func (s *Shell) choose(c *column, name string) {
	for _, o := range s.options {
		if o.name == name {
			c.show(o)
		}
	}
	s.sync()
}

// sync makes everything that depends on the columns agree with them: what each offers and has buttons for, how they are
// laid out, and how the header is: on more lines, to be narrow, for one column.
func (s *Shell) sync() {
	n := len(s.columns)
	boxes := make([]fyne.CanvasObject, n)
	for i, c := range s.columns {
		c.choice.Options, c.choice.Selected = s.offered(c), c.shown.name
		c.choice.Refresh()
		setVisible(c.remove, n > 1)
		setVisible(c.add, i == n-1 && n < maxColumns)
		boxes[i] = c.box
	}
	s.row.Layout = layout.NewGridLayoutWithColumns(n)
	s.row.Objects = boxes
	s.row.Refresh()
	s.dash.Compact(n == 1)
	if s.content != nil {
		s.content.Refresh()
	}
}

// addColumn shows the first page that no column shows, in a new column, and widens the window by one: by what it was
// narrowed by, when a column was taken away, and by as much as a column now is, when not.
func (s *Shell) addColumn() {
	n := len(s.columns)
	if n >= maxColumns {
		return
	}
	for _, o := range s.options {
		if !slices.ContainsFunc(s.columns, func(c *column) bool { return c.shown == o }) {
			room := (s.row.Size().Width + theme.Padding()) / float32(n) // a column and the gap beside it
			if k := len(s.taken); k > 0 {
				room, s.taken = s.taken[k-1], s.taken[:k-1]
			}
			s.columns = append(s.columns, s.newColumn(o))
			s.sync()
			s.grow(room)
			return
		}
	}
}

func (s *Shell) removeColumn(c *column) {
	if len(s.columns) < 2 {
		return
	}
	room := c.box.Size().Width + theme.Padding()
	s.columns = slices.DeleteFunc(s.columns, func(o *column) bool { return o == c })
	s.sync()
	s.taken = append(s.taken, -s.grow(-room))
}

// grow makes the window wider by dw, or narrower when it is below zero, and says by how much it did change, which is
// less than asked when the window would not get as narrow as that.
func (s *Shell) grow(dw float32) float32 {
	before := s.win.Canvas().Size()
	s.win.Resize(fyne.NewSize(before.Width+dw, before.Height))
	return s.win.Canvas().Size().Width - before.Width
}

func setVisible(o fyne.CanvasObject, on bool) {
	if on {
		o.Show()
	} else {
		o.Hide()
	}
}

// Content is what to put in the window.
func (s *Shell) Content() fyne.CanvasObject { return s.content }

func (s *Shell) showDemo(on bool) {
	setVisible(s.demo, on)
	s.content.Refresh()
}
