// Package domain is the monster as the game reports it: a snapshot of its breed, stats and moves.
package domain

const (
	// SlotCount is how many moves the game keeps learned flags and use counters for.
	SlotCount = 24
	// ActiveGroups is the number of active-move slots, one per column of the grid.
	ActiveGroups = 4
	// NoActiveMove is what the game stores for a group with no active move.
	NoActiveMove = 24
	// DrillCount is how many drills a monster has a motivation for.
	DrillCount = 10
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

	// The rest is what the Monster, Life Index and Training pages show.
	Name            string
	Age             int // in weeks
	Lifespan        int // weeks left
	InitialLifespan int
	NatureBase      int // the part of the nature that does not swing
	Fatigue         int
	Stress          int
	Fame            int
	Form            int
	Spoil           int
	Fear            int
	Growth          [StatCount]int  // 0 (E) to 4 (A) and beyond, by stat, in the order of Stats
	TrainBoost      int             // flags: the drills and the places the monster is good at
	CupJellies      int             // a Worm's count
	GoldPeach       bool            // eaten
	SilverPeach     bool            // eaten
	ItemUsed        bool            // an item has been used this week
	Motivation      [DrillCount]int // percent, by drill: Domino, Study, Run, Shoot, Dodge, Endure, Pull, Meditate, Leap, Swim
	PatchedSwim     bool            // the game trains the swim as the patched game does, which the Steam release does
	PlayType        int
	Drug            int // an item id; the drugs are 34 to 41
	DrugWeeks       int
	ItemLike        int // an item id, or a desire id such as Battle
	ItemDislike     int
	Rank            int
	LifeStage       int
	LifeType        int
	ArenaSpeed      int
	GutsRate        int
	Specials        int // flags: the battle specials
	PrizeMoney      int
	Record          Record
}

// Record is what the game keeps of a monster's fights and ailments: counts, each of them.
type Record struct {
	Wins, Losses, KOs                          int
	Injuries, Runaways, Sicknesses, Resistance int
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
