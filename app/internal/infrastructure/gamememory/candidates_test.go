package gamememory

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("attachFirst", func() {
	noGame := errors.New("DuckStation is not running a game")

	It("reads the second DuckStation when the first runs no game", func() {
		game := &image{}
		var tried []int
		got, err := attachFirst([]int{11, 22}, func(pid int) (Memory, error) {
			tried = append(tried, pid)
			if pid == 11 {
				return nil, noGame
			}
			return game, nil
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(BeIdenticalTo(game))
		Expect(tried).To(Equal([]int{11, 22}))
	})

	It("stops at the first that runs a game", func() {
		var tried []int
		_, err := attachFirst([]int{11, 22}, func(pid int) (Memory, error) {
			tried = append(tried, pid)
			return &image{}, nil
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(tried).To(Equal([]int{11}))
	})

	It("says what the first said when none runs a game", func() {
		_, err := attachFirst([]int{11, 22}, func(pid int) (Memory, error) {
			if pid == 11 {
				return nil, noGame
			}
			return nil, errors.New("cannot read it")
		})
		Expect(err).To(MatchError(noGame))
	})
})
