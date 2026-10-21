package moveui

import (
	"errors"

	"fyne.io/fyne/v2/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	"app/internal/core/move/domain"
)

var _ = Describe("Detail", func() {
	BeforeEach(func() { test.NewTempApp(GinkgoTB()) })

	It("shows a move as the Moves page does under its grid: its name, where it stands, and what it still needs", func() {
		d := NewDetail(NewIcons(func(string) ([]byte, error) { return nil, errors.New("no icon") }))
		slap := domain.Move{Name: "Slap", Kind: domain.KindPower, Guts: 18, Description: "Basic tech. Chains from 30 Pats."}

		d.Show(domain.Cell{Move: &slap, CanLearn: true, Requirements: []domain.Requirement{
			{Text: "POW at least 400", Have: "421", Met: true, Stats: []monster.Stat{monster.Pow}},
		}})

		Expect(d.Name()).To(Equal("Slap"))
		Expect(d.d.status.txt.Text).To(Equal("READY TO LEARN"))
		Expect(d.d.guts.Text).To(Equal("18"))
		Expect(d.d.needs).To(HaveLen(1))
		Expect(d.Content()).NotTo(BeNil())
	})

	It("shows another move in place of the first", func() {
		d := NewDetail(NewIcons(func(string) ([]byte, error) { return nil, errors.New("no icon") }))
		slap, kick := domain.Move{Name: "Slap"}, domain.Move{Name: "Kick"}

		d.Show(domain.Cell{Move: &slap})
		d.Show(domain.Cell{Move: &kick, Learned: true})

		Expect(d.Name()).To(Equal("Kick"))
		Expect(d.d.status.txt.Text).To(Equal("LEARNED"))
	})
})
