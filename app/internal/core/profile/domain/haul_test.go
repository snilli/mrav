package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	player "app/internal/core/player/domain"
	"app/internal/core/profile/domain"
)

var _ = Describe("Haul", func() {
	const kawrea = "Kawrea Volcano"
	var m domain.ExpeditionMap
	at := func(x, y, found int) player.Ranch {
		r := player.Ranch{Known: true, Date: 8, Found: found, FoundKnown: true}
		return r.WalkedTo(x, y)
	}

	with := func(r player.Ranch, energy int) player.Ranch {
		r.Energy, r.EnergyKnown = energy, true
		return r
	}

	BeforeEach(func() {
		m = chart(nil,
			"S.A",
			"...",
			"..B",
		)
		m.Name = kawrea
	})

	It("records the item that turns up at the place next to where the monster stands", func() {
		var h domain.Haul
		h.Note(m, at(0, 0, 0))
		h.Note(m, at(1, 0, 46))
		Expect(h.Got).To(Equal(map[string][]int{"A": {46}}))
	})

	It("takes the place under the monster too", func() {
		var h domain.Haul
		h.Note(m, at(0, 0, 0))
		h.Note(m, at(2, 0, 46))
		Expect(h.Got).To(Equal(map[string][]int{"A": {46}}))
	})

	It("gives the find to the place the monster stands on rather than to one before it in the list that is only next to it, as on E of Kawrea with D beside it", func() {
		m := chart(nil, "S.DE")
		m.Name = kawrea
		var h domain.Haul
		h.Note(m, at(0, 0, 0))
		h.Note(m, at(3, 0, 46))
		Expect(h.Got).To(Equal(map[string][]int{"E": {46}}))
	})

	It("records nothing new while the item stays the same, and the next item at the next place", func() {
		var h domain.Haul
		h.Note(m, at(0, 0, 0))
		h.Note(m, at(1, 0, 46))
		h.Note(m, at(1, -1, 46))
		Expect(h.Got).To(Equal(map[string][]int{"A": {46}}))
		h.Note(m, at(1, -2, 18))
		Expect(h.Got).To(Equal(map[string][]int{"A": {46}, "B": {18}}))
	})

	It("does not give an item to a place it turned up far from", func() {
		var h domain.Haul
		h.Note(m, at(0, 0, 0))
		h.Note(m, at(0, -2, 46))
		h.Note(m, at(1, -2, 46))
		Expect(h.Got).To(BeEmpty())
	})

	It("records each find at a place searched again, the same item too, as the game clears the item before each search", func() {
		var h domain.Haul
		h.Note(m, at(1, 0, 164))
		h.Note(m, at(1, 0, 0))
		h.Note(m, at(1, 0, 164))
		h.Note(m, at(1, 0, 0xFFFF))
		h.Note(m, at(1, 0, 18))
		Expect(h.Got).To(Equal(map[string][]int{"A": {164, 164, 18}}))
	})

	It("keeps what it found when the game clears the item before the next search", func() {
		var h domain.Haul
		h.Note(m, at(1, 0, 46))
		h.Note(m, at(1, -1, 0))
		h.Note(m, at(1, -2, 18))
		Expect(h.Got).To(Equal(map[string][]int{"A": {46}, "B": {18}}))
	})

	It("finds again the same item after the game cleared it, at another place", func() {
		var h domain.Haul
		h.Note(m, at(1, 0, 46))
		h.Note(m, at(1, -1, 0))
		h.Note(m, at(1, -2, 46))
		Expect(h.Got).To(Equal(map[string][]int{"A": {46}, "B": {46}}))
	})

	It("keeps what it found when the monster goes onto the map again in the same week, with its energy full again", func() {
		var h domain.Haul
		h.Note(m, with(at(1, 0, 46), 200))
		h.Note(m, with(at(0, 0, 0xFFFF), 280))
		Expect(h.Got).To(Equal(map[string][]int{"A": {46}}))
	})

	It("takes 0xFFFF, which the game holds before a find, as no item", func() {
		var h domain.Haul
		h.Note(m, at(1, 0, 0xFFFF))
		Expect(h.Got).To(BeEmpty())
		h.Note(m, at(1, 0, 46))
		Expect(h.Got).To(Equal(map[string][]int{"A": {46}}))
	})

	It("starts again in another week", func() {
		var h domain.Haul
		h.Note(m, at(1, 0, 46))
		next := at(0, 0, 0)
		next.Date += 48
		h.Note(m, next)
		Expect(h.Got).To(BeEmpty())
	})

	It("starts again on another map, and leaves it be where the game does not say the item or the monster is not on the map", func() {
		var h domain.Haul
		h.Note(m, at(1, 0, 46))
		unknown := at(1, -2, 18)
		unknown.FoundKnown = false
		h.Note(m, unknown)
		away := at(1, -2, 18)
		away.Date = 9
		h.Note(m, away)
		Expect(h.Got).To(Equal(map[string][]int{"A": {46}}))

		other := m
		other.Name = "Parepare Jungle"
		october := at(0, 0, 18)
		october.Date = 24
		h.Note(other, october)
		Expect(h.Got).To(BeEmpty())
	})
})
