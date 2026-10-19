// Package domain is the monster as the game reports it: a snapshot of its breed, stats and moves.
package domain

const (
	// SlotCount is how many moves the game keeps learned flags and use counters for.
	SlotCount = 24
	// ActiveGroups is the number of active-move slots, one per column of the grid.
	ActiveGroups = 4
	// NoActiveMove is what the game stores for a group with no active move.
	NoActiveMove = 24
)

// Genus is a monster's main breed, 0 (Pixie) to 37 (Naga).
type Genus int

// Stat indexes Monster.Stats, in the order the game keeps them.
type Stat int

const (
	Life Stat = iota
	Pow
	Int
	Skill
	Speed
	Defense
	StatCount = iota
)

// Monster is a snapshot of the game's state for the monster being raised.
type Monster struct {
	Genus    Genus
	SubGenus Genus
	Stats    [StatCount]int
	Nature   int // the effective nature, as the game shows it
	Learned  [SlotCount]bool
	Used     [SlotCount]int
	Active   [ActiveGroups]int // a memory slot per group, NoActiveMove when there is none
}

// StatSum adds the named stats; a stat outside the six adds nothing.
func (m Monster) StatSum(stats ...Stat) int {
	sum := 0
	for _, s := range stats {
		if s >= 0 && s < StatCount {
			sum += m.Stats[s]
		}
	}
	return sum
}
