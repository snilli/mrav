package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Effect struct{ ent.Schema }

func (Effect) Fields() []ent.Field {
	return []ent.Field{
		field.String("subject").NotEmpty(),
		field.Enum("version").Values("mr2", "mr2dx"),
		field.Int("genus").Optional().Nillable(),
		field.Int("life_stage").Optional().Nillable(),
		field.Float("stress_factor").Optional().Nillable(),
		field.Int("fatigue"),
		field.Int("stress"),
		field.Int("form_divisor").Optional().Nillable(),
		field.Int("lost"),
		field.Int("form"),
	}
}

func (Effect) Indexes() []ent.Index {
	return []ent.Index{index.Fields("subject", "version", "genus", "life_stage").Unique()}
}
