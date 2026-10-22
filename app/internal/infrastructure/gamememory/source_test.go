package gamememory

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/apperr"
	monster "app/internal/core/monster/domain"
)

var _ = Describe("Source", func() {
	var (
		ctx       context.Context
		mem       *image
		attached  int
		attachErr error
		now       time.Time
		src       *Source
	)

	BeforeEach(func() {
		ctx = context.Background()
		mem = monsterImage(ps1MR2AV, ps1Name, nil)
		attached, attachErr = 0, nil
		src = New(ps1, func() (Memory, error) {
			attached++
			if attachErr != nil {
				return nil, attachErr
			}
			return mem, nil
		})
		now = time.Unix(1000, 0)
		src.now = func() time.Time { return now }
	})

	It("attaches the first time it is asked, and reads the monster every time", func() {
		for range 2 {
			m, err := src.Current(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(m.Genus).To(Equal(monster.Genus(12)))
		}
		Expect(attached).To(Equal(1))
	})

	It("reads the player's side with the same attachment", func() {
		_, err := src.Current(ctx)
		Expect(err).NotTo(HaveOccurred())
		p, err := src.Player(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(p.Money).To(Equal(987654))
		Expect(attached).To(Equal(1))
	})

	It("says why it cannot attach, as not found, and tries again once a second has passed", func() {
		attachErr = errors.New("DuckStation is not running")
		_, err := src.Current(ctx)
		Expect(errors.Is(err, apperr.ErrNotFound)).To(BeTrue())
		Expect(reason(err)).To(Equal("DuckStation is not running"))

		attachErr = nil
		now = now.Add(time.Second)
		m, err := src.Current(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(m.Genus).To(Equal(monster.Genus(12)))
		Expect(attached).To(Equal(2))
	})

	It("does not try to attach again within a second of failing, however many pages ask, and answers with the same reason", func() {
		attachErr = errors.New("DuckStation is not running")
		for range 3 {
			_, err := src.Current(ctx)
			Expect(reason(err)).To(Equal("DuckStation is not running"))
			_, err = src.Player(ctx)
			Expect(reason(err)).To(Equal("DuckStation is not running"))
		}
		Expect(attached).To(Equal(1))

		now = now.Add(time.Second - time.Millisecond)
		_, _ = src.Current(ctx)
		Expect(attached).To(Equal(1))

		now = now.Add(time.Millisecond)
		_, _ = src.Current(ctx)
		Expect(attached).To(Equal(2))
	})

	It("attaches at once when a game that was read has been lost, whatever it has been asked before", func() {
		_, err := src.Current(ctx)
		Expect(err).NotTo(HaveOccurred())

		mem.fail = errors.New("access denied")
		_, err = src.Current(ctx)
		Expect(err).To(HaveOccurred())
		mem.fail = nil
		_, err = src.Current(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(attached).To(Equal(2))
	})

	It("lets go of the game when it can no longer be read, and attaches again when it is asked next", func() {
		_, err := src.Current(ctx)
		Expect(err).NotTo(HaveOccurred())

		mem.fail = errors.New("access denied")
		_, err = src.Current(ctx)
		Expect(errors.Is(err, apperr.ErrNotFound)).To(BeTrue())
		Expect(reason(err)).To(ContainSubstring("access denied"))
		Expect(mem.closed).To(Equal(1))

		mem.fail = nil
		_, err = src.Player(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(attached).To(Equal(2))
	})

	It("says there is no monster when the memory holds none, and stays attached", func() {
		mem.ram = make([]byte, ramSize)
		_, err := src.Current(ctx)
		Expect(errors.Is(err, apperr.ErrNotFound)).To(BeTrue())
		Expect(reason(err)).To(Equal("No monster"))
		Expect(attached).To(Equal(1))
		Expect(mem.closed).To(Equal(0))
	})

	It("still reads the player's side when the game shows no monster", func() {
		mem.ram = make([]byte, ramSize)
		mem.i32(ps1MR2AV.money, 500)
		p, err := src.Player(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(p.Money).To(Equal(500))
	})
})
