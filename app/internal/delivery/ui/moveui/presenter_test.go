package moveui_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/move/domain"
	"app/internal/delivery/ui/moveui"
)

var _ = Describe("NewGridView", func() {
	pat := domain.Move{Col: 1, Row: 1, Name: "Pat", Icon: "Pixie_1"}
	slap := domain.Move{Col: 1, Row: 2, Name: "Slap", Icon: "Pixie_2"}
	grid := func() domain.Grid {
		g := domain.Grid{Genus: domain.Genus{Code: 0, Name: "Pixie"}}
		g.Cells[0][0] = domain.Cell{Move: &pat, Learned: true, Active: true}
		g.Cells[0][1] = domain.Cell{Move: &slap, CanLearn: true}
		g.Active[0] = &pat
		return g
	}

	It("names the genus", func() {
		Expect(moveui.NewGridView(grid()).Genus).To(Equal("Pixie"))
	})

	It("shows an empty cell with the No Move icon", func() {
		Expect(moveui.NewGridView(grid()).Cells[1][0]).To(Equal(moveui.CellView{Empty: true, Icon: moveui.NoMove}))
	})

	It("shows a move with its icon and what the monster has done with it", func() {
		v := moveui.NewGridView(grid())
		Expect(v.Cells[0][0]).To(Equal(moveui.CellView{Icon: "Pixie_1", Learned: true, Active: true}))
		Expect(v.Cells[0][1]).To(Equal(moveui.CellView{Icon: "Pixie_2", CanLearn: true}))
	})

	It("shows the icon of each column's active move, and No Move where it has none", func() {
		v := moveui.NewGridView(grid())
		Expect(v.Active[0]).To(Equal("Pixie_1"))
		Expect(v.Active[1]).To(Equal(moveui.NoMove))
		Expect(v.Active[3]).To(Equal(moveui.NoMove))
	})
})

var _ = Describe("NewDetailView", func() {
	It("says every value of a move in the game's words", func() {
		m := domain.Move{Name: "Slap", Kind: domain.KindPower, Guts: 18, Force: 15, Hit: 9, Wither: 5, Crit: 0, Description: "Basic tech."}
		v := moveui.NewDetailView(domain.Cell{Move: &m, Used: 40, CanLearn: true})
		Expect(v).To(Equal(moveui.DetailView{
			Name: "Slap", Kind: domain.KindPower, Guts: "18", Damage: "D (15)", Hit: "A (9)",
			Withering: "E (5)", Sharpness: "---", Uses: "40", CanLearn: true, Description: "Basic tech.",
		}))
	})

	It("adds the extra row of a move, as a literal", func() {
		m := domain.Move{Name: "Life Steal", Extra: &domain.Extra{Label: "HP Drain:", Text: "100%"}}
		Expect(moveui.NewDetailView(domain.Cell{Move: &m}).Extra).To(Equal("HP Drain: 100%"))
	})

	It("adds the extra row of a move, rated in words", func() {
		scale := domain.ScaleForce
		m := domain.Move{Name: "Refreshment", Extra: &domain.Extra{Label: "Recovery:", Scale: &scale, Raw: 30}}
		Expect(moveui.NewDetailView(domain.Cell{Move: &m}).Extra).To(Equal("Recovery: B (30)"))
	})

	It("lists what the monster still needs to learn the move, and where it stands on each", func() {
		m := domain.Move{Name: "Slap"}
		v := moveui.NewDetailView(domain.Cell{Move: &m, Requirements: []domain.Requirement{
			{Text: "Pow + Int at least 400", Have: "421", Met: true},
			{Text: "Use Pat at least 50 times", Have: "31 / 50"},
			{Text: "The game gives no way to learn it"},
		}})
		Expect(v.Requirements).To(Equal([]moveui.RequirementView{
			{Text: "Pow + Int at least 400", Have: "421", Met: true},
			{Text: "Use Pat at least 50 times", Have: "31 / 50"},
			{Text: "The game gives no way to learn it"},
		}))
	})

	It("tells a learned move apart", func() {
		m := domain.Move{Name: "Pat"}
		Expect(moveui.NewDetailView(domain.Cell{Move: &m, Learned: true}).Learned).To(BeTrue())
	})

	It("shows an empty cell the way the game does", func() {
		Expect(moveui.NewDetailView(domain.Cell{})).To(Equal(moveui.DetailView{
			Name: "No Move", Kind: domain.KindUnknown, Guts: "---", Damage: "---", Hit: "---",
			Withering: "---", Sharpness: "---", Uses: "---", Description: "Click on a move to learn more.",
		}))
	})
})
