package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type LifeStage struct{ ent.Schema }

func (LifeStage) Fields() []ent.Field {
	return []ent.Field{
		field.Int("life_type").Min(0).Max(3),
		field.Int("stage").Min(0).Max(9),
		field.Int("share").Min(1).Max(100),
	}
}

func (LifeStage) Indexes() []ent.Index {
	return []ent.Index{index.Fields("life_type", "stage").Unique()}
}
