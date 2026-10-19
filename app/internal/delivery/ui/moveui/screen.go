package moveui

import (
	"context"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"app/internal/core/move"
	"app/internal/core/move/domain"
	"app/internal/delivery/ui/status"
)

// The game's icons are 48x32 pixels and its marks 16x16; the screen draws everything at twice that.
const (
	zoom  = 2
	iconW = 48 * zoom
	iconH = 32 * zoom
	markW = 16 * zoom
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

func newPicture(is *iconSet) *picture {
	return &picture{img: pixelImage(is.resource(NoMove), iconW, iconH), key: NoMove}
}

func (p *picture) show(key string, is *iconSet) {
	if key == p.key {
		return
	}
	p.img.Resource, p.key = is.resource(key), key
	p.img.Refresh()
}

// cell is one place of the grid: the icon, the selection, and the learned and can-learn marks.
type cell struct {
	widget.BaseWidget
	iconKey   string
	icon      *canvas.Image
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
		highlight: pixelImage(is.resource(Highlight), iconW, iconH),
		check:     pixelImage(is.resource(Checked), markW, markW),
		canGet:    pixelImage(is.resource(CanGet), markW, markW),
		onTap:     onTap,
	}
	c.check.Move(fyne.NewPos(30*zoom, 2*zoom)) // where MR2AV puts the mark
	c.canGet.Move(fyne.NewPos(30*zoom, 2*zoom))
	c.highlight.Hide()
	c.check.Hide()
	c.canGet.Hide()
	c.body = container.NewWithoutLayout(c.icon, c.highlight, c.check, c.canGet)
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
}

// detail is the panel of numbers and words about the selected move.
type detail struct {
	name                                                 *canvas.Text
	guts, damage, hit, withering, sharpness, extra, uses *widget.Label
	learned, canLearn                                    *widget.Check
	description                                          *widget.Label
	needsBox                                             *fyne.Container // the heading and the list; hidden when the move needs nothing
	requirements                                         *fyne.Container
	needs                                                []need
	content                                              fyne.CanvasObject
}

var (
	metColor   color.Color = color.NRGBA{R: 50, G: 205, B: 50, A: 255}
	unmetColor color.Color = color.NRGBA{R: 255, G: 69, A: 255}
)

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

func newDetail() *detail {
	d := &detail{
		name:        canvas.NewText("", theme.Color(theme.ColorNameForeground)),
		guts:        widget.NewLabel(""),
		damage:      widget.NewLabel(""),
		hit:         widget.NewLabel(""),
		withering:   widget.NewLabel(""),
		sharpness:   widget.NewLabel(""),
		extra:       widget.NewLabel(""),
		uses:        widget.NewLabel(""),
		learned:     widget.NewCheck("Learned", nil),
		canLearn:    widget.NewCheck("Can learn", nil),
		description: widget.NewLabel(""),
	}
	d.name.TextSize = 22
	d.name.TextStyle = fyne.TextStyle{Bold: true}
	d.learned.Disable() // they show state; the game, not the viewer, learns moves
	d.canLearn.Disable()
	d.description.Wrapping = fyne.TextWrapWord
	numbers := container.New(layout.NewFormLayout(),
		widget.NewLabel("Guts"), d.guts,
		widget.NewLabel("Damage"), d.damage,
		widget.NewLabel("Hit"), d.hit,
		widget.NewLabel("Withering"), d.withering,
		widget.NewLabel("Sharpness"), d.sharpness,
		widget.NewLabel("Uses"), d.uses,
	)
	d.requirements = container.NewVBox()
	d.needsBox = container.NewVBox(
		widget.NewLabelWithStyle("To learn it", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		d.requirements,
	)
	d.needsBox.Hide()
	d.content = container.NewVBox(d.name, numbers, d.extra, container.NewHBox(d.learned, d.canLearn), d.needsBox, d.description)
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
	d.learned.SetChecked(v.Learned)
	d.canLearn.SetChecked(v.CanLearn)
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

	cells  [domain.Columns][domain.Rows]*cell
	active [domain.Columns]*picture
	genus  *widget.Label
	detail *detail

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
	grid := container.New(layout.NewCustomPaddedHBoxLayout(0), columns...)

	actives := make([]fyne.CanvasObject, domain.Columns)
	for c := range s.active {
		s.active[c] = newPicture(s.icons)
		actives[c] = s.active[c].img
	}
	s.body = container.NewVBox(
		s.genus,
		grid,
		widget.NewLabelWithStyle("Active moves", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.New(layout.NewCustomPaddedHBoxLayout(0), actives...),
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
