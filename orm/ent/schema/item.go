package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Item is one thing the player can carry, as the game's item list has it.
type Item struct{ ent.Schema }

func (Item) Fields() []ent.Field {
	return []ent.Field{
		// code is the game's item id, kept apart from the auto id so 0 is a plain value.
		field.Int("code").NonNegative().Unique().Immutable(),
		field.String("name").NotEmpty(),
		// desire_name is what the item is called when a monster wants it or dislikes it, which is not always
		// its name: every kind of chip is "Disc Chips", and there are things a monster wants that are no item.
		field.String("desire_name").NotEmpty(),
		field.String("kind").NotEmpty(),
		field.Text("description"),
		field.Text("effect"),
		// The prices are text as the game's list has them: a number, "-----" when it is not sold or bought, and
		// "???" when nobody knows.
		field.String("sale_price").NotEmpty(),
		field.String("purchase_price").NotEmpty(),
	}
}
