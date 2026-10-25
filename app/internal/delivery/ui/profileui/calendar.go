package profileui

import (
	"cmp"
	"image/color"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	player "app/internal/core/player/domain"
	"app/internal/core/profile/domain"
)

type CalendarView struct {
	None  string
	Years []CalYearView
}

type CalYearView struct {
	Title  string
	Months []CalMonthView
}

type CalMonthView struct {
	Title string
	Count int
	Weeks []CalWeekView
}

type CalWeekView struct {
	Title  string
	Now    bool
	Events []CalEventView
}

type CalEventView struct {
	Name   string
	Type   string
	Note   string
	Detail string
	Status Box
}

type LegendItem struct {
	Status Box
	Means  string
}

var fitStatus = map[domain.Fit]Box{
	domain.FitReady:   {Text: "Ready", Tone: ToneLimeGreen},
	domain.FitNotYet:  {Text: "Not yet", Tone: ToneOrange},
	domain.FitUnknown: {Text: "Maybe", Tone: ToneYellow},
	domain.FitWarn:    {Text: "Over rank", Tone: ToneOrangeRed},
}

var CalendarLegend = []LegendItem{
	{Status: fitStatus[domain.FitReady], Means: "can go"},
	{Status: fitStatus[domain.FitNotYet], Means: "not there yet"},
	{Status: fitStatus[domain.FitUnknown], Means: "can't see it all"},
	{Status: fitStatus[domain.FitWarn], Means: "rank is past it"},
}

var eventTypes = map[domain.EventKind]string{
	domain.EventCup:    "Tournament",
	domain.EventInvite: "Invite",
	domain.EventTrip:   "Expedition",
	domain.EventRanch:  "Ranch",
	domain.EventShop:   "Shop",
	domain.EventVisit:  "Visitor",
}

func NewCalendarView(p domain.Profile, hidePast bool) CalendarView {
	weeks := domain.Calendar(p)
	if len(weeks) == 0 {
		return CalendarView{None: "The game does not say the date yet"}
	}
	var v CalendarView
	for _, w := range weeks {
		title := strconv.Itoa(w.Date.Year)
		if n := len(v.Years); n == 0 || v.Years[n-1].Title != title {
			v.Years = append(v.Years, CalYearView{Title: title})
		}
		year := &v.Years[len(v.Years)-1]
		name := player.MonthNames[w.Date.Month-1]
		if n := len(year.Months); n == 0 || year.Months[n-1].Title != name {
			year.Months = append(year.Months, CalMonthView{Title: name})
		}
		month := &year.Months[len(year.Months)-1]
		wv := CalWeekView{Title: "wk " + strconv.Itoa(w.Date.Week), Now: w.In == 0}
		for _, e := range w.Events {
			if e.Fit == domain.FitNever || e.Fit == domain.FitWarn && hidePast {
				continue
			}
			wv.Events = append(wv.Events, eventView(e))
		}
		month.Count += len(wv.Events)
		month.Weeks = append(month.Weeks, wv)
	}
	return v
}

func eventView(e domain.Event) CalEventView {
	v := CalEventView{Name: e.Name, Type: eventTypes[e.Kind], Note: e.Need, Detail: e.Detail, Status: fitStatus[e.Fit]}
	switch e.Fit {
	case domain.FitNotYet:
		v.Note = "needs " + e.Need
	case domain.FitUnknown:
		v.Note = cmp.Or(e.Need, "not known") + "?"
	case domain.FitWarn:
		v.Note = ""
	}
	return v
}

func eventTitle(e CalEventView) string {
	if e.Type == eventTypes[domain.EventCup] || e.Type == eventTypes[domain.EventShop] {
		return e.Name
	}
	return e.Type + " · " + e.Name
}

const foldIndent = 16

type CalendarPage struct {
	hidePast *widget.Check
	none     *widget.Label
	years    *fyne.Container
	content  *container.Scroll
	last     *domain.Profile
	shown    *CalendarView
	open     map[string]bool
}

func NewCalendarPage() *CalendarPage {
	c := &CalendarPage{none: Quiet(""), years: container.NewVBox(), open: map[string]bool{}}
	c.hidePast = widget.NewCheck("Hide cups past rank", func(bool) { c.redraw() })
	legend := container.NewGridWithColumns(2)
	for _, l := range CalendarLegend {
		chip := NewToneBox()
		chip.Set(l.Status)
		means := widget.NewLabel(l.Means)
		means.Wrapping = fyne.TextWrapWord
		legend.Add(container.NewBorder(nil, nil, container.NewVBox(chip), nil, means))
	}
	c.content = container.NewVScroll(container.NewVBox(legend, container.NewHBox(layout.NewSpacer(), c.hidePast), widget.NewSeparator(), c.none, c.years))
	return c
}

func (c *CalendarPage) Content() fyne.CanvasObject { return c.content }

func (c *CalendarPage) Set(p *domain.Profile) {
	c.last = p
	c.redraw()
}

func (c *CalendarPage) redraw() {
	v := CalendarView{None: "No game to read"}
	if c.last != nil {
		v = NewCalendarView(*c.last, c.hidePast.Checked)
	}
	if c.shown != nil && reflect.DeepEqual(v, *c.shown) {
		return
	}
	c.shown = &v
	c.none.SetText(v.None)
	setVisible(c.none, v.None != "")
	var objects []fyne.CanvasObject
	for i, y := range v.Years {
		objects = append(objects, c.fold(y.Title, y.Title, true, i == 0, c.months(y, i == 0)))
	}
	c.years.Objects = objects
	c.years.Refresh()
	c.content.Refresh()
}

func (c *CalendarPage) months(y CalYearView, thisYear bool) fyne.CanvasObject {
	var objects []fyne.CanvasObject
	for i, m := range y.Months {
		label := m.Title + " (" + strconv.Itoa(m.Count) + ")"
		objects = append(objects, c.fold(y.Title+" "+m.Title, label, false, thisYear && i < 2, weeksBox(m)))
	}
	return indent(container.NewVBox(objects...))
}

func (c *CalendarPage) fold(key, label string, bold, open bool, body fyne.CanvasObject) fyne.CanvasObject {
	if was, ok := c.open[key]; ok {
		open = was
	}
	box := container.NewVBox()
	head := widget.NewButton("", nil)
	head.Alignment = widget.ButtonAlignLeading
	head.Importance = widget.LowImportance
	set := func(on bool) {
		c.open[key] = on
		mark := "▸ "
		if on {
			mark = "▾ "
		}
		head.SetText(mark + label)
		setVisible(body, on)
		box.Refresh()
		c.years.Refresh()
		c.content.Refresh()
	}
	head.OnTapped = func() { set(!c.open[key]) }
	if bold {
		head.Importance = widget.MediumImportance
	}
	box.Objects = []fyne.CanvasObject{head, body}
	set(open)
	return box
}

func indent(o fyne.CanvasObject) fyne.CanvasObject {
	gap := canvas.NewRectangle(color.Transparent)
	gap.SetMinSize(fyne.NewSize(foldIndent, 0))
	return container.NewBorder(nil, nil, gap, nil, o)
}

func weeksBox(m CalMonthView) fyne.CanvasObject {
	rows := container.NewVBox()
	for _, w := range m.Weeks {
		head := container.NewVBox(widget.NewLabelWithStyle(w.Title, fyne.TextAlignLeading, fyne.TextStyle{Bold: w.Now}))
		if w.Now {
			now := widget.NewLabel("now")
			now.Importance = widget.HighImportance
			head.Add(now)
		}
		events := container.NewVBox()
		for _, e := range w.Events {
			name := widget.NewLabelWithStyle(eventTitle(e), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			name.Wrapping = fyne.TextWrapWord
			chip := NewToneBox()
			chip.Set(e.Status)
			row := container.NewVBox(container.NewBorder(nil, nil, nil, container.NewVBox(chip), name))
			if more := strings.Join(slices.DeleteFunc([]string{e.Detail, e.Note}, func(s string) bool { return s == "" }), " · "); more != "" {
				note := widget.NewLabel(more)
				note.Wrapping = fyne.TextWrapWord
				row.Add(note)
			}
			events.Add(row)
		}
		if len(w.Events) == 0 {
			events.Add(Quiet("Nothing"))
		}
		rows.Add(widget.NewSeparator())
		rows.Add(container.NewBorder(nil, nil, head, nil, events))
	}
	return indent(rows)
}
