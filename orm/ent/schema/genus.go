package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Genus is one monster breed; its moves fill a 4x6 grid.
type Genus struct{ ent.Schema }

func (Genus) Fields() []ent.Field {
	return []ent.Field{
		// code is the game's genus number (0 is Pixie), kept apart from the auto id so 0 is a plain value.
		field.Int("code").NonNegative().Unique().Immutable(),
		field.String("name").NotEmpty().Unique(),
	}
}

func (Genus) Edges() []ent.Edge {
	return []ent.Edge{edge.To("moves", Move.Type)}
}
