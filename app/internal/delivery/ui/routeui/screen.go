package routeui

import (
	"cmp"
	"context"
	"fmt"
	"image/color"
	"slices"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	monster "app/internal/core/monster/domain"
	"app/internal/core/profile"
	"app/internal/core/profile/domain"
	"app/internal/delivery/ui/profileui"
	"app/internal/delivery/ui/status"
)

const (
	cellSize = 17
	cellGap  = 1
)

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
	left := max((size.Width-t.MinSize(objs).Width)/2, 0)
	for i, o := range objs {
		o.Move(fyne.NewPos(left+float32(i%t.cols*(cellSize+cellGap)), float32(i/t.cols*(cellSize+cellGap))))
		o.Resize(fyne.NewSize(cellSize, cellSize))
	}
}

type tableSize struct{ w, h int }

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

var (
	floorFill = color.NRGBA{R: 96, G: 102, B: 112, A: 255}
	wallFill  = color.NRGBA{R: 24, G: 27, B: 33, A: 255}
	rockFill  = color.NRGBA{R: 139, G: 90, B: 43, A: 255}
	startFill = color.NRGBA{R: 65, G: 105, B: 225, A: 255}
	stopFill  = color.NRGBA{R: 50, G: 180, B: 50, A: 255}
	spotFill  = color.NRGBA{R: 70, G: 110, B: 85, A: 255}
	doneFill  = color.NRGBA{R: 120, G: 120, B: 130, A: 255}
	dimFill   = color.NRGBA{R: 60, G: 70, B: 64, A: 255}
	pickFill  = color.NRGBA{R: 0, G: 190, B: 255, A: 255}

	detailFill = color.NRGBA{R: 0, G: 190, B: 255, A: 40}

	markerFill = color.NRGBA{R: 255, G: 105, B: 180, A: 255}

	gateFill     = color.NRGBA{R: 160, G: 200, B: 80, A: 255}
	shutFill     = color.NRGBA{R: 80, G: 92, B: 62, A: 255}
	teleportFill = color.NRGBA{R: 160, G: 110, B: 210, A: 255}
)

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

func newTable(v View, size tableSize, onTap func(string)) (table fyne.CanvasObject, cells, padding []*cell, taps map[string]fyne.Tappable) {
	m := v.Map
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
	labels := make(map[domain.Point]string, len(m.Portals))
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
		case isSpot && v.Order[name] > 0:
			c.bg.FillColor, c.text.Text = stopFill, strconv.Itoa(v.Order[name])
		case isSpot && v.Order != nil && searched(v, name):
			c.bg.FillColor, c.text.Text = doneFill, name
		case isSpot && v.Order != nil:
			c.bg.FillColor, c.text.Text = dimFill, name
		case isSpot && v.Rank[name] > 0:
			c.bg.FillColor, c.text.Text = pickFill, strconv.Itoa(v.Rank[name])
		case isSpot && v.Rank != nil:
			c.bg.FillColor, c.text.Text = dimFill, name
		case isSpot && searched(v, name):
			c.bg.FillColor, c.text.Text = doneFill, name
		case isSpot && taken[p]:
			c.bg.FillColor, c.text.Text = stopFill, name
		case isSpot:
			c.bg.FillColor, c.text.Text = spotFill, name
		case tile == domain.Rock:
			c.bg.FillColor = rockFill
		case labels[p] != "":
			c.bg.FillColor, c.text.Text = teleportFill, labels[p]
		default:
			c.bg.FillColor = floorFill
		}
		if name, ok := spots[p]; ok && v.Shown != "" && name == v.Shown {
			c.bg.StrokeColor, c.bg.StrokeWidth = color.White, 2
		}
		if v.Marker != nil && p == *v.Marker {
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
				if name, ok := spots[p]; ok && onTap != nil && (v.Rank[name] > 0 || v.Order != nil) {
					t := newTapper(canvas.NewRectangle(color.Transparent), func() { onTap(name) })
					c.object.(*fyne.Container).Add(t)
					if taps == nil {
						taps = map[string]fyne.Tappable{}
					}
					taps[name] = t
				}
			} else {
				c = newCell()
				c.bg.FillColor = wallFill
				padding = append(padding, c)
			}
			objects = append(objects, c.object)
		}
	}
	return container.New(tableLayout{cols: size.w}, objects...), cells, padding, taps
}

type tapper struct {
	widget.BaseWidget
	content fyne.CanvasObject
	onTap   func()
}

func newTapper(content fyne.CanvasObject, onTap func()) *tapper {
	t := &tapper{content: content, onTap: onTap}
	t.ExtendBaseWidget(t)
	return t
}

func (t *tapper) CreateRenderer() fyne.WidgetRenderer { return widget.NewSimpleRenderer(t.content) }

func (t *tapper) Tapped(*fyne.PointEvent) { t.onTap() }

const keyColumns = 2

func newKey(v View, mode int) (fyne.CanvasObject, []keyItem) {
	m := v.Map
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
	if mode == modeFind {
		add(pickFill, "1", "A place with a picked item, best first")
		for _, s := range m.Spots {
			if v.Rank[s.Name] == 0 {
				add(dimFill, s.Name, "A place without one")
				break
			}
		}
	}
	if mode == modeRoute && len(v.Route.Spots) > 0 {
		add(stopFill, "1", "A stop, in the order to go")
	}
	for _, s := range m.Spots {
		if mode == modeRoute && searched(v, s.Name) {
			add(doneFill, s.Name, "A place searched")
			break
		}
	}
	for _, s := range m.Spots {
		if mode == modeRoute && !taken[s.At] && !searched(v, s.Name) {
			add(dimFill, s.Name, "A place off the route")
			break
		}
	}

	var rock, teleport bool
	for _, tile := range m.Tiles {
		rock = rock || tile == domain.Rock
		teleport = teleport || tile == domain.Teleport
	}
	if rock {
		add(rockFill, "", "Rock")
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

func searched(v View, name string) bool {
	return len(v.Got[name]) > 0
}

type keyItem struct {
	swatch *cell
	label  *widget.Label
}

type box struct {
	name      string
	v         View
	size      tableSize
	title     *widget.Label
	object    fyne.CanvasObject
	summary   *widget.Label
	key       []keyItem
	teleports *widget.Label
	holder    *fyne.Container
	table     fyne.CanvasObject
	standing  []*profileui.ToneBox
	cells     []*cell
	padding   []*cell
	controls  *fyne.Container
	routePart *fyne.Container
	findPart  *fyne.Container
	keyToggle *widget.Button
	keyBox    *fyne.Container

	taps         map[string]fyne.Tappable
	rows         []*widget.Button
	detailTitle  *widget.Label
	detailItems  *widget.Label
	notYetToggle *widget.Button
	notYetBox    *fyne.Container

	routeTaps  map[string]fyne.Tappable
	routeTitle *widget.Label
	routeNote  *widget.Label
	drops      []*profileui.ToneBox
	routeRows  []*widget.Button
	offRows    []*widget.Button
	offToggle  *widget.Button
	offBox     *fyne.Container
}

type findState struct {
	keyOpen, notYetOpen bool
	selected            string
	onSelect            func(string)
	onNotYet            func()
	offOpen             bool
	stop                string
	onStop              func(string)
	onOff               func()
}

func newBox(v View, size tableSize) *box {
	b := &box{
		name:    v.Name,
		v:       v,
		size:    size,
		title:   widget.NewLabelWithStyle(v.Title(), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		summary: widget.NewLabel(v.Summary),
		holder:  container.NewStack(),
	}
	b.teleports = profileui.Quiet(strings.Join(v.Teleports, "\n"))
	b.teleports.Wrapping = fyne.TextWrapWord
	b.teleports.Hidden = len(v.Teleports) == 0
	b.summary.Wrapping = fyne.TextWrapWord
	b.keyToggle = widget.NewButton("", nil)
	b.keyToggle.Alignment = widget.ButtonAlignLeading
	b.keyToggle.Importance = widget.LowImportance
	b.keyBox = container.NewVBox()
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
	for _, o := range []fyne.CanvasObject{b.summary, b.holder, b.keyToggle, b.keyBox} {
		content.Add(o)
	}
	b.routePart = container.NewVBox()
	b.findPart = container.NewVBox()
	b.controls = container.NewVBox()
	content.Add(b.controls)
	content.Add(b.routePart)
	content.Add(b.findPart)
	b.object = content
	b.draw(modeRoute, FindView{}, findState{})
	return b
}

func (b *box) draw(mode int, f FindView, st findState) {
	lit := b.v
	rank := map[string]int{}
	onTap := st.onSelect
	if mode == modeRoute {
		lit.Order = map[string]int{}
		for i, sp := range b.v.Route.Spots {
			lit.Order[sp.Name] = i + 1
		}
		onTap = st.onStop
		if p, ok := b.routeShown(st); ok {
			lit.Shown = p.Name
		}
	}
	if mode == modeFind {
		lit.Route = domain.Route{}
		for i, p := range f.Places {
			rank[p.Name] = i + 1
		}
		lit.Rank = rank
		if len(f.Places) > 0 {
			lit.Shown = f.Places[0].Name
			for _, p := range f.Places {
				if p.Name == st.selected {
					lit.Shown = p.Name
				}
			}
		}
	}
	b.table, b.cells, b.padding, b.taps = newTable(lit, b.size, onTap)
	b.routeTaps = b.taps
	b.drawRoute(st)
	b.holder.Objects = []fyne.CanvasObject{b.table}
	b.holder.Refresh()

	key, items := newKey(lit, mode)
	b.key = items
	b.keyBox.Objects = []fyne.CanvasObject{key, b.teleports}
	b.keyBox.Hidden = !st.keyOpen
	b.keyToggle.SetText("▸ Map key")
	if st.keyOpen {
		b.keyToggle.SetText("▾ Map key")
	}

	b.findPart.RemoveAll()
	b.rows, b.detailTitle, b.detailItems = nil, nil, nil
	if len(f.Places) == 0 && len(f.NotYet) == 0 {
		b.findPart.Add(profileui.Quiet("Pick items above to see where to go"))
	}
	if len(f.Places) > 0 {
		shown := f.Places[0]
		for _, p := range f.Places {
			if p.Name == st.selected {
				shown = p
			}
		}
		var items []string
		for _, it := range shown.Items {
			items = append(items, it.Item+" "+it.Chance)
		}
		b.detailTitle = widget.NewLabelWithStyle(fmt.Sprintf("%d. %s · %s", rank[shown.Name], placeName(shown), shown.Energy), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		b.detailItems = widget.NewLabel(strings.Join(items, " · "))
		b.detailItems.Wrapping = fyne.TextWrapWord
		b.findPart.Add(container.NewStack(canvas.NewRectangle(detailFill), container.NewVBox(b.detailTitle, b.detailItems)))
		for _, p := range f.Places {
			row := widget.NewButton(fmt.Sprintf("%d. %s · %s", rank[p.Name], placeName(p), p.Energy), nil)
			row.Alignment = widget.ButtonAlignLeading
			row.Importance = widget.LowImportance
			if p.Name == shown.Name {
				row.Importance = widget.MediumImportance
			}
			name := p.Name
			row.OnTapped = func() { st.onSelect(name) }
			b.rows = append(b.rows, row)
			b.findPart.Add(row)
		}
	}
	b.notYetToggle, b.notYetBox = nil, nil
	if n := len(f.NotYet); n > 0 {
		verb := "needs"
		if n > 1 {
			verb = "need"
		}
		mark := "▸ "
		if st.notYetOpen {
			mark = "▾ "
		}
		b.notYetToggle = widget.NewButton(fmt.Sprintf("%s%d more %s a higher stat", mark, n, verb), st.onNotYet)
		b.notYetToggle.Alignment = widget.ButtonAlignLeading
		b.notYetToggle.Importance = widget.LowImportance
		b.notYetBox = container.NewVBox()
		for _, line := range f.NotYet {
			l := profileui.Quiet(line)
			l.Wrapping = fyne.TextWrapWord
			b.notYetBox.Add(l)
		}
		b.notYetBox.Hidden = !st.notYetOpen
		b.findPart.Add(b.notYetToggle)
		b.findPart.Add(b.notYetBox)
	}
	b.routePart.Hidden = mode != modeRoute
	b.findPart.Hidden = mode != modeFind
	b.object.Refresh()
}

func (b *box) places() (stops, done, off []PlaceView) {
	for _, p := range b.v.Places {
		switch {
		case p.Stop > 0:
			stops = append(stops, p)
		case len(p.Got) > 0:
			done = append(done, p)
		default:
			off = append(off, p)
		}
	}
	return stops, done, off
}

func (b *box) routeShown(st findState) (PlaceView, bool) {
	stops, done, off := b.places()
	listed := append(slices.Clone(stops), done...)
	for _, p := range append(slices.Clone(listed), off...) {
		if p.Name == st.stop {
			return p, true
		}
	}
	if len(listed) > 0 {
		return listed[0], true
	}
	return PlaceView{}, false
}

func (b *box) drawRoute(st findState) {
	b.routePart.RemoveAll()
	b.routeTitle, b.routeNote, b.drops, b.routeRows, b.offRows = nil, nil, nil, nil, nil
	stops, done, off := b.places()
	shown, any := b.routeShown(st)
	if any {
		b.drawStop(shown)
	}
	for _, p := range append(slices.Clone(stops), done...) {
		text := fmt.Sprintf("%d. %s · %s", p.Stop, spotName(p), b.energy(p.Name))
		switch {
		case p.Stop == 0:
			text = fmt.Sprintf("✓ %s · got %s", spotName(p), gotText(p.Got))
		case len(p.Got) > 0:
			text += fmt.Sprintf(" · got %d", len(p.Got))
		}
		b.routeRows = append(b.routeRows, b.placeButton(text, p.Name, shown.Name, st))
		b.routePart.Add(b.routeRows[len(b.routeRows)-1])
	}
	b.offToggle, b.offBox = nil, nil
	if n := len(off); n > 0 {
		noun := "place"
		if n > 1 {
			noun = "places"
		}
		mark := "▸ "
		if st.offOpen {
			mark = "▾ "
		}
		b.offToggle = widget.NewButton(fmt.Sprintf("%s%d %s off the route", mark, n, noun), st.onOff)
		b.offToggle.Alignment = widget.ButtonAlignLeading
		b.offToggle.Importance = widget.LowImportance
		b.offBox = container.NewVBox()
		for _, p := range off {
			row := b.placeButton(spotName(p)+" · "+b.offReason(p), p.Name, shown.Name, st)
			b.offRows = append(b.offRows, row)
			b.offBox.Add(row)
		}
		b.offBox.Hidden = !st.offOpen
		b.routePart.Add(b.offToggle)
		b.routePart.Add(b.offBox)
	}
}

func (b *box) placeButton(text, name, shown string, st findState) *widget.Button {
	row := widget.NewButton(text, func() { st.onStop(name) })
	row.Alignment = widget.ButtonAlignLeading
	row.Importance = widget.LowImportance
	if name == shown {
		row.Importance = widget.MediumImportance
	}
	return row
}

func (b *box) offReason(p PlaceView) string {
	switch {
	case strings.HasPrefix(p.Need, "needs "):
		return p.Need
	case b.energy(p.Name) == "out of reach":
		return "out of reach"
	}
	return "left out: the energy is worth more elsewhere"
}

func (b *box) drawStop(p PlaceView) {
	title := fmt.Sprintf("%d. %s · %s", p.Stop, spotName(p), b.energy(p.Name))
	var notes []string
	if len(p.Got) > 0 {
		notes = append(notes, "Got "+gotText(p.Got))
	}
	switch {
	case p.Stop == 0 && len(p.Got) > 0:
		title = fmt.Sprintf("✓ %s · searched", spotName(p))
	case p.Stop == 0:
		title = fmt.Sprintf("%s · off the route", spotName(p))
		notes = append(notes, b.offReason(p))
	default:
		notes = append(notes, slices.DeleteFunc([]string{p.Worth, p.Need}, func(s string) bool { return s == "" })...)
	}
	b.routeTitle = widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	chips := container.NewGridWithColumns(findColumns)
	for _, d := range p.Loot {
		box := profileui.Box{Text: d.Item + " " + d.Chance}
		if n := count(p.Got, d.Item); n > 0 {
			box = profileui.Box{Text: "✓ " + box.Text + times(n), Tone: profileui.ToneLimeGreen}
		}
		chip := profileui.NewToneBox()
		chip.Set(box)
		b.drops = append(b.drops, chip)
		chips.Add(chip)
	}
	b.routeNote = profileui.Quiet(strings.Join(notes, " · "))
	b.routeNote.Wrapping = fyne.TextWrapWord
	if len(p.Loot) == 0 && len(p.Got) == 0 {
		chips.Add(profileui.Quiet("Nothing turns up here for this monster yet"))
	}
	b.routePart.Add(container.NewStack(canvas.NewRectangle(detailFill), container.NewVBox(b.routeTitle, chips, b.routeNote)))
}

func count(got []string, item string) int {
	n := 0
	for _, g := range got {
		if g == item {
			n++
		}
	}
	return n
}

func times(n int) string {
	if n < 2 {
		return ""
	}
	return fmt.Sprintf(" ×%d", n)
}

func gotText(got []string) string {
	var parts []string
	for i, g := range got {
		if !slices.Contains(got[:i], g) {
			parts = append(parts, g+times(count(got, g)))
		}
	}
	return strings.Join(parts, ", ")
}

func (b *box) energy(name string) string {
	way, ok := b.v.Ways[name]
	return walkText(way, ok)
}

func walkText(way domain.Way, ok bool) string {
	switch {
	case !ok:
		return "out of reach"
	case way.Steps == 0:
		return "right here"
	}
	var text strings.Builder
	text.WriteString(fmt.Sprintf("%d energy to walk", way.Steps))
	for _, n := range []struct {
		count int
		noun  string
	}{{way.Rocks, "rock"}, {way.Gates, "gate"}} {
		switch {
		case n.count == 1:
			text.WriteString(" · 1 " + n.noun)
		case n.count > 1:
			text.WriteString(fmt.Sprintf(" · %d %ss", n.count, n.noun))
		}
	}
	return text.String()
}

func spotName(p PlaceView) string {
	if p.Building == "" {
		return p.Name
	}
	return p.Name + " · " + p.Building
}

func placeName(p FindPlace) string {
	if p.Building == "" {
		return p.Name
	}
	return p.Name + " · " + p.Building
}

const (
	modeRoute = iota
	modeFind
)

var modeNames = []string{"Route", "Find items"}

type Screen struct {
	seen status.Seen[domain.Profile]
	svc  profile.ProfileService

	status  *widget.Label
	list    *fyne.Container
	content *fyne.Container
	boxes   []*box
	drawn   string
	haul    domain.Haul

	mode      int
	bar       *fyne.Container
	modes     []*widget.Button
	picker    *fyne.Container
	filter    *widget.Entry
	showItems *widget.Button
	chips     *fyne.Container
	search    *fyne.Container
	listOpen  bool
	state     findState
	checks    *fyne.Container
	picked    map[string]bool
	by        domain.Searcher
}

func NewScreen(svc profile.ProfileService) *Screen {
	s := &Screen{
		svc:    svc,
		status: widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		list:   container.NewVBox(),
		picked: map[string]bool{},
	}
	s.state.onSelect = func(name string) { s.state.selected = name; s.drawBoxes() }
	s.state.onNotYet = func() { s.state.notYetOpen = !s.state.notYetOpen; s.drawBoxes() }
	s.state.onStop = func(name string) { s.state.stop = name; s.drawBoxes() }
	s.state.onOff = func() { s.state.offOpen = !s.state.offOpen; s.drawBoxes() }
	s.status.Hide()
	s.filter = widget.NewEntry()
	s.filter.SetPlaceHolder("Search items")
	s.filter.OnChanged = func(string) { s.drawChecks() }
	s.showItems = widget.NewButton("", func() { s.listOpen = !s.listOpen; s.drawChecks() })
	s.showItems.Alignment = widget.ButtonAlignLeading
	s.showItems.Importance = widget.LowImportance
	s.chips = container.NewGridWithColumns(findColumns)
	s.checks = container.NewGridWithColumns(findColumns)
	all := widget.NewButton("All", func() { s.pickAll(true) })
	none := widget.NewButton("None", func() { s.pickAll(false) })
	s.search = container.NewBorder(nil, nil, nil, container.NewHBox(all, none), s.filter)
	s.picker = container.NewVBox(s.chips, s.showItems, s.search, s.checks)
	s.bar = container.NewGridWithColumns(len(modeNames))
	for i, name := range modeNames {
		b := widget.NewButton(name, func() { s.setMode(i) })
		s.modes = append(s.modes, b)
		s.bar.Add(b)
	}
	s.content = container.NewVBox(s.status, s.list)
	s.content.Hide()
	s.setMode(modeRoute)
	return s
}

func (s *Screen) modeNames() []string {
	var out []string
	for _, b := range s.modes {
		out = append(out, b.Text)
	}
	return out
}

func (s *Screen) setMode(mode int) {
	s.mode = mode
	for i, b := range s.modes {
		b.Importance = widget.LowImportance
		if i == mode {
			b.Importance = widget.HighImportance
		}
		b.Refresh()
	}
	s.picker.Hidden = mode != modeFind
	s.drawChecks()
	s.drawBoxes()
	s.content.Refresh()
}

func (s *Screen) Content() fyne.CanvasObject { return s.content }

func (s *Screen) Load(ctx context.Context) {
	p, err := s.svc.Current(ctx)
	fyne.Do(func() {
		if s.seen.Changed(p, err) {
			s.apply(p.Maps, p, err)
		}
	})
}

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
	size := tableSizeOf(maps)
	mon := p.Monster
	by := domain.Searcher{Intelligence: mon.Stats[monster.Int], Skill: mon.Stats[monster.Skill], GoodNature: mon.Nature > 0}
	fmt.Fprint(&built, size.w, "x", size.h, ";", by, ";")
	for _, m := range maps {
		if !m.Open(p.Player.Ranch, p.Monster) {
			continue
		}
		v, err := NewView(p, m, s.got(m, p))
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", m.Name, err))
			continue
		}
		views = append(views, v)
		fmt.Fprint(&built, v.Name, "|", v.Summary, "|", v.Places, "|", v.Route.Path, "|", v.Standing, "|", v.Got, "|", v.Intelligence, "|", v.Ways, ";")
	}
	drawn := built.String()
	s.status.SetText(strings.Join(problems, "\n"))
	if len(problems) > 0 {
		s.status.Show()
	} else {
		s.status.Hide()
	}
	if len(views) == 0 && len(problems) == 0 {
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
	s.by = by
	for _, v := range views {
		b := newBox(v, size)
		b.keyToggle.OnTapped = func() { s.state.keyOpen = !s.state.keyOpen; s.drawBoxes() }
		s.boxes = append(s.boxes, b)
		s.list.Add(b.object)
	}
	if len(s.boxes) > 0 {
		s.boxes[0].controls.Objects = []fyne.CanvasObject{s.bar, s.picker}
	}
	s.drawChecks()
	s.drawBoxes()
	s.list.Refresh()
}

func (s *Screen) got(m domain.ExpeditionMap, p domain.Profile) map[string][]int {
	s.haul.Note(m, p.Player.Ranch)
	if s.haul.Map != m.Name {
		return nil
	}
	return s.haul.Got
}

func (s *Screen) clear() {
	s.list.RemoveAll()
	s.boxes, s.drawn = nil, ""
}
