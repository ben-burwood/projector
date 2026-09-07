package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"projector/internal/provider"
)

type detectRequest struct {
	URL string `json:"url"`
}

type detectResponse struct {
	SourceType       string            `json:"source_type"`
	SourceIdentifier string            `json:"source_identifier"`
	SourceHost       string            `json:"source_host"`
	Prefill          provider.Metadata `json:"prefill"`
}

func (a *API) handleDetect(w http.ResponseWriter, r *http.Request) {
	var req detectRequest
	if err := decode(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	req.URL = strings.TrimSpace(req.URL)
	if req.URL == "" {
		writeErr(w, http.StatusBadRequest, "url is required")
		return
	}
	resp := a.detect(r.Context(), req.URL)
	writeJSON(w, http.StatusOK, resp)
}

// detect resolves a URL to a source ref and best-effort prefill metadata.
// Prefill failures are non-fatal — the caller can still save the bookmark.
func (a *API) detect(ctx context.Context, rawURL string) detectResponse {
	ref, err := a.registry.Detect(rawURL)
	if err != nil {
		return detectResponse{SourceType: provider.Website, Prefill: provider.Metadata{HomeURL: rawURL}}
	}
	resp := detectResponse{
		SourceType:       ref.Type,
		SourceIdentifier: ref.Identifier,
		SourceHost:       ref.Host,
		Prefill:          provider.Metadata{HomeURL: rawURL},
	}
	if p := a.registry.Get(ref.Type); p != nil {
		pctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if meta, err := p.Prefill(pctx, ref); err == nil {
			if meta.HomeURL == "" {
				meta.HomeURL = rawURL
			}
			resp.Prefill = meta
		}
	}
	return resp
}

// refFromProject builds a provider.Ref from a stored project.
func refFromProject(sourceType, identifier, host string) provider.Ref {
	return provider.Ref{Type: sourceType, Identifier: identifier, Host: host}
}

// isNotPollable reports whether err means the source has no release feed.
func isNotPollable(err error) bool {
	return errors.Is(err, provider.ErrNotPollable)
}
