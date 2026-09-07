package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// giteaProvider covers Codeberg and any self-hosted Gitea/Forgejo instance.
type giteaProvider struct {
	hc     *httpClient
	tokens map[string]string // host -> token
}

func newGitea(hc *httpClient, tokensSpec string) *giteaProvider {
	return &giteaProvider{hc: hc, tokens: parseHostTokens(tokensSpec)}
}

func (g *giteaProvider) Type() SourceType { return Gitea }

func (g *giteaProvider) host(ref Ref) string {
	if ref.Host != "" {
		return ref.Host
	}
	return "codeberg.org"
}

func (g *giteaProvider) header(ref Ref) http.Header {
	h := http.Header{}
	if tok := g.tokens[g.host(ref)]; tok != "" {
		h.Set("Authorization", "token "+tok)
	}
	return h
}

func (g *giteaProvider) Detect(u *url.URL) (Ref, bool) {
	// Only auto-detect Codeberg; self-hosted Gitea is added manually with a host.
	if hostname(u) != "codeberg.org" {
		return Ref{}, false
	}
	parts := pathParts(u)
	if len(parts) < 2 {
		return Ref{}, false
	}
	owner, repo := parts[0], strings.TrimSuffix(parts[1], ".git")
	return Ref{Type: Gitea, Identifier: owner + "/" + repo, Host: "codeberg.org"}, true
}

func (g *giteaProvider) repoAPI(ref Ref) string {
	return fmt.Sprintf("https://%s/api/v1/repos/%s", g.host(ref), ref.Identifier)
}

type giteaRepo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Website     string `json:"website"`
	HTMLURL     string `json:"html_url"`
}

func (g *giteaProvider) Prefill(ctx context.Context, ref Ref) (Metadata, error) {
	res, err := g.hc.getCond(ctx, g.repoAPI(ref), "", "", g.header(ref))
	if err != nil {
		return Metadata{}, err
	}
	var r giteaRepo
	if err := decodeJSON(res, &r); err != nil {
		return Metadata{}, err
	}
	home := r.Website
	if home == "" {
		home = r.HTMLURL
	}
	name := r.Name
	if name == "" {
		name = lastPathSegment(ref.Identifier)
	}
	return Metadata{Name: name, Description: r.Description, HomeURL: home}, nil
}

type giteaRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	HTMLURL     string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
	Draft       bool      `json:"draft"`
}

func (g *giteaProvider) Fetch(ctx context.Context, req FetchRequest) (FetchResult, error) {
	limit := clampLimit(req.Limit)
	u := fmt.Sprintf("%s/releases?limit=%d", g.repoAPI(req.Ref), limit)
	res, err := g.hc.getCond(ctx, u, req.ETag, req.LastModified, g.header(req.Ref))
	if err != nil {
		return FetchResult{}, err
	}
	if res.NotModified {
		return FetchResult{NotModified: true, ETag: res.ETag, LastModified: res.LastModified}, nil
	}
	var raw []giteaRelease
	if err := decodeJSON(res, &raw); err != nil {
		return FetchResult{}, err
	}
	out := make([]Release, 0, len(raw))
	for _, r := range raw {
		if r.Draft {
			continue
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
			PublishedAt: r.PublishedAt.UTC(),
		})
	}
	return FetchResult{ETag: res.ETag, LastModified: res.LastModified, Releases: out}, nil
}

// parseHostTokens parses "host=token,host2=token2" into a map.
func parseHostTokens(spec string) map[string]string {
	m := make(map[string]string)
	for pair := range strings.SplitSeq(spec, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		if k, v, ok := strings.Cut(pair, "="); ok {
			m[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return m
}
