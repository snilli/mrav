package domain_test

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"os"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	"app/internal/core/move/repository/db/moverepo"
	"app/internal/core/profile/domain"
	"app/internal/testutil"
)

type golden struct {
	Name    string
	Genus   int
	Sub     int
	Stats   [monster.StatCount]int
	Spoil   int
	Fear    int
	Nature  int
	Learned []int
	Used    map[string]int
	Places  map[string][][2]float64
}

var _ = Describe("LearnChances against the calculator of legendcup.com", Label("integration"), func() {
	It("lists what the page lists, with the chances it gives, for the moves of the shipped database", func() {
		ctx := context.Background()
		client, err := testutil.OpenShipped(ctx)
		if errors.Is(err, testutil.ErrNoShippedDB) {
			Skip(err.Error())
		}
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)
		repo := moverepo.New(client)

		raw, err := os.ReadFile("testdata/learn_golden.json")
		Expect(err).NotTo(HaveOccurred())
		var file struct{ Cases []golden }
		Expect(json.Unmarshal(raw, &file)).To(Succeed())
		Expect(file.Cases).NotTo(BeEmpty())

		for _, c := range file.Cases {
			moves, err := repo.ListMoves(ctx, monster.Genus(c.Genus), monster.MR2DX)
			Expect(err).NotTo(HaveOccurred(), c.Name)
			m := monster.Monster{
				Genus: monster.Genus(c.Genus), SubGenus: monster.Genus(c.Sub),
				Stats: c.Stats, Spoil: c.Spoil, Fear: c.Fear, Nature: c.Nature,
			}
			for _, slot := range c.Learned {
				m.Learned[slot] = true
			}
			for slot, uses := range c.Used {
				n, _ := strconv.Atoi(slot)
				m.Used[n] = uses
			}

			got := domain.LearnChances(m, moves)
			for place := 1; place <= len(got); place++ {
				var have [][2]float64
				for _, l := range got[place-1] {
					have = append(have, [2]float64{float64(l.Move.MemorySlot), math.Round(l.Chance*10) / 10})
				}
				Expect(have).To(Equal(c.Places[strconv.Itoa(place)]), "%s, place %d", c.Name, place)
			}
		}
	})
})
