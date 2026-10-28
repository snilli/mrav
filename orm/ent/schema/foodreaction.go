package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type FoodReaction struct{ ent.Schema }

func (FoodReaction) Fields() []ent.Field {
	return []ent.Field{
		field.Int("genus").NonNegative(),
		field.Enum("food").Values("potato", "fish", "meat", "milk", "cup_jelly", "tablet"),
		field.Enum("stage").Values("any", "baby", "grown"),
		field.Enum("reaction").Values("like", "neutral", "dislike"),
	}
}

func (FoodReaction) Indexes() []ent.Index {
	return []ent.Index{index.Fields("genus", "food", "stage").Unique()}
}
