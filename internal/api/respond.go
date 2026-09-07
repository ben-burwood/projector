package api

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"projector/internal/store"
)

// writeJSON encodes v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encode response", "err", err)
	}
}

// errBody is the standard error envelope.
type errBody struct {
	Error string `json:"error"`
}

// writeErr writes a JSON error with the given status.
func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errBody{Error: msg})
}

// writeStoreErr maps store errors to HTTP responses.
func writeStoreErr(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	slog.Error("store error", "err", err)
	writeErr(w, http.StatusInternalServerError, "internal error")
}

// decode reads a JSON request body into v, rejecting unknown fields and
// oversized bodies.
func decode(r *http.Request, v any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20)) // 1 MiB cap
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
