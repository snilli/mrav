package domain

import player "app/internal/core/player/domain"

const noItem = 0xFFFF

type Haul struct {
	Map  string
	Date int
	Got  map[string][]int
	Last int
}

func (h *Haul) Note(m ExpeditionMap, r player.Ranch) {
	if !r.FoundKnown || !m.Here(r) {
		return
	}
	if h.Map != m.Name || h.Date != r.Date {
		*h = Haul{Map: m.Name, Date: r.Date}
	}
	if r.Found == h.Last {
		return
	}
	h.Last = r.Found
	if r.Found == 0 || r.Found == noItem {
		return
	}
	at, _ := m.Stand(r)
	best, far := "", 2
	for _, s := range m.Spots {
		if d := max(s.At.X-at.X, at.X-s.At.X, s.At.Y-at.Y, at.Y-s.At.Y); d < far {
			best, far = s.Name, d
		}
	}
	if best == "" {
		return
	}
	if h.Got == nil {
		h.Got = map[string][]int{}
	}
	h.Got[best] = append(h.Got[best], r.Found)
}
