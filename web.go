package main

// SPA serving. The go:embed directive must live at the module root because
// embed cannot traverse "..", and frontend/dist is under the root.

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:frontend/dist
var embeddedDist embed.FS

// spaHandler serves the built Vue SPA embedded into the binary. Real asset
// requests are served from the embedded FS; any other non-/api path falls back
// to index.html so client-side routing works on refresh/deep-link.
func spaHandler() (http.Handler, error) {
	dist, err := fs.Sub(embeddedDist, "frontend/dist")
	if err != nil {
		return nil, err
	}

	fileServer := http.FileServerFS(dist)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Never let the SPA shadow the API.
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		if r.URL.Path != "/" && assetExists(dist, r.URL.Path) {
			fileServer.ServeHTTP(w, r)
			return
		}
		serveIndex(w, r, dist)
	}), nil
}

func assetExists(dist fs.FS, urlPath string) bool {
	name := strings.TrimPrefix(urlPath, "/")
	if name == "" {
		return false
	}
	f, err := dist.Open(name)
	if err != nil {
		return false
	}
	defer f.Close()
	if st, err := f.Stat(); err == nil && st.IsDir() {
		return false
	}
	return true
}

func serveIndex(w http.ResponseWriter, r *http.Request, dist fs.FS) {
	data, err := fs.ReadFile(dist, "index.html")
	if err != nil {
		http.Error(w, "SPA not built (frontend/dist/index.html missing)", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}
