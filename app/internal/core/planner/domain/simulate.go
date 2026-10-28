package domain

import (
	"math"

	monster "app/internal/core/monster/domain"
	profile "app/internal/core/profile/domain"
)

func clamp(v int) int { return max(0, min(100, v)) }

func ceil(v float64) int { return int(math.Ceil(v)) }

func simulate(rows []Row, reactions Reactions, effects Effects, s Settings) (weeks, lived int) {
	ftg, str, form := s.Fatigue, s.Stress, s.Form
	frozen := 0
	for i := range rows {
		r := &rows[i]
		switch {
		case r.Kind == FoodRow:
			ftg, str = monthly(r.Food, ftg, str, s, reactions)
			form = clampForm(form + effects.of(foodSubject(r.Food), s).Form)
			r.Shown, r.Fatigue, r.Stress, r.Form = true, ftg, str, form
		case r.Item == NoItem && r.Action == NoAction:
			frozen++
		default:
			var lost int
			ftg, str, lost = week(rows, i, ftg, str, s, effects)
			form = clampForm(form + effects.of(itemSubject(r.Item), s).Form)
			form = clampForm(form + effects.of(actionSubject(r.Action), s).Form)
			r.Form = form
			weeks += lost
			if r.Action == Freeze {
				frozen++
			}
			r.ShowLife, r.Total, r.Life = true, weeks, s.Lifespan-(r.Month-1)*4-r.Week+frozen+weeks
		}
	}
	last := rows[len(rows)-1]
	return weeks, (last.Month-1)*4 + last.Week - frozen
}

func monthly(food Food, ftg, str int, s Settings, reactions Reactions) (int, int) {
	for range s.Pots {
		d := int(math.Floor(float64(str) * 0.03))
		if d == 0 && s.Version == monster.MR2DX {
			d = 1
		}
		str -= d
	}
	for range s.Sculptures {
		ftg -= int(math.Floor(float64(ftg) * 0.03))
	}
	var reaction Reaction
	switch food {
	case s.Like:
		reaction = Like
	case s.Dislike:
		reaction = Dislike
	default:
		reaction = reactions.Of(s.Main, food, s.LifeStage)
	}
	return clamp(ftg), clamp(str + FoodStress(food, reaction))
}

func week(rows []Row, at, ftg, str int, s Settings, effects Effects) (int, int, int) {
	r := &rows[at]
	dx := s.Version == monster.MR2DX
	item, action := effects.of(itemSubject(r.Item), s), effects.of(actionSubject(r.Action), s)
	ftg, str = item.apply(ftg, str, s.Form)
	ftg, str = clamp(ftg), clamp(str)
	ftg, str = action.apply(ftg, str, s.Form)
	ftg, str = clamp(ftg), clamp(str)
	lost := item.Lost + action.Lost

	r.Shown, r.Fatigue, r.Stress = true, ftg, str
	r.LI = ftg + 2*str
	away := (dx && r.Action == Away) || r.Action == Freeze
	r.ShowLI = !away

	d := profile.LifeIndexOf(ftg, str).WeeksLost
	if r.Action == BattleWin || r.Action == BattleOK || r.Action == BattleLose {
		lost -= 3 * (d + 1)
	}
	if !dx && r.Action == Errantry && followedByOther(rows, at) {
		lost -= d + 1
	}
	lost -= d

	if away {
		return ftg, str, 0
	}
	r.Lost = lost
	r.ShowLost = r.Action != NoAction || r.Item == GoldPeach || r.Item == SilverPeach || r.Item == Paradoxine || r.Item == Troron
	return ftg, str, lost
}

func followedByOther(rows []Row, at int) bool {
	month, week := rows[at].Month, rows[at].Week+1
	if week == 5 {
		month, week = month+1, 1
	}
	for _, r := range rows {
		if r.Month == month && r.Week == week {
			return r.Action != Errantry
		}
	}
	return false
}
