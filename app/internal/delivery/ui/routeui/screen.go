package routeui

import (
	"cmp"
	"context"
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"app/internal/core/profile"
	"app/internal/core/profile/domain"
	"app/internal/delivery/ui/profileui"
	"app/internal/delivery/ui/status"
)

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
	if t.cols <= 0 {
		return fyne.Size{}
	}
	rows := (len(objs) + t.cols - 1) / t.cols
	return fyne.NewSize(float32(t.cols*(cellSize+cellGap)-cellGap), float32(rows*(cellSize+cellGap)-cellGap))
}

func (t tableLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	if t.cols <= 0 {
		return
	}
	left := max((size.Width-t.MinSize(objs).Width)/2, 0) // the table is in the middle of what it is given
	for i, o := range objs {
		o.Move(fyne.NewPos(left+float32(i%t.cols*(cellSize+cellGap)), float32(i/t.cols*(cellSize+cellGap))))
		o.Resize(fyne.NewSize(cellSize, cellSize))
	}
}

// tableSize is the size, in tiles, of the table that every map is drawn on: as wide and as tall as the biggest of the maps, so
// that the maps are as big as each other, with a smaller one in the middle of it and walls round it.
type tableSize struct{ w, h int }

// rowsOf is how many rows of tiles a map has.
func rowsOf(m domain.ExpeditionMap) int {
	if m.Width <= 0 {
		return 0
	}
	return (len(m.Tiles) + m.Width - 1) / m.Width
}

func tableSizeOf(maps []domain.ExpeditionMap) tableSize {
	var t tableSize
	for _, m := range maps {
		t.w, t.h = max(t.w, m.Width), max(t.h, rowsOf(m))
	}
	return t
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

// newTable is the tiles of a map as a table of the size given, with the route and the places marked in it. The map is in the
// middle of the table, and the tiles round it are walls, which are padding and not the map's own tiles: cells are those only.
func newTable(v View, size tableSize) (table fyne.CanvasObject, cells, padding []*cell) {
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
	labels := make(map[domain.Point]string, len(m.Portals)) // T1, E1, ... as the guide writes them on its maps
	for _, p := range m.Portals {
		labels[p.At] = p.Label
	}

	tileCell := func(p domain.Point, tile domain.Terrain) *cell {
		c := newCell()
		switch name, isSpot := spots[p]; {
		case tile == domain.Wall:
			c.bg.FillColor = wallFill
		case tile == domain.Teleport:
			c.bg.FillColor, c.text.Text = teleportFill, cmp.Or(labels[p], "T")
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
			c.bg.FillColor, c.text.Text = routeFill, labels[p] // a place to land, when it is one, keeps its label
		case labels[p] != "":
			c.bg.FillColor, c.text.Text = teleportFill, labels[p]
		default:
			c.bg.FillColor = floorFill
		}
		if v.Marker != nil && p == *v.Marker { // where the monster is
			c.bg.FillColor, c.text.Text, c.text.Color = markerFill, "★", color.Black
			c.bg.StrokeColor, c.bg.StrokeWidth = color.White, 2
		}
		return c
	}

	offX, offY := (size.w-m.Width)/2, (size.h-rowsOf(m))/2
	objects := make([]fyne.CanvasObject, 0, size.w*size.h)
	for y := 0; y < size.h; y++ {
		for x := 0; x < size.w; x++ {
			p := domain.Point{X: x - offX, Y: y - offY}
			var c *cell
			if i := p.Y*m.Width + p.X; p.X >= 0 && p.X < m.Width && p.Y >= 0 && i < len(m.Tiles) {
				c = tileCell(p, m.Tiles[i])
				cells = append(cells, c)
			} else { // outside the map
				c = newCell()
				c.bg.FillColor = wallFill
				padding = append(padding, c)
			}
			objects = append(objects, c.object)
		}
	}
	return container.New(tableLayout{cols: size.w}, objects...), cells, padding
}

// keyColumns is how many tiles of the key, with their words, stand side by side.
const keyColumns = 2

// newKey is the key to a map: each kind of tile that the map has, drawn as it is drawn on the map, with a few words beside it.
// A kind that the map does not have is left out.
func newKey(v View) (fyne.CanvasObject, []keyItem) {
	m := v.Map
	onRoute := make(map[domain.Point]bool, len(v.Route.Path))
	for _, p := range v.Route.Path {
		onRoute[p] = true
	}
	taken := make(map[domain.Point]bool, len(v.Route.Spots))
	for _, s := range v.Route.Spots {
		taken[s.At] = true
	}

	var items []keyItem
	add := func(fill color.Color, text, label string) *cell {
		c := newCell()
		c.bg.FillColor, c.text.Text = fill, text
		items = append(items, keyItem{swatch: c, label: widget.NewLabel(label)})
		return c
	}

	add(startFill, "S", "Start")
	if v.Marker != nil {
		c := add(markerFill, "★", "Where the monster is")
		c.text.Color = color.Black
		c.bg.StrokeColor, c.bg.StrokeWidth = color.White, 2
	}
	if len(v.Route.Spots) > 0 {
		add(routeFill, "", "The route")
		add(stopFill, v.Route.Spots[0].Name, "A place the route searches")
	}
	for _, s := range m.Spots {
		if !taken[s.At] {
			add(spotFill, s.Name, "A place it leaves out")
			break
		}
	}

	var loose, broken bool // rocks that the route leaves, and rocks that it breaks
	teleport := false
	for i, tile := range m.Tiles {
		p := domain.Point{X: i % m.Width, Y: i / m.Width}
		switch {
		case tile == domain.Rock && onRoute[p]:
			broken = true
		case tile == domain.Rock:
			loose = true
		case tile == domain.Teleport:
			teleport = true
		}
	}
	if loose {
		add(rockFill, "", "Rock")
	}
	if broken {
		add(rockFill, "•", "Rock the route breaks")
	}

	var open, shut *domain.Lock
	for i, l := range m.Locks {
		switch {
		case v.Intelligence >= l.MinINT && open == nil:
			open = &m.Locks[i]
		case v.Intelligence < l.MinINT && shut == nil:
			shut = &m.Locks[i]
		}
	}
	if open != nil {
		add(gateFill, strconv.Itoa(open.MinINT/100), fmt.Sprintf("Hidden way, needs %d INT", open.MinINT))
	}
	if shut != nil {
		add(shutFill, strconv.Itoa(shut.MinINT/100), "Hidden way, INT too low")
	}

	if teleport {
		label := "T"
		for _, p := range m.Portals {
			if len(p.To) > 0 {
				label = p.Label
				break
			}
		}
		add(teleportFill, label, "Teleport, not on the route")
	}
	for _, p := range m.Portals {
		if len(p.To) == 0 {
			add(teleportFill, p.Label, "Where a teleport lands")
			break
		}
	}

	objects := make([]fyne.CanvasObject, len(items))
	for i, k := range items {
		objects[i] = container.NewHBox(container.NewCenter(k.swatch.object), k.label)
	}
	return container.NewGridWithColumns(keyColumns, objects...), items
}

// lootColumns is how many items of a place are written side by side.
const lootColumns = 3

// placeRow is a place under the map: its letter, its building, where it stands in the route and what it is worth, and
// what a search of it may turn up, with the chance of each.
type placeRow struct {
	letter   *profileui.ToneBox
	building *widget.Label
	note     *widget.Label
	loot     []*widget.Label
	empty    *widget.Label // says that nothing turns up there, when nothing does
	object   fyne.CanvasObject
}

func newPlaceRow(p PlaceView) *placeRow {
	r := &placeRow{
		letter:   profileui.NewToneBox(),
		building: widget.NewLabelWithStyle(p.Building, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		note:     profileui.Quiet(""),
		empty:    profileui.Quiet("Nothing turns up here for this monster yet"),
	}
	tone, note := profileui.ToneGray, "not on the route"
	if p.Stop > 0 {
		tone, note = profileui.ToneLimeGreen, fmt.Sprintf("stop %d", p.Stop)
	}
	if p.Worth != "" {
		note += " · " + p.Worth
	}
	r.letter.Set(profileui.Box{Text: p.Name, Tone: tone})
	r.note.SetText(note)

	cells := make([]fyne.CanvasObject, len(p.Loot))
	for i, d := range p.Loot {
		l := widget.NewLabelWithStyle(d.Item+" "+d.Chance, fyne.TextAlignLeading, fyne.TextStyle{Bold: d.Story})
		r.loot = append(r.loot, l)
		cells[i] = l
	}
	body := container.NewVBox(container.NewHBox(r.letter, r.building, layout.NewSpacer(), r.note))
	if len(cells) > 0 {
		r.empty.Hide()
		body.Add(container.NewGridWithColumns(lootColumns, cells...))
	}
	body.Add(r.empty)
	body.Add(widget.NewSeparator())
	r.object = body
	return r
}

// keyItem is a tile of the key to the map, the tile itself as it is drawn there, and a few words beside it.
type keyItem struct {
	swatch *cell
	label  *widget.Label
}

// box is one map: a heading that says what the route is worth, the monster's standing on the trip, the table, and the places.
type box struct {
	name      string
	title     *widget.Label
	object    fyne.CanvasObject
	summary   *widget.Label
	key       []keyItem
	teleports *widget.Label // where each teleport leads, under the key to the map; hidden where the map has none
	table     fyne.CanvasObject
	standing  []*profileui.ToneBox // the energy, the place and the steps of the monster on the trip, above the map
	places    []*placeRow          // the places of the map, the route's first
	cells     []*cell              // the tiles of the map, in its order
	padding   []*cell              // the walls that make the table as big as the biggest map's
}

func newBox(v View, size tableSize) *box {
	b := &box{
		name:    v.Name,
		title:   widget.NewLabelWithStyle(v.Title(), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		summary: widget.NewLabel(v.Summary),
	}
	b.teleports = profileui.Quiet(strings.Join(v.Teleports, "\n"))
	b.teleports.Wrapping = fyne.TextWrapWord
	b.teleports.Hidden = len(v.Teleports) == 0
	b.summary.Wrapping = fyne.TextWrapWord
	table, cells, padding := newTable(v, size)
	b.table, b.cells, b.padding = table, cells, padding
	key, items := newKey(v)
	b.key = items
	content := container.NewVBox(b.title)
	if len(v.Standing) > 0 {
		row := container.NewHBox()
		for _, s := range v.Standing {
			box := profileui.NewToneBox().Large()
			box.Set(s)
			b.standing = append(b.standing, box)
			row.Add(box)
		}
		content.Add(container.NewCenter(row))
	}
	for _, o := range []fyne.CanvasObject{b.summary, table, key, b.teleports} {
		content.Add(o)
	}
	if len(v.Places) > 0 {
		content.Add(widget.NewSeparator())
		content.Add(widget.NewLabelWithStyle("Places", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
		content.Add(profileui.Quiet("What a search of each may turn up, and the chance of each, for this monster"))
	}
	for _, p := range v.Places {
		row := newPlaceRow(p)
		b.places = append(b.places, row)
		content.Add(row.object)
	}
	b.object = content
	return b
}

// Screen is the Route panel.
type Screen struct {
	seen status.Seen[domain.Profile] // what apply last drew, so that the same is not drawn again
	svc  profile.ProfileService

	status  *widget.Label
	list    *fyne.Container // the panel of each map
	content *fyne.Container
	boxes   []*box
	drawn   string // what the panels were built from, so that they are built again only when it changes
}

// NewScreen is the Route panel of the maps given. It shows the map of the trip that is on offer to the monster this
// week, which is the one whose week has come and that the monster is fit for.
func NewScreen(svc profile.ProfileService) *Screen {
	s := &Screen{
		svc:    svc,
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
	p, err := s.svc.Current(ctx)
	fyne.Do(func() {
		if s.seen.Changed(p, err) {
			s.apply(p.Maps, p, err)
		}
	})
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
	size := tableSizeOf(maps) // of all the maps, so that the one shown is as big as it is when they all are
	fmt.Fprint(&built, size.w, "x", size.h, ";")
	for _, m := range maps {
		if !m.Open(p.Player.Ranch, p.Monster) {
			continue
		}
		v, err := NewView(p, m)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", m.Name, err))
			continue
		}
		views = append(views, v)
		fmt.Fprint(&built, v.Name, "|", v.Summary, "|", v.Places, "|", v.Route.Path, "|", v.Standing, ";") // Standing says where the monster is
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

	s.clear()
	s.drawn = drawn
	for _, v := range views {
		b := newBox(v, size)
		s.boxes = append(s.boxes, b)
		s.list.Add(b.object)
	}
	s.list.Refresh()
}

// clear takes every panel away.
func (s *Screen) clear() {
	s.list.RemoveAll()
	s.boxes, s.drawn = nil, ""
}
