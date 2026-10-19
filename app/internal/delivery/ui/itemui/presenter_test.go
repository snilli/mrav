package itemui_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/item/domain"
	"app/internal/delivery/ui/itemui"
)

var _ = Describe("itemui presenter", func() {
	banana := domain.Item{
		Code: 29, Name: "Magic Banana", Kind: "Mysterious Banana", Description: "Its effect differs each time.",
		Effect: "[effect]", SalePrice: "100", PurchasePrice: "200",
	}

	It("labels a slot with the name of its item, and a blank slot with dashes", func() {
		Expect(itemui.SlotName(domain.Slot{Item: &banana})).To(Equal("Magic Banana"))
		Expect(itemui.SlotName(domain.Slot{})).To(Equal("-----"))
	})

	It("says everything the list says of an item", func() {
		Expect(itemui.NewDetailView(&domain.Slot{Index: 3, Item: &banana})).To(Equal(itemui.DetailView{
			Name: "Magic Banana", Kind: "Mysterious Banana", Description: "Its effect differs each time.",
			Effect: "[effect]", Sale: "100", Purchase: "200",
		}))
	})

	It("shows a blank slot the way the game does", func() {
		Expect(itemui.NewDetailView(&domain.Slot{Index: 1})).To(Equal(itemui.DetailView{
			Kind: "No Item", Description: "No item selected.", Effect: "This is a blank slot.", Sale: "-----", Purchase: "-----",
		}))
	})

	It("labels a place the house has no room for as locked, and says why when it is looked at", func() {
		Expect(itemui.SlotName(domain.Slot{Locked: true})).To(Equal("locked"))
		Expect(itemui.NewDetailView(&domain.Slot{Index: 15, Locked: true})).To(Equal(itemui.DetailView{
			Kind: "Locked", Description: "The house has no room for this place yet.",
			Effect: "Upgrade the house for more places.", Sale: "-----", Purchase: "-----",
		}))
	})

	It("shows nothing selected when no slot is", func() {
		Expect(itemui.NewDetailView(nil)).To(Equal(itemui.DetailView{
			Kind: "No Item", Description: "No item selected.", Effect: "No item selected.", Sale: "-----", Purchase: "-----",
		}))
	})
})
