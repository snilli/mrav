package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ExpeditionPlace is a building of an expedition's map where the monster can search for items.
type ExpeditionPlace struct{ ent.Schema }

func (ExpeditionPlace) Fields() []ent.Field {
	return []ent.Field{
		field.Int("map_id"),
		// name is the letter that the place has on its map.
		field.String("name").NotEmpty(),
		field.Int("x").NonNegative(),
		field.Int("y").NonNegative(),
		field.String("building").NotEmpty(),
		// stat is the stat that a search of the building goes by.
		field.Enum("stat").Values("int", "skill"),
	}
}

func (ExpeditionPlace) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("map", ExpeditionMap.Type).Ref("places").Field("map_id").Unique().Required(),
		edge.To("chances", LootChance.Type),
	}
}

func (ExpeditionPlace) Indexes() []ent.Index {
	// the places of a map, by the letter: one lookup by the map, in the order that they are read
	return []ent.Index{index.Fields("map_id", "name").Unique()}
}
