// Package domain is the Items page: the game's list of items, and what the player carries of it. Pure Go.
package domain

// EmptyFrom is the first id that is no item: the game marks an empty slot with an id from here up.
const EmptyFrom = 178

// Item is one thing in the game's list of items. The prices are text as the list has them: a number, "-----"
// when the item is not sold or bought, "???" when nobody knows.
type Item struct {
	Code          int
	Name          string
	DesireName    string // what the item is called when a monster wants it or dislikes it
	Kind          string
	Description   string
	Effect        string
	SalePrice     string
	PurchasePrice string
}

// Slot is one of the places the player carries an item in. Item is nil when the place is empty.
type Slot struct {
	Index int
	Item  *Item
}

// Fill puts the item of each id in its slot. A slot stays empty for an id that marks one, and for an id that the
// list does not have.
func Fill(ids []int, items []Item) []Slot {
	byCode := make(map[int]Item, len(items))
	for _, it := range items {
		byCode[it.Code] = it
	}
	slots := make([]Slot, len(ids))
	for i, id := range ids {
		slots[i].Index = i
		if id >= EmptyFrom {
			continue
		}
		if it, ok := byCode[id]; ok {
			slots[i].Item = &it
		}
	}
	return slots
}
