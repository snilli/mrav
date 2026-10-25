package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"orm/stored"
)

type CalendarEvent struct{ ent.Schema }

func (CalendarEvent) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("version").Values("mr2", "mr2dx"),
		field.Int("month").Min(1).Max(12),
		field.Int("week").Min(1).Max(4),
		field.String("name").NotEmpty(),
		field.Enum("kind").Values("cup", "invite", "trip", "ranch", "visit"),
		field.Int("purse").NonNegative().Default(0),
		field.String("prize").Default(""),
		field.String("unseen").Default(""),
		field.String("note").Default(""),
		field.JSON("conditions", []stored.Condition{}).Optional(),
	}
}

func (CalendarEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("version", "month", "week", "name").Unique(),
	}
}
