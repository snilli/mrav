package plannerui

import (
	"context"
	"fmt"
	"image/color"
	"maps"
	"slices"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"app/internal/core/planner"
	"app/internal/core/planner/domain"
	"app/internal/delivery/ui/profileui"
	"app/internal/delivery/ui/status"
)

const note = "Starts from the monster in the game this week.\nRest is worked out at the least that it can give.\nChanging a week works the whole plan out again."

type Screen struct {
	seen status.Seen[domain.Current]
	svc  planner.PlannerService

	reactions domain.Reactions
	effects   domain.Effects
	tracked
	plan     domain.Plan
	updating bool
	current  domain.Current
	live     domain.Settings
	aside    *tracked
	waiting  bool
	acting   bool

	status *widget.Label
	body   fyne.CanvasObject

	table *fyne.Container
	frame *canvas.Rectangle
	slots []*slot

	lived, weeks, average *widget.Label
	problems              *widget.Label
	effect                *widget.Label
	clear                 *widget.Button
	tracking              *widget.Label

	content fyne.CanvasObject
}

func NewScreen(svc planner.PlannerService) *Screen {
	s := &Screen{
		svc:      svc,
		settings: domain.DefaultSettings(), foods: map[int]domain.Food{}, changed: -1,
		live:     domain.DefaultSettings(),
		status:   widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		lived:    widget.NewLabel(""),
		weeks:    widget.NewLabel(""),
		average:  widget.NewLabel(""),
		problems: widget.NewLabel(""),
		effect:   widget.NewLabel(""),
		tracking: widget.NewLabel(""),
	}
	s.status.Hide()
	s.problems.Wrapping = fyne.TextWrapWord
	s.problems.Hide()
	s.effect.Wrapping = fyne.TextWrapWord
	s.effect.Hide()
	s.tracking.Hide()

	s.table = container.New(layout.NewCustomPaddedVBoxLayout(0), newHeading())
	s.frame = canvas.NewRectangle(color.Transparent)
	s.frame.StrokeColor, s.frame.StrokeWidth = profileui.LineColor, profileui.LineWidth
	s.clear = widget.NewButton("Clear", s.reset)

	summary := container.NewBorder(nil, nil, container.NewVBox(s.lived, s.weeks, s.average), container.NewVBox(s.clear))
	s.body = container.NewVBox(s.tracking, container.NewStack(s.table, s.frame), s.effect, summary, s.problems, widget.NewLabel(note))
	s.content = container.NewPadded(container.NewVBox(s.status, s.body))
	s.recompute()
	return s
}

func (s *Screen) Content() fyne.CanvasObject { return s.content }

func (s *Screen) Load(ctx context.Context) {
	current, err := s.svc.Current(ctx)
	fyne.Do(func() {
		if s.seen.Changed(current, err) {
			s.apply(current, err)
		}
	})
}

func (s *Screen) apply(current domain.Current, err error) {
	if err != nil {
		s.status.SetText(status.Message(err))
		s.status.Show()
		s.body.Hide()
		s.content.Refresh()
		return
	}
	s.status.Hide()
	s.body.Show()
	s.content.Refresh()

	s.acting = domain.Acting(s.acting, s.current, current)
	s.reactions, s.effects, s.current, s.live = current.Reactions, current.Effects, current, current.Settings
	if s.track.Another(current) {
		aside := s.tracked
		s.forget()
		s.aside = &aside
	} else if s.aside != nil && !s.track.Started && current.Dated && current.Who == s.aside.track.Who {
		s.tracked, s.aside = *s.aside, nil
	}
	weeks, ok := s.track.Weeks(current)
	s.waiting = s.track.Started && !ok
	switch {
	case !ok:
	case weeks < 0:
		s.back(current.Date)
		s.follow(current.Date - s.track.Date)
	case weeks > 0:
		s.follow(weeks)
	}
	if !s.acting && (!s.track.Started || ok && s.track.Date == current.Date) {
		s.settings, s.used = current.Settings, current.Settings.ItemUsed
		s.age()
	}
	s.recompute()
}

func (s *Screen) back(date int) {
	for len(s.marks) > 0 && s.track.Date > date {
		m := s.marks[len(s.marks)-1]
		s.marks = s.marks[:len(s.marks)-1]
		s.steps = append(slices.Clone(m.passed), s.steps...)
		foods := maps.Clone(m.foods)
		for month, food := range s.foods {
			foods[month+m.month-1] = food
		}
		s.foods = foods
		if s.changed >= 0 {
			s.changed += len(m.passed)
		}
		s.done, s.settings, s.track.Date = s.done[:m.done], m.settings, m.date
	}
	s.passing, s.used, s.aged = 0, true, true
}

func (s *Screen) follow(weeks int) {
	s.passing = weeks
	if weeks == 0 {
		return
	}
	s.plan = domain.Build(s.reactions, s.effects, s.settings, s.steps, s.foods)
	passed, ok := s.plan.Passed(weeks)
	if !ok {
		return
	}
	m := mark{date: s.track.Date, passed: slices.Clone(s.steps[:min(passed.Step, len(s.steps))]), month: passed.Month, done: len(s.done), settings: s.settings, foods: map[int]domain.Food{}}
	for month, food := range s.foods {
		if month < passed.Month {
			m.foods[month] = food
		}
	}
	s.marks = append(s.marks, m)
	for _, r := range passed.Played(s.used, s.live).Rows {
		s.done = append(s.done, doneRow{row: r, month: s.settings.MonthName(r.Month)})
	}
	s.steps = s.steps[min(passed.Step, len(s.steps)):]
	foods := map[int]domain.Food{}
	for month, food := range s.foods {
		if month >= passed.Month {
			foods[month-passed.Month+1] = food
		}
	}
	s.foods = foods
	s.changed = max(s.changed-passed.Step, -1)
	s.settings, s.track.Date, s.passing, s.aged = s.live, s.current.Date, 0, false
}

func (s *Screen) age() {
	if s.aged || len(s.marks) == 0 {
		return
	}
	start := s.marks[len(s.marks)-1].settings.Lifespan
	for i := len(s.done) - 1; i >= 0; i-- {
		if s.done[i].row.Kind != domain.FoodRow {
			s.done[i].row, s.aged = s.done[i].row.Aged(start, s.live.Lifespan)
			return
		}
	}
}

func (s *Screen) begin() {
	if s.track.Started || !s.current.Dated {
		return
	}
	s.track = domain.Tracking{Started: true, Start: s.current.Date, Date: s.current.Date, Who: s.current.Who}
	s.from = fmt.Sprintf("%s wk %d", s.settings.MonthName(1), s.plan.Rows[0].Week)
	s.aside = nil
}

func (s *Screen) forget() {
	s.tracked = tracked{settings: s.settings, foods: map[int]domain.Food{}, changed: -1}
	s.aside, s.waiting = nil, false
}

func (s *Screen) quietly(do func()) {
	was := s.updating
	s.updating = true
	do()
	s.updating = was
}

func (s *Screen) recompute() {
	s.plan = domain.Build(s.reactions, s.effects, s.settings, s.steps, s.foods)
	s.render()
}

func (s *Screen) reset() {
	s.forget()
	s.settings = s.live
	s.recompute()
}

func (s *Screen) edit(sl *slot) {
	if s.updating {
		return
	}
	s.begin()
	step := domain.Step{Item: itemOf(sl.item.Selected), Action: actionOf(sl.action.Selected)}
	s.steps = domain.SetStep(s.steps, sl.row.Step, step)
	s.changed = sl.row.Step
	if step == (domain.Step{}) {
		s.changed = -1
	}
	s.recompute()
}

func (s *Screen) eat(sl *slot) {
	if s.updating {
		return
	}
	s.begin()
	s.foods[sl.row.Month] = foodOf(sl.food.Selected)
	s.recompute()
}

func (s *Screen) render() {
	s.quietly(func() {
		used := 0
		next := func() *slot {
			if used == len(s.slots) {
				sl := newSlot(s.edit, s.eat)
				s.slots = append(s.slots, sl)
				s.table.Add(sl.object)
			}
			sl := s.slots[used]
			used++
			sl.object.Show()
			return sl
		}
		heading := ""
		week := func(r domain.Row, month string, done bool) {
			if month != heading {
				heading = month
				s.fillHeader(next(), month)
			}
			s.fillWeek(next(), r, done)
		}
		for _, d := range s.done {
			week(d.row, d.month, true)
		}
		passing := s.passing
		for _, r := range s.plan.Rows {
			done := passing > 0
			if done && r.Kind != domain.FoodRow {
				passing--
			}
			week(r, s.settings.MonthName(r.Month), done)
		}
		for _, sl := range s.slots[used:] {
			sl.object.Hide()
		}
	})

	s.tracking.Hide()
	if s.track.Started {
		weeks := s.track.Date - s.track.Start + s.passing
		unit := "weeks"
		if weeks == 1 {
			unit = "week"
		}
		text := fmt.Sprintf("Plan from %s · %d %s done", s.from, weeks, unit)
		if s.waiting {
			text += " · waiting for " + s.track.Who.Name
		}
		s.tracking.SetText(text)
		s.tracking.Show()
	}

	t := s.plan.Totals
	s.lived.SetText(fmt.Sprintf("Weeks lived: %d", t.Lived))
	if t.Weeks > 0 {
		s.weeks.SetText(fmt.Sprintf("Weeks gained: %d", t.Weeks))
	} else {
		s.weeks.SetText(fmt.Sprintf("Weeks lost: %d", -t.Weeks))
	}
	average := fmt.Sprintf("%.1f", t.Average)
	if t.Average > 0 {
		average = "+" + average
	}
	s.average.SetText("Average: " + average + " wk/month")

	var problems []string
	for _, p := range s.plan.Problems {
		for _, r := range s.plan.Rows {
			if r.Step == p.Step && r.Kind == domain.StepRow {
				problems = append(problems, fmt.Sprintf("%s, week %d: %s", s.settings.MonthName(r.Month), r.Week, p.Reason.Text()))
				break
			}
		}
	}
	s.effect.Hide()
	for _, r := range s.plan.Rows {
		if r.Kind == domain.StepRow && r.Step == s.changed && r.Item != domain.NoItem {
			s.effect.SetText(fmt.Sprintf("%s, week %d, %s: %s", s.settings.MonthName(r.Month), r.Week, r.Item.Name(), s.effects.Text(r.Item, s.settings)))
			s.effect.Show()
		}
	}
	s.problems.SetText(strings.Join(problems, "\n"))
	if len(problems) > 0 {
		s.problems.Show()
	} else {
		s.problems.Hide()
	}
	s.table.Refresh()
	s.body.Refresh()
}

type doneRow struct {
	row   domain.Row
	month string
}

type tracked struct {
	settings domain.Settings
	steps    []domain.Step
	foods    map[int]domain.Food
	changed  int
	track    domain.Tracking
	done     []doneRow
	passing  int
	from     string
	marks    []mark
	used     bool
	aged     bool
}

type mark struct {
	date     int
	passed   []domain.Step
	month    int
	done     int
	settings domain.Settings
	foods    map[int]domain.Food
}
