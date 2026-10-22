package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type Cocoon struct{ ent.Schema }

func (Cocoon) Fields() []ent.Field {
	return []ent.Field{
		field.Int("guts").NonNegative(),
		field.JSON("growth", []int{}),
		field.String("name").NotEmpty().Unique(),
	}
}
