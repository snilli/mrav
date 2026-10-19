package fakeplayer_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	item "app/internal/core/item/domain"
	player "app/internal/core/player/domain"
	"app/internal/infrastructure/fakeplayer"
)

var _ = Describe("Source", func() {
	It("starts as a player with no money, no breeder and an empty place for every item", func() {
		p, err := fakeplayer.New().Player(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(p.Money).To(BeZero())
		Expect(p.Breeder).To(BeEmpty())
		for slot, id := range p.Items {
			Expect(id).To(BeNumerically(">=", item.EmptyFrom), "slot %d", slot)
		}
	})

	It("shows what an update made of the player", func() {
		src := fakeplayer.New()
		src.Update(func(p *player.Player) { p.Money, p.Items[2], p.Breeder = 1500, 29, "Bob" })

		p, err := src.Player(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(p.Money).To(Equal(1500))
		Expect(p.Items[2]).To(Equal(29))
		Expect(p.Breeder).To(Equal("Bob"))
	})

	It("gives a copy, so that a reader cannot change the source's player", func() {
		src := fakeplayer.New()
		p, _ := src.Player(context.Background())
		p.Items[0], p.Money = 5, 99

		again, _ := src.Player(context.Background())
		Expect(again.Items[0]).To(BeNumerically(">=", item.EmptyFrom))
		Expect(again.Money).To(BeZero())
	})
})
