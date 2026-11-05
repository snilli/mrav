package drillui

import (
	"cmp"
	"context"
	"fmt"
	"image/color"
	"slices"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"app/internal/core/drill"
	"app/internal/core/drill/domain"
	monster "app/internal/core/monster/domain"
	"app/internal/delivery/ui/profileui"
	"app/internal/delivery/ui/status"
)

const (
	noEffect   = "N/A"
	bonusMark  = "+1"
	legendText = "specialty: its numbers are the ones with the bonus"
	note       = "Mot is the monster's motivation for the drill\nChance is how likely the chosen result is"

	firstPlace  = 9
	firstResult = 1
)

var (
	recommendedFill color.Color = color.NRGBA{R: 50, G: 205, B: 50, A: 70}
	mainColor       color.Color = color.NRGBA{R: 0x8c, G: 0xcf, B: 0x95, A: 0xff}
)

var (
	columns     = profileui.Columns{104, 60, 56, 0, 0, 0, 0, 0, 0}
	resultNames = []string{"Cheat", "Success", "Great"}
	results     = []domain.Result{domain.Cheat, domain.Success, domain.Great}
)

type row struct {
	name   *widget.Label
	bonus  *fyne.Container
	tag    *canvas.Rectangle
	mark   *canvas.Text
	title  *fyne.Container
	keen   *widget.Label
	chance *widget.Label
	cells  [monster.StatCount]*canvas.Text
	bg     *canvas.Rectangle
	object fyne.CanvasObject
}

func newTag() (*fyne.Container, *canvas.Rectangle, *canvas.Text) {
	tag := canvas.NewRectangle(color.White)
	tag.CornerRadius = 3
	mark := canvas.NewText(bonusMark, color.Black)
	mark.TextStyle = fyne.TextStyle{Bold: true}
	return container.NewStack(tag, container.New(layout.NewCustomPaddedLayout(0, 0, 2, 2), mark)), tag, mark
}

type legend struct {
	tag  *canvas.Rectangle
	mark *canvas.Text
	text *widget.Label
}

func newRow(name string) *row {
	r := &row{
		name:   widget.NewLabelWithStyle(name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		keen:   widget.NewLabelWithStyle("", fyne.TextAlignTrailing, fyne.TextStyle{}),
		chance: widget.NewLabelWithStyle("", fyne.TextAlignTrailing, fyne.TextStyle{}),
		bg:     canvas.NewRectangle(color.Transparent),
	}
	r.bonus, r.tag, r.mark = newTag()
	r.bonus.Hide()
	r.title = container.New(layout.NewCustomPaddedHBoxLayout(2), r.name, container.NewCenter(r.bonus))
	objects := []fyne.CanvasObject{r.title, r.keen, r.chance}
	for stat := range r.cells {
		r.cells[stat] = canvas.NewText("", theme.Color(theme.ColorNameForeground))
		objects = append(objects, container.NewCenter(r.cells[stat]))
	}
	r.object = container.NewStack(r.bg, profileui.NewGrid(columns), container.New(columns, objects...))
	return r
}

func (r *row) show(g domain.DrillGains, result int) {
	r.name.SetText(g.Drill.Name)
	r.bonus.Hidden = !g.Specialty
	r.title.Refresh()
	r.keen.Importance = widget.MediumImportance
	if profileui.MotivationTone(g.Motivation) == profileui.ToneOrangeRed {
		r.keen.Importance = widget.DangerImportance
	}
	r.keen.SetText(strconv.Itoa(g.Motivation) + "%")
	r.chance.SetText(strconv.FormatFloat(g.Chances.Of(results[result]), 'f', -1, 64) + "%")
	r.bg.FillColor = color.Transparent
	if g.Recommended {
		r.bg.FillColor = recommendedFill
	}
	r.bg.Refresh()

	for _, c := range r.cells {
		c.Text, c.Color = "", theme.Color(theme.ColorNameForeground)
	}
	var effects []domain.Effect
	if result < len(g.Outcomes) {
		effects = g.Outcomes[result].Effects
	}
	if len(effects) == 0 {
		r.cells[g.Drill.Primary].Text = noEffect
	}
	for _, e := range effects {
		c := r.cells[e.Stat]
		c.Text = rangeText(e.Low, e.High)
		switch {
		case e.High < 0:
			c.Color = theme.Color(theme.ColorNameError)
		case e.Stat == g.Drill.Primary:
			c.Color = mainColor
		}
	}
	for _, c := range r.cells {
		c.Refresh()
	}
}

type Screen struct {
	seen status.Seen[domain.Training]
	svc  drill.DrillService

	training domain.Training
	order    order
	onChance bool
	byChance *widget.Button
	status   *widget.Label
	result   *widget.RadioGroup
	sort     *widget.Button
	heads    [monster.StatCount]*canvas.Text
	frame    *canvas.Rectangle
	rows     []*row
	table    *fyne.Container
	legend   legend
	note     *widget.Label
	places   *widget.Label
	body     fyne.CanvasObject
	content  fyne.CanvasObject
}

func NewScreen(svc drill.DrillService) *Screen {
	s := &Screen{
		svc:    svc,
		status: widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		note:   widget.NewLabel(note),
		places: widget.NewLabel("Good places: none"),
	}
	s.status.Hide()
	badge, tag, mark := newTag()
	s.legend = legend{tag: tag, mark: mark, text: widget.NewLabel(legendText)}
	key := container.New(layout.NewCustomPaddedLayout(0, 0, theme.InnerPadding(), 0),
		container.New(layout.NewCustomPaddedHBoxLayout(2), container.NewCenter(badge), s.legend.text))

	s.sort = widget.NewButton("Mot", func() { s.cycle(false) })
	s.sort.Importance = widget.LowImportance
	s.byChance = widget.NewButton("Chance", func() { s.cycle(true) })
	s.byChance.Importance = widget.LowImportance
	head := []fyne.CanvasObject{widget.NewLabelWithStyle("Drill", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), s.sort, s.byChance}
	for stat := range s.heads {
		s.heads[stat] = canvas.NewText(profileui.StatNames[stat], profileui.StatColors[stat])
		s.heads[stat].TextStyle = fyne.TextStyle{Bold: true}
		head = append(head, container.NewCenter(s.heads[stat]))
	}
	s.table = container.New(layout.NewCustomPaddedVBoxLayout(0), container.NewStack(profileui.NewGrid(columns), container.New(columns, head...)))
	for _, d := range domain.Drills() {
		r := newRow(d.Name)
		s.rows = append(s.rows, r)
		s.table.Add(r.object)
	}

	s.result = widget.NewRadioGroup(resultNames, func(string) { s.render() })
	s.result.Horizontal, s.result.Required = true, true
	s.result.SetSelected(resultNames[firstResult])

	s.frame = canvas.NewRectangle(color.Transparent)
	s.frame.StrokeColor, s.frame.StrokeWidth = profileui.LineColor, profileui.LineWidth
	s.body = container.NewVBox(s.result, container.NewStack(s.table, s.frame), key, s.note, s.places)
	s.content = container.NewPadded(container.NewVBox(s.status, s.body))
	return s
}

func (s *Screen) Content() fyne.CanvasObject { return s.content }

func (s *Screen) Load(ctx context.Context) {
	t, err := s.svc.Current(ctx)
	fyne.Do(func() {
		if s.seen.Changed(t, err) {
			s.apply(t, err)
		}
	})
}

func (s *Screen) apply(t domain.Training, err error) {
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

	s.training = t
	s.render()

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

type order int

const (
	pageOrder order = iota
	highFirst
	lowFirst
)

func (o order) icon() fyne.Resource {
	switch o {
	case highFirst:
		return theme.MenuDropDownIcon()
	case lowFirst:
		return theme.MenuDropUpIcon()
	}
	return nil
}

func (s *Screen) cycle(chance bool) {
	if chance != s.onChance {
		s.onChance, s.order = chance, pageOrder
	}
	s.order = (s.order + 1) % (lowFirst + 1)
	s.sort.SetIcon(nil)
	s.byChance.SetIcon(nil)
	if s.onChance {
		s.byChance.SetIcon(s.order.icon())
	} else {
		s.sort.SetIcon(s.order.icon())
	}
	s.render()
}

func (s *Screen) render() {
	result := max(slices.Index(resultNames, s.result.Selected), 0)
	drills := slices.Clone(s.training.Drills)
	key := func(g domain.DrillGains) float64 {
		if s.onChance {
			return g.Chances.Of(results[result])
		}
		return float64(g.Motivation)
	}
	switch s.order {
	case highFirst:
		slices.SortStableFunc(drills, func(a, b domain.DrillGains) int { return cmp.Compare(key(b), key(a)) })
	case lowFirst:
		slices.SortStableFunc(drills, func(a, b domain.DrillGains) int { return cmp.Compare(key(a), key(b)) })
	}
	for i, g := range drills[:min(len(drills), len(s.rows))] {
		s.rows[i].show(g, result)
	}
	s.table.Refresh()
}

func rangeText(low, high int) string {
	if low == high {
		return strconv.Itoa(low)
	}
	return fmt.Sprintf("%d~%d", low, high)
}
