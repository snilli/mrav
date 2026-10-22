package trainingrepo_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/training"
	"app/internal/core/training/domain"
	"app/internal/core/training/repository/db/trainingrepo"
	"app/internal/testutil"
)

var _ = Describe("trainingrepo", Label("integration"), func() {
	var ctx context.Context

	BeforeEach(func() { ctx = context.Background() })

	It("lists the gains ordered by life stage, growth rate and kind", func() {
		client, err := testutil.NewClient(ctx)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)
		client.TrainingGain.Create().SetLifeStage(5).SetGrowthRate(1).SetKind(3).SetText("b").SaveX(ctx)
		client.TrainingGain.Create().SetLifeStage(0).SetGrowthRate(4).SetKind(0).SetText("a").SaveX(ctx)
		client.TrainingGain.Create().SetLifeStage(5).SetGrowthRate(1).SetKind(1).SetText("c").SaveX(ctx)
		var repo training.TrainingRepository = trainingrepo.New(client)

		got, err := repo.ListGains(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal([]domain.Gain{
			{LifeStage: 0, GrowthRate: 4, Kind: 0, Text: "a"},
			{LifeStage: 5, GrowthRate: 1, Kind: 1, Text: "c"},
			{LifeStage: 5, GrowthRate: 1, Kind: 3, Text: "b"},
		}))
	})

	It("lists none from an empty table, without failing", func() {
		client, err := testutil.NewClient(ctx)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)

		got, err := trainingrepo.New(client).ListGains(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(BeEmpty())
	})

	It("reads the shipped database: every life stage, growth rate and kind has its text", func() {
		client, err := testutil.OpenShipped(ctx)
		if errors.Is(err, testutil.ErrNoShippedDB) {
			Skip(err.Error())
		}
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)

		got, err := trainingrepo.New(client).ListGains(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(HaveLen(10 * 5 * 3))
		gains := domain.NewGains(got)
		Expect(gains.Text(0, 0, domain.KindLight)).To(Equal("1-3 (1/1/2/3)"))
		Expect(gains.Text(9, 0, domain.KindHeavy)).To(Equal("3-4 (3/3/3/4)"))
		Expect(gains.Text(4, 4, domain.KindHeavy)).To(Equal("20 (20/20/20/20)"))
		Expect(gains.Text(6, 4, domain.KindHeavy)).To(Equal("19-20 (19/20/20/20)"))
	})
})
