package moveui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	"app/internal/core/move/domain"
)

type Icons struct{ set *iconSet }

func NewIcons(load IconFunc) *Icons {
	return &Icons{set: &iconSet{load: load, cache: map[string]fyne.Resource{}}}
}

func (i *Icons) Image(m domain.Move, w, h float32) *canvas.Image {
	return pixelImage(i.set.tinted(m.Icon, m.Kind), w, h)
}
