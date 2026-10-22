package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"orm/stored"
)

type Move struct{ ent.Schema }

func (Move) Fields() []ent.Field {
	return []ent.Field{
		field.Int("genus_id"),
		field.Int("memory_slot").Min(0).Max(23),
		field.Int("col").Min(1).Max(4),
		field.Int("row").Min(1).Max(6),
		field.String("name").NotEmpty(),
		field.Enum("kind").Values("power", "intelligence", "unknown"),
		field.Int("guts"),
		field.Int("force"),
		field.Int("hit"),
		field.Int("wither"),
		field.Int("crit"),
		field.JSON("extra", &stored.Extra{}).Optional(),
		field.Text("description"),
		field.Bool("unlock_never"),
		field.JSON("unlock", []stored.Condition{}),
		field.JSON("acquire", []stored.Acquire{}).Optional(),
		field.String("icon").NotEmpty(),
	}
}

func (Move) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("genus", Genus.Type).Ref("moves").Field("genus_id").Unique().Required(),
	}
}

func (Move) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("genus_id", "memory_slot").Unique(),
		index.Fields("genus_id", "col", "row").Unique(),
	}
}
