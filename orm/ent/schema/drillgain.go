package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type DrillGain struct{ ent.Schema }

func (DrillGain) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("result").Values("cheat", "success", "great"),
		field.Enum("role").Values("light", "primary", "secondary", "drop"),
		field.Bool("bonus"),
		field.Int("life_stage").Min(0).Max(9),
		field.Int("growth_rate").Min(0).Max(4),
		field.JSON("slots", [][]int{}),
	}
}

func (DrillGain) Indexes() []ent.Index {
	return []ent.Index{index.Fields("result", "role", "bonus", "life_stage", "growth_rate").Unique()}
}
