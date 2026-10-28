package profileui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

const LineWidth float32 = 1

var LineColor color.Color = color.NRGBA{R: 0x8a, G: 0x90, B: 0x98, A: 0xff}

type Grid struct{ Columns Columns }

func NewGrid(columns Columns) *fyne.Container {
	lines := make([]fyne.CanvasObject, len(columns))
	for i := range lines {
		lines[i] = canvas.NewRectangle(LineColor)
	}
	return container.New(Grid{columns}, lines...)
}

func (g Grid) Layout(lines []fyne.CanvasObject, size fyne.Size) {
	edges := g.Columns.Edges(size.Width)
	for i, x := range edges {
		lines[i].Move(fyne.NewPos(x-LineWidth/2, 0))
		lines[i].Resize(fyne.NewSize(LineWidth, size.Height))
	}
	bottom := lines[len(edges)]
	bottom.Move(fyne.NewPos(0, size.Height-LineWidth))
	bottom.Resize(fyne.NewSize(size.Width, LineWidth))
}

func (Grid) MinSize([]fyne.CanvasObject) fyne.Size { return fyne.Size{} }
