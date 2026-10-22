package profileui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const pipSize = 12

type Pips struct {
	widget.BaseWidget
	row         *fyne.Container
	dots        []*canvas.Rectangle
	done, total int
}

func NewPips() *Pips {
	p := &Pips{row: container.NewHBox()}
	p.ExtendBaseWidget(p)
	return p
}

func (p *Pips) Done() int  { return p.done }
func (p *Pips) Total() int { return p.total }

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
		p.row.Add(container.NewCenter(dot))
	}
	p.row.Refresh()
	p.Refresh()
}

func (p *Pips) CreateRenderer() fyne.WidgetRenderer { return widget.NewSimpleRenderer(p.row) }
