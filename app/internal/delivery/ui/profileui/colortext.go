package profileui

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	monster "app/internal/core/monster/domain"
)

var StatColors = [monster.StatCount]color.Color{
	monster.Life:    color.NRGBA{R: 254, G: 254, B: 145, A: 255},
	monster.Pow:     color.NRGBA{R: 238, G: 134, B: 131, A: 255},
	monster.Int:     color.NRGBA{R: 130, G: 209, B: 114, A: 255},
	monster.Skill:   color.NRGBA{R: 238, G: 134, B: 248, A: 255},
	monster.Speed:   color.NRGBA{R: 140, G: 230, B: 232, A: 255},
	monster.Defense: color.NRGBA{R: 127, G: 127, B: 246, A: 255},
}

type Segment struct {
	Text  string
	Color color.Color
}

type ColorText struct {
	widget.BaseWidget
	texts []*canvas.Text
}

func NewColorText(segments ...Segment) *ColorText {
	c := &ColorText{}
	c.ExtendBaseWidget(c)
	c.Set(segments...)
	return c
}

func (c *ColorText) Set(segments ...Segment) {
	c.texts = c.texts[:0]
	for _, s := range segments {
		fill := s.Color
		if fill == nil {
			fill = theme.Color(theme.ColorNameForeground)
		}
		c.texts = append(c.texts, canvas.NewText(s.Text, fill))
	}
	c.Refresh()
}

func (c *ColorText) Segments() []Segment {
	segments := make([]Segment, len(c.texts))
	for i, t := range c.texts {
		segments[i] = Segment{Text: t.Text, Color: t.Color}
	}
	return segments
}

func (c *ColorText) String() string {
	var b strings.Builder
	for _, t := range c.texts {
		b.WriteString(t.Text)
	}
	return b.String()
}

func (c *ColorText) CreateRenderer() fyne.WidgetRenderer { return &colorTextRenderer{c: c} }

type colorTextRenderer struct{ c *ColorText }

func (r *colorTextRenderer) Layout(fyne.Size) {
	pad := theme.InnerPadding()
	x := pad
	for _, t := range r.c.texts {
		m := t.MinSize()
		t.Move(fyne.NewPos(x, pad))
		t.Resize(m)
		x += m.Width
	}
}

func (r *colorTextRenderer) MinSize() fyne.Size {
	pad := theme.InnerPadding()
	var w, h float32
	for _, t := range r.c.texts {
		m := t.MinSize()
		w += m.Width
		h = max(h, m.Height)
	}
	return fyne.NewSize(w+2*pad, h+2*pad)
}

func (r *colorTextRenderer) Refresh() {
	r.Layout(r.c.Size())
	for _, t := range r.c.texts {
		t.Refresh()
	}
}

func (r *colorTextRenderer) Objects() []fyne.CanvasObject {
	objects := make([]fyne.CanvasObject, len(r.c.texts))
	for i, t := range r.c.texts {
		objects[i] = t
	}
	return objects
}

func (r *colorTextRenderer) Destroy() {}
