package moveui

import (
	"fyne.io/fyne/v2"

	"app/internal/core/move/domain"
)

// Detail is the details of one move, as the Moves page shows them under its grid, for the other pages that show a move:
// its name and where it stands, its numbers, its description, and what it still needs.
type Detail struct{ d *detail }

func NewDetail(icons *Icons) *Detail { return &Detail{d: newDetail(icons.set)} }

// Content is what to put in a window.
func (p *Detail) Content() fyne.CanvasObject { return p.d.content }

// Show shows a move, where it stands for the monster, and what the monster still needs for it.
func (p *Detail) Show(c domain.Cell) { p.d.show(NewDetailView(c)) }

// Name is the name of the move shown.
func (p *Detail) Name() string { return p.d.name.Text }
