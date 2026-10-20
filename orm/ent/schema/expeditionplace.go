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
		edge.From("map", ExpeditionMap.Type).Ref("places").Unique().Required(),
		edge.To("chances", LootChance.Type).StorageKey(edge.Column("place_id")),
	}
}

func (ExpeditionPlace) Indexes() []ent.Index {
	return []ent.Index{index.Fields("name").Edges("map").Unique()}
}
