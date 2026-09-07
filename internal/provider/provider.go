// Package provider fetches release/version information live from external
// sources (GitHub, GitLab, etc.). Nothing is persisted — results are held only
// in an in-memory TTL cache (see cache.go).
package provider

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// ErrNotPollable is returned for sources that have no live release feed
// (e.g. plain website bookmarks).
var ErrNotPollable = errors.New("source does not support release fetching")

// SourceType identifies a kind of source.
type SourceType = string

const (
	GitHub    SourceType = "github"
	GitLab    SourceType = "gitlab"
	Gitea     SourceType = "gitea"
	DockerHub SourceType = "dockerhub"
	PyPI      SourceType = "pypi"
	Crates    SourceType = "crates"
	Website   SourceType = "website"
)

// Ref uniquely identifies what to fetch for a project.
type Ref struct {
	Type       SourceType
	Identifier string // owner/repo | package | image | feed URL
	Host       string // e.g. codeberg.org, self-hosted host; empty = provider default
}

// Metadata is prefill information for the add-by-URL flow.
type Metadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	HomeURL     string `json:"home_url"`
}

// Release is a single version/release of a project. Newest first when returned.
type Release struct {
	ExternalID  string    `json:"external_id"` // stable dedup key (tag/version/guid)
	Version     string    `json:"version"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	URL         string    `json:"url"`
	PublishedAt time.Time `json:"published_at"`
}

// FetchRequest carries conditional-request state so providers can revalidate.
type FetchRequest struct {
	Ref          Ref
	ETag         string
	LastModified string
	Limit        int
}

// FetchResult is a provider's response, including revalidation metadata.
type FetchResult struct {
	NotModified  bool
	ETag         string
	LastModified string
	Releases     []Release
}

// Provider fetches releases for one kind of source.
type Provider interface {
	Type() SourceType
	// Detect returns a Ref if this provider recognises the URL.
	Detect(u *url.URL) (Ref, bool)
	// Prefill returns metadata for the add-by-URL confirmation form.
	Prefill(ctx context.Context, ref Ref) (Metadata, error)
	// Fetch returns releases live, honouring conditional-request headers.
	Fetch(ctx context.Context, req FetchRequest) (FetchResult, error)
}

// Registry holds the set of providers and routes detection/dispatch.
type Registry struct {
	byType map[SourceType]Provider
	// order defines detection precedence.
	order []Provider
}

// NewRegistry builds a registry from the given providers (detection tried in order).
func NewRegistry(providers ...Provider) *Registry {
	r := &Registry{byType: make(map[SourceType]Provider)}
	for _, p := range providers {
		r.byType[p.Type()] = p
		r.order = append(r.order, p)
	}
	return r
}

// Get returns the provider for a source type, or nil.
func (r *Registry) Get(t SourceType) Provider { return r.byType[t] }

// Detect parses raw and returns the first matching provider's Ref. If no
// provider matches, it returns a website Ref (bookmark only).
func (r *Registry) Detect(raw string) (Ref, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Ref{}, fmt.Errorf("empty url")
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return Ref{}, fmt.Errorf("invalid url: %w", err)
	}
	u.Host = strings.ToLower(u.Host)
	for _, p := range r.order {
		if ref, ok := p.Detect(u); ok {
			return ref, nil
		}
	}
	return Ref{Type: Website, Identifier: raw}, nil
}

// hostname strips a leading "www." and any port from a URL host.
func hostname(u *url.URL) string {
	h := u.Hostname()
	return strings.TrimPrefix(strings.ToLower(h), "www.")
}

// indexOf returns the index of s in parts, or -1.
func indexOf(parts []string, s string) int {
	for i, p := range parts {
		if p == s {
			return i
		}
	}
	return -1
}

// lastPathSegment returns the final segment of a slash-separated identifier.
func lastPathSegment(id string) string {
	if i := strings.LastIndexByte(id, '/'); i >= 0 {
		return id[i+1:]
	}
	return id
}

// clampLimit bounds a requested release count to a sane range.
func clampLimit(n int) int {
	if n <= 0 {
		return 20
	}
	if n > 100 {
		return 100
	}
	return n
}

// pathParts returns the non-empty, slash-separated segments of a URL path.
func pathParts(u *url.URL) []string {
	var out []string
	for seg := range strings.SplitSeq(u.Path, "/") {
		if seg != "" {
			out = append(out, seg)
		}
	}
	return out
}
