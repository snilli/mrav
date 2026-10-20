package moveui_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	"app/internal/core/move/domain"
	"app/internal/delivery/ui/moveui"
)

var _ = Describe("NewGridView", func() {
	pat := domain.Move{Col: 1, Row: 1, Name: "Pat", Icon: "Pixie_1"}
	slap := domain.Move{Col: 1, Row: 2, Name: "Slap", Icon: "Pixie_2"}
	bolt := domain.Move{Col: 1, Row: 3, Name: "Bolt", Icon: "Pixie_3"}
	grid := func() domain.Grid {
		g := domain.Grid{Genus: domain.Genus{Code: 0, Name: "Pixie"}}
		g.Cells[0][0] = domain.Cell{Move: &pat, Learned: true, Active: true}
		g.Cells[0][1] = domain.Cell{Move: &slap, CanLearn: true}
		g.Cells[0][2] = domain.Cell{Move: &bolt}
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

	It("gives each cell the kind of its move, and the nature that the move asks for", func() {
		kind := domain.Move{Col: 1, Row: 4, Name: "Zap", Icon: "Pixie_4", Kind: domain.KindIntelligence,
			Unlock: domain.UnlockRule{All: []domain.Condition{domain.NatureAtMost{Max: -20}}}}
		g := grid()
		g.Cells[0][3] = domain.Cell{Move: &kind}
		g.Active[1] = &kind
		v := moveui.NewGridView(g)
		Expect(v.Cells[0][3].Kind).To(Equal(domain.KindIntelligence))
		Expect(v.Cells[0][3].Nature).To(Equal(domain.NatureBad))
		Expect(v.Cells[0][0].Nature).To(Equal(domain.NatureNone))
		Expect(v.ActiveKind[1]).To(Equal(domain.KindIntelligence))
	})

	It("shows a move that is neither learned nor ready to learn as locked", func() {
		Expect(moveui.NewGridView(grid()).Cells[0][2]).To(Equal(moveui.CellView{Icon: "Pixie_3", Locked: true}))
	})

	It("counts the moves learned, ready to learn and not yet, and leaves the empty cells out", func() {
		v := moveui.NewGridView(grid())
		Expect(v.Learned).To(Equal(1))
		Expect(v.Ready).To(Equal(1))
		Expect(v.Locked).To(Equal(1))
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
			Withering: "E (5)", Sharpness: "---", Uses: "40", CanLearn: true, Status: "READY TO LEARN", Description: "Basic tech.",
		}))
	})

	It("gives the nature that the move asks for, to show beside its name", func() {
		good := domain.Move{Name: "Pat", Unlock: domain.UnlockRule{All: []domain.Condition{domain.NatureAtLeast{Min: 50}}}}
		Expect(moveui.NewDetailView(domain.Cell{Move: &good}).Nature).To(Equal(domain.NatureGood))
		plain := domain.Move{Name: "Pat"}
		Expect(moveui.NewDetailView(domain.Cell{Move: &plain}).Nature).To(Equal(domain.NatureNone))
	})

	It("says in words where the move stands: learned, ready to learn, or not yet", func() {
		m := domain.Move{Name: "Slap"}
		Expect(moveui.NewDetailView(domain.Cell{Move: &m, Learned: true}).Status).To(Equal("LEARNED"))
		Expect(moveui.NewDetailView(domain.Cell{Move: &m, CanLearn: true}).Status).To(Equal("READY TO LEARN"))
		Expect(moveui.NewDetailView(domain.Cell{Move: &m}).Status).To(Equal("NOT YET"))
		Expect(moveui.NewDetailView(domain.Cell{}).Status).To(BeEmpty())
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
			{Text: "Pow + Int at least 400", Have: "421", Met: true, Stats: []monster.Stat{monster.Pow, monster.Int}},
			{Text: "Use Pat at least 50 times", Have: "31 / 50"},
			{Text: "The game gives no way to learn it"},
		}})
		Expect(v.Requirements).To(Equal([]moveui.RequirementView{
			{Text: "Pow + Int at least 400", Have: "421", Met: true, Stats: []monster.Stat{monster.Pow, monster.Int}},
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
