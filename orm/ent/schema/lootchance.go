package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type LootChance struct{ ent.Schema }

func (LootChance) Fields() []ent.Field {
	return []ent.Field{
		field.Int("place_id"),
		field.Bool("good"),
		field.Int("tier").NonNegative(),
		field.Int("slot").NonNegative(),
		field.Int("min_stat").NonNegative(),
		field.Float("pct").Min(0).Max(100),
	}
}

func (LootChance) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("place", ExpeditionPlace.Type).Ref("chances").Field("place_id").Unique().Required(),
		edge.To("item", Item.Type).Unique().Required().StorageKey(edge.Column("item_id")),
	}
}

func (LootChance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("good", "tier", "slot").Edges("place").Unique(),
		index.Fields("place_id", "good", "min_stat"),
	}
}
