package config_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/config"
)

var _ = Describe("Load", func() {
	unset := func(name string) {
		if v, ok := os.LookupEnv(name); ok {
			Expect(os.Unsetenv(name)).To(Succeed())
			DeferCleanup(os.Setenv, name, v)
		}
	}

	It("defaults the log level and the database path", func() {
		unset("LOG_LEVEL")
		unset("DB_PATH")
		cfg, err := config.Load()
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.LogLevel).To(Equal("info"))
		Expect(cfg.DBPath).To(Equal("../data/mrav.sqlite"))
	})

	It("refreshes every half second unless told otherwise, and refuses a rate too fast to be useful", func() {
		unset("REFRESH_MS")
		cfg, err := config.Load()
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.RefreshMS).To(Equal(500))

		GinkgoT().Setenv("REFRESH_MS", "10")
		_, err = config.Load()
		Expect(err).To(HaveOccurred())
	})

	It("reads whichever game is running unless told otherwise, and knows only the games it can read", func() {
		unset("GAME")
		cfg, err := config.Load()
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.Game).To(Equal("auto"))

		for _, game := range []string{"demo", "duckstation", "mr2dx", "auto"} {
			GinkgoT().Setenv("GAME", game)
			cfg, err = config.Load()
			Expect(err).NotTo(HaveOccurred(), game)
			Expect(cfg.Game).To(Equal(game))
		}

		GinkgoT().Setenv("GAME", "epsxe")
		_, err = config.Load()
		Expect(err).To(HaveOccurred())
	})

	It("reads the database path from the environment", func() {
		GinkgoT().Setenv("DB_PATH", "elsewhere.sqlite")
		cfg, err := config.Load()
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.DBPath).To(Equal("elsewhere.sqlite"))
	})
})
