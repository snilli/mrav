package profilerepo

import (
	"context"
	"fmt"

	"orm/ent"
	entcalendar "orm/ent/calendarevent"
	"orm/stored"

	monster "app/internal/core/monster/domain"
	"app/internal/core/profile/domain"
)

var eventKinds = map[entcalendar.Kind]domain.EventKind{
	entcalendar.KindCup:    domain.EventCup,
	entcalendar.KindInvite: domain.EventInvite,
	entcalendar.KindTrip:   domain.EventTrip,
	entcalendar.KindRanch:  domain.EventRanch,
	entcalendar.KindVisit:  domain.EventVisit,
}

var condTypes = map[string]domain.CondType{
	stored.RankAtLeast:        domain.CondRankAtLeast,
	stored.RankExactly:        domain.CondRankExactly,
	stored.CupGrade:           domain.CondCupGrade,
	stored.TrainerRankAtLeast: domain.CondTrainerRankAtLeast,
	stored.YearAtLeast:        domain.CondYearAtLeast,
	stored.AgeUnder:           domain.CondAgeUnder,
	stored.AgeAtLeast:         domain.CondAgeAtLeast,
	stored.NatureAtLeast:      domain.CondNatureAtLeast,
	stored.NatureAtMost:       domain.CondNatureAtMost,
	stored.Expedition:         domain.CondExpedition,
	stored.PhoenixOpen:        domain.CondPhoenixOpen,
	stored.PhoenixDone:        domain.CondPhoenixDone,
	stored.IMaYear:            domain.CondIMaYear,
	stored.FacilityNext:       domain.CondFacilityNext,
}

func (r *Repository) ListCalendar(ctx context.Context, version monster.Version) ([]domain.CalendarEntry, error) {
	return r.calendar.Get(version, func() ([]domain.CalendarEntry, error) { return r.loadCalendar(ctx, version) })
}

func (r *Repository) loadCalendar(ctx context.Context, version monster.Version) ([]domain.CalendarEntry, error) {
	rows, err := r.client.CalendarEvent.Query().
		Where(entcalendar.VersionEQ(entcalendar.Version(version))).
		Order(ent.Asc(entcalendar.FieldMonth), ent.Asc(entcalendar.FieldWeek), ent.Asc(entcalendar.FieldName)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.CalendarEntry, len(rows))
	for i, e := range rows {
		out[i] = domain.CalendarEntry{
			Month: e.Month, Week: e.Week, Name: e.Name, Kind: eventKinds[e.Kind],
			Purse: e.Purse, Prize: e.Prize, Unseen: e.Unseen, Note: e.Note,
		}
		for _, c := range e.Conditions {
			t, ok := condTypes[c.Type]
			if !ok {
				return nil, fmt.Errorf("calendar entry %s: unknown condition %q", e.Name, c.Type)
			}
			cond := domain.EntryCondition{Type: t}
			if c.Min != nil {
				cond.Value = *c.Min
			}
			out[i].When = append(out[i].When, cond)
		}
	}
	return out, nil
}
