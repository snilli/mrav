package profileui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// pipSize is the size of a dot of Pips.
const pipSize = 12

// Pips shows how far something is as a row of dots, one for each step: the ones that are done are lit, and the rest are
// rings.
type Pips struct {
	widget.BaseWidget
	row         *fyne.Container
	dots        []*canvas.Rectangle // round ones: a rectangle can say its smallest size, a circle cannot
	done, total int
}

func NewPips() *Pips {
	p := &Pips{row: container.NewHBox()}
	p.ExtendBaseWidget(p)
	return p
}

// Done is how many dots are lit, and Total how many there are.
func (p *Pips) Done() int  { return p.done }
func (p *Pips) Total() int { return p.total }

// Set lights done dots of total. What is done is kept between nothing and the total.
func (p *Pips) Set(done, total int) {
	total = max(total, 0)
	done = min(max(done, 0), total)
	p.done, p.total = done, total
	for len(p.dots) < total {
		dot := canvas.NewRectangle(color.Transparent)
		dot.CornerRadius, dot.StrokeWidth = pipSize/2, 1.5
		dot.SetMinSize(fyne.NewSize(pipSize, pipSize))
		p.dots = append(p.dots, dot)
	}
	p.dots = p.dots[:total]
	p.row.RemoveAll()
	for i, dot := range p.dots {
		if i < done {
			dot.FillColor, dot.StrokeColor = toneColors[ToneAccent], toneColors[ToneAccent]
		} else {
			dot.FillColor, dot.StrokeColor = color.Transparent, theme.Color(theme.ColorNameDisabled)
		}
		p.row.Add(container.NewCenter(dot)) // in the middle of the height of the row, at its own size
	}
	p.row.Refresh()
	p.Refresh()
}

func (p *Pips) CreateRenderer() fyne.WidgetRenderer { return widget.NewSimpleRenderer(p.row) }
