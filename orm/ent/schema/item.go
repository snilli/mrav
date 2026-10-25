package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Item struct{ ent.Schema }

func (Item) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("version").Values("mr2", "mr2dx"),
		field.Int("code").NonNegative().Immutable(),
		field.String("name").NotEmpty(),
		field.String("desire_name").NotEmpty(),
		field.String("kind").NotEmpty(),
		field.Text("description"),
		field.Text("effect"),
		field.String("sale_price").NotEmpty(),
		field.String("purchase_price").NotEmpty(),
	}
}

func (Item) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("version", "code").Unique(),
	}
}
