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

// Extra is moves.extra: the optional "HP Drain: 100%" row. Text is a literal; Scale and Raw name a
// value the game rates in words.
type Extra struct {
	Label string `json:"label"`
	Text  string `json:"text,omitempty"`
	Scale *int   `json:"scale,omitempty"`
	Raw   int    `json:"raw,omitempty"`
}
