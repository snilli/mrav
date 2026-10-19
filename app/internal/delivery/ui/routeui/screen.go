package routeui

import (
	"context"
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"app/internal/core/profile"
	"app/internal/core/profile/domain"
	"app/internal/delivery/ui/profileui"
	"app/internal/delivery/ui/status"
)

// Maps are the maps of expeditions that a route is planned on. There are none where no map is known.
type Maps func() []domain.ExpeditionMap

// A tile of the table is cellSize on a side, with cellGap between.
const (
	cellSize = 17
	cellGap  = 1
)

// tableLayout puts its objects in rows of cols tiles, and asks for no more than the table is. Nothing in it
// scrolls: the table is as wide as a column of the window, and a scroll of its own would take the wheel away from
// the column, which has to scroll to the panels below.
type tableLayout struct{ cols int }

func (t tableLayout) MinSize(objs []fyne.CanvasObject) fyne.Size {
	rows := (len(objs) + t.cols - 1) / t.cols
	return fyne.NewSize(float32(t.cols*(cellSize+cellGap)-cellGap), float32(rows*(cellSize+cellGap)-cellGap))
}

func (t tableLayout) Layout(objs []fyne.CanvasObject, _ fyne.Size) {
	for i, o := range objs {
		o.Move(fyne.NewPos(float32(i%t.cols*(cellSize+cellGap)), float32(i/t.cols*(cellSize+cellGap))))
		o.Resize(fyne.NewSize(cellSize, cellSize))
	}
}

// The fills of the tiles of the table.
var (
	floorFill = color.NRGBA{R: 96, G: 102, B: 112, A: 255}
	wallFill  = color.NRGBA{R: 24, G: 27, B: 33, A: 255}
	rockFill  = color.NRGBA{R: 139, G: 90, B: 43, A: 255}
	routeFill = color.NRGBA{R: 235, G: 150, B: 15, A: 255}
	startFill = color.NRGBA{R: 65, G: 105, B: 225, A: 255}
	stopFill  = color.NRGBA{R: 50, G: 180, B: 50, A: 255} // a place that the route takes
	spotFill  = color.NRGBA{R: 70, G: 110, B: 85, A: 255} // a place that it leaves out

	// markerFill is the tile that the monster is on: the colour of the box that gives its place above the map.
	markerFill = color.NRGBA{R: 255, G: 105, B: 180, A: 255}

	gateFill     = color.NRGBA{R: 160, G: 200, B: 80, A: 255} // a gate that the monster can find the way through
	shutFill     = color.NRGBA{R: 80, G: 92, B: 62, A: 255}   // a gate that it cannot
	teleportFill = color.NRGBA{R: 160, G: 110, B: 210, A: 255}
)

const legend = "S start · orange: the route · green letters: places, bright where the route searches · brown: rocks · " +
	"light green number: a hidden way that wants that many hundreds of INT, dim when the monster has too little · " +
	"purple T: a teleport, which the route does not use · dark: walls"

// cell is a tile of the table: its fill, and the letter on it where it has one.
type cell struct {
	bg     *canvas.Rectangle
	text   *canvas.Text
	object fyne.CanvasObject
}

func newCell() *cell {
	c := &cell{bg: canvas.NewRectangle(color.Transparent), text: canvas.NewText("", color.White)}
	c.bg.SetMinSize(fyne.NewSize(cellSize, cellSize))
	c.text.TextSize = 11
	c.text.TextStyle.Bold = true
	c.text.Alignment = fyne.TextAlignCenter
	c.object = container.NewStack(c.bg, container.NewCenter(c.text))
	return c
}

// newTable is the tiles of a map as a table, with the route and the places marked in it.
func newTable(v View) (fyne.CanvasObject, []*cell) {
	m := v.Map
	onRoute := make(map[domain.Point]bool, len(v.Route.Path))
	for _, p := range v.Route.Path {
		onRoute[p] = true
	}
	taken := make(map[domain.Point]bool, len(v.Route.Spots))
	for _, s := range v.Route.Spots {
		taken[s.At] = true
	}
	spots := make(map[domain.Point]string, len(m.Spots))
	for _, s := range m.Spots {
		spots[s.At] = s.Name
	}
	locks := make(map[domain.Point]int, len(m.Locks))
	for _, l := range m.Locks {
		locks[l.At] = l.MinINT
	}

	cells := make([]*cell, len(m.Tiles))
	objects := make([]fyne.CanvasObject, len(m.Tiles))
	for i, tile := range m.Tiles {
		p := domain.Point{X: i % m.Width, Y: i / m.Width}
		c := newCell()
		switch name, isSpot := spots[p]; {
		case tile == domain.Wall:
			c.bg.FillColor = wallFill
		case tile == domain.Teleport:
			c.bg.FillColor, c.text.Text = teleportFill, "T"
		case tile == domain.Gate:
			c.text.Text = strconv.Itoa(locks[p] / 100)
			c.bg.FillColor = shutFill
			if v.Intelligence >= locks[p] {
				c.bg.FillColor = gateFill
			}
		case p == m.Start:
			c.bg.FillColor, c.text.Text = startFill, "S"
		case isSpot && taken[p]:
			c.bg.FillColor, c.text.Text = stopFill, name
		case isSpot:
			c.bg.FillColor, c.text.Text = spotFill, name
		case tile == domain.Rock:
			c.bg.FillColor = rockFill
			if onRoute[p] {
				c.text.Text = "•" // the route breaks through it
			}
		case onRoute[p]:
			c.bg.FillColor = routeFill
		default:
			c.bg.FillColor = floorFill
		}
		if v.Marker != nil && p == *v.Marker { // where the monster is
			c.bg.FillColor, c.text.Text, c.text.Color = markerFill, "★", color.Black
			c.bg.StrokeColor, c.bg.StrokeWidth = color.White, 2
		}
		cells[i], objects[i] = c, c.object
	}
	return container.New(tableLayout{cols: m.Width}, objects...), cells
}

// box is the panel of one map, closed until it is opened, with its title saying what the route is worth.
type box struct {
	name     string
	panel    *widget.Accordion
	item     *widget.AccordionItem
	summary  *widget.Label
	standing []*profileui.ToneBox // the energy, the place and the steps of the monster on the trip, above the map
	stops    *widget.Label
	credit   *widget.Label // where the map is from
	cells    []*cell
}

func newBox(v View, open bool) *box {
	b := &box{name: v.Name, summary: widget.NewLabel(v.Summary), stops: widget.NewLabel(strings.Join(v.Stops, "\n"))}
	b.summary.Wrapping = fyne.TextWrapWord
	if len(v.Stops) == 0 {
		b.stops.Hide()
	}
	b.credit = profileui.Quiet(v.Map.Source)
	b.credit.Wrapping = fyne.TextWrapWord
	if v.Map.Source == "" {
		b.credit.Hide()
	}
	var table fyne.CanvasObject
	table, b.cells = newTable(v)
	key := profileui.Quiet(legend)
	key.Wrapping = fyne.TextWrapWord
	content := container.NewVBox()
	if len(v.Standing) > 0 {
		row := container.NewHBox()
		for _, s := range v.Standing {
			box := profileui.NewToneBox().Large()
			box.Set(s)
			b.standing = append(b.standing, box)
			row.Add(box)
		}
		content.Add(row)
	}
	for _, o := range []fyne.CanvasObject{b.summary, table, key, b.stops, b.credit} {
		content.Add(o)
	}
	b.item = widget.NewAccordionItem(v.Title(), content)
	b.item.Open = open
	b.panel = widget.NewAccordion(b.item)
	return b
}

// Screen is the Route panel.
type Screen struct {
	svc  profile.ProfileService
	maps Maps
	all  bool // every map is shown, as the demo does; otherwise only the one whose trip the monster is on

	status  *widget.Label
	list    *fyne.Container // the panel of each map
	content *fyne.Container
	boxes   []*box
	drawn   string // what the panels were built from, so that they are built again only when it changes
}

// NewScreen is the Route panel of the maps given. In a game it shows the map of the trip that the monster is on,
// when it is on one; with all, as the demo has it, it shows them all.
func NewScreen(svc profile.ProfileService, maps Maps, all bool) *Screen {
	s := &Screen{
		svc:    svc,
		maps:   maps,
		all:    all,
		status: widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		list:   container.NewVBox(),
	}
	s.status.Hide()
	s.content = container.NewVBox(s.status, s.list)
	s.content.Hide() // until there is a map to show
	return s
}

// Content is what to put in a window.
func (s *Screen) Content() fyne.CanvasObject { return s.content }

// Load plans the routes for the monster as it is now and shows them. It is safe to call from any goroutine.
func (s *Screen) Load(ctx context.Context) {
	maps := s.maps()
	if len(maps) == 0 {
		fyne.Do(s.content.Hide)
		return
	}
	p, err := s.svc.Current(ctx)
	fyne.Do(func() { s.apply(maps, p, err) })
}

// apply shows the routes on the maps, or why there are none. It must run on the UI thread.
func (s *Screen) apply(maps []domain.ExpeditionMap, p domain.Profile, err error) {
	if err != nil {
		s.content.Show()
		s.status.SetText(status.Message(err))
		s.status.Show()
		s.clear()
		return
	}

	var views []View
	var problems []string
	var built strings.Builder
	for _, m := range maps {
		if !s.all && !m.Here(p.Player.Ranch) {
			continue
		}
		v, err := NewView(p, m)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", m.Name, err))
			continue
		}
		views = append(views, v)
		fmt.Fprint(&built, v.Name, "|", v.Map.Source, "|", v.Summary, "|", v.Stops, "|", v.Route.Path, ";")
	}
	drawn := built.String()
	s.status.SetText(strings.Join(problems, "\n"))
	if len(problems) > 0 {
		s.status.Show()
	} else {
		s.status.Hide()
	}
	if len(views) == 0 && len(problems) == 0 { // no map has its trip on: there is nothing to show
		s.clear()
		s.content.Hide()
		return
	}
	s.content.Show()
	if drawn == s.drawn && len(s.boxes) == len(views) {
		return
	}

	open := make(map[string]bool, len(s.boxes)) // a panel that was open stays open when the route changes under it
	for _, b := range s.boxes {
		open[b.name] = b.item.Open
	}
	s.clear()
	s.drawn = drawn
	for _, v := range views {
		was, known := open[v.Name]
		if !known { // the panel of the trip that the monster is on opens when it appears, and the others do not
			was = v.Marker != nil
		}
		b := newBox(v, was)
		s.boxes = append(s.boxes, b)
		s.list.Add(b.panel)
	}
	s.list.Refresh()
}

// clear takes every panel away.
func (s *Screen) clear() {
	s.list.RemoveAll()
	s.boxes, s.drawn = nil, ""
}
