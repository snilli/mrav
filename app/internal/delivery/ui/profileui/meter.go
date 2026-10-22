package profileui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

const (
	markerWidth = 4
	barHeight   = 10
)

var trackColor color.Color = color.NRGBA{R: 0x2f, G: 0x33, B: 0x38, A: 0xff}

type Meter struct {
	widget.BaseWidget
	track  *canvas.Rectangle
	fill   *canvas.Rectangle
	ratio  float64
	marker bool
	color  color.Color
}

func NewMeter() *Meter { return newMeter(false) }

func NewGauge() *Meter { return newMeter(true) }

func newMeter(marker bool) *Meter {
	m := &Meter{
		track:  canvas.NewRectangle(trackColor),
		fill:   canvas.NewRectangle(toneColors[ToneAccent]),
		marker: marker,
	}
	m.track.CornerRadius, m.fill.CornerRadius = 4, 4
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

func (m *Meter) Ratio() float64 { return m.ratio }

func (m *Meter) CreateRenderer() fyne.WidgetRenderer { return &meterRenderer{m: m} }

type meterRenderer struct{ m *Meter }

func (r *meterRenderer) Layout(size fyne.Size) {
	m := r.m
	y := (size.Height - barHeight) / 2
	m.track.Move(fyne.NewPos(0, y))
	m.track.Resize(fyne.NewSize(size.Width, barHeight))
	if m.marker {
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
	r.m.fill.Refresh()
}

func (r *meterRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.m.track, r.m.fill}
}

func (r *meterRenderer) Destroy() {}
