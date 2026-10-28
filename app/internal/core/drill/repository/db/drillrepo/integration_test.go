package drillrepo_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"orm/ent/drillgain"

	"app/internal/core/drill"
	"app/internal/core/drill/domain"
	"app/internal/core/drill/repository/db/drillrepo"
	"app/internal/testutil"
)

var _ = Describe("drillrepo", Label("integration"), func() {
	var ctx context.Context

	BeforeEach(func() { ctx = context.Background() })

	It("lists the cells ordered by result, role, bonus, life stage and growth rate, with their choices", func() {
		client, err := testutil.NewClient(ctx)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)
		client.DrillGain.Create().SetResult(drillgain.ResultSuccess).SetRole(drillgain.RolePrimary).SetBonus(true).SetLifeStage(5).SetGrowthRate(1).SetSlots([][]int{{3}, {4}}).SaveX(ctx)
		client.DrillGain.Create().SetResult(drillgain.ResultCheat).SetRole(drillgain.RoleLight).SetBonus(false).SetLifeStage(0).SetGrowthRate(4).SetSlots([][]int{{1}, {2}}).SaveX(ctx)
		client.DrillGain.Create().SetResult(drillgain.ResultSuccess).SetRole(drillgain.RolePrimary).SetBonus(false).SetLifeStage(5).SetGrowthRate(1).SetSlots([][]int{{5, 6}}).SaveX(ctx)
		var repo drill.DrillRepository = drillrepo.New(client)

		got, err := repo.ListGains(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal([]domain.Cell{
			{Result: domain.Cheat, Role: domain.Light, Bonus: false, LifeStage: 0, GrowthRate: 4, Slots: [][]int{{1}, {2}}},
			{Result: domain.Success, Role: domain.Primary, Bonus: false, LifeStage: 5, GrowthRate: 1, Slots: [][]int{{5, 6}}},
			{Result: domain.Success, Role: domain.Primary, Bonus: true, LifeStage: 5, GrowthRate: 1, Slots: [][]int{{3}, {4}}},
		}))
	})

	It("lists none from an empty table, without failing", func() {
		client, err := testutil.NewClient(ctx)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)

		got, err := drillrepo.New(client).ListGains(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(BeEmpty())
	})

	It("reads the shipped database: 1050 cells, a pair where the page gives a pair, and the drop below zero", func() {
		client, err := testutil.OpenShipped(ctx)
		if errors.Is(err, testutil.ErrNoShippedDB) {
			Skip(err.Error())
		}
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)

		got, err := drillrepo.New(client).ListGains(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(HaveLen(1050))
		gains := domain.NewGains(got)
		Expect(gains.Slots(domain.Success, domain.Light, false, 0, 0)).To(Equal([][]int{{1}, {1}, {2}, {3}}))
		Expect(gains.Slots(domain.Great, domain.Primary, false, 0, 0)).To(Equal([][]int{{4, 8}, {6, 10}}))
		Expect(gains.Slots(domain.Success, domain.Drop, false, 9, 4)).To(Equal([][]int{{-2}, {-3}, {-2}, {-3}}))
	})
})
