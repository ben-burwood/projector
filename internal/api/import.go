package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"projector/internal/provider"
	"projector/internal/store"
)

type importStarsRequest struct {
	Username string  `json:"username"`
	InUse    *bool   `json:"in_use"`
	TagIDs   []int64 `json:"tag_ids"`
}

type importResult struct {
	Total    int `json:"total"`
	Imported int `json:"imported"`
	Skipped  int `json:"skipped"`
}

// handleImportGitHubStars bookmarks every repo a GitHub user has starred,
// skipping ones already tracked.
func (a *API) handleImportGitHubStars(w http.ResponseWriter, r *http.Request) {
	var req importStarsRequest
	if err := decode(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	lister, ok := a.registry.Get(provider.GitHub).(provider.StarLister)
	if !ok {
		writeErr(w, http.StatusNotImplemented, "github star import unavailable")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()
	stars, err := lister.Starred(ctx, strings.TrimPrefix(strings.TrimSpace(req.Username), "@"))
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	inUse := boolOr(req.InUse, false)
	ps := make([]store.Project, 0, len(stars))
	for _, s := range stars {
		ps = append(ps, store.Project{
			Name:             s.Name,
			URL:              s.HTMLURL,
			SourceType:       provider.GitHub,
			SourceIdentifier: s.FullName,
			InUse:            inUse,
		})
	}
	n, err := a.store.ImportProjects(ps, req.TagIDs)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, importResult{Total: len(ps), Imported: n, Skipped: len(ps) - n})
}
