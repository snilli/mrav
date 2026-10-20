package domain

import monster "app/internal/core/monster/domain"

// Gate is a stat total that the chance of learning a move rises with. The total is the monster's loyalty and its Stats. Below Min
// the chance is no more than the auto learn share of what is left; it climbs from there to all that is left at Max.
type Gate struct {
	Stats    []monster.Stat
	Min, Max int
}

// Chain says that a move is taught only to a monster that has used the move in memory slot Slot Uses times.
type Chain struct{ Slot, Uses int }

// Acquire is a way that the errantry teaches a move: a place, from 1 (Papas) to 5 (Kawrea), and what it asks of the monster. The
// entries of a place are tried by Priority, and each takes its chance from what the ones before it left.
type Acquire struct {
	Place, Priority int
	Auto            int // the percent of what is left that is given however short the stats are
	Gates           []Gate
	Nature          int // at least this when it is above 0, at most this when it is below, nothing when it is 0
	Chain           *Chain
	SubIn, SubNotIn []monster.Genus // the sub-genera that it is only for, and the ones that it is not for
}
