package profileui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

const (
	markerWidth = 4
	barHeight   = 10
	baseWidth   = 2
	baseReach   = 5
)

var (
	trackColor color.Color = color.NRGBA{R: 0x2f, G: 0x33, B: 0x38, A: 0xff}
	bandColor  color.Color = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x30}
)

type Meter struct {
	widget.BaseWidget
	track    *canvas.Rectangle
	fill     *canvas.Rectangle
	band     *canvas.Rectangle
	plus     *canvas.Rectangle
	base     *canvas.Rectangle
	span     Span
	hint     string
	hintText *widget.Label
	popup    *widget.PopUp
	ratio    float64
	marker   bool
	color    color.Color
}

func NewMeter() *Meter { return newMeter(false) }

func NewGauge() *Meter { return newMeter(true) }

func newMeter(marker bool) *Meter {
	m := &Meter{
		track:  canvas.NewRectangle(trackColor),
		fill:   canvas.NewRectangle(toneColors[ToneAccent]),
		band:   canvas.NewRectangle(bandColor),
		plus:   canvas.NewRectangle(toneColors[ToneLimeGreen]),
		base:   canvas.NewRectangle(color.White),
		marker: marker,
	}
	m.track.CornerRadius, m.fill.CornerRadius, m.band.CornerRadius = 4, 4, 4
	m.band.Hide()
	m.plus.Hide()
	m.base.Hide()
	m.ExtendBaseWidget(m)
	return m
}

func (m *Meter) Colored(c color.Color) *Meter {
	m.color, m.fill.FillColor = c, c
	return m
}

func (m *Meter) Set(ratio float64, tone Tone) {
	m.ratio = min(max(ratio, 0), 1)
	switch fill, ok := toneColors[tone]; {
	case m.band.Visible():
		m.fill.FillColor = toneColors[ToneRed]
	case m.marker:
		m.fill.FillColor = color.White
	case m.color != nil:
		m.fill.FillColor = m.color
	case ok:
		m.fill.FillColor = fill
	default:
		m.fill.FillColor = toneColors[ToneAccent]
	}
	m.Refresh()
}

func (m *Meter) SetSpan(s Span) {
	m.span = s
	m.band.Show()
	m.plus.Show()
	m.base.Show()
	m.Refresh()
}

func (m *Meter) Ratio() float64 { return m.ratio }

func (m *Meter) CreateRenderer() fyne.WidgetRenderer { return &meterRenderer{m: m} }

type meterRenderer struct{ m *Meter }

func (r *meterRenderer) Layout(size fyne.Size) {
	m := r.m
	y := (size.Height - barHeight) / 2
	m.track.Move(fyne.NewPos(0, y))
	m.track.Resize(fyne.NewSize(size.Width, barHeight))
	m.band.Move(fyne.NewPos(float32(m.span.Low)*size.Width, y))
	m.band.Resize(fyne.NewSize(float32(m.span.High-m.span.Low)*size.Width, barHeight))
	m.base.Move(fyne.NewPos(float32(m.span.Base)*(size.Width-baseWidth), y-baseReach))
	m.base.Resize(fyne.NewSize(baseWidth, barHeight+2*baseReach))
	switch {
	case m.band.Visible():
		from, to, zero := min(m.span.Base, m.ratio), max(m.span.Base, m.ratio), m.span.Zero
		piece(m.fill, from, min(to, zero), size.Width, y)
		piece(m.plus, max(from, zero), to, size.Width, y)
		return
	case m.marker:
		m.fill.Move(fyne.NewPos(float32(m.ratio)*(size.Width-markerWidth), y))
		m.fill.Resize(fyne.NewSize(markerWidth, barHeight))
		return
	}
	m.fill.Move(fyne.NewPos(0, y))
	m.fill.Resize(fyne.NewSize(float32(m.ratio)*size.Width, barHeight))
}

func (r *meterRenderer) MinSize() fyne.Size { return fyne.NewSize(60, barHeight) }

func (r *meterRenderer) Refresh() {
	r.Layout(r.m.Size())
	r.m.track.Refresh()
	r.m.band.Refresh()
	r.m.base.Refresh()
	r.m.fill.Refresh()
	r.m.plus.Refresh()
}

func (r *meterRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.m.track, r.m.band, r.m.fill, r.m.plus, r.m.base}
}

func (r *meterRenderer) Destroy() {}

func (m *Meter) Hint(text string) { m.hint = text }

func (m *Meter) MouseIn(e *desktop.MouseEvent) {
	if m.hint == "" {
		return
	}
	c := fyne.CurrentApp().Driver().CanvasForObject(m)
	if c == nil {
		return
	}
	if m.popup == nil {
		m.hintText = widget.NewLabel("")
		m.popup = widget.NewPopUp(m.hintText, c)
	}
	m.hintText.SetText(m.hint)
	m.popup.ShowAtPosition(e.AbsolutePosition.Add(fyne.NewPos(12, 16)))
}

func (m *Meter) MouseMoved(e *desktop.MouseEvent) {
	if m.popup != nil && m.popup.Visible() {
		m.popup.Move(e.AbsolutePosition.Add(fyne.NewPos(12, 16)))
	}
}

func (m *Meter) MouseOut() {
	if m.popup != nil {
		m.popup.Hide()
	}
}

func piece(r *canvas.Rectangle, from, to float64, width, y float32) {
	r.Move(fyne.NewPos(float32(from)*width, y))
	r.Resize(fyne.NewSize(float32(max(to-from, 0))*width, barHeight))
}
