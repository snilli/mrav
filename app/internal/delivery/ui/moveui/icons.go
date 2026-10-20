package moveui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	"app/internal/core/move/domain"
)

// Icons draws the icon of a move as the Moves page does, for the other pages that show a move: coloured by the kind of the
// move, each icon made once.
type Icons struct{ set *iconSet }

func NewIcons(load IconFunc) *Icons {
	return &Icons{set: &iconSet{load: load, cache: map[string]fyne.Resource{}}}
}

// Image is the icon of a move, w by h, its pixels not smoothed. A move whose icon is missing is an empty cell, and
// nothing at all when the empty cell is missing too.
func (i *Icons) Image(m domain.Move, w, h float32) *canvas.Image {
	return pixelImage(i.set.tinted(m.Icon, m.Kind), w, h)
}
