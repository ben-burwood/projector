package provider

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type dockerHubProvider struct{ hc *httpClient }

func newDockerHub(hc *httpClient) *dockerHubProvider { return &dockerHubProvider{hc: hc} }

func (d *dockerHubProvider) Type() SourceType { return DockerHub }

func (d *dockerHubProvider) Detect(u *url.URL) (Ref, bool) {
	host := hostname(u)
	if host != "hub.docker.com" && host != "docker.io" && host != "registry.hub.docker.com" {
		return Ref{}, false
	}
	parts := pathParts(u)
	// hub.docker.com/r/{ns}/{img}, hub.docker.com/_/{img} (official)
	var id string
	switch {
	case len(parts) >= 3 && parts[0] == "r":
		id = parts[1] + "/" + parts[2]
	case len(parts) >= 2 && parts[0] == "_":
		id = "library/" + parts[1]
	case len(parts) >= 3 && parts[0] == "layers":
		id = parts[1] + "/" + parts[2]
	case len(parts) == 1:
		id = "library/" + parts[0]
	case len(parts) >= 2:
		id = parts[0] + "/" + parts[1]
	default:
		return Ref{}, false
	}
	return Ref{Type: DockerHub, Identifier: id}, true
}

func (d *dockerHubProvider) repoAPI(id string) string {
	return "https://hub.docker.com/v2/repositories/" + id
}

type dhRepo struct {
	Description     string `json:"description"`
	FullDescription string `json:"full_description"`
}

func (d *dockerHubProvider) Prefill(ctx context.Context, ref Ref) (Metadata, error) {
	res, err := d.hc.getCond(ctx, d.repoAPI(ref.Identifier), "", "", nil)
	if err != nil {
		return Metadata{}, err
	}
	var r dhRepo
	if err := decodeJSON(res, &r); err != nil {
		return Metadata{}, err
	}
	name := strings.TrimPrefix(ref.Identifier, "library/")
	return Metadata{
		Name:        name,
		Description: r.Description,
		HomeURL:     "https://hub.docker.com/r/" + ref.Identifier,
	}, nil
}

type dhTagsResponse struct {
	Results []struct {
		Name        string    `json:"name"`
		LastUpdated time.Time `json:"last_updated"`
	} `json:"results"`
}

func (d *dockerHubProvider) Fetch(ctx context.Context, req FetchRequest) (FetchResult, error) {
	limit := clampLimit(req.Limit)
	u := fmt.Sprintf("%s/tags?page_size=%d&ordering=last_updated", d.repoAPI(req.Ref.Identifier), limit)
	res, err := d.hc.getCond(ctx, u, req.ETag, req.LastModified, nil)
	if err != nil {
		return FetchResult{}, err
	}
	if res.NotModified {
		return FetchResult{NotModified: true, ETag: res.ETag, LastModified: res.LastModified}, nil
	}
	var raw dhTagsResponse
	if err := decodeJSON(res, &raw); err != nil {
		return FetchResult{}, err
	}
	out := make([]Release, 0, len(raw.Results))
	for _, t := range raw.Results {
		out = append(out, Release{
			ExternalID:  t.Name,
			Version:     t.Name,
			Title:       t.Name,
			URL:         fmt.Sprintf("https://hub.docker.com/r/%s/tags?name=%s", req.Ref.Identifier, url.QueryEscape(t.Name)),
			PublishedAt: t.LastUpdated.UTC(),
		})
	}
	return FetchResult{ETag: res.ETag, LastModified: res.LastModified, Releases: out}, nil
}
