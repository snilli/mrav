package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Song struct{ ent.Schema }

func (Song) Fields() []ent.Field {
	return []ent.Field{
		field.Int("monster_id").Positive(),
		field.Int("main_genus").NonNegative(),
		field.Int("sub_genus").NonNegative(),
		field.String("title").NotEmpty(),
		field.String("artist").NotEmpty(),
		field.Bool("has_offset"),
		field.JSON("stats", []int{}),
		field.JSON("gains", []int{}),
		field.Int("lifespan").Positive(),
		field.Int("nature"),
		field.Int("life_type").Min(0).Max(3),
		field.Int("guts").Positive(),
		field.String("arena").NotEmpty(),
		field.String("training_bonus"),
		field.String("extra_techs"),
		field.String("battle_specials"),
		field.Bool("enemy"),
	}
}

func (Song) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("monster_id").Unique(),
		index.Fields("main_genus", "sub_genus"),
	}
}
