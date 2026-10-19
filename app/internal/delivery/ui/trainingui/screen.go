// Package trainingui is the Training panel: each drill with what it gives the monster, how keen the monster is for
// it and whether it is a specialty, and the places the monster is good at.
package trainingui

import (
	"context"
	"fmt"
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

	footnote = "A flat +1 is granted to the primary stat!\n" +
		"Greats reward +1/+3/+5/+7 to the primary stat.\n" +
		"(Cannot exceed +15 for Light/+20 for Heavy)"
)

// recommendedFill is the green behind a drill that is a specialty and that the monster is keen on.
var recommendedFill color.Color = color.NRGBA{R: 50, G: 205, B: 50, A: 70}

// columns are the columns of every row of the table: the drill, what it gives, the bar of the motivation and its
// percent, and the star of a specialty.
var columns = profileui.Columns{158, 0, 90, 48, 24}

// row is a drill in the table. Its name is a button that shows every number of the drill.
type row struct {
	name    *widget.Button
	gain    *widget.Label
	bar     *profileui.Meter
	percent *widget.Label
	star    *widget.Label
	bg      *canvas.Rectangle
	object  fyne.CanvasObject
}

func newRow(name string, onTap func()) *row {
	r := &row{
		name:    widget.NewButton(name, onTap),
		gain:    widget.NewLabel(""),
		bar:     profileui.NewMeter(),
		percent: widget.NewLabel(""),
		star:    widget.NewLabel(""),
		bg:      canvas.NewRectangle(color.Transparent),
	}
	r.name.Alignment = widget.ButtonAlignLeading
	r.name.Importance = widget.LowImportance
	r.object = container.NewStack(r.bg, container.New(columns, r.name, r.gain, r.bar, r.percent, r.star))
	return r
}

// Screen is the Training panel.
type Screen struct {
	svc      training.TrainingService
	last     domain.Training
	loaded   bool // last holds a page
	selected int  // the row whose numbers are shown, or -1

	status  *widget.Label
	body    fyne.CanvasObject // the panel, which is hidden for what to say in its place
	panel   *widget.Accordion
	item    *widget.AccordionItem
	content fyne.CanvasObject
	rows    []*row
	places  *widget.Label
	detail  *widget.Label
}

func NewScreen(svc training.TrainingService) *Screen {
	s := &Screen{
		svc:      svc,
		selected: -1,
		status:   widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		places:   widget.NewLabel(""),
		detail:   widget.NewLabel(""),
	}
	s.status.Hide()
	s.detail.Wrapping = fyne.TextWrapWord

	bold := func(text string) *widget.Label {
		return widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	}
	table := container.NewVBox(container.New(columns, bold("Drill"), bold("Gain"), bold("Motivation"), bold(""), bold(star)))
	for i, d := range domain.Drills() {
		r := newRow(d.Name, func() { s.selected = i; s.showDetail() })
		s.rows = append(s.rows, r)
		table.Add(r.object)
	}

	// A panel that is closed until it is opened. The window scrolls it along with the other panels of its column.
	s.item = widget.NewAccordionItem("Training", container.NewVBox(table, s.places, widget.NewSeparator(), s.detail))
	s.panel = widget.NewAccordion(s.item)
	s.body = s.panel
	s.content = container.NewPadded(container.NewVBox(s.status, s.body))
	s.places.SetText("Good places: none")
	s.showDetail()
	return s
}

// Content is what to put in a window.
func (s *Screen) Content() fyne.CanvasObject { return s.content }

// Load asks the service for the current page and shows it. It is safe to call from any goroutine.
func (s *Screen) Load(ctx context.Context) {
	t, err := s.svc.Current(ctx)
	fyne.Do(func() { s.apply(t, err) })
}

// apply shows a page, or why there is none. It must run on the UI thread.
func (s *Screen) apply(t domain.Training, err error) {
	if err != nil {
		s.loaded = false
		s.status.SetText(status.Message(err))
		s.status.Show()
		s.body.Hide()
		return
	}
	s.last, s.loaded = t, true
	s.status.Hide()
	s.body.Show()

	for i, g := range t.Drills[:min(len(t.Drills), len(s.rows))] {
		r := s.rows[i]
		r.gain.SetText(strings.Join(gainParts(g, true), " / "))
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

	recommended := 0
	for _, g := range t.Drills {
		if g.Recommended {
			recommended++
		}
	}
	s.item.Title = "Training"
	if recommended > 0 {
		s.item.Title = fmt.Sprintf("Training · %d recommended", recommended)
	}
	s.panel.Refresh()

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
	s.showDetail()
}

// showDetail marks the row that is selected and shows every number of its drill above the footnote.
func (s *Screen) showDetail() {
	text := footnote
	if s.loaded && s.selected >= 0 && s.selected < len(s.last.Drills) {
		text = strings.Join(gainParts(s.last.Drills[s.selected], false), "\n") + "\n\n" + footnote
	}
	s.detail.SetText(text)
	for i, r := range s.rows {
		r.name.Importance = widget.LowImportance
		if i == s.selected {
			r.name.Importance = widget.HighImportance
		}
		r.name.Refresh()
	}
}

// gainParts are what a drill gives, a stat to a part: the primary stat, and for a heavy drill the other. Short
// leaves out the numbers in brackets, which say how the gain falls for each outcome of the drill.
func gainParts(g domain.DrillGains, short bool) []string {
	parts := []string{gainText(g.Drill.Primary, g.Primary, short)}
	if g.Drill.Heavy {
		parts = append(parts, gainText(g.Drill.Secondary, g.Secondary, short))
	}
	return parts
}

func gainText(stat monster.Stat, text string, short bool) string {
	if text == noGain {
		return noGain
	}
	if short {
		text, _, _ = strings.Cut(text, " (")
	}
	return profileui.StatNames[stat] + " +" + text
}
