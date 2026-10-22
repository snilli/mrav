package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type Item struct{ ent.Schema }

func (Item) Fields() []ent.Field {
	return []ent.Field{
		field.Int("code").NonNegative().Unique().Immutable(),
		field.String("name").NotEmpty(),
		field.String("desire_name").NotEmpty(),
		field.String("kind").NotEmpty(),
		field.Text("description"),
		field.Text("effect"),
		field.String("sale_price").NotEmpty(),
		field.String("purchase_price").NotEmpty(),
	}
}
