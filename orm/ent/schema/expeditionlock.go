package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ExpeditionLock is what a gate tile of an expedition's map wants: the intelligence that finding the way through
// it takes.
type ExpeditionLock struct{ ent.Schema }

func (ExpeditionLock) Fields() []ent.Field {
	return []ent.Field{
		field.Int("map_id"),
		field.Int("x").NonNegative(),
		field.Int("y").NonNegative(),
		field.Int("min_int").Positive(),
	}
}

func (ExpeditionLock) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("map", ExpeditionMap.Type).Ref("locks").Field("map_id").Unique().Required(),
	}
}

func (ExpeditionLock) Indexes() []ent.Index {
	// the locks of a map are read by the map, in the order of the tiles
	return []ent.Index{index.Fields("map_id", "x", "y").Unique()}
}
