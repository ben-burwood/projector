package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const githubAPI = "https://api.github.com"

type githubProvider struct {
	hc    *httpClient
	token string
}

func newGitHub(hc *httpClient, token string) *githubProvider {
	return &githubProvider{hc: hc, token: token}
}

func (g *githubProvider) Type() SourceType { return GitHub }

func (g *githubProvider) header() http.Header {
	h := http.Header{}
	h.Set("Accept", "application/vnd.github+json")
	h.Set("X-GitHub-Api-Version", "2022-11-28")
	if g.token != "" {
		h.Set("Authorization", "Bearer "+g.token)
	}
	return h
}

func (g *githubProvider) Detect(u *url.URL) (Ref, bool) {
	if hostname(u) != "github.com" {
		return Ref{}, false
	}
	parts := pathParts(u)
	if len(parts) < 2 {
		return Ref{}, false
	}
	owner, repo := parts[0], strings.TrimSuffix(parts[1], ".git")
	// Skip GitHub's reserved routes that look like owner/repo but aren't.
	switch owner {
	case "features", "about", "pricing", "marketplace", "sponsors", "settings", "notifications":
		return Ref{}, false
	}
	return Ref{Type: GitHub, Identifier: owner + "/" + repo}, true
}

type ghRepo struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Homepage    string `json:"homepage"`
	HTMLURL     string `json:"html_url"`
}

func (g *githubProvider) Prefill(ctx context.Context, ref Ref) (Metadata, error) {
	res, err := g.hc.getCond(ctx, githubAPI+"/repos/"+ref.Identifier, "", "", g.header())
	if err != nil {
		return Metadata{}, err
	}
	var r ghRepo
	if err := decodeJSON(res, &r); err != nil {
		return Metadata{}, err
	}
	home := r.Homepage
	if home == "" {
		home = r.HTMLURL
	}
	name := r.Name
	if name == "" {
		name = ref.Identifier
	}
	return Metadata{Name: name, Description: r.Description, HomeURL: home}, nil
}

type ghRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	HTMLURL     string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
	CreatedAt   time.Time `json:"created_at"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
}

type ghTag struct {
	Name   string `json:"name"`
	Commit struct {
		SHA string `json:"sha"`
	} `json:"commit"`
}

func (g *githubProvider) Fetch(ctx context.Context, req FetchRequest) (FetchResult, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	relURL := fmt.Sprintf("%s/repos/%s/releases?per_page=%d", githubAPI, req.Ref.Identifier, limit)
	res, err := g.hc.getCond(ctx, relURL, req.ETag, req.LastModified, g.header())
	if err != nil {
		return FetchResult{}, err
	}
	if res.NotModified {
		return FetchResult{NotModified: true, ETag: res.ETag, LastModified: res.LastModified}, nil
	}

	var raw []ghRelease
	if err := decodeJSON(res, &raw); err != nil {
		return FetchResult{}, err
	}

	// Repos with no formal releases (empty array) fall back to tags.
	if len(raw) == 0 {
		return g.fetchTags(ctx, req)
	}

	out := make([]Release, 0, len(raw))
	for _, r := range raw {
		if r.Draft {
			continue
		}
		pub := r.PublishedAt
		if pub.IsZero() {
			pub = r.CreatedAt
		}
		title := r.Name
		if title == "" {
			title = r.TagName
		}
		out = append(out, Release{
			ExternalID:  r.TagName,
			Version:     r.TagName,
			Title:       title,
			Body:        r.Body,
			URL:         r.HTMLURL,
			PublishedAt: pub.UTC(),
		})
	}
	return FetchResult{ETag: res.ETag, LastModified: res.LastModified, Releases: out}, nil
}

// fetchTags is the fallback for repos without GitHub Releases. Tags carry no
// date or notes, so those fields are left empty.
func (g *githubProvider) fetchTags(ctx context.Context, req FetchRequest) (FetchResult, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	tagURL := fmt.Sprintf("%s/repos/%s/tags?per_page=%d", githubAPI, req.Ref.Identifier, limit)
	res, err := g.hc.getCond(ctx, tagURL, "", "", g.header())
	if err != nil {
		return FetchResult{}, err
	}
	var raw []ghTag
	if err := decodeJSON(res, &raw); err != nil {
		return FetchResult{}, err
	}
	out := make([]Release, 0, len(raw))
	for _, t := range raw {
		out = append(out, Release{
			ExternalID: t.Name,
			Version:    t.Name,
			Title:      t.Name,
			URL:        fmt.Sprintf("https://github.com/%s/releases/tag/%s", req.Ref.Identifier, url.PathEscape(t.Name)),
		})
	}
	return FetchResult{ETag: res.ETag, LastModified: res.LastModified, Releases: out}, nil
}
