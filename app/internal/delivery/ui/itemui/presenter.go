package itemui

import "app/internal/core/item/domain"

const blank = "-----"

type DetailView struct {
	Name        string
	Kind        string
	Description string
	Effect      string
	Sale        string
	Purchase    string
}

func SlotName(s domain.Slot) string {
	switch {
	case s.Locked:
		return "locked"
	case s.Item == nil:
		return blank
	}
	return s.Item.Name
}

func NewDetailView(s *domain.Slot) DetailView {
	if s != nil && s.Locked {
		return DetailView{
			Kind: "Locked", Description: "The house has no room for this place yet.",
			Effect: "Upgrade the house for more places.", Sale: blank, Purchase: blank,
		}
	}
	if s != nil && s.Item != nil {
		it := s.Item
		return DetailView{
			Name: it.Name, Kind: it.Kind, Description: it.Description, Effect: it.Effect,
			Sale: it.SalePrice, Purchase: it.PurchasePrice,
		}
	}
	v := DetailView{Kind: "No Item", Description: "No item selected.", Effect: "No item selected.", Sale: blank, Purchase: blank}
	if s != nil {
		v.Effect = "This is a blank slot."
	}
	return v
}
