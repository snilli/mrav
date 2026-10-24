package stored

const (
	StatSum       = "stat_sum"
	MoveUsed      = "move_used"
	NatureAtLeast = "nature_at_least"
	NatureAtMost  = "nature_at_most"
	SubGenusIn    = "sub_genus_in"
	SubGenusNotIn = "sub_genus_not_in"
)

type Condition struct {
	Type  string `json:"type"`
	Stats []int  `json:"stats,omitempty"`
	Slot  *int   `json:"slot,omitempty"`
	Min   *int   `json:"min,omitempty"`
	Max   *int   `json:"max,omitempty"`
	IDs   []int  `json:"ids,omitempty"`
}

type Gate struct {
	Stats []int `json:"stats"`
	Min   int   `json:"min"`
	Max   int   `json:"max"`
}

type Chain struct {
	Slot int `json:"slot"`
	Uses int `json:"uses"`
}

type Acquire struct {
	Place    int    `json:"place"`
	Priority int    `json:"priority"`
	Auto     int    `json:"auto,omitempty"`
	Gates    []Gate `json:"gates,omitempty"`
	Nature   int    `json:"nature,omitempty"`
	Chain    *Chain `json:"chain,omitempty"`
	SubIn    []int  `json:"sub_in,omitempty"`
	SubNotIn []int  `json:"sub_not_in,omitempty"`
}

type Portal struct {
	Label string   `json:"label"`
	X     int      `json:"x"`
	Y     int      `json:"y"`
	To    []string `json:"to,omitempty"`
}

type Extra struct {
	Label string `json:"label"`
	Text  string `json:"text,omitempty"`
	Raw   int    `json:"raw,omitempty"`
}
