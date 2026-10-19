package itemui

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"app/internal/core/item"
	"app/internal/core/item/domain"
	player "app/internal/core/player/domain"
	"app/internal/delivery/ui/status"
)

// Screen is the Items page: a button for each place the player carries an item in, and the details of the
// item in the place that was tapped.
type Screen struct {
	svc      item.ItemService
	slots    []domain.Slot
	selected int // the place tapped; -1 for none

	status      *widget.Label
	body        fyne.CanvasObject
	content     fyne.CanvasObject
	strip       *container.Scroll
	buttons     [player.ItemSlots]*widget.Button
	name        *widget.Label
	kind        *widget.Label
	description *widget.Label
	effect      *widget.Label
	sale        *widget.Label
	purchase    *widget.Label
}

func NewScreen(svc item.ItemService) *Screen {
	s := &Screen{
		svc:         svc,
		selected:    -1,
		status:      widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		name:        widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		kind:        widget.NewLabel(""),
		description: widget.NewLabel(""),
		effect:      widget.NewLabel(""),
		sale:        widget.NewLabel(""),
		purchase:    widget.NewLabel(""),
	}
	s.status.Hide()
	s.description.Wrapping, s.effect.Wrapping = fyne.TextWrapWord, fyne.TextWrapWord

	buttons := make([]fyne.CanvasObject, player.ItemSlots)
	for i := range s.buttons {
		s.buttons[i] = widget.NewButton(blank, func() { s.selected = i; s.showDetail() })
		buttons[i] = s.buttons[i]
	}
	// Two rows, as the game lays them out, in a scroll, so that long names never make the window wider.
	s.strip = container.NewHScroll(container.NewGridWithColumns(player.ItemSlots/2, buttons...))
	s.body = container.NewVBox(
		s.strip,
		container.NewHBox(s.name, s.kind, layout.NewSpacer(), widget.NewLabel("Buy"), s.purchase, widget.NewLabel("Sell"), s.sale),
		s.description,
		s.effect,
	)
	// No vertical scroll: the page is a strip along the bottom of the window, which would shrink it to nothing.
	s.content = container.NewPadded(container.NewVBox(s.status, s.body))
	s.showDetail()
	return s
}

// Content is what to put in a window.
func (s *Screen) Content() fyne.CanvasObject { return s.content }

// Load asks the service for what the player carries and shows it. It is safe to call from any goroutine.
func (s *Screen) Load(ctx context.Context) {
	slots, err := s.svc.Inventory(ctx)
	fyne.Do(func() { s.apply(slots, err) })
}

// apply shows the slots, or why there are none. It must run on the UI thread.
func (s *Screen) apply(slots []domain.Slot, err error) {
	if err != nil {
		s.status.SetText(status.Message(err))
		s.status.Show()
		s.body.Hide()
		return
	}
	s.slots = slots
	s.status.Hide()
	s.body.Show()
	for i, b := range s.buttons {
		if i < len(slots) {
			b.SetText(SlotName(slots[i]))
			if slots[i].Locked {
				b.Disable()
			} else {
				b.Enable()
			}
		}
	}
	s.showDetail()
}

// showDetail shows the details of the place tapped, from the slots last shown.
func (s *Screen) showDetail() {
	var slot *domain.Slot
	if s.selected >= 0 && s.selected < len(s.slots) {
		slot = &s.slots[s.selected]
	}
	v := NewDetailView(slot)
	s.name.SetText(v.Name)
	s.kind.SetText(v.Kind)
	s.description.SetText(v.Description)
	s.effect.SetText(v.Effect)
	s.sale.SetText(v.Sale)
	s.purchase.SetText(v.Purchase)
}
