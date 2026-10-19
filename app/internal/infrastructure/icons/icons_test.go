package icons_test

import (
	"bytes"
	"context"
	"image/png"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"orm"
	entmove "orm/ent/move"

	"app/internal/infrastructure/icons"
)

var _ = Describe("icons", func() {
	size := func(key string) (int, int) {
		b, err := icons.Bytes(key)
		Expect(err).NotTo(HaveOccurred())
		cfg, err := png.DecodeConfig(bytes.NewReader(b))
		Expect(err).NotTo(HaveOccurred())
		return cfg.Width, cfg.Height
	}

	DescribeTable("serves the overlays at the size the screen draws them",
		func(key string, w, h int) {
			gotW, gotH := size(key)
			Expect([]int{gotW, gotH}).To(Equal([]int{w, h}))
		},
		Entry("an empty cell", "No_Move", 48, 32),
		Entry("the selection", "hilight", 48, 32),
		Entry("learned", "skillchecked", 16, 16),
		Entry("can learn", "skillget", 16, 16),
	)

	It("serves a move icon as a 48x32 PNG", func() {
		w, h := size("Pixie_1")
		Expect([]int{w, h}).To(Equal([]int{48, 32}))
	})

	It("fails for an icon it does not have, naming it", func() {
		_, err := icons.Bytes("Pixie_99")
		Expect(err).To(MatchError(ContainSubstring("Pixie_99")))
	})

	It("has the icon of every move in the imported database", Label("integration"), func() {
		const path = "../../../../data/mrav.sqlite"
		if _, err := os.Stat(path); err != nil {
			Skip("no " + path + ": nothing has been imported yet")
		}
		ctx := context.Background()
		client, err := orm.Open(ctx, path)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)

		keys, err := client.Move.Query().Select(entmove.FieldIcon).Strings(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(keys).NotTo(BeEmpty())
		for _, k := range keys {
			_, err := icons.Bytes(k)
			Expect(err).NotTo(HaveOccurred(), k)
		}
	})
})
