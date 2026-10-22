package itemrepo_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/item"
	"app/internal/core/item/domain"
	"app/internal/core/item/repository/db/itemrepo"
	"app/internal/testutil"
)

var _ = Describe("itemrepo", Label("integration"), func() {
	var ctx context.Context

	BeforeEach(func() { ctx = context.Background() })

	It("lists the items ordered by code, with every field", func() {
		client, err := testutil.NewClient(ctx)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)
		client.Item.Create().SetCode(29).SetName("Magic Banana").SetDesireName("Magic Banana").SetKind("Mysterious Banana").
			SetDescription("Its effect differs each time.").SetEffect("txt_Banana").SetSalePrice("100").SetPurchasePrice("200").SaveX(ctx)
		client.Item.Create().SetCode(0).SetName("Potato").SetDesireName("Potato").SetKind("Animal Foods").
			SetDescription("Some like it.").SetEffect("").SetSalePrice("-----").SetPurchasePrice("10").SaveX(ctx)
		var repo item.ItemRepository = itemrepo.New(client)

		got, err := repo.ListItems(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal([]domain.Item{
			{Code: 0, Name: "Potato", DesireName: "Potato", Kind: "Animal Foods", Description: "Some like it.", SalePrice: "-----", PurchasePrice: "10"},
			{Code: 29, Name: "Magic Banana", DesireName: "Magic Banana", Kind: "Mysterious Banana", Description: "Its effect differs each time.", Effect: "txt_Banana", SalePrice: "100", PurchasePrice: "200"},
		}))
	})

	It("reads the shipped database: the whole list, with its unknown prices kept as they are", func() {
		client, err := testutil.OpenShipped(ctx)
		if errors.Is(err, testutil.ErrNoShippedDB) {
			Skip(err.Error())
		}
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)

		got, err := itemrepo.New(client).ListItems(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(HaveLen(179))
		for i, it := range got {
			Expect(it.Code).To(Equal(i))
		}
		Expect(got[0].Name).To(Equal("Potato"))
		Expect(got[0].Kind).To(Equal("Animal Foods"))
		Expect(got[0].SalePrice).To(Equal("-----"))
		Expect(got[0].PurchasePrice).To(Equal("10"))
		Expect(got[29].SalePrice).To(Equal("100"))
		Expect(got[60].PurchasePrice).To(Equal("???"))
		Expect(got[65].Effect).To(Equal("Decreases Stress by 1."))
		Expect(got[130].DesireName).To(Equal("Nothing"))
		Expect(got[131].DesireName).To(Equal("Battle"))
		Expect(got[172].Name).To(Equal("Reward"))
		Expect(got[178].Name).To(Equal("No item"))
	})
})
