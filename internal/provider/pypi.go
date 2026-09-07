package provider

import (
	"context"
	"net/url"
	"sort"
	"time"
)

type pypiProvider struct{ hc *httpClient }

func newPyPI(hc *httpClient) *pypiProvider { return &pypiProvider{hc: hc} }

func (p *pypiProvider) Type() SourceType { return PyPI }

func (p *pypiProvider) Detect(u *url.URL) (Ref, bool) {
	if hostname(u) != "pypi.org" {
		return Ref{}, false
	}
	parts := pathParts(u)
	if len(parts) < 2 || parts[0] != "project" {
		return Ref{}, false
	}
	return Ref{Type: PyPI, Identifier: parts[1]}, true
}

func (p *pypiProvider) apiURL(name string) string {
	return "https://pypi.org/pypi/" + url.PathEscape(name) + "/json"
}

type pypiResponse struct {
	Info struct {
		Name        string `json:"name"`
		Summary     string `json:"summary"`
		HomePage    string `json:"home_page"`
		ProjectURL  string `json:"project_url"`
		Version     string `json:"version"`
		ProjectURLs map[string]string `json:"project_urls"`
	} `json:"info"`
	Releases map[string][]struct {
		UploadTime time.Time `json:"upload_time_iso_8601"`
		Yanked     bool      `json:"yanked"`
	} `json:"releases"`
}

func (p *pypiProvider) fetchJSON(ctx context.Context, ref Ref, etag, lastMod string) (pypiResponse, condResponse, error) {
	res, err := p.hc.getCond(ctx, p.apiURL(ref.Identifier), etag, lastMod, nil)
	if err != nil {
		return pypiResponse{}, res, err
	}
	if res.NotModified {
		return pypiResponse{}, res, nil
	}
	var raw pypiResponse
	if err := decodeJSON(res, &raw); err != nil {
		return pypiResponse{}, res, err
	}
	return raw, res, nil
}

func (p *pypiProvider) Prefill(ctx context.Context, ref Ref) (Metadata, error) {
	raw, _, err := p.fetchJSON(ctx, ref, "", "")
	if err != nil {
		return Metadata{}, err
	}
	home := raw.Info.HomePage
	if home == "" {
		home = raw.Info.ProjectURLs["Homepage"]
	}
	if home == "" {
		home = raw.Info.ProjectURL
	}
	name := raw.Info.Name
	if name == "" {
		name = ref.Identifier
	}
	return Metadata{Name: name, Description: raw.Info.Summary, HomeURL: home}, nil
}

func (p *pypiProvider) Fetch(ctx context.Context, req FetchRequest) (FetchResult, error) {
	raw, res, err := p.fetchJSON(ctx, req.Ref, req.ETag, req.LastModified)
	if err != nil {
		return FetchResult{}, err
	}
	if res.NotModified {
		return FetchResult{NotModified: true, ETag: res.ETag, LastModified: res.LastModified}, nil
	}

	// Build one release per version, dated by its earliest file upload.
	type verRel struct {
		ver  string
		when time.Time
	}
	rels := make([]verRel, 0, len(raw.Releases))
	for ver, files := range raw.Releases {
		var when time.Time
		yanked := true
		for _, f := range files {
			if !f.Yanked {
				yanked = false
			}
			if when.IsZero() || (!f.UploadTime.IsZero() && f.UploadTime.Before(when)) {
				when = f.UploadTime
			}
		}
		if len(files) == 0 || yanked {
			continue
		}
		rels = append(rels, verRel{ver: ver, when: when})
	}
	// Newest first by upload time.
	sort.Slice(rels, func(i, j int) bool { return rels[i].when.After(rels[j].when) })

	limit := clampLimit(req.Limit)
	if len(rels) > limit {
		rels = rels[:limit]
	}
	out := make([]Release, 0, len(rels))
	for _, r := range rels {
		out = append(out, Release{
			ExternalID:  r.ver,
			Version:     r.ver,
			Title:       raw.Info.Name + " " + r.ver,
			URL:         "https://pypi.org/project/" + raw.Info.Name + "/" + r.ver + "/",
			PublishedAt: r.when.UTC(),
		})
	}
	return FetchResult{ETag: res.ETag, LastModified: res.LastModified, Releases: out}, nil
}
