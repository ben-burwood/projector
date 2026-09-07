// Package config loads the optional provider auth tokens from the environment.
// Everything else is a compile-time constant (see the main, api and provider
// packages): the app always runs in its container, so there is nothing to tune.
package config

import "os"

// Config holds the optional provider auth tokens.
type Config struct {
	GitHubToken string // optional, raises GitHub rate limit
	GitLabToken string // optional, for private/self-hosted GitLab
	GiteaTokens string // optional, host=token,host=token for self-hosted Gitea
}

// Load reads the provider tokens from the environment.
func Load() Config {
	return Config{
		GitHubToken: os.Getenv("PROJECTOR_GITHUB_TOKEN"),
		GitLabToken: os.Getenv("PROJECTOR_GITLAB_TOKEN"),
		GiteaTokens: os.Getenv("PROJECTOR_GITEA_TOKENS"),
	}
}
