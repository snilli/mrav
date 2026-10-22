package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type TrainingGain struct{ ent.Schema }

func (TrainingGain) Fields() []ent.Field {
	return []ent.Field{
		field.Int("life_stage").Min(0).Max(9),
		field.Int("growth_rate").Min(0).Max(4),
		field.Int("kind").Min(0).Max(3),
		field.String("text").NotEmpty(),
	}
}

func (TrainingGain) Indexes() []ent.Index {
	return []ent.Index{index.Fields("life_stage", "growth_rate", "kind").Unique()}
}
