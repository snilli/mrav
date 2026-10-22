package errantryui

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"app/internal/core/move"
	movedomain "app/internal/core/move/domain"
	"app/internal/core/profile"
	"app/internal/core/profile/domain"
	"app/internal/delivery/ui/moveui"
	"app/internal/delivery/ui/profileui"
	"app/internal/delivery/ui/status"
)

var columns = profileui.Columns{140, 0}

var iconSize = fyne.NewSize(48, 32)

type lineLayout struct{}

func (lineLayout) MinSize(objs []fyne.CanvasObject) fyne.Size {
	min := objs[0].MinSize()
	return fyne.NewSize(min.Width, max(min.Height, iconSize.Height))
}

func (lineLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	h := objs[0].MinSize().Height
	objs[0].Resize(fyne.NewSize(size.Width, h))
	objs[0].Move(fyne.NewPos(0, (size.Height-h)/2))
}

type row struct {
	place    *widget.Label
	weeks    [domain.ErrantryWeeks]*profileui.ToneBox
	gains    *profileui.ColorText
	ready    *fyne.Container
	none     *widget.Label
	noneLine fyne.CanvasObject
	object   fyne.CanvasObject
}

func newRow(l domain.ErrantryLocale) *row {
	r := &row{
		place: widget.NewLabelWithStyle(l.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		gains: profileui.NewColorText(),
		ready: container.NewGridWrap(iconSize),
		none:  profileui.Quiet("No tech ready"),
	}
	chances := make([]fyne.CanvasObject, len(r.weeks))
	for week := range r.weeks {
		r.weeks[week] = profileui.NewToneBox()
		chances[week] = r.weeks[week]
	}
	left := container.NewVBox(r.place, r.gains)
	r.noneLine = container.New(lineLayout{}, r.none)
	right := container.NewVBox(container.NewGridWithColumns(len(chances), chances...), r.ready, r.noneLine)
	r.object = container.NewVBox(container.New(columns, left, right), widget.NewSeparator())
	return r
}

type Screen struct {
	seen     status.Seen[loaded]
	profiles profile.ProfileService
	moves    move.MoveService
	icons    *moveui.Icons

	status   *widget.Label
	body     fyne.CanvasObject
	content  fyne.CanvasObject
	rows     []*row
	detail   *moveui.Detail
	selected *moveKey
}

func NewScreen(profiles profile.ProfileService, moves move.MoveService, icons *moveui.Icons) *Screen {
	s := &Screen{
		profiles: profiles, moves: moves, icons: icons,
		status: widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		detail: moveui.NewDetail(icons),
	}
	s.status.Hide()

	bold := func(text string) *widget.Label {
		return widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	}
	table := container.NewVBox(container.New(columns, bold("Place"),
		container.NewGridWithColumns(domain.ErrantryWeeks, bold("Wk 1"), bold("Wk 2"), bold("Wk 3"), bold("Wk 4"))))
	for _, l := range domain.ErrantryLocales() {
		r := newRow(l)
		s.rows = append(s.rows, r)
		table.Add(r.object)
	}
	s.detail.Content().Hide()
	s.body = container.NewVBox(table, s.detail.Content())
	s.body.Hide()
	s.content = container.NewPadded(container.NewVBox(s.status, s.body))
	return s
}

func (s *Screen) Content() fyne.CanvasObject { return s.content }

func (s *Screen) Load(ctx context.Context) {
	p, err := s.profiles.Current(ctx)
	var g movedomain.Grid
	if err == nil {
		g, err = s.moves.CurrentGrid(ctx)
	}
	fyne.Do(func() {
		if s.seen.Changed(loaded{p, g}, err) {
			s.apply(p, g, err)
		}
	})
}

type loaded struct {
	p domain.Profile
	g movedomain.Grid
}

func (s *Screen) choose(key moveKey, cell movedomain.Cell) {
	s.selected = &key
	for _, r := range s.rows {
		for _, o := range r.ready.Objects {
			button := o.(*iconButton)
			button.Select(button.key == key)
		}
	}
	s.detail.Show(cell)
	s.detail.Content().Show()
	// Show only clears Hidden: the parent's Refresh is what lays the details out and redraws.
	s.body.Refresh()
}

func (s *Screen) apply(p domain.Profile, g movedomain.Grid, err error) {
	if err != nil {
		s.status.SetText(status.Message(err))
		s.status.Show()
		s.body.Hide()
		return
	}
	s.status.Hide()
	s.body.Show()

	v := NewView(p, g)
	found := false
	for i, r := range s.rows {
		for week, box := range v.Rows[i].Weeks {
			r.weeks[week].Set(box)
		}
		r.gains.Set(v.Rows[i].Gains...)
		keep := len(r.ready.Objects) == len(v.Rows[i].Ready)
		for j, ready := range v.Rows[i].Ready {
			keep = keep && r.ready.Objects[j].(*iconButton).key == keyOf(ready.Cell)
		}
		if !keep {
			r.ready.RemoveAll()
		}
		for j, ready := range v.Rows[i].Ready {
			cell := ready.Cell
			key := keyOf(cell)
			choose := func() { s.choose(key, cell) }
			var button *iconButton
			if keep {
				button = r.ready.Objects[j].(*iconButton)
				button.chance.Set(ready.Chance)
				button.onTap = choose
			} else {
				button = newIconButton(key, s.icons.Image(*cell.Move, iconSize.Width, iconSize.Height), ready.Chance, choose)
				r.ready.Add(button)
			}
			selected := s.selected != nil && *s.selected == key
			button.Select(selected)
			if selected {
				s.detail.Show(cell)
				found = true
			}
		}
		if len(v.Rows[i].Ready) == 0 {
			r.ready.Hide()
			r.none.Show()
			r.noneLine.Show()
		} else {
			r.ready.Show()
			r.none.Hide()
			r.noneLine.Hide()
		}
	}
	if !found {
		s.selected = nil
		s.detail.Content().Hide()
	}
}
