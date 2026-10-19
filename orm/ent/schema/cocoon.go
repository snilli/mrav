package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Cocoon names a cocooned Worm from its guts rate and its growth rates.
type Cocoon struct{ ent.Schema }

func (Cocoon) Fields() []ent.Field {
	return []ent.Field{
		field.Int("guts").NonNegative(),
		// growth is the six growth rates in the order Life, Pow, Int, Skill, Speed, Defense.
		field.JSON("growth", []int{}),
		field.String("name").NotEmpty().Unique(),
	}
}
