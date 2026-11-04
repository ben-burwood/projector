package api

import (
	"context"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"projector/internal/provider"
	"projector/internal/store"
)

// feedItem pairs a release with the project it came from.
type feedItem struct {
	ProjectID   int64            `json:"project_id"`
	ProjectName string           `json:"project_name"`
	SourceType  string           `json:"source_type"`
	SourceIdent string           `json:"source_identifier"`
	Release     provider.Release `json:"release"`
}

type feedError struct {
	ProjectID   int64  `json:"project_id"`
	ProjectName string `json:"project_name"`
	Error       string `json:"error"`
}

type feedResponse struct {
	Items   []feedItem  `json:"items"`
	Errors  []feedError `json:"errors"`
	Page    int         `json:"page"`
	Limit   int         `json:"limit"`
	HasMore bool        `json:"has_more"`
}

const perReleaseFetchLimit = 10

// fetchConcurrency bounds how many provider fetches the feed fan-out runs at once.
const fetchConcurrency = 6

// handleFeed aggregates releases live across the (filtered) project set. Fetches
// run concurrently through a bounded worker pool; per-project errors are
// reported inline rather than failing the whole feed.
func (a *API) handleFeed(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page := atoiDefault(q.Get("page"), 1)
	if page < 1 {
		page = 1
	}
	limit := atoiDefault(q.Get("limit"), 30)
	if limit < 1 || limit > 100 {
		limit = 30
	}

	// Only in-use, pollable projects appear in the feed, unless ?all=1 asks to
	// include not-in-use ones too.
	f := store.ProjectFilter{
		Tag:    q.Get("tag"),
		Source: q.Get("source"),
	}
	if all := parseBoolParam(q.Get("all")); all == nil || !*all {
		inUse := true
		f.InUse = &inUse
	}
	projects, err := a.store.ListProjects(f)
	if err != nil {
		writeStoreErr(w, err)
		return
	}

	items, errs := a.aggregate(r.Context(), projects)

	// Sort newest first; releases with no date sink to the bottom.
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Release.PublishedAt.After(items[j].Release.PublishedAt)
	})

	// In-memory pagination over the merged result.
	start := (page - 1) * limit
	end := start + limit
	total := len(items)
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	resp := feedResponse{
		Items:   items[start:end],
		Errors:  errs,
		Page:    page,
		Limit:   limit,
		HasMore: end < total,
	}
	if resp.Items == nil {
		resp.Items = []feedItem{}
	}
	if resp.Errors == nil {
		resp.Errors = []feedError{}
	}
	writeJSON(w, http.StatusOK, resp)
}

// aggregate fetches releases for all pollable projects concurrently.
func (a *API) aggregate(ctx context.Context, projects []store.Project) ([]feedItem, []feedError) {
	sem := make(chan struct{}, fetchConcurrency)
	var (
		mu    sync.Mutex
		items []feedItem
		errs  []feedError
		wg    sync.WaitGroup
	)

	for _, p := range projects {
		if a.registry.Get(p.SourceType) == nil {
			continue // website / unpollable
		}
		wg.Add(1)
		go func(p store.Project) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			fctx, cancel := context.WithTimeout(ctx, provider.HTTPTimeout+5*time.Second)
			defer cancel()

			ref := refFromProject(p.SourceType, p.SourceIdentifier, p.SourceHost)
			releases, err := a.releases.Releases(fctx, ref, perReleaseFetchLimit)

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, feedError{ProjectID: p.ID, ProjectName: p.Name, Error: err.Error()})
				return
			}
			for _, rel := range releases {
				items = append(items, feedItem{
					ProjectID:   p.ID,
					ProjectName: p.Name,
					SourceType:  p.SourceType,
					SourceIdent: p.SourceIdentifier,
					Release:     rel,
				})
			}
		}(p)
	}
	wg.Wait()
	return items, errs
}

// handleRefreshFeed clears the entire release cache.
func (a *API) handleRefreshFeed(w http.ResponseWriter, r *http.Request) {
	a.releases.InvalidateAll()
	w.WriteHeader(http.StatusNoContent)
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
