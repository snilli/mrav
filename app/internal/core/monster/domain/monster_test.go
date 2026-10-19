package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/monster/domain"
)

var _ = Describe("Monster.StatSum", func() {
	m := domain.Monster{Stats: [domain.StatCount]int{
		domain.Life: 100, domain.Pow: 200, domain.Int: 300, domain.Skill: 400, domain.Speed: 500, domain.Defense: 600,
	}}

	DescribeTable("adds the named stats",
		func(stats []domain.Stat, want int) { Expect(m.StatSum(stats...)).To(Equal(want)) },
		Entry("none", []domain.Stat{}, 0),
		Entry("one", []domain.Stat{domain.Pow}, 200),
		Entry("two", []domain.Stat{domain.Int, domain.Speed}, 800),
		Entry("all six", []domain.Stat{domain.Life, domain.Pow, domain.Int, domain.Skill, domain.Speed, domain.Defense}, 2100),
		Entry("a stat past the sixth adds nothing", []domain.Stat{domain.Stat(9), domain.Pow}, 200),
		Entry("a negative stat adds nothing", []domain.Stat{domain.Stat(-1), domain.Pow}, 200),
	)
})
