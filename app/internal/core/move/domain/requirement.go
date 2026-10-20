package domain

import (
	"fmt"
	"strconv"
	"strings"

	monster "app/internal/core/monster/domain"
)

// Requirement is one condition of an unlock rule in words, with what the monster has and whether it is met, so
// that a move that cannot be learned yet says what it still asks.
type Requirement struct {
	Text string // "Use Bolt at least 50 times"
	Have string // "49 / 50": what the monster has
	Met  bool
	// Stats are the stats that a text starts with, "Pow + Int at least 400", in the order of its words, so that they can be
	// shown in their colours.
	Stats []monster.Stat
}

// Names turn the numbers in a rule into words. A missing function leaves the number: "move 9", "#24".
type Names struct {
	Move  func(slot int) string
	Genus func(g monster.Genus) string
}

func (n Names) move(slot int) string {
	if n.Move != nil {
		if name := n.Move(slot); name != "" {
			return name
		}
	}
	return "move " + strconv.Itoa(slot)
}

func (n Names) genus(g monster.Genus) string {
	if n.Genus != nil {
		if name := n.Genus(g); name != "" {
			return name
		}
	}
	return "#" + strconv.Itoa(int(g))
}

func (n Names) genera(ids []monster.Genus) string {
	names := make([]string, len(ids))
	for i, id := range ids {
		names[i] = n.genus(id)
	}
	return strings.Join(names, " or ")
}

var statNames = [monster.StatCount]string{"Life", "Pow", "Int", "Skill", "Speed", "Def"}

func (c StatSumAtLeast) Require(m monster.Monster, _ Names) Requirement {
	var names []string
	var stats []monster.Stat
	for _, s := range c.Stats {
		if s >= 0 && s < monster.StatCount { // a stat outside the six adds nothing, as in StatSum
			names = append(names, statNames[s])
			stats = append(stats, s)
		}
	}
	return Requirement{
		Text:  fmt.Sprintf("%s at least %d", strings.Join(names, " + "), c.Min),
		Have:  strconv.Itoa(m.StatSum(c.Stats...)),
		Met:   c.Holds(m),
		Stats: stats,
	}
}

func (c MoveUsedAtLeast) Require(m monster.Monster, n Names) Requirement {
	used := 0
	if c.Slot >= 0 && c.Slot < monster.SlotCount {
		used = m.Used[c.Slot]
	}
	return Requirement{
		Text: fmt.Sprintf("Use %s at least %d times", n.move(c.Slot), c.Min),
		Have: fmt.Sprintf("%d / %d", used, c.Min),
		Met:  c.Holds(m),
	}
}

func (c NatureAtLeast) Require(m monster.Monster, _ Names) Requirement {
	return Requirement{Text: fmt.Sprintf("Nature at least %d", c.Min), Have: strconv.Itoa(m.Nature), Met: c.Holds(m)}
}

func (c NatureAtMost) Require(m monster.Monster, _ Names) Requirement {
	return Requirement{Text: fmt.Sprintf("Nature at most %d", c.Max), Have: strconv.Itoa(m.Nature), Met: c.Holds(m)}
}

func (c SubGenusIn) Require(m monster.Monster, n Names) Requirement {
	return Requirement{Text: "Sub-genus is " + n.genera(c.IDs), Have: n.genus(m.SubGenus), Met: c.Holds(m)}
}

func (c SubGenusNotIn) Require(m monster.Monster, n Names) Requirement {
	return Requirement{Text: "Sub-genus is not " + n.genera(c.IDs), Have: n.genus(m.SubGenus), Met: c.Holds(m)}
}

// Requirements lists what the rule asks of a monster, condition by condition, met or not. A move the game gives
// no way to learn has one requirement that cannot be met; a move with no condition has none.
func (r UnlockRule) Requirements(m monster.Monster, n Names) []Requirement {
	if r.Never {
		return []Requirement{{Text: "The game gives no way to learn it"}}
	}
	out := make([]Requirement, 0, len(r.All))
	for _, c := range r.All {
		out = append(out, c.Require(m, n))
	}
	return out
}
