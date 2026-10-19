package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	player "app/internal/core/player/domain"
	"app/internal/core/profile/domain"
)

var _ = Describe("BreedUnlocks", func() {
	names := map[monster.Genus]string{0: "Pixie", 1: "Dragon", 36: "Worm"}

	It("lists every main breed in the game's order, with whether it is unlocked", func() {
		list := domain.BreedUnlocks(names, [player.BreedCount]bool{0: true, 36: true})
		Expect(list).To(HaveLen(player.BreedCount))
		Expect(list[0]).To(Equal(domain.BreedUnlock{Name: "Pixie", Unlocked: true}))
		Expect(list[1]).To(Equal(domain.BreedUnlock{Name: "Dragon"}))
		Expect(list[36]).To(Equal(domain.BreedUnlock{Name: "Worm", Unlocked: true}))
	})

	It("calls a breed the names lack by its code", func() {
		Expect(domain.BreedUnlocks(names, [player.BreedCount]bool{})[2]).To(Equal(domain.BreedUnlock{Name: "#2"}))
	})
})
