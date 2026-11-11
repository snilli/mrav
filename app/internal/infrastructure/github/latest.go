package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"app/internal/core/release/domain"
)

const API = "https://api.github.com"

func Latest(ctx context.Context, api string) (domain.Release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, api+"/repos/snilli/mrav/releases/latest", nil)
	if err != nil {
		return domain.Release{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return domain.Release{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return domain.Release{}, fmt.Errorf("latest release: %s", res.Status)
	}
	var body struct {
		Tag string `json:"tag_name"`
		URL string `json:"html_url"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return domain.Release{}, fmt.Errorf("latest release: %w", err)
	}
	return domain.Release{Tag: body.Tag, URL: body.URL}, nil
}
