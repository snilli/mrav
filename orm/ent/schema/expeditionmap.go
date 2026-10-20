package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	"orm/stored"
)

// ExpeditionMap is the map of one of the trips that come round every year, tile by tile.
type ExpeditionMap struct{ ent.Schema }

func (ExpeditionMap) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().Unique(),
		// rows is the tiles row by row, a newline after each: # a wall, . a floor, S the start, o a rock, g a gate that
		// wants intelligence (its lock says how much), T a teleport, and a letter a place.
		field.Text("rows").NotEmpty(),
		// start is the tile that the trip starts on, as a column and a row of rows.
		field.Int("start_x").NonNegative(),
		field.Int("start_y").NonNegative(),
		// portals are the teleports of the map and the places where they land, with the labels of the guide.
		field.JSON("portals", []stored.Portal{}).Optional(),
	}
}

func (ExpeditionMap) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("places", ExpeditionPlace.Type),
		edge.To("locks", ExpeditionLock.Type),
	}
}
