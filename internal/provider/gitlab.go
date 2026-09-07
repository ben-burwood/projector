package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type gitlabProvider struct {
	hc    *httpClient
	token string
}

func newGitLab(hc *httpClient, token string) *gitlabProvider {
	return &gitlabProvider{hc: hc, token: token}
}

func (g *gitlabProvider) Type() SourceType { return GitLab }

func (g *gitlabProvider) header() http.Header {
	h := http.Header{}
	if g.token != "" {
		h.Set("PRIVATE-TOKEN", g.token)
	}
	return h
}

func (g *gitlabProvider) host(ref Ref) string {
	if ref.Host != "" {
		return ref.Host
	}
	return "gitlab.com"
}

func (g *gitlabProvider) Detect(u *url.URL) (Ref, bool) {
	if hostname(u) != "gitlab.com" {
		return Ref{}, false
	}
	parts := pathParts(u)
	if len(parts) < 2 {
		return Ref{}, false
	}
	// GitLab projects can be nested (group/subgroup/project). Trim known
	// non-project trailing segments (/-/tree, /-/releases, etc.).
	if i := indexOf(parts, "-"); i >= 2 {
		parts = parts[:i]
	}
	path := strings.TrimSuffix(strings.Join(parts, "/"), ".git")
	return Ref{Type: GitLab, Identifier: path}, true
}

func (g *gitlabProvider) projectAPI(ref Ref) string {
	return fmt.Sprintf("https://%s/api/v4/projects/%s", g.host(ref), url.PathEscape(ref.Identifier))
}

type glProject struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	WebURL      string `json:"web_url"`
}

func (g *gitlabProvider) Prefill(ctx context.Context, ref Ref) (Metadata, error) {
	res, err := g.hc.getCond(ctx, g.projectAPI(ref), "", "", g.header())
	if err != nil {
		return Metadata{}, err
	}
	var p glProject
	if err := decodeJSON(res, &p); err != nil {
		return Metadata{}, err
	}
	name := p.Name
	if name == "" {
		name = lastPathSegment(ref.Identifier)
	}
	return Metadata{Name: name, Description: p.Description, HomeURL: p.WebURL}, nil
}

type glRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ReleasedAt  time.Time `json:"released_at"`
	Links       struct {
		Self string `json:"self"`
	} `json:"_links"`
}

func (g *gitlabProvider) Fetch(ctx context.Context, req FetchRequest) (FetchResult, error) {
	limit := clampLimit(req.Limit)
	u := fmt.Sprintf("%s/releases?per_page=%d", g.projectAPI(req.Ref), limit)
	res, err := g.hc.getCond(ctx, u, req.ETag, req.LastModified, g.header())
	if err != nil {
		return FetchResult{}, err
	}
	if res.NotModified {
		return FetchResult{NotModified: true, ETag: res.ETag, LastModified: res.LastModified}, nil
	}
	var raw []glRelease
	if err := decodeJSON(res, &raw); err != nil {
		return FetchResult{}, err
	}
	out := make([]Release, 0, len(raw))
	for _, r := range raw {
		title := r.Name
		if title == "" {
			title = r.TagName
		}
		out = append(out, Release{
			ExternalID:  r.TagName,
			Version:     r.TagName,
			Title:       title,
			Body:        r.Description,
			URL:         r.Links.Self,
			PublishedAt: r.ReleasedAt.UTC(),
		})
	}
	return FetchResult{ETag: res.ETag, LastModified: res.LastModified, Releases: out}, nil
}
