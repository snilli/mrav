package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Genus struct{ ent.Schema }

func (Genus) Fields() []ent.Field {
	return []ent.Field{
		field.Int("code").NonNegative().Unique().Immutable(),
		field.String("name").NotEmpty().Unique(),
	}
}

func (Genus) Edges() []ent.Edge {
	return []ent.Edge{edge.To("moves", Move.Type)}
}
