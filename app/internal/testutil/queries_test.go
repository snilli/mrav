package testutil_test

import (
	"context"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/item/repository/db/itemrepo"
	monster "app/internal/core/monster/domain"
	"app/internal/core/move/repository/db/moverepo"
	"app/internal/core/profile/repository/db/profilerepo"
	"app/internal/core/training/repository/db/trainingrepo"
	"app/internal/testutil"
)

type statement struct {
	query string
	args  []any
}

var _ = Describe("the queries of the viewer", Label("integration"), func() {
	It("are all served by an index, on the database that ships, because every page asks for them twice a second", func() {
		ctx := context.Background()
		var sent []statement
		client, err := testutil.OpenShippedTraced(ctx, func(query string, args []any) {
			sent = append(sent, statement{query, args})
		})
		if errors.Is(err, testutil.ErrNoShippedDB) {
			Skip(err.Error())
		}
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)

		moves, items, gains, profiles := moverepo.New(client), itemrepo.New(client), trainingrepo.New(client), profilerepo.New(client)
		askAll := func() {
			genera, err := moves.ListGenera(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(genera).NotTo(BeEmpty())
			_, err = moves.GetGenus(ctx, genera[0].Code)
			Expect(err).NotTo(HaveOccurred())
			for _, version := range []monster.Version{monster.MR2, monster.MR2DX} {
				_, err = moves.ListMoves(ctx, genera[0].Code, version)
				Expect(err).NotTo(HaveOccurred())
			}
			_, err = items.ListItems(ctx)
			Expect(err).NotTo(HaveOccurred())
			_, err = gains.ListGains(ctx)
			Expect(err).NotTo(HaveOccurred())
			_, err = profiles.ListGenusNames(ctx)
			Expect(err).NotTo(HaveOccurred())
			_, err = profiles.ListBreedRules(ctx, int(genera[0].Code))
			Expect(err).NotTo(HaveOccurred())
			_, err = profiles.ListCocoonRules(ctx)
			Expect(err).NotTo(HaveOccurred())
			_, err = profiles.ListDesireNames(ctx)
			Expect(err).NotTo(HaveOccurred())
			maps, err := profiles.ListExpeditionMaps(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(maps).NotTo(BeEmpty())
		}
		askAll()

		Expect(len(sent)).To(BeNumerically(">=", 12), "every repository was asked, and a map takes more than one statement")
		first := len(sent)
		askAll()
		Expect(sent[first:]).To(BeEmpty(), "the data never changes while the viewer runs, so it is read once")

		sent = sent[:first]
		var slow []string
		for _, s := range sent {
			plan, err := testutil.Plan(ctx, client.DB, s.query, s.args...)
			Expect(err).NotTo(HaveOccurred(), s.query)
			if steps := testutil.Unindexed(s.query, plan); len(steps) > 0 {
				slow = append(slow, fmt.Sprintf("%s\n    not served: %v\n    plan: %v", s.query, steps, plan))
			}
		}
		Expect(slow).To(BeEmpty())
	})
})
