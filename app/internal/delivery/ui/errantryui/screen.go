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

// columns are the columns of every row of the table: the words about the place, and then the chances of the four weeks, over
// the moves that are ready to learn, which share what is left.
var columns = profileui.Columns{140, 0}

// iconSize is the size of the icon of a move, as the game draws it.
var iconSize = fyne.NewSize(48, 32)

// lineLayout keeps a line as high as an icon for its one object and puts the object in the middle of it, so that a place with
// no move to show is as high as one with a line of icons.
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

// row is a place in the table: its name and what it would give on the left, and on the right the chances of its weeks and under them the
// moves that it teaches that are ready to learn.
type row struct {
	place    *widget.Label
	weeks    [domain.ErrantryWeeks]*profileui.ToneBox
	gains    *profileui.ColorText
	ready    *fyne.Container   // the icons of the moves that are ready to learn, as many to a line as there is room for
	none     *widget.Label     // says that there are none, when there are none
	noneLine fyne.CanvasObject // the line that keeps the room of a line of icons for it
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

// Screen is the Errantry panel.
type Screen struct {
	seen     status.Seen[loaded] // what apply last drew, so that the same is not drawn again
	profiles profile.ProfileService
	moves    move.MoveService
	icons    *moveui.Icons

	status   *widget.Label
	body     fyne.CanvasObject // the table and the details, hidden until there is a monster, and for what to say in its place
	content  fyne.CanvasObject
	rows     []*row
	detail   *moveui.Detail // the details of the move whose icon was tapped, under the table
	selected *moveKey       // the move whose icon was tapped, while it is still ready to learn
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
	s.detail.Content().Hide() // until the icon of a move is tapped
	s.body = container.NewVBox(table, s.detail.Content())
	s.body.Hide() // until there is a monster
	s.content = container.NewPadded(container.NewVBox(s.status, s.body))
	return s
}

// Content is what to put in a window.
func (s *Screen) Content() fyne.CanvasObject { return s.content }

// Load asks the services for the monster as it is now and for its moves, and shows the chances for it. It is safe to call
// from any goroutine.
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

// loaded is what Load gave apply.
type loaded struct {
	p domain.Profile
	g movedomain.Grid
}

// choose selects a move: its icon is marked, and its details are shown.
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
}

// apply shows the chances and the moves, or why there are none. It must run on the UI thread.
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
	found := false // the move that was selected is still ready
	for i, r := range s.rows {
		for week, box := range v.Rows[i].Weeks {
			r.weeks[week].Set(box)
		}
		r.gains.Set(v.Rows[i].Gains...)
		// The icons are kept while the moves are the same ones, and only told the new chances: an icon made again on each
		// load would lose a tap whose press and release fall either side of it.
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
				s.detail.Show(cell) // with what the monster has now
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
