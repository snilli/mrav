// Package demoui is the demo panel: it sets the monster and the player of the fake sources by hand, or to
// random values, so that the viewer can be tried without a game running.
package demoui

import (
	"context"
	"math/rand/v2"
	"slices"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	monster "app/internal/core/monster/domain"
	"app/internal/core/move/domain"
	player "app/internal/core/player/domain"
	profiledomain "app/internal/core/profile/domain"
	"app/internal/infrastructure/fakemonster"
	"app/internal/infrastructure/fakeplayer"
)

var statNames = [monster.StatCount]string{"Life", "Pow", "Int", "Skill", "Speed", "Def"}

type Panel struct {
	monsters  *fakemonster.Source
	players   *fakeplayer.Source
	genera    []domain.Genus
	usesValue int
	rng       *rand.Rand

	genus                       *widget.Select
	tripX, tripY, steps, energy *widget.Entry  // where the monster is on a trip's map, how far it has walked, its energy left
	rank                        *widget.Select // the grade of the monster, E to S
	month                       *widget.Select // the month and the week of the ranch's date, which the trips and the Route page go by
	week                        *widget.Select
	stats                       [monster.StatCount]*widget.Entry
	nature                      *widget.Entry
	subGenus                    *widget.Entry
	uses                        *widget.Entry
	learnAll                    *widget.Button
	forgetAll                   *widget.Button
	random                      *widget.Button

	growth                                                                             [monster.StatCount]*widget.Entry
	age, lifespan, initial, fatigue, stress, fame, form, spoil, fear, lifeStage, money *widget.Entry
	content                                                                            fyne.CanvasObject
}

func New(monsters *fakemonster.Source, players *fakeplayer.Source, genera []domain.Genus) *Panel {
	p := &Panel{
		monsters: monsters, players: players, genera: genera,
		rng: rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0)),
	}

	names := make([]string, len(genera))
	for i, g := range genera {
		names[i] = g.Name
	}
	p.genus = widget.NewSelect(names, p.selectGenus)

	items := []*widget.FormItem{widget.NewFormItem("Genus", p.genus)}
	for stat := range p.stats {
		p.stats[stat] = p.numberEntry(func(m *monster.Monster, n int) { m.Stats[stat] = n })
		items = append(items, widget.NewFormItem(statNames[stat], p.stats[stat]))
	}
	p.nature = p.numberEntry(func(m *monster.Monster, n int) { m.Nature = n })
	p.subGenus = p.numberEntry(func(m *monster.Monster, n int) { m.SubGenus = monster.Genus(n) })
	p.uses = p.numberEntry(nil)
	p.uses.OnChanged = func(text string) {
		if n, ok := number(text); ok {
			p.usesValue = n
		}
	}
	ranks := make([]string, 6)
	for i := range ranks {
		ranks[i] = profiledomain.RankLetter(i)
	}
	p.rank = widget.NewSelect(ranks, func(letter string) {
		if i := slices.Index(ranks, letter); i >= 0 {
			p.monsters.Update(func(m *monster.Monster) { m.Rank = i })
		}
	})
	months := player.MonthNames[:]
	p.month = widget.NewSelect(months, func(name string) {
		if i := slices.Index(months, name); i >= 0 { // the week is left as it is
			p.players.Update(func(pl *player.Player) { pl.Ranch = pl.Ranch.InWeek(i+1, pl.Ranch.Calendar().Week) })
		}
	})
	weeks := []string{"1", "2", "3", "4"}
	p.week = widget.NewSelect(weeks, func(name string) {
		if i := slices.Index(weeks, name); i >= 0 { // and the month
			p.players.Update(func(pl *player.Player) { pl.Ranch = pl.Ranch.InWeek(pl.Ranch.Calendar().Month, i+1) })
		}
	})
	p.tripX = p.playerEntry(func(pl *player.Player, n int) {
		_, y := pl.Ranch.Walked()
		pl.Ranch = pl.Ranch.WalkedTo(n, y)
	})
	p.tripY = p.playerEntry(func(pl *player.Player, n int) {
		x, _ := pl.Ranch.Walked()
		pl.Ranch = pl.Ranch.WalkedTo(x, n)
	})
	p.steps = p.playerEntry(func(pl *player.Player, n int) { pl.Ranch.Steps = n })
	p.energy = p.playerEntry(func(pl *player.Player, n int) { pl.Ranch.Energy, pl.Ranch.EnergyKnown = n, true })
	items = append(items,
		widget.NewFormItem("Nature", p.nature),
		widget.NewFormItem("Sub-genus", p.subGenus),
		widget.NewFormItem("Rank", p.rank),
		widget.NewFormItem("Month and week", container.NewGridWithColumns(2, p.month, p.week)),
		widget.NewFormItem("On a trip: right, up", container.NewGridWithColumns(2, p.tripX, p.tripY)),
		widget.NewFormItem("Steps walked", p.steps),
		widget.NewFormItem("Energy left", p.energy),
		widget.NewFormItem("Uses of each move", p.uses),
	)

	more := []*widget.FormItem{}
	field := func(label string, apply func(m *monster.Monster, n int)) *widget.Entry {
		e := p.numberEntry(apply)
		more = append(more, widget.NewFormItem(label, e))
		return e
	}
	p.age = field("Age (weeks)", func(m *monster.Monster, n int) { m.Age = n })
	p.lifespan = field("Lifespan", func(m *monster.Monster, n int) { m.Lifespan = n })
	p.initial = field("Initial lifespan", func(m *monster.Monster, n int) { m.InitialLifespan = n })
	p.fatigue = field("Fatigue", func(m *monster.Monster, n int) { m.Fatigue = n })
	p.stress = field("Stress", func(m *monster.Monster, n int) { m.Stress = n })
	p.fame = field("Fame", func(m *monster.Monster, n int) { m.Fame = n })
	p.form = field("Form", func(m *monster.Monster, n int) { m.Form = n })
	p.spoil = field("Spoil", func(m *monster.Monster, n int) { m.Spoil = n })
	p.fear = field("Fear", func(m *monster.Monster, n int) { m.Fear = n })
	p.lifeStage = field("Life stage (0-9)", func(m *monster.Monster, n int) { m.LifeStage = n })
	for stat := range p.growth {
		p.growth[stat] = field(statNames[stat]+" growth (0-4)", func(m *monster.Monster, n int) { m.Growth[stat] = n })
	}
	p.money = p.playerEntry(func(pl *player.Player, n int) { pl.Money = n })
	more = append(more, widget.NewFormItem("Player money", p.money))

	p.learnAll = widget.NewButton("Learn all moves", p.learnEveryMove)
	p.forgetAll = widget.NewButton("Forget all moves", p.forgetEveryMove)
	p.random = widget.NewButton("Randomize everything", p.Randomize)

	p.content = container.NewVScroll(container.NewVBox(
		widget.NewLabelWithStyle("Demo monster", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		p.random,
		widget.NewForm(items...),
		container.NewHBox(p.learnAll, p.forgetAll),
		widget.NewAccordion(widget.NewAccordionItem("More numbers", widget.NewForm(more...))),
	))
	if len(names) > 0 {
		p.genus.SetSelected(names[0])
	}
	return p
}

// Content is what to put in a window.
func (p *Panel) Content() fyne.CanvasObject { return p.content }

// number reads what was typed: empty is zero, and text that is not a whole number is not accepted.
func number(text string) (int, bool) {
	if text == "" {
		return 0, true
	}
	n, err := strconv.Atoi(text)
	return n, err == nil
}

// numberEntry is an entry that applies each number typed to the monster; text that is not a number
// leaves the monster as it was.
func (p *Panel) numberEntry(apply func(m *monster.Monster, n int)) *widget.Entry {
	e := widget.NewEntry()
	e.SetPlaceHolder("0")
	e.OnChanged = func(text string) {
		if n, ok := number(text); ok && apply != nil {
			p.monsters.Update(func(m *monster.Monster) { apply(m, n) })
		}
	}
	return e
}

// playerEntry is numberEntry for the player.
func (p *Panel) playerEntry(apply func(pl *player.Player, n int)) *widget.Entry {
	e := widget.NewEntry()
	e.SetPlaceHolder("0")
	e.OnChanged = func(text string) {
		if n, ok := number(text); ok {
			p.players.Update(func(pl *player.Player) { apply(pl, n) })
		}
	}
	return e
}

func (p *Panel) selectGenus(name string) {
	for _, g := range p.genera {
		if g.Name == name {
			p.monsters.Update(func(m *monster.Monster) { m.Genus = g.Code })
			return
		}
	}
}

// learnEveryMove teaches all 24 slots, with the typed use count, and makes the first slot of each
// column the active move.
func (p *Panel) learnEveryMove() {
	p.monsters.Update(func(m *monster.Monster) {
		for slot := range m.Learned {
			m.Learned[slot], m.Used[slot] = true, p.usesValue
		}
		for column := range m.Active {
			m.Active[column] = column * domain.Rows
		}
	})
}

func (p *Panel) forgetEveryMove() {
	p.monsters.Update(func(m *monster.Monster) {
		m.Learned = [monster.SlotCount]bool{}
		m.Used = [monster.SlotCount]int{}
		for column := range m.Active {
			m.Active[column] = monster.NoActiveMove
		}
	})
}

// Randomize gives the monster and the player random values, and shows them in the boxes. It is what the
// button does, and what the demo starts with, so that its pages have something to show.
func (p *Panel) Randomize() {
	ctx := context.Background()
	m, _ := p.monsters.Current(ctx)
	pl, _ := p.players.Player(ctx)
	randomize(p.rng, p.genera, &m, &pl)
	p.monsters.Update(func(dst *monster.Monster) { *dst = m })
	p.players.Update(func(dst *player.Player) { *dst = pl })
	p.sync(m, pl)
}

// sync shows a monster and a player in the boxes. Setting a box applies its number again, which changes
// nothing.
func (p *Panel) sync(m monster.Monster, pl player.Player) {
	show := func(e *widget.Entry, n int) { e.SetText(strconv.Itoa(n)) }
	for _, g := range p.genera {
		if g.Code == m.Genus {
			p.genus.SetSelected(g.Name)
		}
	}
	for stat := range p.stats {
		show(p.stats[stat], m.Stats[stat])
		show(p.growth[stat], m.Growth[stat])
	}
	show(p.nature, m.Nature)
	show(p.subGenus, int(m.SubGenus))
	show(p.age, m.Age)
	show(p.lifespan, m.Lifespan)
	show(p.initial, m.InitialLifespan)
	show(p.fatigue, m.Fatigue)
	show(p.stress, m.Stress)
	show(p.fame, m.Fame)
	show(p.form, m.Form)
	show(p.spoil, m.Spoil)
	show(p.fear, m.Fear)
	show(p.lifeStage, m.LifeStage)
	show(p.money, pl.Money)
	x, y := pl.Ranch.Walked()
	show(p.tripX, x)
	show(p.tripY, y)
	show(p.steps, pl.Ranch.Steps)
	show(p.energy, pl.Ranch.Energy)
	cal := pl.Ranch.Calendar()
	p.rank.SetSelected(profiledomain.RankLetter(m.Rank))
	p.month.SetSelected(player.MonthNames[cal.Month-1])
	p.week.SetSelected(strconv.Itoa(cal.Week))
}
