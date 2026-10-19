package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	"app/internal/core/move/domain"
)

var _ = Describe("NewGrid", func() {
	genus := domain.Genus{Code: 0, Name: "Pixie"}
	moves := func() []domain.Move {
		return []domain.Move{
			{Genus: 0, MemorySlot: 0, Col: 1, Row: 1, Name: "Pat", Icon: "Pixie_1"},
			{Genus: 0, MemorySlot: 1, Col: 1, Row: 2, Name: "Slap", Icon: "Pixie_2",
				Unlock: domain.UnlockRule{All: []domain.Condition{domain.MoveUsedAtLeast{Slot: 0, Min: 30}}}},
			{Genus: 0, MemorySlot: 12, Col: 3, Row: 1, Name: "Bolt", Icon: "Pixie_11"},
			{Genus: 0, MemorySlot: 13, Col: 3, Row: 2, Name: "Lightning", Icon: "Pixie_12",
				Unlock: domain.UnlockRule{All: []domain.Condition{domain.MoveUsedAtLeast{Slot: 12, Min: 50}}}},
		}
	}
	newMonster := func() monster.Monster {
		none := monster.NoActiveMove
		return monster.Monster{Active: [monster.ActiveGroups]int{none, none, none, none}}
	}

	It("places each move in its cell and leaves the others empty", func() {
		g := domain.NewGrid(genus, moves(), newMonster())
		Expect(g.Genus).To(Equal(genus))
		Expect(g.Cells[0][0].Move.Name).To(Equal("Pat"))
		Expect(g.Cells[0][1].Move.Name).To(Equal("Slap"))
		Expect(g.Cells[2][0].Move.Name).To(Equal("Bolt"))
		Expect(g.Cells[2][1].Move.Name).To(Equal("Lightning"))
		Expect(g.Cells[1][0].Move).To(BeNil())
		Expect(g.Cells[3][5].Move).To(BeNil())
	})

	It("reads learned and used from the memory slot of the move, not from its cell", func() {
		m := newMonster()
		m.Learned[13], m.Used[13] = true, 7
		g := domain.NewGrid(genus, moves(), m)
		Expect(g.Cells[2][1].Learned).To(BeTrue())
		Expect(g.Cells[2][1].Used).To(Equal(7))
		Expect(g.Cells[0][1].Learned).To(BeFalse())
		Expect(g.Cells[0][1].Used).To(Equal(0))
	})

	It("lets a monster learn what it has not learned and meets the rule of", func() {
		m := newMonster()
		m.Learned[0] = true
		m.Used[0], m.Used[12] = 30, 49
		g := domain.NewGrid(genus, moves(), m)
		Expect(g.Cells[0][0].CanLearn).To(BeFalse(), "Pat is already learned")
		Expect(g.Cells[0][1].CanLearn).To(BeTrue(), "Slap: Pat used 30 times")
		Expect(g.Cells[2][0].CanLearn).To(BeTrue(), "Bolt has no condition")
		Expect(g.Cells[2][1].CanLearn).To(BeFalse(), "Lightning: Bolt used 49 times, not 50")
	})

	It("finds the active move of each column by its memory slot", func() {
		m := newMonster()
		m.Active = [monster.ActiveGroups]int{0, monster.NoActiveMove, 13, monster.NoActiveMove}
		g := domain.NewGrid(genus, moves(), m)
		Expect(g.Active[0].Name).To(Equal("Pat"))
		Expect(g.Active[1]).To(BeNil())
		Expect(g.Active[2].Name).To(Equal("Lightning"))
		Expect(g.Active[3]).To(BeNil())
		Expect(g.Cells[0][0].Active).To(BeTrue())
		Expect(g.Cells[2][1].Active).To(BeTrue())
		Expect(g.Cells[0][1].Active).To(BeFalse())
	})

	It("ignores an active slot outside the range of its column, as the game does", func() {
		m := newMonster()
		m.Active = [monster.ActiveGroups]int{12, 0, monster.NoActiveMove, monster.NoActiveMove}
		g := domain.NewGrid(genus, moves(), m)
		Expect(g.Active[0]).To(BeNil(), "slot 12 belongs to the third column")
		Expect(g.Active[1]).To(BeNil(), "slot 0 belongs to the first column")
	})

	It("ignores an active slot that no move has", func() {
		m := newMonster()
		m.Active[0] = 5
		Expect(domain.NewGrid(genus, moves(), m).Active[0]).To(BeNil())
	})

	It("leaves out a move with an impossible cell or slot instead of failing", func() {
		bad := []domain.Move{
			{MemorySlot: 0, Col: 9, Row: 1, Name: "Off the grid"},
			{MemorySlot: 40, Col: 1, Row: 1, Name: "Off the slots"},
			{MemorySlot: 2, Col: 1, Row: 3, Name: "Fine"},
		}
		var g domain.Grid
		Expect(func() { g = domain.NewGrid(genus, bad, newMonster()) }).NotTo(Panic())
		Expect(g.Cells[0][2].Move.Name).To(Equal("Fine"))
		Expect(g.Cells[0][0].Move).To(BeNil())
	})
})
