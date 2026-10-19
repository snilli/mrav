// Package itemui is the Items page: the places the player carries items in, and what the game's list says of each.
package itemui

import "app/internal/core/item/domain"

const blank = "-----"

// DetailView is what the details of one slot show, every value already in the game's words.
type DetailView struct {
	Name        string
	Kind        string
	Description string
	Effect      string
	Sale        string
	Purchase    string
}

// SlotName is what a slot's button says: the name of its item, or dashes for an empty place.
func SlotName(s domain.Slot) string {
	if s.Item == nil {
		return blank
	}
	return s.Item.Name
}

// NewDetailView says what a slot holds. A nil slot is none selected.
func NewDetailView(s *domain.Slot) DetailView {
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
