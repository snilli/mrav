package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Breed struct{ ent.Schema }

func (Breed) Fields() []ent.Field {
	return []ent.Field{
		field.Int("main_genus").NonNegative(),
		field.Int("sub_genus").NonNegative(),
		field.Int("seq").NonNegative(),
		field.String("name").NotEmpty(),
		field.String("otherwise").Default(""),
		field.Int("guts").Optional().Nillable().NonNegative(),
		field.Bool("match_guts").Default(false),
		field.JSON("growth", []int{}).Optional(),
	}
}

func (Breed) Indexes() []ent.Index {
	return []ent.Index{index.Fields("main_genus", "sub_genus", "seq").Unique()}
}
