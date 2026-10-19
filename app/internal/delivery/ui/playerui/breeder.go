package playerui

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"app/internal/core/player"
	playerdomain "app/internal/core/player/domain"
	"app/internal/delivery/ui/status"
)

// BreederScreen is the Breeder page: the breeder's name. MR2AV can also change it; this viewer only reads.
type BreederScreen struct {
	svc player.PlayerService

	status  *widget.Label
	name    *widget.Label
	content fyne.CanvasObject
}

func NewBreederScreen(svc player.PlayerService) *BreederScreen {
	s := &BreederScreen{
		svc:    svc,
		status: widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		name:   widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	}
	s.status.Hide()
	s.content = container.NewPadded(container.NewVBox(widget.NewLabel("Breeder name"), s.status, s.name))
	return s
}

// Content is what to put in a window.
func (s *BreederScreen) Content() fyne.CanvasObject { return s.content }

// Load asks the service for the player and shows the breeder. It is safe to call from any goroutine.
func (s *BreederScreen) Load(ctx context.Context) {
	p, err := s.svc.Current(ctx)
	fyne.Do(func() { s.apply(p, err) })
}

// apply shows the breeder's name, or why there is none. It must run on the UI thread.
func (s *BreederScreen) apply(p playerdomain.Player, err error) {
	if err != nil {
		s.status.SetText(status.Message(err))
		s.status.Show()
		s.name.Hide()
		return
	}
	s.status.Hide()
	s.name.Show()
	s.name.SetText(BreederText(p))
}
