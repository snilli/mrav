package moverepo

import (
	"fmt"

	"orm/ent"
	"orm/stored"

	monster "app/internal/core/monster/domain"
	movedomain "app/internal/core/move/domain"
)

func toMove(e *ent.Move, genus monster.Genus) (movedomain.Move, error) {
	conds, err := domainConditions(e.Unlock)
	if err != nil {
		return movedomain.Move{}, err
	}
	acquire, err := domainAcquire(e.Acquire)
	if err != nil {
		return movedomain.Move{}, err
	}
	return movedomain.Move{
		Genus:       genus,
		MemorySlot:  e.MemorySlot,
		Col:         e.Col,
		Row:         e.Row,
		Name:        e.Name,
		Kind:        movedomain.Kind(e.Kind),
		Guts:        e.Guts,
		Force:       e.Force,
		Hit:         e.Hit,
		Wither:      e.Wither,
		Crit:        e.Crit,
		Extra:       domainExtra(e.Extra),
		Description: e.Description,
		Unlock:      movedomain.UnlockRule{Never: e.UnlockNever, All: conds},
		Acquire:     acquire,
		Icon:        e.Icon,
	}, nil
}

// errantryPlaces is how many places the errantry has, which an entry of moves.acquire names from 1.
const errantryPlaces = 5

// domainAcquire reads the JSON shape of moves.acquire back into the places that teach a move. No entries is nil. A place that
// the errantry does not have, or a gate that does not rise, is an error.
func domainAcquire(es []stored.Acquire) ([]movedomain.Acquire, error) {
	if len(es) == 0 {
		return nil, nil
	}
	out := make([]movedomain.Acquire, len(es))
	for i, e := range es {
		if e.Place < 1 || e.Place > errantryPlaces {
			return nil, fmt.Errorf("acquire entry %d: place %d is not one of the %d places", i, e.Place, errantryPlaces)
		}
		a := movedomain.Acquire{
			Place: e.Place, Priority: e.Priority, Auto: e.Auto, Nature: e.Nature,
			SubIn: genera(e.SubIn), SubNotIn: genera(e.SubNotIn),
		}
		if e.Chain != nil {
			a.Chain = &movedomain.Chain{Slot: e.Chain.Slot, Uses: e.Chain.Uses}
		}
		for _, g := range e.Gates {
			if g.Max <= g.Min {
				return nil, fmt.Errorf("acquire entry %d: gate from %d to %d does not rise", i, g.Min, g.Max)
			}
			stats := make([]monster.Stat, len(g.Stats))
			for k, s := range g.Stats {
				stats[k] = monster.Stat(s)
			}
			a.Gates = append(a.Gates, movedomain.Gate{Stats: stats, Min: g.Min, Max: g.Max})
		}
		out[i] = a
	}
	return out, nil
}

func domainExtra(e *stored.Extra) *movedomain.Extra {
	if e == nil {
		return nil
	}
	out := &movedomain.Extra{Label: e.Label, Text: e.Text, Raw: e.Raw}
	if e.Scale != nil {
		s := movedomain.Scale(*e.Scale)
		out.Scale = &s
	}
	return out
}

func genera(ids []int) []monster.Genus {
	if len(ids) == 0 {
		return nil
	}
	out := make([]monster.Genus, len(ids))
	for i, id := range ids {
		out[i] = monster.Genus(id)
	}
	return out
}

// domainConditions reads the JSON shape of moves.unlock back into conditions: the inverse of what the
// importer stores. No conditions is nil, the unconditional rule. A shape it does not know is an error.
func domainConditions(cs []stored.Condition) ([]movedomain.Condition, error) {
	if len(cs) == 0 {
		return nil, nil
	}
	out := make([]movedomain.Condition, 0, len(cs))
	for _, c := range cs {
		switch c.Type {
		case stored.StatSum:
			if c.Min == nil {
				return nil, fmt.Errorf("%s condition has no min", c.Type)
			}
			stats := make([]monster.Stat, len(c.Stats))
			for i, s := range c.Stats {
				stats[i] = monster.Stat(s)
			}
			out = append(out, movedomain.StatSumAtLeast{Stats: stats, Min: *c.Min})
		case stored.MoveUsed:
			if c.Slot == nil || c.Min == nil {
				return nil, fmt.Errorf("%s condition has no slot or no min", c.Type)
			}
			out = append(out, movedomain.MoveUsedAtLeast{Slot: *c.Slot, Min: *c.Min})
		case stored.NatureAtLeast:
			if c.Min == nil {
				return nil, fmt.Errorf("%s condition has no min", c.Type)
			}
			out = append(out, movedomain.NatureAtLeast{Min: *c.Min})
		case stored.NatureAtMost:
			if c.Max == nil {
				return nil, fmt.Errorf("%s condition has no max", c.Type)
			}
			out = append(out, movedomain.NatureAtMost{Max: *c.Max})
		case stored.SubGenusIn:
			out = append(out, movedomain.SubGenusIn{IDs: genera(c.IDs)})
		case stored.SubGenusNotIn:
			out = append(out, movedomain.SubGenusNotIn{IDs: genera(c.IDs)})
		default:
			return nil, fmt.Errorf("unknown unlock condition type %q", c.Type)
		}
	}
	return out, nil
}
