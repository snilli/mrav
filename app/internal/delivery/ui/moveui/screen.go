package moveui

import (
	"context"
	"fmt"
	"image/color"
	"slices"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	monster "app/internal/core/monster/domain"
	"app/internal/core/move"
	"app/internal/core/move/domain"
	"app/internal/delivery/ui/profileui"
	"app/internal/delivery/ui/status"
)

const (
	zoom       = 1.5
	iconW      = 48 * zoom
	iconH      = 32 * zoom
	activeW    = 48
	activeH    = 32
	frameWidth = 3
	ribbonH    = 14
)

var (
	metColor    color.Color = color.NRGBA{R: 50, G: 205, B: 50, A: 255}
	readyColor  color.Color = color.NRGBA{R: 255, G: 159, B: 28, A: 255}
	chanceColor color.Color = color.NRGBA{R: 79, G: 195, B: 247, A: 255}
	lockedColor color.Color = color.NRGBA{R: 150, G: 150, B: 150, A: 255}
	unmetColor  color.Color = color.NRGBA{R: 255, G: 69, A: 255}
	dimColor    color.Color = color.NRGBA{A: 150}
)

type IconFunc func(key string) ([]byte, error)

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

func (is *iconSet) tinted(key string, kind domain.Kind) fyne.Resource {
	if kind != domain.KindPower && kind != domain.KindIntelligence {
		return is.resource(key)
	}
	id := key + "|" + string(kind)
	if r, ok := is.cache[id]; ok {
		return r
	}
	b, err := is.load(key)
	if err != nil {
		return is.resource(key)
	}
	b, err = tint(b, kind)
	if err != nil {
		return is.resource(key)
	}
	r := fyne.NewStaticResource(id+".png", b)
	is.cache[id] = r
	return r
}

func (is *iconSet) nature(n domain.Nature) fyne.Resource {
	b := natureIcon(n)
	if b == nil {
		return nil
	}
	id := fmt.Sprintf("nature|%d", n)
	if r, ok := is.cache[id]; ok {
		return r
	}
	r := fyne.NewStaticResource(id+".png", b)
	is.cache[id] = r
	return r
}

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

type picture struct {
	img  *canvas.Image
	key  string
	kind domain.Kind
}

func newPicture(is *iconSet, w, h float32) *picture {
	return &picture{img: pixelImage(is.resource(NoMove), w, h), key: NoMove}
}

func (p *picture) show(key string, kind domain.Kind, is *iconSet) {
	if key == p.key && kind == p.kind {
		return
	}
	p.img.Resource, p.key, p.kind = is.tinted(key, kind), key, kind
	p.img.Refresh()
}

type cell struct {
	widget.BaseWidget
	iconKey   string
	iconKind  domain.Kind
	icon      *canvas.Image
	nature    *canvas.Image
	dim       *canvas.Rectangle
	frame     *canvas.Rectangle
	ribbon    *fyne.Container
	ribbonBg  *canvas.Rectangle
	ribbonTxt *canvas.Text
	highlight *canvas.Image
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
		nature:    pixelImage(is.nature(domain.NatureGood), natureW, natureH),
		onTap:     onTap,
	}
	c.dim.Resize(fyne.NewSize(iconW, iconH))
	c.frame.StrokeWidth = frameWidth
	c.frame.Resize(fyne.NewSize(iconW, iconH))

	c.ribbonBg = canvas.NewRectangle(readyColor)
	c.ribbonBg.Move(fyne.NewPos(0, iconH-ribbonH))
	c.ribbonBg.Resize(fyne.NewSize(iconW, ribbonH))
	c.ribbonTxt = canvas.NewText("READY", color.Black)
	c.ribbonTxt.TextSize, c.ribbonTxt.TextStyle = 10, fyne.TextStyle{Bold: true}
	c.centreRibbonText()
	c.ribbon = container.NewWithoutLayout(c.ribbonBg, c.ribbonTxt)
	c.ribbon.Resize(fyne.NewSize(iconW, iconH))

	c.nature.Move(fyne.NewPos(2*zoom, 2*zoom))
	for _, o := range []fyne.CanvasObject{c.dim, c.frame, c.ribbon, c.highlight, c.nature} {
		o.Hide()
	}
	c.body = container.NewWithoutLayout(c.icon, c.dim, c.frame, c.ribbon, c.highlight, c.nature)
	c.ExtendBaseWidget(c)
	return c
}

func (c *cell) centreRibbonText() {
	want := c.ribbonTxt.MinSize()
	c.ribbonTxt.Move(fyne.NewPos((iconW-want.Width)/2, iconH-ribbonH+(ribbonH-want.Height)/2))
}

func (c *cell) CreateRenderer() fyne.WidgetRenderer { return widget.NewSimpleRenderer(c.body) }
func (c *cell) MinSize() fyne.Size                  { return fyne.NewSize(iconW, iconH) }
func (c *cell) Tapped(*fyne.PointEvent)             { c.onTap() }

func (c *cell) show(v CellView, is *iconSet) {
	if v.Icon != c.iconKey || v.Kind != c.iconKind {
		c.icon.Resource, c.iconKey, c.iconKind = is.tinted(v.Icon, v.Kind), v.Icon, v.Kind
		c.icon.Refresh()
	}
	if r := is.nature(v.Nature); r != nil && r != c.nature.Resource {
		c.nature.Resource = r
		c.nature.Refresh()
	}
	setVisible(c.nature, v.Nature != domain.NatureNone)
	setVisible(c.dim, v.Locked)
	setVisible(c.ribbon, v.CanLearn || v.Chance != "")
	text, fill := "READY", readyColor
	if v.Chance != "" {
		text, fill = v.Chance, chanceColor
	}
	if c.ribbonTxt.Text != text {
		c.ribbonTxt.Text = text
		c.centreRibbonText()
		c.ribbonTxt.Refresh()
	}
	if c.ribbonBg.FillColor != fill {
		c.ribbonBg.FillColor = fill
		c.ribbonBg.Refresh()
	}
	stroke := readyColor
	switch {
	case v.Learned:
		stroke = metColor
	case v.Chance != "":
		stroke = chanceColor
	}
	setVisible(c.frame, v.Learned || v.CanLearn || v.Chance != "")
	c.frame.StrokeColor = stroke
	c.frame.Refresh()
}

type legend struct {
	learned, ready, chance, locked *widget.Label
	object                         fyne.CanvasObject
}

func newLegend() *legend {
	l := &legend{learned: widget.NewLabel(""), ready: widget.NewLabel(""), chance: widget.NewLabel(""), locked: widget.NewLabel("")}
	item := func(c color.Color, label *widget.Label) fyne.CanvasObject {
		swatch := canvas.NewRectangle(c)
		swatch.SetMinSize(fyne.NewSize(12, 12))
		return container.NewHBox(container.NewCenter(swatch), label)
	}
	l.object = container.NewHBox(item(metColor, l.learned), item(readyColor, l.ready), item(chanceColor, l.chance), item(lockedColor, l.locked))
	l.show(GridView{})
	return l
}

func (l *legend) show(v GridView) {
	l.learned.SetText(fmt.Sprintf("Learned %d", v.Learned))
	l.ready.SetText(fmt.Sprintf("Ready to learn %d", v.Ready))
	l.chance.SetText(fmt.Sprintf("Errantry chance %d", v.Chance))
	l.locked.SetText(fmt.Sprintf("Not yet %d", v.Locked))
}

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

type need struct {
	mark    *canvas.Text
	label   *widget.Label
	colored *profileui.ColorText
	object  fyne.CanvasObject
}

func (n need) words() string {
	if n.colored != nil {
		return n.colored.String()
	}
	return n.label.Text
}

func statSegments(r RequirementView) []profileui.Segment {
	lead, rest, found := strings.Cut(r.Text, " at least ")
	words := strings.Split(lead, " + ")
	if len(r.Stats) == 0 || !found || len(words) != len(r.Stats) {
		return nil
	}
	var segments []profileui.Segment
	for i, word := range words {
		if r.Stats[i] < 0 || r.Stats[i] >= monster.StatCount {
			return nil
		}
		if i > 0 {
			segments = append(segments, profileui.Segment{Text: " + "})
		}
		segments = append(segments, profileui.Segment{Text: word, Color: profileui.StatColors[r.Stats[i]]})
	}
	tail := " at least " + rest
	if r.Have != "" {
		tail += " (" + r.Have + ")"
	}
	return append(segments, profileui.Segment{Text: tail})
}

func newNeed(r RequirementView) need {
	n := need{mark: canvas.NewText("✗", unmetColor)}
	if r.Met {
		n.mark.Text, n.mark.Color = "✓", metColor
	}
	if segments := statSegments(r); segments != nil {
		n.colored = profileui.NewColorText(segments...)
		n.object = container.NewBorder(nil, nil, n.mark, nil, n.colored)
		return n
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

type detail struct {
	name                                                 *canvas.Text
	nature                                               *canvas.Image
	icons                                                *iconSet
	status                                               *badge
	guts, damage, hit, withering, sharpness, extra, uses *widget.Label
	description                                          *widget.Label
	needsBox                                             *fyne.Container
	requirements                                         *fyne.Container
	needs                                                []need
	content                                              fyne.CanvasObject
}

func newDetail(is *iconSet) *detail {
	d := &detail{
		icons:       is,
		nature:      pixelImage(is.nature(domain.NatureGood), natureW*zoom, natureH*zoom),
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
		widget.NewLabelWithStyle("To unlock it", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		d.requirements,
	)
	d.needsBox.Hide()
	d.nature.Hide()
	d.content = container.NewVBox(
		container.NewBorder(nil, nil, container.NewCenter(d.nature), container.NewCenter(d.status.object), d.name),
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
		return color.NRGBA{R: 0xDA, G: 0xA5, B: 0x20, A: 0xFF}
	case domain.KindIntelligence:
		return color.NRGBA{R: 0x3C, G: 0xB3, B: 0x71, A: 0xFF}
	}
	return theme.Color(theme.ColorNameForeground)
}

func (d *detail) show(v DetailView) {
	d.name.Text, d.name.Color = v.Name, nameColor(v.Kind)
	d.name.Refresh()
	if r := d.icons.nature(v.Nature); r != nil && r != d.nature.Resource {
		d.nature.Resource = r
		d.nature.Refresh()
	}
	setVisible(d.nature, v.Nature != domain.NatureNone)
	switch {
	case v.Learned:
		d.status.set(v.Status, metColor)
	case v.CanLearn:
		d.status.set(v.Status, readyColor)
	case v.Chance != "":
		d.status.set(v.Status, chanceColor)
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

type Screen struct {
	seen  status.Seen[domain.Grid]
	svc   move.MoveService
	icons *iconSet

	cells [domain.Columns][domain.Rows]*cell

	cellGrid  fyne.CanvasObject
	activeRow fyne.CanvasObject
	active    [domain.Columns]*picture
	genus     *widget.Label
	legend    *legend
	detail    *detail

	body    fyne.CanvasObject
	content fyne.CanvasObject
	grid    domain.Grid
	hasSel  bool
	selCol  int
	selRow  int
}

func NewScreen(svc move.MoveService, icons IconFunc) *Screen {
	is := &iconSet{load: icons, cache: map[string]fyne.Resource{}}
	s := &Screen{
		svc:    svc,
		icons:  is,
		genus:  widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		legend: newLegend(),
		detail: newDetail(is),
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
	slices.Reverse(columns)
	s.cellGrid = container.New(layout.NewCustomPaddedHBoxLayout(0), columns...)

	pictures := make([]fyne.CanvasObject, domain.Columns)
	for c := range s.active {
		s.active[c] = newPicture(s.icons, activeW, activeH)
		pictures[c] = s.active[c].img
	}
	slices.Reverse(pictures)
	actives := append([]fyne.CanvasObject{profileui.Quiet("Active moves")}, pictures...)
	s.activeRow = container.NewHBox(actives...)
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

func (s *Screen) Content() fyne.CanvasObject { return s.content }

func (s *Screen) Load(ctx context.Context) {
	g, err := s.svc.CurrentGrid(ctx)
	fyne.Do(func() {
		if s.seen.Changed(g, err) {
			s.apply(g, err)
		}
	})
}

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
		s.active[c].show(v.Active[c], v.ActiveKind[c], s.icons)
	}
	s.legend.show(v)
	s.showSelection()
}

func (s *Screen) selectCell(col, row int) {
	s.hasSel, s.selCol, s.selRow = true, col, row
	s.showSelection()
}

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
