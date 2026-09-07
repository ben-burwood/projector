package provider

import (
	"context"
	"net/url"
	"time"
)

type cratesProvider struct{ hc *httpClient }

func newCrates(hc *httpClient) *cratesProvider { return &cratesProvider{hc: hc} }

func (c *cratesProvider) Type() SourceType { return Crates }

func (c *cratesProvider) Detect(u *url.URL) (Ref, bool) {
	if hostname(u) != "crates.io" {
		return Ref{}, false
	}
	parts := pathParts(u)
	if len(parts) < 2 || parts[0] != "crates" {
		return Ref{}, false
	}
	return Ref{Type: Crates, Identifier: parts[1]}, true
}

func (c *cratesProvider) apiURL(name string) string {
	return "https://crates.io/api/v1/crates/" + url.PathEscape(name)
}

type cratesResponse struct {
	Crate struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Homepage    string `json:"homepage"`
		Repository  string `json:"repository"`
	} `json:"crate"`
	Versions []struct {
		Num       string    `json:"num"`
		CreatedAt time.Time `json:"created_at"`
		Yanked    bool      `json:"yanked"`
	} `json:"versions"`
}

func (c *cratesProvider) get(ctx context.Context, ref Ref, etag, lastMod string) (cratesResponse, condResponse, error) {
	res, err := c.hc.getCond(ctx, c.apiURL(ref.Identifier), etag, lastMod, nil)
	if err != nil {
		return cratesResponse{}, res, err
	}
	if res.NotModified {
		return cratesResponse{}, res, nil
	}
	var raw cratesResponse
	if err := decodeJSON(res, &raw); err != nil {
		return cratesResponse{}, res, err
	}
	return raw, res, nil
}

func (c *cratesProvider) Prefill(ctx context.Context, ref Ref) (Metadata, error) {
	raw, _, err := c.get(ctx, ref, "", "")
	if err != nil {
		return Metadata{}, err
	}
	home := raw.Crate.Homepage
	if home == "" {
		home = raw.Crate.Repository
	}
	if home == "" {
		home = "https://crates.io/crates/" + ref.Identifier
	}
	name := raw.Crate.Name
	if name == "" {
		name = ref.Identifier
	}
	return Metadata{Name: name, Description: raw.Crate.Description, HomeURL: home}, nil
}

func (c *cratesProvider) Fetch(ctx context.Context, req FetchRequest) (FetchResult, error) {
	raw, res, err := c.get(ctx, req.Ref, req.ETag, req.LastModified)
	if err != nil {
		return FetchResult{}, err
	}
	if res.NotModified {
		return FetchResult{NotModified: true, ETag: res.ETag, LastModified: res.LastModified}, nil
	}
	limit := clampLimit(req.Limit)
	out := make([]Release, 0, limit)
	for _, v := range raw.Versions {
		if v.Yanked {
			continue
		}
		out = append(out, Release{
			ExternalID:  v.Num,
			Version:     v.Num,
			Title:       raw.Crate.Name + " " + v.Num,
			URL:         "https://crates.io/crates/" + raw.Crate.Name + "/" + v.Num,
			PublishedAt: v.CreatedAt.UTC(),
		})
		if len(out) >= limit {
			break
		}
	}
	return FetchResult{ETag: res.ETag, LastModified: res.LastModified, Releases: out}, nil
}
