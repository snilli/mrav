package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/item/domain"
)

var _ = Describe("Fill", func() {
	potato := domain.Item{Code: 0, Name: "Potato", Kind: "Animal Foods"}
	banana := domain.Item{Code: 29, Name: "Magic Banana", Kind: "Mysterious Banana"}
	catalog := []domain.Item{potato, banana}

	It("puts the item of each id in its slot, numbering the slots from 0", func() {
		slots := domain.Fill([]int{29, 0}, catalog)
		Expect(slots).To(HaveLen(2))
		Expect(slots[0].Index).To(Equal(0))
		Expect(*slots[0].Item).To(Equal(banana))
		Expect(slots[1].Index).To(Equal(1))
		Expect(*slots[1].Item).To(Equal(potato))
	})

	It("leaves a slot empty for the ids from 178 up, which is how the game marks one", func() {
		slots := domain.Fill([]int{178, 255}, catalog)
		Expect(slots[0].Item).To(BeNil())
		Expect(slots[1].Item).To(BeNil())
		Expect(slots[1].Index).To(Equal(1))
	})

	It("leaves a slot empty for an id the catalog does not have", func() {
		Expect(domain.Fill([]int{5}, catalog)[0].Item).To(BeNil())
	})

	It("does not let a slot change the catalog", func() {
		slots := domain.Fill([]int{0}, catalog)
		slots[0].Item.Name = "Changed"
		Expect(catalog[0].Name).To(Equal("Potato"))
	})
})

var _ = Describe("Lock", func() {
	It("locks the places the house has no room for, from the first one past what it gives", func() {
		slots := domain.Lock(domain.Fill(make([]int, 20), nil), 14)
		for i, s := range slots {
			Expect(s.Locked).To(Equal(i >= 14), "slot %d", i)
		}
	})

	It("locks none when the house gives all of them, or more", func() {
		for _, usable := range []int{20, 25} {
			for _, s := range domain.Lock(domain.Fill(make([]int, 20), nil), usable) {
				Expect(s.Locked).To(BeFalse())
			}
		}
	})
})
