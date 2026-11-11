package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/release/domain"
)

var _ = Describe("Newer", func() {
	DescribeTable("whether the latest release is worth telling about",
		func(running, latest string, want bool) {
			Expect(domain.Newer(running, latest)).To(Equal(want))
		},
		Entry("a newer patch", "v0.5.0", "v0.5.1", true),
		Entry("a newer minor beats a bigger patch", "v0.4.9", "v0.5.0", true),
		Entry("a newer major", "v0.9.9", "v1.0.0", true),
		Entry("10 is newer than 9, not older", "v0.9.0", "v0.10.0", true),
		Entry("the same release", "v0.5.0", "v0.5.0", false),
		Entry("an older release", "v0.5.1", "v0.5.0", false),
		Entry("a dev build never asks", "dev", "v0.5.1", false),
		Entry("a dirty build never asks", "v0.5.0-dirty", "v0.5.1", false),
		Entry("a build past a tag never asks", "v0.5.0-3-gabc1234", "v0.5.1", false),
		Entry("a bare commit never asks", "abc1234", "v0.5.1", false),
		Entry("a latest tag that is not a release", "v0.5.0", "nightly", false),
		Entry("a pre-release latest", "v0.5.0", "v0.6.0-rc1", false),
		Entry("no latest at all", "v0.5.0", "", false),
	)
})

var _ = Describe("IsRelease", func() {
	DescribeTable("whether a build came from a release tag",
		func(version string, want bool) {
			Expect(domain.IsRelease(version)).To(Equal(want))
		},
		Entry("a release", "v0.5.0", true),
		Entry("a dev build", "dev", false),
		Entry("a dirty build", "v0.5.0-dirty", false),
		Entry("a build past a tag", "v0.5.0-3-gabc1234", false),
	)
})
