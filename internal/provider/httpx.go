package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const userAgent = "projector/1.0 (+https://github.com/ben-burwood/projector)"

// httpClient wraps a shared *http.Client with conditional-request helpers.
type httpClient struct {
	c *http.Client
}

func newHTTPClient(timeout time.Duration) *httpClient {
	return &httpClient{c: &http.Client{Timeout: timeout}}
}

// condResponse is the outcome of a conditional GET.
type condResponse struct {
	NotModified  bool
	StatusCode   int
	Body         []byte
	ETag         string
	LastModified string
	Header       http.Header
}

// getJSON performs a conditional GET and decodes a 200 body into out. On 304 it
// returns NotModified without touching out. header lets callers add auth, etc.
func (h *httpClient) getCond(ctx context.Context, url, etag, lastMod string, header http.Header) (condResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return condResponse{}, err
	}
	req.Header.Set("User-Agent", userAgent)
	for k, vs := range header {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	if lastMod != "" {
		req.Header.Set("If-Modified-Since", lastMod)
	}

	// Log every outbound request. URLs may carry query params but never auth —
	// tokens live in headers, which are intentionally not logged.
	start := time.Now()
	slog.DebugContext(ctx, "http request", "method", http.MethodGet, "url", url,
		"conditional", etag != "" || lastMod != "")

	resp, err := h.c.Do(req)
	if err != nil {
		slog.WarnContext(ctx, "http request failed", "method", http.MethodGet, "url", url,
			"dur", time.Since(start).String(), "err", err)
		return condResponse{}, err
	}
	defer resp.Body.Close()

	out := condResponse{
		StatusCode:   resp.StatusCode,
		ETag:         resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("Last-Modified"),
		Header:       resp.Header,
	}
	if resp.StatusCode == http.StatusNotModified {
		out.NotModified = true
		slog.InfoContext(ctx, "http response", "method", http.MethodGet, "url", url,
			"status", resp.StatusCode, "not_modified", true, "dur", time.Since(start).String())
		return out, nil
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20)) // 8 MiB cap
	if err != nil {
		slog.WarnContext(ctx, "http response read failed", "method", http.MethodGet, "url", url,
			"status", resp.StatusCode, "dur", time.Since(start).String(), "err", err)
		return out, err
	}
	out.Body = body

	level := slog.LevelInfo
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		level = slog.LevelWarn
	}
	slog.Log(ctx, level, "http response", "method", http.MethodGet, "url", url,
		"status", resp.StatusCode, "bytes", len(body), "dur", time.Since(start).String())

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return out, &httpError{Status: resp.StatusCode, URL: url, Body: string(body)}
	}
	return out, nil
}

// decodeJSON unmarshals a condResponse body into v.
func decodeJSON(r condResponse, v any) error {
	if err := json.Unmarshal(r.Body, v); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}
	return nil
}

// httpError describes a non-2xx response.
type httpError struct {
	Status int
	URL    string
	Body   string
}

func (e *httpError) Error() string {
	msg := e.Body
	if len(msg) > 200 {
		msg = msg[:200]
	}
	switch e.Status {
	case http.StatusNotFound:
		return "not found (404)"
	case http.StatusForbidden:
		return "forbidden (403) — rate limited or private?"
	case http.StatusUnauthorized:
		return "unauthorized (401)"
	}
	return fmt.Sprintf("http %d", e.Status)
}
