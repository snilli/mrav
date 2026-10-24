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
		field.Enum("version").Values("mr2", "mr2dx"),
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
		index.Fields("version", "genus_id", "memory_slot").Unique(),
		index.Fields("version", "genus_id", "col", "row").Unique(),
	}
}
