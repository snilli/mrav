package memo_test

import (
	"errors"
	"sync"
	"sync/atomic"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/infrastructure/memo"
)

var _ = Describe("Cache", func() {
	var c *memo.Cache[int, string]

	BeforeEach(func() { c = &memo.Cache[int, string]{} })

	It("asks for a value the first time only, whatever number of times it is wanted", func() {
		calls := 0
		load := func() (string, error) { calls++; return "one", nil }
		for range 3 {
			got, err := c.Get(1, load)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal("one"))
		}
		Expect(calls).To(Equal(1))
	})

	It("keeps a value for each key", func() {
		got1, _ := c.Get(1, func() (string, error) { return "one", nil })
		got2, _ := c.Get(2, func() (string, error) { return "two", nil })
		again, _ := c.Get(1, func() (string, error) { return "changed", nil })
		Expect([]string{got1, got2, again}).To(Equal([]string{"one", "two", "one"}))
	})

	It("does not keep a failure, so that the next call tries again", func() {
		_, err := c.Get(1, func() (string, error) { return "", errors.New("disk gone") })
		Expect(err).To(MatchError("disk gone"))
		got, err := c.Get(1, func() (string, error) { return "back", nil })
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal("back"))
	})

	It("is ready to use as it is declared, and several callers at once ask once", func() {
		var calls atomic.Int32
		var wg sync.WaitGroup
		for range 8 {
			wg.Go(func() {
				defer GinkgoRecover()
				got, err := c.Get(7, func() (string, error) { calls.Add(1); return "seven", nil })
				Expect(err).NotTo(HaveOccurred())
				Expect(got).To(Equal("seven"))
			})
		}
		wg.Wait()
		Expect(calls.Load()).To(BeEquivalentTo(1))
	})
})
