package moveui

import (
	"context"
	"fmt"
	"image/color"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"app/internal/core/move"
	"app/internal/core/move/domain"
	"app/internal/delivery/ui/profileui"
	"app/internal/delivery/ui/status"
)

// The game's icons are 48x32 pixels and its marks 16x16; the grid draws them one and a half times that, which
// leaves room under it for what a move needs, and the row of active moves draws them as they are.
const (
	zoom       = 1.5
	iconW      = 48 * zoom
	iconH      = 32 * zoom
	markW      = 16 * zoom
	activeW    = 48
	activeH    = 32
	frameWidth = 3
	ribbonH    = 14
)

// The colours that say where a move stands: learned, ready to learn, and not yet.
var (
	metColor    color.Color = color.NRGBA{R: 50, G: 205, B: 50, A: 255}
	readyColor  color.Color = color.NRGBA{R: 255, G: 159, B: 28, A: 255}
	lockedColor color.Color = color.NRGBA{R: 150, G: 150, B: 150, A: 255}
	unmetColor  color.Color = color.NRGBA{R: 255, G: 69, A: 255}
	dimColor    color.Color = color.NRGBA{A: 150} // laid over a move that is not yet, to push it back
)

// IconFunc returns the PNG of an icon by key.
type IconFunc func(key string) ([]byte, error)

// iconSet turns icon keys into Fyne resources, once each. A missing icon shows as an empty cell.
type iconSet struct {
	load  IconFunc
	cache map[string]fyne.Resource
}

func (is *iconSet) resource(key string) fyne.Resource {
	if r, ok := is.cache[key]; ok {
		return r
	}
	b, err := is.load(key)
	if err != nil {
		if key == NoMove {
			return nil
		}
		return is.resource(NoMove)
	}
	r := fyne.NewStaticResource(key+".png", b)
	is.cache[key] = r
	return r
}

// pixelImage draws a resource without smoothing: these are pixel icons.
func pixelImage(r fyne.Resource, w, h float32) *canvas.Image {
	img := canvas.NewImageFromResource(r)
	img.ScaleMode = canvas.ImageScalePixels
	img.SetMinSize(fyne.NewSize(w, h))
	img.Resize(fyne.NewSize(w, h))
	return img
}

func setVisible(o fyne.CanvasObject, on bool) {
	if on == o.Visible() {
		return
	}
	if on {
		o.Show()
	} else {
		o.Hide()
	}
}

// picture is an image that remembers which icon it shows.
type picture struct {
	img *canvas.Image
	key string
}

func newPicture(is *iconSet, w, h float32) *picture {
	return &picture{img: pixelImage(is.resource(NoMove), w, h), key: NoMove}
}

func (p *picture) show(key string, is *iconSet) {
	if key == p.key {
		return
	}
	p.img.Resource, p.key = is.resource(key), key
	p.img.Refresh()
}

// cell is one place of the grid: the icon, and what says where the move stands: a green frame when it is learned,
// an orange frame and a ribbon when it is ready to learn, a dimming when it is not yet; with the selection, and
// the game's own marks for learned and can-learn.
type cell struct {
	widget.BaseWidget
	iconKey   string
	icon      *canvas.Image
	dim       *canvas.Rectangle
	frame     *canvas.Rectangle
	ribbon    *fyne.Container
	highlight *canvas.Image
	check     *canvas.Image
	canGet    *canvas.Image
	body      *fyne.Container
	onTap     func()
}

func newCell(is *iconSet, onTap func()) *cell {
	c := &cell{
		iconKey:   NoMove,
		icon:      pixelImage(is.resource(NoMove), iconW, iconH),
		dim:       canvas.NewRectangle(dimColor),
		frame:     canvas.NewRectangle(color.Transparent),
		highlight: pixelImage(is.resource(Highlight), iconW, iconH),
		check:     pixelImage(is.resource(Checked), markW, markW),
		canGet:    pixelImage(is.resource(CanGet), markW, markW),
		onTap:     onTap,
	}
	c.dim.Resize(fyne.NewSize(iconW, iconH))
	c.frame.StrokeWidth = frameWidth
	c.frame.Resize(fyne.NewSize(iconW, iconH))

	ribbonBg := canvas.NewRectangle(readyColor)
	ribbonBg.Move(fyne.NewPos(0, iconH-ribbonH))
	ribbonBg.Resize(fyne.NewSize(iconW, ribbonH))
	ribbonTxt := canvas.NewText("READY", color.Black)
	ribbonTxt.TextSize, ribbonTxt.TextStyle = 10, fyne.TextStyle{Bold: true}
	want := ribbonTxt.MinSize()
	ribbonTxt.Move(fyne.NewPos((iconW-want.Width)/2, iconH-ribbonH+(ribbonH-want.Height)/2))
	c.ribbon = container.NewWithoutLayout(ribbonBg, ribbonTxt)
	c.ribbon.Resize(fyne.NewSize(iconW, iconH))

	c.check.Move(fyne.NewPos(30*zoom, 2*zoom)) // where MR2AV puts the mark
	c.canGet.Move(fyne.NewPos(30*zoom, 2*zoom))
	for _, o := range []fyne.CanvasObject{c.dim, c.frame, c.ribbon, c.highlight, c.check, c.canGet} {
		o.Hide()
	}
	c.body = container.NewWithoutLayout(c.icon, c.dim, c.frame, c.ribbon, c.highlight, c.check, c.canGet)
	c.ExtendBaseWidget(c)
	return c
}

func (c *cell) CreateRenderer() fyne.WidgetRenderer { return widget.NewSimpleRenderer(c.body) }
func (c *cell) MinSize() fyne.Size                  { return fyne.NewSize(iconW, iconH) }
func (c *cell) Tapped(*fyne.PointEvent)             { c.onTap() }

func (c *cell) show(v CellView, is *iconSet) {
	if v.Icon != c.iconKey {
		c.icon.Resource, c.iconKey = is.resource(v.Icon), v.Icon
		c.icon.Refresh()
	}
	setVisible(c.check, v.Learned)
	setVisible(c.canGet, v.CanLearn)
	setVisible(c.dim, v.Locked)
	setVisible(c.ribbon, v.CanLearn)
	setVisible(c.frame, v.Learned || v.CanLearn)
	c.frame.StrokeColor = readyColor
	if v.Learned {
		c.frame.StrokeColor = metColor
	}
	c.frame.Refresh()
}

// legend counts the moves in each state, beside the colour that marks it.
type legend struct {
	learned, ready, locked *widget.Label
	object                 fyne.CanvasObject
}

func newLegend() *legend {
	l := &legend{learned: widget.NewLabel(""), ready: widget.NewLabel(""), locked: widget.NewLabel("")}
	item := func(c color.Color, label *widget.Label) fyne.CanvasObject {
		swatch := canvas.NewRectangle(c)
		swatch.SetMinSize(fyne.NewSize(12, 12))
		return container.NewHBox(container.NewCenter(swatch), label)
	}
	l.object = container.NewHBox(item(metColor, l.learned), item(readyColor, l.ready), item(lockedColor, l.locked))
	l.show(GridView{})
	return l
}

func (l *legend) show(v GridView) {
	l.learned.SetText(fmt.Sprintf("Learned %d", v.Learned))
	l.ready.SetText(fmt.Sprintf("Ready to learn %d", v.Ready))
	l.locked.SetText(fmt.Sprintf("Not yet %d", v.Locked))
}

// badge is a few words on a coloured ground, hidden when there are none.
type badge struct {
	bg     *canvas.Rectangle
	txt    *canvas.Text
	object *fyne.Container
}

func newBadge() *badge {
	b := &badge{bg: canvas.NewRectangle(color.Transparent), txt: canvas.NewText("", color.Black)}
	b.bg.CornerRadius = 4
	b.txt.TextStyle = fyne.TextStyle{Bold: true}
	b.object = container.NewStack(b.bg, container.NewPadded(b.txt))
	b.object.Hide()
	return b
}

func (b *badge) set(text string, fill color.Color) {
	b.txt.Text, b.bg.FillColor = text, fill
	b.txt.Refresh()
	b.bg.Refresh()
	setVisible(b.object, text != "")
}

// need is one line of the list of what a move asks of the monster: a mark for met or not, and the words.
type need struct {
	mark   *canvas.Text
	label  *widget.Label
	object fyne.CanvasObject
}

func newNeed(r RequirementView) need {
	n := need{mark: canvas.NewText("✗", unmetColor)}
	if r.Met {
		n.mark.Text, n.mark.Color = "✓", metColor
	}
	text := r.Text
	if r.Have != "" {
		text += " (" + r.Have + ")"
	}
	n.label = widget.NewLabel(text)
	n.label.Wrapping = fyne.TextWrapWord
	n.object = container.NewBorder(nil, nil, n.mark, nil, n.label)
	return n
}

// detail is the panel of numbers and words about the selected move.
type detail struct {
	name                                                 *canvas.Text
	status                                               *badge
	guts, damage, hit, withering, sharpness, extra, uses *widget.Label
	description                                          *widget.Label
	needsBox                                             *fyne.Container // the heading and the list; hidden when the move needs nothing
	requirements                                         *fyne.Container
	needs                                                []need
	content                                              fyne.CanvasObject
}

func newDetail() *detail {
	d := &detail{
		name:        canvas.NewText("", theme.Color(theme.ColorNameForeground)),
		status:      newBadge(),
		guts:        widget.NewLabel(""),
		damage:      widget.NewLabel(""),
		hit:         widget.NewLabel(""),
		withering:   widget.NewLabel(""),
		sharpness:   widget.NewLabel(""),
		extra:       widget.NewLabel(""),
		uses:        widget.NewLabel(""),
		description: widget.NewLabel(""),
	}
	d.name.TextSize = 22
	d.name.TextStyle = fyne.TextStyle{Bold: true}
	d.description.Wrapping = fyne.TextWrapWord
	pair := func(name string, value *widget.Label) fyne.CanvasObject {
		return container.NewHBox(profileui.Quiet(name), value)
	}
	numbers := container.NewGridWithColumns(3,
		pair("Guts", d.guts), pair("Damage", d.damage), pair("Hit", d.hit),
		pair("Withering", d.withering), pair("Sharpness", d.sharpness), pair("Uses", d.uses),
	)
	d.requirements = container.NewVBox()
	d.needsBox = container.NewVBox(
		widget.NewLabelWithStyle("To learn it", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		d.requirements,
	)
	d.needsBox.Hide()
	d.content = container.NewVBox(
		container.NewBorder(nil, nil, nil, container.NewCenter(d.status.object), d.name),
		numbers,
		d.extra,
		d.needsBox,
		d.description,
	)
	return d
}

func nameColor(k domain.Kind) color.Color {
	switch k {
	case domain.KindPower:
		return color.NRGBA{R: 0xDA, G: 0xA5, B: 0x20, A: 0xFF} // goldenrod, as in MR2AV
	case domain.KindIntelligence:
		return color.NRGBA{R: 0x3C, G: 0xB3, B: 0x71, A: 0xFF}
	}
	return theme.Color(theme.ColorNameForeground)
}

func (d *detail) show(v DetailView) {
	d.name.Text, d.name.Color = v.Name, nameColor(v.Kind)
	d.name.Refresh()
	switch {
	case v.Learned:
		d.status.set(v.Status, metColor)
	case v.CanLearn:
		d.status.set(v.Status, readyColor)
	default:
		d.status.set(v.Status, lockedColor)
	}
	d.guts.SetText(v.Guts)
	d.damage.SetText(v.Damage)
	d.hit.SetText(v.Hit)
	d.withering.SetText(v.Withering)
	d.sharpness.SetText(v.Sharpness)
	d.uses.SetText(v.Uses)
	d.extra.SetText(v.Extra)
	if v.Extra == "" {
		d.extra.Hide()
	} else {
		d.extra.Show()
	}
	d.description.SetText(v.Description)

	d.needs = nil
	d.requirements.RemoveAll()
	for _, r := range v.Requirements {
		n := newNeed(r)
		d.needs = append(d.needs, n)
		d.requirements.Add(n.object)
	}
	setVisible(d.needsBox, len(d.needs) > 0)
}

// Screen is the Move Viewer: the grid of moves, the active move of each column, and the details of
// the selected move.
type Screen struct {
	svc   move.MoveService
	icons *iconSet

	cells [domain.Columns][domain.Rows]*cell

	cellGrid  fyne.CanvasObject // the columns of cells, side by side
	activeRow fyne.CanvasObject // the label and the active moves
	active    [domain.Columns]*picture
	genus     *widget.Label
	legend    *legend
	detail    *detail

	body    fyne.CanvasObject
	content fyne.CanvasObject
	grid    domain.Grid // the grid last shown, for the details of the selected cell
	hasSel  bool
	selCol  int
	selRow  int
}

func NewScreen(svc move.MoveService, icons IconFunc) *Screen {
	s := &Screen{
		svc:    svc,
		icons:  &iconSet{load: icons, cache: map[string]fyne.Resource{}},
		genus:  widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		legend: newLegend(),
		detail: newDetail(),
	}

	columns := make([]fyne.CanvasObject, domain.Columns)
	for c := range s.cells {
		cells := make([]fyne.CanvasObject, domain.Rows)
		for r := range s.cells[c] {
			s.cells[c][r] = newCell(s.icons, func() { s.selectCell(c, r) })
			cells[r] = s.cells[c][r]
		}
		columns[c] = container.New(layout.NewCustomPaddedVBoxLayout(0), cells...)
	}
	slices.Reverse(columns) // the last column is shown first, and the active moves under the columns follow
	s.cellGrid = container.New(layout.NewCustomPaddedHBoxLayout(0), columns...)

	pictures := make([]fyne.CanvasObject, domain.Columns)
	for c := range s.active {
		s.active[c] = newPicture(s.icons, activeW, activeH)
		pictures[c] = s.active[c].img
	}
	slices.Reverse(pictures)
	actives := append([]fyne.CanvasObject{profileui.Quiet("Active moves")}, pictures...)
	s.activeRow = container.NewHBox(actives...)
	// What shows the moves is in the middle of the panel; the text of the selected move under it starts at the left.
	s.body = container.NewVBox(
		container.NewCenter(s.genus),
		container.NewCenter(s.legend.object),
		container.NewCenter(s.cellGrid),
		container.NewCenter(s.activeRow),
		widget.NewSeparator(),
		s.detail.content,
	)
	s.content = container.NewVScroll(container.NewPadded(s.body))

	s.detail.show(NewDetailView(domain.Cell{}))
	return s
}

// Content is what to put in a window.
func (s *Screen) Content() fyne.CanvasObject { return s.content }

// Load asks the service for the current grid and shows it. It is safe to call from any goroutine.
func (s *Screen) Load(ctx context.Context) {
	g, err := s.svc.CurrentGrid(ctx)
	fyne.Do(func() { s.apply(g, err) })
}

// apply shows a grid, or why there is none. It must run on the UI thread.
func (s *Screen) apply(g domain.Grid, err error) {
	if err != nil {
		g = domain.Grid{}
		s.genus.SetText(status.Message(err))
	} else {
		s.genus.SetText(g.Genus.Name)
	}
	s.grid = g
	v := NewGridView(g)
	for c := range s.cells {
		for r := range s.cells[c] {
			s.cells[c][r].show(v.Cells[c][r], s.icons)
		}
		s.active[c].show(v.Active[c], s.icons)
	}
	s.legend.show(v)
	s.showSelection()
}

func (s *Screen) selectCell(col, row int) {
	s.hasSel, s.selCol, s.selRow = true, col, row
	s.showSelection()
}

// showSelection highlights the selected cell and shows its details from the grid last shown.
func (s *Screen) showSelection() {
	for c := range s.cells {
		for r := range s.cells[c] {
			setVisible(s.cells[c][r].highlight, s.hasSel && c == s.selCol && r == s.selRow)
		}
	}
	if s.hasSel {
		s.detail.show(NewDetailView(s.grid.Cells[s.selCol][s.selRow]))
	}
}
