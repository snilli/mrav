package domain_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/profile/domain"
	"app/internal/core/profile/repository/db/profilerepo"
	"app/internal/testutil"
)

var _ = Describe("PlanRoute on the shipped maps", Label("integration"), func() {
	It("costs on every map what walking its path and searching its places costs, with no step back onto a tile it stood on", func() {
		ctx := context.Background()
		client, err := testutil.OpenShipped(ctx)
		if errors.Is(err, testutil.ErrNoShippedDB) {
			Skip(err.Error())
		}
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)
		maps, err := profilerepo.New(client).ListExpeditionMaps(ctx)
		Expect(err).NotTo(HaveOccurred())
		w := domain.Walker{Energy: 700, BreakChance: 1, Intelligence: 300}
		for _, m := range maps {
			m = m.For(domain.Searcher{Intelligence: 300, Skill: 300})
			r, err := domain.PlanRoute(m, w)
			Expect(err).NotTo(HaveOccurred(), m.Name)
			Expect(r.Spots).NotTo(BeEmpty(), m.Name)
			Expect(r.Cost).To(BeNumerically("~", walked(m, w, r), 1e-9), m.Name)
		}
	})
})
