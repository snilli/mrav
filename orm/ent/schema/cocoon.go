package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Cocoon struct{ ent.Schema }

func (Cocoon) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("version").Values("mr2", "mr2dx"),
		field.Int("guts").NonNegative(),
		field.JSON("growth", []int{}),
		field.String("name").NotEmpty(),
	}
}

func (Cocoon) Indexes() []ent.Index {
	return []ent.Index{index.Fields("version")}
}
