package database

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("resolve", func() {
	var dir, path string
	built := []byte("the database that was built in")

	BeforeEach(func() {
		dir = filepath.Join(GinkgoT().TempDir(), "cache")
		path = filepath.Join(GinkgoT().TempDir(), "mrav.sqlite")
	})

	It("opens the file at the path when there is one, and writes nothing", func() {
		Expect(os.WriteFile(path, []byte("mine"), 0o644)).To(Succeed())
		got, err := resolve(path, built, dir)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(path))
		Expect(dir).NotTo(BeADirectory())
	})

	It("writes the database that was built in, and opens that, when there is no file at the path", func() {
		got, err := resolve(path, built, dir)
		Expect(err).NotTo(HaveOccurred())
		Expect(filepath.Dir(got)).To(Equal(dir))
		Expect(os.ReadFile(got)).To(Equal(built))
		Expect(path).NotTo(BeAnExistingFile())
	})

	It("names the file after what is in it, so that a program with other data does not share it", func() {
		a, err := resolve(path, built, dir)
		Expect(err).NotTo(HaveOccurred())
		b, err := resolve(path, []byte("another build's data"), dir)
		Expect(err).NotTo(HaveOccurred())
		Expect(a).NotTo(Equal(b))
	})

	It("leaves the file it wrote alone the next time, which another copy of the viewer may have open", func() {
		got, err := resolve(path, built, dir)
		Expect(err).NotTo(HaveOccurred())
		before, err := os.Stat(got)
		Expect(err).NotTo(HaveOccurred())

		again, err := resolve(path, built, dir)
		Expect(err).NotTo(HaveOccurred())
		Expect(again).To(Equal(got))
		after, err := os.Stat(again)
		Expect(err).NotTo(HaveOccurred())
		Expect(os.SameFile(before, after)).To(BeTrue())
		Expect(after.ModTime()).To(Equal(before.ModTime()))
	})

	It("writes it again when what is there is cut short, as a write that was interrupted leaves it", func() {
		got, err := resolve(path, built, dir)
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(got, built[:5], 0o644)).To(Succeed())

		again, err := resolve(path, built, dir)
		Expect(err).NotTo(HaveOccurred())
		Expect(os.ReadFile(again)).To(Equal(built))
	})

	It("gives back the path when nothing was built in, so that Open says what is missing", func() {
		got, err := resolve(path, nil, dir)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(path))
		Expect(dir).NotTo(BeADirectory())
	})

	It("says so when the built-in database cannot be written", func() {
		blocker := filepath.Join(GinkgoT().TempDir(), "file")
		Expect(os.WriteFile(blocker, nil, 0o644)).To(Succeed())
		_, err := resolve(path, built, filepath.Join(blocker, "cache"))
		Expect(err).To(MatchError(ContainSubstring("built-in database")))
	})
})
