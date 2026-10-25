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
	for _, s := range m.Spots {
		if Near(s.At, at) {
			if h.Got == nil {
				h.Got = map[string][]int{}
			}
			h.Got[s.Name] = append(h.Got[s.Name], r.Found)
			return
		}
	}
}
