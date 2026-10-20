// Package stored holds the shapes kept inside JSON columns, so the writer and the reader of a column
// agree on them in one place.
package stored

// Condition types of an unlock rule.
const (
	StatSum       = "stat_sum"
	MoveUsed      = "move_used"
	NatureAtLeast = "nature_at_least"
	NatureAtMost  = "nature_at_most"
	SubGenusIn    = "sub_genus_in"
	SubGenusNotIn = "sub_genus_not_in"
)

// Condition is one entry of moves.unlock; which fields are set depends on Type.
type Condition struct {
	Type  string `json:"type"`
	Stats []int  `json:"stats,omitempty"`
	Slot  *int   `json:"slot,omitempty"`
	Min   *int   `json:"min,omitempty"`
	Max   *int   `json:"max,omitempty"`
	IDs   []int  `json:"ids,omitempty"`
}

// Gate is a stat total that the chance of learning a move rises with: nothing is asked below Min, and the whole of what is
// left is given from Max up. The total is the monster's loyalty and the stats in Stats, numbered like the stats of a monster.
type Gate struct {
	Stats []int `json:"stats"`
	Min   int   `json:"min"`
	Max   int   `json:"max"`
}

// Chain says that a move is taught only to a monster that has used the move in memory slot Slot Uses times.
type Chain struct {
	Slot int `json:"slot"`
	Uses int `json:"uses"`
}

// Acquire is one entry of moves.acquire: a place of the errantry that teaches the move, and what it asks. Places are 1 to 5,
// Papas to Kawrea. Entries of a place are tried by Priority, and each takes its chance from what the ones before it left.
// Nature, when set, is the nature that the monster needs: at least Nature when it is above 0, at most when below. SubIn and
// SubNotIn are sub-genus codes, like the ids of the sub-genus conditions of moves.unlock.
type Acquire struct {
	Place    int    `json:"place"`
	Priority int    `json:"priority"`
	Auto     int    `json:"auto,omitempty"` // the percent of what is left that is given however short the stats are
	Gates    []Gate `json:"gates,omitempty"`
	Nature   int    `json:"nature,omitempty"`
	Chain    *Chain `json:"chain,omitempty"`
	SubIn    []int  `json:"sub_in,omitempty"`
	SubNotIn []int  `json:"sub_not_in,omitempty"`
}

// Portal is an entry of expedition_maps.portals: a teleport (T1, T2, ...) or a place where a teleport lands (E1, E2, ...), as
// the guide labels them on its maps. A teleport leads, at random, to one of the labels in To: a place to land, or another
// teleport. A place to land has no To.
type Portal struct {
	Label string   `json:"label"`
	X     int      `json:"x"`
	Y     int      `json:"y"`
	To    []string `json:"to,omitempty"`
}

// Extra is moves.extra: the optional "HP Drain: 100%" row. Text is a literal; Scale and Raw name a
// value the game rates in words.
type Extra struct {
	Label string `json:"label"`
	Text  string `json:"text,omitempty"`
	Scale *int   `json:"scale,omitempty"`
	Raw   int    `json:"raw,omitempty"`
}
