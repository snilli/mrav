package itemrepo_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	entitem "orm/ent/item"

	"app/internal/core/item"
	"app/internal/core/item/domain"
	"app/internal/core/item/repository/db/itemrepo"
	monster "app/internal/core/monster/domain"
	"app/internal/testutil"
)

var _ = Describe("itemrepo", Label("integration"), func() {
	var ctx context.Context

	BeforeEach(func() { ctx = context.Background() })

	It("lists the items ordered by code, with every field", func() {
		client, err := testutil.NewClient(ctx)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)
		client.Item.Create().SetVersion(entitem.VersionMr2dx).SetCode(29).SetName("Magic Banana").SetDesireName("Magic Banana").SetKind("Mysterious Banana").
			SetDescription("Its effect differs each time.").SetEffect("txt_Banana").SetSalePrice(100).SetPurchasePrice(200).SaveX(ctx)
		client.Item.Create().SetVersion(entitem.VersionMr2dx).SetCode(0).SetName("Potato").SetDesireName("Potato").SetKind("Animal Foods").
			SetDescription("Some like it.").SetEffect("").SetPurchasePrice(10).SaveX(ctx)
		var repo item.ItemRepository = itemrepo.New(client)

		got, err := repo.ListItems(ctx, monster.MR2DX)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal([]domain.Item{
			{Code: 0, Name: "Potato", DesireName: "Potato", Kind: "Animal Foods", Description: "Some like it.", PurchasePrice: new(10)},
			{Code: 29, Name: "Magic Banana", DesireName: "Magic Banana", Kind: "Mysterious Banana", Description: "Its effect differs each time.", Effect: "txt_Banana", SalePrice: new(100), PurchasePrice: new(200)},
		}))
	})

	It("lists the items of the game it is asked for, and none of the other game's", func() {
		client, err := testutil.NewClient(ctx)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)
		for _, v := range []entitem.Version{entitem.VersionMr2, entitem.VersionMr2dx} {
			client.Item.Create().SetVersion(v).SetCode(6).SetName("Sculpture").SetDesireName("Sculpture").SetKind("Sculpture of Artemis").
				SetDescription("").SetEffect("Sculpture " + string(v)).SetSalePrice(1500).SaveX(ctx)
		}
		repo := itemrepo.New(client)

		mr2, err := repo.ListItems(ctx, monster.MR2)
		Expect(err).NotTo(HaveOccurred())
		mr2dx, err := repo.ListItems(ctx, monster.MR2DX)
		Expect(err).NotTo(HaveOccurred())
		Expect(mr2).To(HaveLen(1))
		Expect(mr2[0].Effect).To(Equal("Sculpture mr2"))
		Expect(mr2dx).To(HaveLen(1))
		Expect(mr2dx[0].Effect).To(Equal("Sculpture mr2dx"))
	})

	It("reads the shipped database: the whole list of each game, a price a number or nothing", func() {
		client, err := testutil.OpenShipped(ctx)
		if errors.Is(err, testutil.ErrNoShippedDB) {
			Skip(err.Error())
		}
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)
		repo := itemrepo.New(client)

		for _, version := range []monster.Version{monster.MR2, monster.MR2DX} {
			got, err := repo.ListItems(ctx, version)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(HaveLen(179), string(version))
			for i, it := range got {
				Expect(it.Code).To(Equal(i))
			}
			Expect(got[0].Name).To(Equal("Potato"))
			Expect(got[0].Kind).To(Equal("Animal Foods"))
			Expect(got[0].SalePrice).To(BeNil())
			Expect(got[0].PurchasePrice).To(HaveValue(Equal(10)))
			Expect(got[29].SalePrice).To(HaveValue(Equal(100)))
			Expect(got[60].PurchasePrice).To(BeNil(), "the source did not know it")
			Expect(got[65].Effect).To(Equal("Decreases Stress by 1."))
			Expect(got[130].DesireName).To(Equal("Nothing"))
			Expect(got[131].DesireName).To(Equal("Battle"))
			Expect(got[172].Name).To(Equal("Reward"))
			Expect(got[178].Name).To(Equal("No item"))

			var sale, purchase int
			for _, it := range got {
				if it.SalePrice != nil {
					sale++
				}
				if it.PurchasePrice != nil {
					purchase++
				}
			}
			Expect(sale).To(Equal(149), "the items that have a sale price, "+string(version))
			Expect(purchase).To(Equal(56), "the items that have a purchase price, "+string(version))
		}
	})

	It("reads on the shipped database the items whose effect MR2DX changed", func() {
		client, err := testutil.OpenShipped(ctx)
		if errors.Is(err, testutil.ErrNoShippedDB) {
			Skip(err.Error())
		}
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)
		repo := itemrepo.New(client)
		mr2, err := repo.ListItems(ctx, monster.MR2)
		Expect(err).NotTo(HaveOccurred())
		mr2dx, err := repo.ListItems(ctx, monster.MR2DX)
		Expect(err).NotTo(HaveOccurred())

		Expect(mr2[6].Effect).To(Equal("-3% Fatigue (Floored) and +1 Spoil at the start of each month."))
		Expect(mr2dx[6].Effect).To(Equal("-3% Fatigue (Floored) at the start of each month."))
		Expect(mr2[7].Effect).To(Equal("-3% Stress (Floored) and +1 Fear at the start of each month."))
		Expect(mr2dx[7].Effect).To(Equal("-3% Stress (Floored) and +1 Fear at the start of each month. When 3% is less than 1, it still gives -1 Stress."))
		Expect(mr2[28].Effect).To(ContainSubstring("+1 week lifespan"))
		Expect(mr2dx[28].Effect).To(ContainSubstring("-30 Fatigue"))
		Expect(mr2dx[28].Effect).NotTo(ContainSubstring("lifespan"))
		Expect(mr2[85].Effect).To(ContainSubstring("ForwardGolem"))
		Expect(mr2dx[85].Effect).To(ContainSubstring("Dream Golem"))
		for _, code := range []int{156, 163} {
			Expect(mr2dx[code].Effect).To(ContainSubstring("bugged"), mr2dx[code].Name)
			Expect(mr2[code].Effect).NotTo(ContainSubstring("bugged"), mr2[code].Name)
		}
		for code := range mr2 {
			if code != 6 && code != 7 && code != 28 && code != 29 && code != 30 && code != 85 && code != 102 && code != 156 && code != 163 {
				Expect(mr2dx[code]).To(Equal(mr2[code]), mr2[code].Name)
			}
		}
	})

	It("ships no item with a placeholder for its effect, and Griever and Nageel as the item list of LegendCup has them", func() {
		client, err := testutil.OpenShipped(ctx)
		if errors.Is(err, testutil.ErrNoShippedDB) {
			Skip(err.Error())
		}
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)
		repo := itemrepo.New(client)

		for _, version := range []monster.Version{monster.MR2, monster.MR2DX} {
			got, err := repo.ListItems(ctx, version)
			Expect(err).NotTo(HaveOccurred())
			for _, it := range got {
				Expect(it.Effect).NotTo(BeElementOf("[effect]", "[find what this does]"), string(version)+" "+it.Name)
			}
			Expect(got[33].Effect).To(ContainSubstring("40%"))
			Expect(got[39].Effect).To(ContainSubstring("-20 weeks"))
		}
	})
})
