package errantryui

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	movedomain "app/internal/core/move/domain"
	"app/internal/core/profile/domain"
	"app/internal/delivery/ui/profileui"
)

func baby() domain.Profile {
	m := monster.Monster{Lifespan: 100, Growth: [monster.StatCount]int{2, 2, 2, 2, 2, 2}}
	m.Learned[4] = true
	return domain.Profile{Monster: m}
}

func gate(stat monster.Stat) []movedomain.Gate {
	return []movedomain.Gate{{Stats: []monster.Stat{stat}, Min: 250, Max: 400}}
}

var (
	ray     = movedomain.Move{MemorySlot: 1, Name: "Ray", Icon: "Pixie_1", Description: "Sharp tech. INT should be over 250 to unlock.", Acquire: []movedomain.Acquire{{Place: 1, Priority: 1, Auto: 3, Gates: gate(monster.Int)}}}
	slap    = movedomain.Move{MemorySlot: 2, Name: "Slap", Icon: "Pixie_2", Description: "Basic tech. Chains from 30 Pats.", Acquire: []movedomain.Acquire{{Place: 3, Priority: 1, Chain: &movedomain.Chain{Slot: 9, Uses: 30}}}}
	flame   = movedomain.Move{MemorySlot: 3, Name: "Flame", Icon: "Pixie_3", Description: "Heavy tech. POW should be over 250 to unlock.", Acquire: []movedomain.Acquire{{Place: 2, Priority: 1, Auto: 3, Gates: gate(monster.Pow)}}}
	bang    = movedomain.Move{MemorySlot: 4, Name: "Bang", Icon: "Pixie_4", Description: "Special tech. INT + SPD should total over 800 to unlock.", Acquire: []movedomain.Acquire{{Place: 5, Priority: 1}}}
	megaray = movedomain.Move{MemorySlot: 5, Name: "Megaray", Icon: "Pixie_5", Description: "Sharp tech. Chains from 50 Rays.", Acquire: []movedomain.Acquire{{Place: 1, Priority: 2, Chain: &movedomain.Chain{Slot: 1, Uses: 50}}}}
)

func grid() movedomain.Grid {
	var g movedomain.Grid
	g.Cells[0][0] = movedomain.Cell{Move: &ray, CanLearn: true}
	g.Cells[0][1] = movedomain.Cell{Move: &slap}
	g.Cells[1][0] = movedomain.Cell{Move: &flame}
	g.Cells[1][1] = movedomain.Cell{Move: &megaray}
	g.Cells[2][0] = movedomain.Cell{Move: &bang, Learned: true}
	return g
}

var _ = Describe("NewView", func() {
	It("has a row for each place, in the order of the game, with the stat that it trains", func() {
		v := NewView(baby(), movedomain.Grid{})

		Expect(v.Rows).To(HaveLen(5))
		Expect(v.Rows[0].Place).To(Equal("Papas Mountain"))
		Expect(v.Rows[0].Stat).To(Equal(monster.Speed))
		Expect(v.Rows[4].Place).To(Equal("Kawrea Volcano"))
		Expect(v.Rows[4].Stat).To(Equal(monster.Defense))
	})

	It("gives the chance of each week in whole percent, on a tone that gets warmer as the chance falls", func() {
		row := NewView(baby(), movedomain.Grid{}).Rows[0]

		Expect(row.Weeks).To(Equal([4]profileui.Box{
			{Text: "100%", Tone: profileui.ToneLimeGreen},
			{Text: "69%", Tone: profileui.ToneLightGreen},
			{Text: "21%", Tone: profileui.ToneOrange},
			{Text: "0%", Tone: profileui.ToneOrangeRed},
		}))
	})

	It("says what the trip would give: the stat of the place and the Life, each named in its colour", func() {
		row := NewView(baby(), movedomain.Grid{}).Rows[0]

		var text string
		colours := map[string]any{}
		for _, seg := range row.Gains {
			text += seg.Text
			colours[seg.Text] = seg.Color
		}
		Expect(text).To(Equal("SPD +2-5 · LIF +3"))
		Expect(colours).To(HaveKeyWithValue("SPD", profileui.StatColors[monster.Speed]))
		Expect(colours).To(HaveKeyWithValue("LIF", profileui.StatColors[monster.Life]))
	})

	names := func(row RowView) []string {
		var out []string
		for _, r := range row.Ready {
			out = append(out, r.Cell.Move.Name+" "+r.Chance.Text)
		}
		return out
	}

	It("lists, for each place, the moves that it may teach with their chances, and none that are learned or chained from a move not used enough", func() {
		v := NewView(baby(), grid())

		Expect(names(v.Rows[0])).To(Equal([]string{"Ray 3%"}), "Megaray needs Ray used 50 times")
		Expect(names(v.Rows[1])).To(Equal([]string{"Flame 3%"}))
		Expect(names(v.Rows[2])).To(BeEmpty(), "Slap needs the move in slot 9 used 30 times")
		Expect(names(v.Rows[3])).To(BeEmpty())
		Expect(names(v.Rows[4])).To(BeEmpty(), "Bang is learned already")
	})

	It("lists a basic move too, once the move that it chains from has been used enough, and gives it the chance that is left", func() {
		p := baby()
		p.Monster.Used[9] = 30

		Expect(names(NewView(p, grid()).Rows[2])).To(Equal([]string{"Slap 100%"}))
	})

	It("gives the cell of the grid with each move, which says where it stands", func() {
		Expect(NewView(baby(), grid()).Rows[0].Ready[0].Cell).To(Equal(grid().Cells[0][0]))
	})

	It("tones the chance of a move like the chance of a week", func() {
		p := baby()
		p.Monster.Stats[monster.Int] = 400
		Expect(NewView(p, grid()).Rows[0].Ready[0].Chance).To(Equal(profileui.Box{Text: "100%", Tone: profileui.ToneLimeGreen}))
	})

	It("leaves out a move whose chance is nothing to one decimal", func() {
		none := movedomain.Move{MemorySlot: 6, Name: "None", Acquire: []movedomain.Acquire{{Place: 4, Priority: 1, Gates: gate(monster.Int)}}}
		var g movedomain.Grid
		g.Cells[0][0] = movedomain.Cell{Move: &none}

		Expect(names(NewView(baby(), g).Rows[3])).To(BeEmpty())
	})

	DescribeTable("tones a chance by how likely it is",
		func(percent float64, want profileui.Tone) { Expect(chanceTone(percent)).To(Equal(want)) },
		Entry("sure", 100.0, profileui.ToneLimeGreen),
		Entry("likely", 60.0, profileui.ToneLightGreen),
		Entry("even", 59.9, profileui.ToneLightYellow),
		Entry("even, at the low end", 30.0, profileui.ToneLightYellow),
		Entry("unlikely", 29.9, profileui.ToneOrange),
		Entry("barely", 3.4, profileui.ToneOrange),
		Entry("none", 0.0, profileui.ToneOrangeRed),
	)
})
