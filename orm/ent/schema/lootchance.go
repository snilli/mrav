package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// LootChance is the chance of finding an item in a place, in percent, for a tier of the stat that the search goes
// by: the tier is the stat from min_stat up to the next tier's. A place has one table of chances for a monster whose
// nature is above 0 and another for one whose nature is not.
type LootChance struct{ ent.Schema }

func (LootChance) Fields() []ent.Field {
	return []ent.Field{
		field.Int("place_id"),
		// good is whether this is the table for a nature above 0.
		field.Bool("good"),
		field.Int("tier").NonNegative(),
		// slot is the row of the table that the chance is in. An item can have two rows, which add up.
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
		// the pages ask for the maps twice a second, and each asks for the chances of a place at the tier of a stat: the
		// highest min_stat that the stat reaches, and then the rows at that min_stat
		index.Fields("place_id", "good", "min_stat"),
	}
}
