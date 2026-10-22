package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type ExpeditionPlace struct{ ent.Schema }

func (ExpeditionPlace) Fields() []ent.Field {
	return []ent.Field{
		field.Int("map_id"),
		field.String("name").NotEmpty(),
		field.Int("x").NonNegative(),
		field.Int("y").NonNegative(),
		field.String("building").NotEmpty(),
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
	return []ent.Index{index.Fields("map_id", "name").Unique()}
}
