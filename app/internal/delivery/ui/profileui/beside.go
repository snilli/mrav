package profileui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Beside is the layout of a row of two parts. Side by side, the first is at the left edge and the second at the right
// edge, each as wide as it asks, or each half of the row when Even. Stacked, the second is under the first, which makes
// the row as narrow as its wider part.
type Beside struct {
	Stacked bool
	Even    bool
}

func (l Beside) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	a, b := objs[0], objs[1]
	ma, mb := a.MinSize(), b.MinSize()
	pad := theme.Padding()
	if l.Stacked {
		a.Move(fyne.NewPos(0, 0))
		a.Resize(fyne.NewSize(size.Width, ma.Height))
		b.Move(fyne.NewPos(0, ma.Height+pad))
		b.Resize(fyne.NewSize(size.Width, mb.Height))
		return
	}
	wa, wb, h := ma.Width, mb.Width, max(ma.Height, mb.Height)
	if l.Even {
		wa = (size.Width - pad) / 2
		wb = wa
	}
	a.Move(fyne.NewPos(0, 0))
	a.Resize(fyne.NewSize(wa, h))
	b.Move(fyne.NewPos(size.Width-wb, 0))
	b.Resize(fyne.NewSize(wb, h))
}

func (l Beside) MinSize(objs []fyne.CanvasObject) fyne.Size {
	ma, mb := objs[0].MinSize(), objs[1].MinSize()
	pad := theme.Padding()
	switch {
	case l.Stacked:
		return fyne.NewSize(max(ma.Width, mb.Width), ma.Height+mb.Height+pad)
	case l.Even:
		return fyne.NewSize(2*max(ma.Width, mb.Width)+pad, max(ma.Height, mb.Height))
	}
	return fyne.NewSize(ma.Width+mb.Width+pad, max(ma.Height, mb.Height))
}
