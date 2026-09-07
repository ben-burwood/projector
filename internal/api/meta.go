package api

import "net/http"

type metaResponse struct {
	Version string `json:"version"`
	Cache   any    `json:"cache"`
}

func (a *API) handleMeta(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, metaResponse{
		Version: a.version,
		Cache:   a.releases.Stats(),
	})
}
