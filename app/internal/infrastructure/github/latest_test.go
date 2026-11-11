package github_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/release/domain"
	"app/internal/infrastructure/github"
)

var _ = Describe("Latest", func() {
	serve := func(status int, body string) *httptest.Server {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Expect(r.URL.Path).To(Equal("/repos/snilli/mrav/releases/latest"))
			w.WriteHeader(status)
			_, _ = w.Write([]byte(body))
		}))
		DeferCleanup(srv.Close)
		return srv
	}

	It("reads the tag and the page of the latest release", func() {
		srv := serve(http.StatusOK, `{"tag_name":"v0.5.1","html_url":"https://github.com/snilli/mrav/releases/tag/v0.5.1","name":"x"}`)
		got, err := github.Latest(context.Background(), srv.URL)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(domain.Release{Tag: "v0.5.1", URL: "https://github.com/snilli/mrav/releases/tag/v0.5.1"}))
	})

	It("fails on a status other than OK", func() {
		srv := serve(http.StatusForbidden, `{"message":"rate limit"}`)
		_, err := github.Latest(context.Background(), srv.URL)
		Expect(err).To(MatchError(ContainSubstring("403")))
	})

	It("fails on a body that is not JSON", func() {
		srv := serve(http.StatusOK, `<html>`)
		_, err := github.Latest(context.Background(), srv.URL)
		Expect(err).To(HaveOccurred())
	})

	It("gives up when the context ends", func() {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			<-r.Context().Done()
		}))
		DeferCleanup(srv.Close)
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		DeferCleanup(cancel)
		_, err := github.Latest(ctx, srv.URL)
		Expect(err).To(HaveOccurred())
	})
})
