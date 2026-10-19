package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Breed is one rule that names a breed from a monster's main and sub genus and its growth rates. A breed has one
// rule, or several, tried in the order of seq, when it comes in variants.
type Breed struct{ ent.Schema }

func (Breed) Fields() []ent.Field {
	return []ent.Field{
		field.Int("main_genus").NonNegative(),
		field.Int("sub_genus").NonNegative(),
		field.Int("seq").NonNegative(),
		field.String("name").NotEmpty(),
		// otherwise is the name when no rule of the breed matches, usually name with "(N/S)"; empty means name.
		field.String("otherwise").Default(""),
		// guts is the breed's guts rate, which the guts box checks; match_guts makes the rule need it.
		field.Int("guts").Optional().Nillable().NonNegative(),
		field.Bool("match_guts").Default(false),
		// growth is the six growth rates in the order Life, Pow, Int, Skill, Speed, Defense; absent means any.
		field.JSON("growth", []int{}).Optional(),
	}
}

func (Breed) Indexes() []ent.Index {
	return []ent.Index{index.Fields("main_genus", "sub_genus", "seq").Unique()}
}
