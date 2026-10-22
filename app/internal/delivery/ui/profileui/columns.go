package profileui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type Columns []float32

func (c Columns) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	pad := theme.Padding()
	fixed, flexible := float32(len(c)-1)*pad, 0
	for _, w := range c {
		if w > 0 {
			fixed += w
		} else {
			flexible++
		}
	}
	var share float32
	if flexible > 0 {
		share = max(size.Width-fixed, 0) / float32(flexible)
	}

	var x float32
	for i, o := range objs[:min(len(objs), len(c))] {
		w := c[i]
		if w <= 0 {
			w = share
		}
		o.Move(fyne.NewPos(x, 0))
		o.Resize(fyne.NewSize(w, size.Height))
		x += w + pad
	}
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
