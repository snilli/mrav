package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	"orm/stored"
)

type ExpeditionMap struct{ ent.Schema }

func (ExpeditionMap) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().Unique(),
		field.Text("rows").NotEmpty(),
		field.Int("start_x").NonNegative(),
		field.Int("start_y").NonNegative(),
		field.JSON("portals", []stored.Portal{}).Optional(),
	}
}

func (ExpeditionMap) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("places", ExpeditionPlace.Type),
		edge.To("locks", ExpeditionLock.Type),
	}
}
