package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// TrainingGain is what a drill gives a monster of a life stage, in words, for one growth rate of the stat that
// gains: "3-6 (3/4/5/6)" is a range and the four values the drill's results come to.
type TrainingGain struct{ ent.Schema }

func (TrainingGain) Fields() []ent.Field {
	return []ent.Field{
		field.Int("life_stage").Min(0).Max(9),
		field.Int("growth_rate").Min(0).Max(4),
		// kind is 0 for the stat a light drill trains, 1 for the main stat of a heavy drill, 3 for its other one.
		field.Int("kind").Min(0).Max(3),
		field.String("text").NotEmpty(),
	}
}

func (TrainingGain) Indexes() []ent.Index {
	return []ent.Index{index.Fields("life_stage", "growth_rate", "kind").Unique()}
}
