// Package trainingui is the Training panel: each drill with what it gives the monster, how keen the monster is for
// it and whether it is a specialty, and the places the monster is good at.
package trainingui

import (
	"context"
	"image/color"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	monster "app/internal/core/monster/domain"
	"app/internal/core/training"
	"app/internal/core/training/domain"
	"app/internal/delivery/ui/profileui"
	"app/internal/delivery/ui/status"
)

const (
	noGain = "N/A"
	star   = "★"

	// firstPlace is the first of the specialties that are places rather than drills: the pool, then the places
	// of the map.
	firstPlace = 9
)

// recommendedFill is the green behind a drill that is a specialty and that the monster is keen on.
var recommendedFill color.Color = color.NRGBA{R: 50, G: 205, B: 50, A: 70}

// columns are the columns of every row of the table: the drill, what it gives, the bar of the motivation and its
// percent, and the star of a specialty.
var columns = profileui.Columns{80, 0, 90, 48, 24}

// row is a drill in the table.
type row struct {
	name    *widget.Label
	gain    *profileui.ColorText
	bar     *profileui.Meter
	percent *widget.Label
	star    *widget.Label
	bg      *canvas.Rectangle
	object  fyne.CanvasObject
}

func newRow(name string) *row {
	r := &row{
		name:    widget.NewLabelWithStyle(name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		gain:    profileui.NewColorText(),
		bar:     profileui.NewMeter(),
		percent: widget.NewLabel(""),
		star:    widget.NewLabel(""),
		bg:      canvas.NewRectangle(color.Transparent),
	}
	r.object = container.NewStack(r.bg, container.New(columns, r.name, r.gain, r.bar, r.percent, r.star))
	return r
}

// Screen is the Training panel.
type Screen struct {
	seen status.Seen[domain.Training] // what apply last drew, so that the same is not drawn again
	svc  training.TrainingService

	status  *widget.Label
	body    fyne.CanvasObject // the table, which is hidden for what to say in its place
	content fyne.CanvasObject
	rows    []*row
	places  *widget.Label
}

func NewScreen(svc training.TrainingService) *Screen {
	s := &Screen{
		svc:    svc,
		status: widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		places: widget.NewLabel(""),
	}
	s.status.Hide()

	bold := func(text string) *widget.Label {
		return widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	}
	table := container.NewVBox(container.New(columns, bold("Drill"), bold("Gain"), bold("Motivation"), bold(""), bold(star)))
	for _, d := range domain.Drills() {
		r := newRow(d.Name)
		s.rows = append(s.rows, r)
		table.Add(r.object)
	}

	s.body = container.NewVBox(table, s.places)
	s.content = container.NewPadded(container.NewVBox(s.status, s.body))
	s.places.SetText("Good places: none")
	return s
}

// Content is what to put in a window.
func (s *Screen) Content() fyne.CanvasObject { return s.content }

// Load asks the service for the current page and shows it. It is safe to call from any goroutine.
func (s *Screen) Load(ctx context.Context) {
	t, err := s.svc.Current(ctx)
	fyne.Do(func() {
		if s.seen.Changed(t, err) {
			s.apply(t, err)
		}
	})
}

// apply shows a page, or why there is none. It must run on the UI thread.
func (s *Screen) apply(t domain.Training, err error) {
	if err != nil {
		s.status.SetText(status.Message(err))
		s.status.Show()
		s.body.Hide()
		return
	}
	s.status.Hide()
	s.body.Show()

	for i, g := range t.Drills[:min(len(t.Drills), len(s.rows))] {
		r := s.rows[i]
		r.gain.Set(gainSegments(g)...)
		r.bar.Set(float64(g.Motivation)/100, profileui.MotivationTone(g.Motivation))
		r.percent.SetText(strconv.Itoa(g.Motivation) + "%")
		r.star.SetText("")
		r.bg.FillColor = color.Transparent
		if g.Specialty {
			r.star.SetText(star)
		}
		if g.Recommended {
			r.bg.FillColor = recommendedFill
		}
		r.bg.Refresh()
	}

	var places []string
	for bit := firstPlace; bit < domain.SpecialtyCount; bit++ {
		if t.Specialties[bit] {
			places = append(places, domain.SpecialtyNames[bit])
		}
	}
	if len(places) == 0 {
		places = []string{"none"}
	}
	s.places.SetText("Good places: " + strings.Join(places, ", "))
}

// gainSegments are what a drill gives, a stat to a part, the stat named in its colour: the primary stat, and for a heavy
// drill the other.
func gainSegments(g domain.DrillGains) []profileui.Segment {
	segments := gainText(g.Drill.Primary, g.Primary)
	if g.Drill.Heavy {
		segments = append(segments, profileui.Segment{Text: " / "})
		segments = append(segments, gainText(g.Drill.Secondary, g.Secondary)...)
	}
	return segments
}

// gainText is a stat and what each outcome of the drill gives it: the table says "9-12 (9/10/11/12)", the range and then
// the four outcomes, and the outcomes say all of it.
func gainText(stat monster.Stat, text string) []profileui.Segment {
	if text == noGain {
		return []profileui.Segment{{Text: noGain}}
	}
	if _, outcomes, found := strings.Cut(text, " ("); found {
		text = "(" + outcomes
	}
	return []profileui.Segment{{Text: profileui.StatNames[stat], Color: profileui.StatColors[stat]}, {Text: " " + text}}
}
