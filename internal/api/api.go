// Package api wires the HTTP handlers for the JSON API under /api.
package api

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"projector/internal/provider"
	"projector/internal/store"
)

// API holds handler dependencies.
type API struct {
	store    *store.Store
	registry *provider.Registry
	releases *provider.Service
	version  string
}

// New constructs the API handler set.
func New(st *store.Store, reg *provider.Registry, releases *provider.Service, version string) *API {
	return &API{store: st, registry: reg, releases: releases, version: version}
}

// Handler returns the fully-wired /api router with middleware applied. The
// returned handler expects to be mounted at the root and matches "/api/..."
// paths itself.
func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/meta", a.handleMeta)

	// Projects
	mux.HandleFunc("GET /api/projects", a.handleListProjects)
	mux.HandleFunc("POST /api/projects", a.handleCreateProject)
	mux.HandleFunc("GET /api/projects/{id}", a.handleGetProject)
	mux.HandleFunc("PATCH /api/projects/{id}", a.handleUpdateProject)
	mux.HandleFunc("DELETE /api/projects/{id}", a.handleDeleteProject)
	mux.HandleFunc("GET /api/projects/{id}/releases", a.handleProjectReleases)
	mux.HandleFunc("POST /api/projects/{id}/refresh", a.handleRefreshProject)
	mux.HandleFunc("PUT /api/projects/{id}/tags", a.handleSetProjectTags)
	mux.HandleFunc("POST /api/projects/{id}/tags", a.handleAddProjectTag)
	mux.HandleFunc("DELETE /api/projects/{id}/tags/{tagId}", a.handleRemoveProjectTag)

	// Tags
	mux.HandleFunc("GET /api/tags", a.handleListTags)
	mux.HandleFunc("POST /api/tags", a.handleCreateTag)
	mux.HandleFunc("PATCH /api/tags/{id}", a.handleUpdateTag)
	mux.HandleFunc("DELETE /api/tags/{id}", a.handleDeleteTag)

	// Detection + live feed
	mux.HandleFunc("POST /api/detect", a.handleDetect)
	mux.HandleFunc("GET /api/feed", a.handleFeed)
	mux.HandleFunc("POST /api/feed/refresh", a.handleRefreshFeed)

	// Import
	mux.HandleFunc("POST /api/import/github-stars", a.handleImportGitHubStars)

	return chain(mux, recoverMW, logMW)
}

// middleware is a standard HTTP middleware.
type middleware func(http.Handler) http.Handler

func chain(h http.Handler, mws ...middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func recoverMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				slog.Error("panic in handler", "err", v, "stack", string(debug.Stack()))
				writeErr(w, http.StatusInternalServerError, "internal error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func logMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		slog.Info("request",
			"method", r.Method, "path", r.URL.Path,
			"status", sw.status, "dur", time.Since(start).String())
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (w *statusWriter) WriteHeader(code int) {
	if !w.wrote {
		w.status = code
		w.wrote = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	w.wrote = true
	return w.ResponseWriter.Write(b)
}
