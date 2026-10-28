package domain

const EmptyFrom = 178

type Item struct {
	Code          int
	Name          string
	DesireName    string
	Kind          string
	Description   string
	Effect        string
	SalePrice     *int
	PurchasePrice *int
}

type Slot struct {
	Index  int
	Item   *Item
	Locked bool
}

func Lock(slots []Slot, usable int) []Slot {
	for i := range slots {
		slots[i].Locked = i >= usable
	}
	return slots
}

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
