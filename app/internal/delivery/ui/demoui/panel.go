// Package demoui is a temporary panel that sets the monster of a fake source by hand, so the viewer can
// be tried before the game's memory can be read. It goes when the real memory adapter lands.
package demoui

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	monster "app/internal/core/monster/domain"
	"app/internal/core/move/domain"
	"app/internal/infrastructure/fakemonster"
)

var statNames = [monster.StatCount]string{"Life", "Pow", "Int", "Skill", "Speed", "Def"}

type Panel struct {
	src       *fakemonster.Source
	genera    []domain.Genus
	usesValue int

	genus     *widget.Select
	stats     [monster.StatCount]*widget.Entry
	nature    *widget.Entry
	subGenus  *widget.Entry
	uses      *widget.Entry
	learnAll  *widget.Button
	forgetAll *widget.Button
	content   fyne.CanvasObject
}

func New(src *fakemonster.Source, genera []domain.Genus) *Panel {
	p := &Panel{src: src, genera: genera}

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
	items = append(items,
		widget.NewFormItem("Nature", p.nature),
		widget.NewFormItem("Sub-genus", p.subGenus),
		widget.NewFormItem("Uses of each move", p.uses),
	)
	p.learnAll = widget.NewButton("Learn all moves", p.learnEveryMove)
	p.forgetAll = widget.NewButton("Forget all moves", p.forgetEveryMove)

	p.content = container.NewVBox(
		widget.NewLabelWithStyle("Demo monster", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewForm(items...),
		container.NewHBox(p.learnAll, p.forgetAll),
	)
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
			p.src.Update(func(m *monster.Monster) { apply(m, n) })
		}
	}
	return e
}

func (p *Panel) selectGenus(name string) {
	for _, g := range p.genera {
		if g.Name == name {
			p.src.Update(func(m *monster.Monster) { m.Genus = g.Code })
			return
		}
	}
}

// learnEveryMove teaches all 24 slots, with the typed use count, and makes the first slot of each
// column the active move.
func (p *Panel) learnEveryMove() {
	p.src.Update(func(m *monster.Monster) {
		for slot := range m.Learned {
			m.Learned[slot], m.Used[slot] = true, p.usesValue
		}
		for column := range m.Active {
			m.Active[column] = column * domain.Rows
		}
	})
}

func (p *Panel) forgetEveryMove() {
	p.src.Update(func(m *monster.Monster) {
		m.Learned = [monster.SlotCount]bool{}
		m.Used = [monster.SlotCount]int{}
		for column := range m.Active {
			m.Active[column] = monster.NoActiveMove
		}
	})
}
