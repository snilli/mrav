package gamememory

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/apperr"
	monster "app/internal/core/monster/domain"
)

var _ = Describe("Auto", func() {
	var (
		ctx                 context.Context
		first, second       *image
		firstErr, secondErr error
		auto                *Auto
	)

	BeforeEach(func() {
		ctx = context.Background()
		first, second = monsterImage(ps1MR2AV, ps1Name, nil), monsterImage(ps1MR2AV, ps1Name, nil)
		second.i32(ps1MR2AV.money, 500) // so that the two games can be told apart: the first has 987654
		firstErr, secondErr = nil, nil
		attach := func(mem *image, err *error) Attach {
			return func() (Memory, error) {
				if *err != nil {
					return nil, *err
				}
				return mem, nil
			}
		}
		auto = NewAuto(New(ps1, attach(first, &firstErr)), New(ps1, attach(second, &secondErr)))
	})

	It("reads the first game when it is running", func() {
		p, err := auto.Player(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(p.Money).To(Equal(987654))
		m, err := auto.Current(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(m.Genus).To(Equal(monster.Genus(12)))
	})

	It("reads the next game when the first is not running", func() {
		firstErr = errors.New("DuckStation is not running")
		p, err := auto.Player(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(p.Money).To(Equal(500))
		_, err = auto.Current(ctx)
		Expect(err).NotTo(HaveOccurred())
	})

	It("moves on to the next game when the one it was reading is lost", func() {
		_, err := auto.Player(ctx)
		Expect(err).NotTo(HaveOccurred())

		first.fail = errors.New("access denied")
		p, err := auto.Player(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(p.Money).To(Equal(500))
	})

	It("stays with a game it is attached to, even when that game shows no monster", func() {
		first.ram = make([]byte, ramSize)
		_, err := auto.Current(ctx)
		Expect(errors.Is(err, apperr.ErrNotFound)).To(BeTrue())
		Expect(reason(err)).To(Equal("No monster")) // not what the second game, which has one, would say
	})

	It("says why no game could be read, naming each", func() {
		firstErr, secondErr = errors.New("DuckStation is not running"), errors.New("MR2DX is not running")
		_, err := auto.Current(ctx)
		Expect(errors.Is(err, apperr.ErrNotFound)).To(BeTrue())
		Expect(reason(err)).To(Equal("DuckStation is not running; MR2DX is not running"))
		_, err = auto.Player(ctx)
		Expect(reason(err)).To(Equal("DuckStation is not running; MR2DX is not running"))
	})
})
