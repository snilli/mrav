package plannerui

import (
	"fmt"
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"app/internal/core/planner/domain"
	"app/internal/delivery/ui/profileui"
)

const (
	noItem   = "No item"
	noAction = "No action"

	noPlaceHolder = " "
)

var (
	columns                   = profileui.Columns{32, 0, 0, 28, 28, 28, 30, 28, 32, 28}
	gainColor     color.Color = color.NRGBA{R: 0x8c, G: 0xcf, B: 0x95, A: 0xff}
	monthFill     color.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 14}
	columnHeading             = [...]string{"Wk", "Item", "Action", "Ftg", "Str", "LI", "Form", "Lost", "Total", "Life"}
	foodNames                 = names(domain.Foods())
)

type slot struct {
	object fyne.CanvasObject
	header bool
	row    domain.Row

	head    *canvas.Text
	headBox *fyne.Container
	rule    *fyne.Container
	grid    *fyne.Container
	cells   *fyne.Container

	week               *widget.Label
	item, action, food *widget.Select
	away               *widget.Label
	ftg, str, li, lost *canvas.Text
	total, life, form  *canvas.Text
}

func newSlot(edit, eat func(*slot)) *slot {
	sl := &slot{
		head:   canvas.NewText("", theme.Color(theme.ColorNameForeground)),
		week:   widget.NewLabel(""),
		item:   widget.NewSelect(nil, nil),
		action: widget.NewSelect(nil, nil),
		food:   widget.NewSelect(foodNames, nil),
		away:   widget.NewLabel(""),
	}
	sl.head.TextStyle = fyne.TextStyle{Bold: true}
	for _, sel := range []*widget.Select{sl.item, sl.action, sl.food} {
		sel.PlaceHolder = noPlaceHolder
	}
	sl.item.OnChanged = func(string) { edit(sl) }
	sl.action.OnChanged = func(string) { edit(sl) }
	sl.food.OnChanged = func(string) { eat(sl) }

	number := func() *canvas.Text { return canvas.NewText("", theme.Color(theme.ColorNameForeground)) }
	sl.ftg, sl.str, sl.li, sl.lost = number(), number(), number(), number()
	sl.total, sl.life, sl.form = number(), number(), number()

	sl.cells = container.New(columns,
		sl.week, container.NewStack(sl.item, sl.food), container.NewStack(sl.action, sl.away),
		container.NewCenter(sl.ftg), container.NewCenter(sl.str), container.NewCenter(sl.li), container.NewCenter(sl.form), container.NewCenter(sl.lost),
		container.NewCenter(sl.total), container.NewCenter(sl.life),
	)
	sl.grid = profileui.NewGrid(columns)
	sl.rule = profileui.NewGrid(profileui.Columns{0})
	sl.headBox = container.NewStack(canvas.NewRectangle(monthFill), container.New(layout.NewCustomPaddedLayout(2, 2, 8, 0), sl.head))
	sl.object = container.NewStack(sl.grid, sl.cells, sl.rule, sl.headBox)
	return sl
}

func newHeading() fyne.CanvasObject {
	cells := make([]fyne.CanvasObject, len(columnHeading))
	for i, title := range columnHeading {
		text := canvas.NewText(title, theme.Color(theme.ColorNameForeground))
		text.TextStyle = fyne.TextStyle{Bold: true}
		cells[i] = container.NewCenter(text)
	}
	return container.NewStack(profileui.NewGrid(columns), container.New(columns, cells...))
}

func (sl *slot) setHeader(on bool) {
	sl.header = on
	for _, o := range []fyne.CanvasObject{sl.headBox, sl.rule} {
		setVisible(o, on)
	}
	for _, o := range []fyne.CanvasObject{sl.cells, sl.grid} {
		setVisible(o, !on)
	}
}

func setVisible(o fyne.CanvasObject, on bool) {
	if on {
		o.Show()
	} else {
		o.Hide()
	}
}

func names[T interface{ Name() string }](list []T) []string {
	out := make([]string, len(list))
	for i, v := range list {
		out[i] = v.Name()
	}
	return out
}

func (s *Screen) fillHeader(sl *slot, month int) {
	sl.setHeader(true)
	sl.head.Text = fmt.Sprintf("Month %d", month)
	sl.head.Refresh()
}

func (s *Screen) fillWeek(sl *slot, r domain.Row) {
	sl.setHeader(false)
	sl.row = r

	label := strconv.Itoa(r.Week) + "wk"
	if r.Kind == domain.FoodRow {
		label = "food"
	}
	sl.week.SetText(label)

	switch r.Kind {
	case domain.FoodRow:
		setVisible(sl.item, false)
		setVisible(sl.action, false)
		setVisible(sl.food, true)
		setVisible(sl.away, true)
		sl.away.SetText("Monthly food")
		sl.food.SetSelected(r.Food.Name())
	case domain.FollowRow:
		setVisible(sl.item, false)
		setVisible(sl.food, false)
		setVisible(sl.action, false)
		setVisible(sl.away, true)
		sl.away.SetText(r.Action.Name())
	default:
		setVisible(sl.item, true)
		setVisible(sl.action, true)
		setVisible(sl.food, false)
		setVisible(sl.away, false)

		var step domain.Step
		if r.Step < len(s.steps) {
			step = s.steps[r.Step]
		}
		sl.item.Options = names(domain.ItemsFor(s.settings, s.steps, r.Step))
		sl.item.SetSelected(step.Item.Name())
		options := []string{noAction}
		for _, a := range domain.Actions() {
			if r.Offers(a) || a == step.Action {
				options = append(options, a.Name())
			}
		}
		sl.action.Options = options
		sl.action.SetSelected(actionName(step.Action))
		sl.item.Refresh()
		sl.action.Refresh()
	}

	foreground := theme.Color(theme.ColorNameForeground)
	set := func(t *canvas.Text, text string, c color.Color) {
		t.Text, t.Color = text, c
		t.Refresh()
	}
	number := func(shown bool, n int) string {
		if !shown {
			return ""
		}
		return strconv.Itoa(n)
	}
	set(sl.ftg, number(r.Shown, r.Fatigue), foreground)
	set(sl.str, number(r.Shown, r.Stress), foreground)
	set(sl.li, number(r.ShowLI, r.LI), foreground)
	set(sl.form, number(r.Shown, r.Form), foreground)

	signed := func(shown bool, n int) (string, color.Color) {
		switch {
		case !shown:
			return "", foreground
		case n > 0:
			return "+" + strconv.Itoa(n), gainColor
		case n < 0:
			return strconv.Itoa(n), theme.Color(theme.ColorNameError)
		}
		return "0", foreground
	}
	lost, colour := signed(r.ShowLost, r.Lost)
	set(sl.lost, lost, colour)
	total, colour := signed(r.ShowLife, r.Total)
	set(sl.total, total, colour)
	colour = foreground
	if r.Life < 0 {
		colour = theme.Color(theme.ColorNameError)
	}
	set(sl.life, number(r.ShowLife, r.Life), colour)
}

func actionName(a domain.Action) string {
	if a == domain.NoAction {
		return noAction
	}
	return a.Name()
}

func itemOf(name string) domain.Item {
	for _, it := range domain.Items() {
		if it.Name() == name {
			return it
		}
	}
	return domain.NoItem
}

func actionOf(name string) domain.Action {
	for _, a := range domain.Actions() {
		if a.Name() == name {
			return a
		}
	}
	return domain.NoAction
}

func foodOf(name string) domain.Food {
	for _, f := range domain.Foods() {
		if f.Name() == name {
			return f
		}
	}
	return domain.NoFood
}
