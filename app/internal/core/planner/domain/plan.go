package domain

import "math"

type Kind int

const (
	StepRow Kind = iota
	FollowRow
	FoodRow
	NextRow
)

type Row struct {
	Month, Week int
	Kind        Kind
	Step        int
	Item        Item
	Action      Action
	Food        Food
	Shown       bool
	Fatigue     int
	Stress      int
	LI          int
	ShowLI      bool
	Lost        int
	ShowLost    bool
	Total       int
	Life        int
	ShowLife    bool
	Form        int
}

func (r Row) Offers(a Action) bool { return a != Expedition || r.Week == 1 }

type Reason int

const (
	PeachUsed Reason = iota
	ExpeditionWeek
	PeachEaten
)

func (r Reason) Text() string {
	switch r {
	case PeachUsed:
		return "that peach was given earlier in the plan, so it is left out"
	case PeachEaten:
		return "the monster has eaten that peach already, so it is left out"
	}
	return "an expedition has to begin on the first week of a month, so it is left out"
}

type Problem struct {
	Step   int
	Reason Reason
}

type Totals struct {
	Lived   int
	Weeks   int
	Average float64
}

func TotalsOf(weeks, lived int) Totals {
	t := Totals{Lived: lived, Weeks: weeks}
	if lived > 0 {
		t.Average = math.Floor(float64(weeks)/float64(lived)*40+0.5) / 10
	}
	return t
}

type Plan struct {
	Rows     []Row
	Problems []Problem
	Totals   Totals
}

type builder struct {
	settings Settings
	rows     []Row
	problems []Problem
}

func (b *builder) cur() *Row { return &b.rows[len(b.rows)-1] }

func (b *builder) advance(skipFood bool) {
	last := b.cur()
	month, week := last.Month, last.Week+1
	if last.Week == 4 {
		month, week = last.Month+1, 1
		if !skipFood {
			b.rows = append(b.rows, Row{Month: month, Kind: FoodRow, Step: -1})
		}
	}
	b.rows = append(b.rows, Row{Month: month, Week: week, Kind: NextRow})
}

func (b *builder) follow(step int, action Action, skipFood bool) {
	b.advance(skipFood)
	c := b.cur()
	c.Kind, c.Step, c.Action = FollowRow, step, action
}

func (b *builder) has(item Item) bool {
	for _, r := range b.rows {
		if r.Item == item {
			return true
		}
	}
	return false
}

func (b *builder) apply(at int, step Step) {
	c := b.cur()
	c.Kind, c.Step = StepRow, at
	switch {
	case b.settings.ate(step.Item):
		b.problems = append(b.problems, Problem{Step: at, Reason: PeachEaten})
	case step.Item.isPeach() && b.has(step.Item):
		b.problems = append(b.problems, Problem{Step: at, Reason: PeachUsed})
	default:
		c.Item = step.Item
	}

	switch step.Action {
	case Errantry:
		c.Action = Errantry
		for range 3 {
			b.follow(at, Errantry, true)
		}
		b.advance(false)
	case Expedition:
		if c.Week != 1 {
			b.problems = append(b.problems, Problem{Step: at, Reason: ExpeditionWeek})
			c.Item = NoItem
			b.advance(false)
			return
		}
		c.Action = Expedition
		for range 3 {
			b.follow(at, Away, false)
		}
		b.advance(false)
	case Freeze:
		c.Action, c.Item = Freeze, NoItem
		inLastWeek := c.Week == 4
		b.advance(inLastWeek)
	default:
		c.Action = step.Action
		b.advance(false)
	}
}

func Build(reactions Reactions, effects Effects, settings Settings, steps []Step, foods map[int]Food) Plan {
	b := &builder{settings: settings, rows: []Row{{Month: 1, Week: 1, Kind: NextRow}}}
	for at, step := range steps {
		if !step.blank() {
			b.apply(at, step)
		}
	}
	b.cur().Step = len(steps)
	for i := range b.rows {
		if b.rows[i].Kind == FoodRow {
			if food := foods[b.rows[i].Month]; food >= Potato && food <= Tablet {
				b.rows[i].Food = food
			}
		}
	}
	weeks, lived := simulate(b.rows, reactions, effects, settings)
	return Plan{Rows: b.rows, Problems: b.problems, Totals: TotalsOf(weeks, lived)}
}
