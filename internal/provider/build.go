package provider

import "time"

// Tunables are compile-time constants — the app runs with no configuration.
const (
	// CacheTTL is how long a live release fetch is held in the in-memory cache.
	CacheTTL = 15 * time.Minute
	// HTTPTimeout bounds each outbound request to a provider API.
	HTTPTimeout = 20 * time.Second
)

// Options configures the provider registry.
type Options struct {
	GitHubToken string
	GitLabToken string
	GiteaTokens string // "host=token,host=token" for self-hosted Gitea
}

// Build constructs the registry (all supported providers) and the caching
// release Service that sits in front of it.
func Build(opts Options) (*Registry, *Service) {
	hc := newHTTPClient(HTTPTimeout)

	// Detection precedence follows registration order: host-specific providers
	// first, then the registry falls back to a plain website bookmark for
	// anything unrecognised.
	reg := NewRegistry(
		newGitHub(hc, opts.GitHubToken),
		newGitLab(hc, opts.GitLabToken),
		newGitea(hc, opts.GiteaTokens),
		newDockerHub(hc),
		newPyPI(hc),
		newCrates(hc),
	)
	svc := NewService(reg, CacheTTL)
	return reg, svc
}
