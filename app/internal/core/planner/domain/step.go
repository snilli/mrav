package domain

import "slices"

type Item int

const (
	NoItem Item = iota
	MintLeaf
	NutsOil
	MagicBananaPlus
	MagicBananaZero
	MagicBananaMinus
	Irritater
	Mango
	Paradoxine
	Troron
	Candy
	SmokedSnake
	AppleCake
	Powder
	SweetJelly
	SourJelly
	StarPrune
	GoldPeach
	SilverPeach
	Griever

	itemCount = iota
)

var itemNames = [itemCount]string{
	"No item", "Mint Leaf", "Nuts Oil", "MB +10", "MB +0", "MB -10", "Irritater", "Mango", "Paradoxine", "Troron",
	"Candy", "Smoked Snake", "Apple Cake", "Powder", "Sweet Jelly", "Sour Jelly", "Star Prune", "Gold Peach", "Silver Peach", "Griever",
}

func Items() []Item {
	items := make([]Item, itemCount)
	for i := range items {
		items[i] = Item(i)
	}
	return items
}

func (i Item) Name() string { return itemNames[i] }

func (i Item) isPeach() bool { return i == GoldPeach || i == SilverPeach }

func ItemsFor(settings Settings, steps []Step, at int) []Item {
	var items []Item
	for _, item := range Items() {
		taken := item.isPeach() && (settings.ate(item) || slices.ContainsFunc(steps, func(s Step) bool { return s.Item == item && s.Action != Freeze }))
		if taken && (at < 0 || at >= len(steps) || steps[at].Item != item) {
			continue
		}
		items = append(items, item)
	}
	return items
}

type Action int

const (
	NoAction Action = iota
	Rest
	LightDrill
	HardDrill
	Errantry
	BattleWin
	BattleOK
	BattleLose
	Expedition
	Freeze
	Away
)

var actionNames = [...]string{"No action", "Rest", "Light Drill", "Hard Drill", "Errantry", "Battle (Win)", "Battle (OK)", "Battle (Lose)", "Expedition", "Freeze", "-"}

func Actions() []Action {
	return []Action{Rest, LightDrill, HardDrill, Errantry, BattleWin, BattleOK, BattleLose, Expedition, Freeze}
}

func (a Action) Name() string { return actionNames[a] }

type Step struct {
	Item   Item
	Action Action
}

func (s Step) blank() bool { return s.Item == NoItem && s.Action == NoAction }

func SetStep(steps []Step, at int, step Step) []Step {
	switch {
	case at < 0 || at > len(steps):
		return steps
	case at == len(steps):
		if step.blank() {
			return steps
		}
		return append(slices.Clone(steps), step)
	case step.blank():
		return slices.Delete(slices.Clone(steps), at, at+1)
	}
	out := slices.Clone(steps)
	out[at] = step
	return out
}
