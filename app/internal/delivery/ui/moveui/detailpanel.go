package moveui

import (
	"fyne.io/fyne/v2"

	"app/internal/core/move/domain"
)

type Detail struct{ d *detail }

func NewDetail(icons *Icons) *Detail { return &Detail{d: newDetail(icons.set)} }

func (p *Detail) Content() fyne.CanvasObject { return p.d.content }

func (p *Detail) Show(c domain.Cell) { p.d.show(NewDetailView(c)) }

func (p *Detail) Name() string { return p.d.name.Text }
