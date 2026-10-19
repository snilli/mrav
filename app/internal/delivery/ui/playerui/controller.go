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

// ControllerScreen is the Controller page: the buttons held on controller 1, which is a view of what the
// game is reading, so it wants a faster refresh than the other pages.
type ControllerScreen struct {
	svc player.PlayerService

	status   *widget.Label
	body     fyne.CanvasObject
	content  fyne.CanvasObject
	shoulder *widget.Label
	face     *widget.Label
	misc     *widget.Label
	dpad     *widget.Label
}

func NewControllerScreen(svc player.PlayerService) *ControllerScreen {
	s := &ControllerScreen{
		svc:      svc,
		status:   widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		shoulder: widget.NewLabel(""),
		face:     widget.NewLabel(""),
		misc:     widget.NewLabel(""),
		dpad:     widget.NewLabel(""),
	}
	s.status.Hide()
	s.body = widget.NewForm(
		widget.NewFormItem("Shoulder", s.shoulder),
		widget.NewFormItem("Face", s.face),
		widget.NewFormItem("Middle", s.misc),
		widget.NewFormItem("D-Pad", s.dpad),
	)
	s.content = container.NewPadded(container.NewVBox(s.status, s.body))
	return s
}

// Content is what to put in a window.
func (s *ControllerScreen) Content() fyne.CanvasObject { return s.content }

// Load asks the service for the player and shows the buttons held. It is safe to call from any goroutine.
func (s *ControllerScreen) Load(ctx context.Context) {
	p, err := s.svc.Current(ctx)
	fyne.Do(func() { s.apply(p, err) })
}

// apply shows the buttons held, or why there are none to show. It must run on the UI thread.
func (s *ControllerScreen) apply(p playerdomain.Player, err error) {
	if err != nil {
		s.status.SetText(status.Message(err))
		s.status.Show()
		s.body.Hide()
		return
	}
	s.status.Hide()
	s.body.Show()
	v := NewControllerView(p)
	s.shoulder.SetText(v.Shoulder)
	s.face.SetText(v.Face)
	s.misc.SetText(v.Misc)
	s.dpad.SetText(v.DPad)
}
