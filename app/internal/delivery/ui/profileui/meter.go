package profileui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

const (
	markerWidth = 4  // how wide a gauge's marker is
	barHeight   = 10 // how tall every bar is, whatever the row around it
)

// trackColor is the empty part of a bar: dark, so that it shows in either theme.
var trackColor color.Color = color.NRGBA{R: 0x2f, G: 0x33, B: 0x38, A: 0xff}

// Meter is a bar that fills in proportion to a value, in the colour of a tone. As a gauge it shows a marker at the
// value instead of a fill.
type Meter struct {
	widget.BaseWidget
	track  *canvas.Rectangle
	fill   *canvas.Rectangle
	ratio  float64
	marker bool
	color  color.Color // the fill's colour whatever the tone, when it has one
}

func NewMeter() *Meter { return newMeter(false) }

// NewGauge is a Meter that marks the value instead of filling up to it, for a value that runs from low to high
// with the good in the middle or at one end, such as a nature.
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

// Colored makes the meter fill in a colour of its own, and not in the colour of the tone it is given.
func (m *Meter) Colored(c color.Color) *Meter {
	m.color, m.fill.FillColor = c, c
	return m
}

// Set fills the meter to a ratio of 0 to 1, which it keeps to, in the colour of a tone: the accent colour for
// none, or the colour the meter was given. A gauge's marker is always white.
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

// Ratio is how full the meter is, from 0 to 1.
func (m *Meter) Ratio() float64 { return m.ratio }

func (m *Meter) CreateRenderer() fyne.WidgetRenderer { return &meterRenderer{m: m} }

type meterRenderer struct{ m *Meter }

// Layout keeps the bar barHeight tall in the middle of whatever height a row gives it.
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
