package errantryui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	monster "app/internal/core/monster/domain"
	movedomain "app/internal/core/move/domain"
	"app/internal/delivery/ui/profileui"
)

type moveKey struct {
	genus monster.Genus
	slot  int
}

func keyOf(cell movedomain.Cell) moveKey {
	return moveKey{genus: cell.Move.Genus, slot: cell.Move.MemorySlot}
}

type iconButton struct {
	widget.BaseWidget
	key      moveKey
	image    *canvas.Image
	chance   *profileui.ToneBox
	frame    *canvas.Rectangle
	selected bool
	onTap    func()
}

func newIconButton(key moveKey, image *canvas.Image, chance profileui.Box, onTap func()) *iconButton {
	b := &iconButton{key: key, image: image, onTap: onTap, chance: profileui.NewToneBox().Compact(), frame: canvas.NewRectangle(color.Transparent)}
	b.chance.Set(chance)
	b.frame.StrokeWidth = 2
	b.frame.StrokeColor = theme.Color(theme.ColorNamePrimary)
	b.frame.Hide()
	b.ExtendBaseWidget(b)
	return b
}

func (b *iconButton) Tapped(*fyne.PointEvent) { b.onTap() }

func (b *iconButton) Select(on bool) {
	b.selected = on
	if on {
		b.frame.Show()
	} else {
		b.frame.Hide()
	}
	b.frame.Refresh()
}

func (b *iconButton) CreateRenderer() fyne.WidgetRenderer {
	corner := container.NewVBox(layout.NewSpacer(), container.NewHBox(layout.NewSpacer(), b.chance))
	return widget.NewSimpleRenderer(container.NewStack(b.image, b.frame, corner))
}
