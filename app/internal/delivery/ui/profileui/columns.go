package profileui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type Columns []float32

func (c Columns) widths(total float32) []float32 {
	fixed, flexible := float32(len(c)-1)*theme.Padding(), 0
	for _, w := range c {
		if w > 0 {
			fixed += w
		} else {
			flexible++
		}
	}
	var share float32
	if flexible > 0 {
		share = max(total-fixed, 0) / float32(flexible)
	}

	widths := make([]float32, len(c))
	for i, w := range c {
		widths[i] = w
		if w <= 0 {
			widths[i] = share
		}
	}
	return widths
}

func (c Columns) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	pad := theme.Padding()
	widths := c.widths(size.Width)

	var x float32
	for i, o := range objs[:min(len(objs), len(c))] {
		o.Move(fyne.NewPos(x, 0))
		o.Resize(fyne.NewSize(widths[i], size.Height))
		x += widths[i] + pad
	}
}

func (c Columns) Edges(width float32) []float32 {
	pad := theme.Padding()
	var edges []float32
	var x float32
	for _, w := range c.widths(width)[:max(len(c)-1, 0)] {
		x += w
		edges = append(edges, x+pad/2)
		x += pad
	}
	return edges
}

func (c Columns) MinSize(objs []fyne.CanvasObject) fyne.Size {
	size := fyne.NewSize(float32(len(c)-1)*theme.Padding(), 0)
	for i, w := range c {
		size.Width += w
		if i < len(objs) {
			want := objs[i].MinSize()
			if w <= 0 {
				size.Width += want.Width
			}
			size.Height = max(size.Height, want.Height)
		}
	}
	return size
}
