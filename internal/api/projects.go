package api

import (
	"net/http"
	"strconv"
	"strings"

	"projector/internal/store"
)

// parseID reads a path value as an int64.
func parseID(r *http.Request, name string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func parseBoolParam(v string) *bool {
	switch strings.ToLower(v) {
	case "1", "true", "yes":
		b := true
		return &b
	case "0", "false", "no":
		b := false
		return &b
	}
	return nil
}

func (a *API) handleListProjects(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := store.ProjectFilter{
		Tag:    q.Get("tag"),
		Source: q.Get("source"),
		Query:  q.Get("q"),
		InUse:  parseBoolParam(q.Get("in_use")),
	}
	projects, err := a.store.ListProjects(f)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	if projects == nil {
		projects = []store.Project{}
	}
	writeJSON(w, http.StatusOK, projects)
}

// createProjectRequest accepts either a fully-specified project or a bare URL
// (detection/prefill happens in a later phase; for now url+name is required).
type createProjectRequest struct {
	Name             string  `json:"name"`
	URL              string  `json:"url"`
	SourceType       string  `json:"source_type"`
	SourceIdentifier string  `json:"source_identifier"`
	SourceHost       string  `json:"source_host"`
	InUse            *bool   `json:"in_use"`
	TagIDs           []int64 `json:"tag_ids"`
}

func (a *API) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var req createProjectRequest
	if err := decode(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	req.URL = strings.TrimSpace(req.URL)
	req.Name = strings.TrimSpace(req.Name)
	if req.URL == "" {
		writeErr(w, http.StatusBadRequest, "url is required")
		return
	}
	// Detect-on-create: a bare {url} (no source_type) gets auto-classified and
	// prefilled from the provider.
	if req.SourceType == "" {
		d := a.detect(r.Context(), req.URL)
		req.SourceType = d.SourceType
		req.SourceIdentifier = d.SourceIdentifier
		req.SourceHost = d.SourceHost
		if req.Name == "" {
			req.Name = d.Prefill.Name
		}
	}
	if req.SourceType == "" {
		req.SourceType = "website"
	}
	if req.Name == "" {
		req.Name = req.URL
	}

	p := store.Project{
		Name:             req.Name,
		URL:              req.URL,
		SourceType:       req.SourceType,
		SourceIdentifier: req.SourceIdentifier,
		SourceHost:       req.SourceHost,
		InUse:            boolOr(req.InUse, true),
	}
	created, err := a.store.CreateProject(p)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	if len(req.TagIDs) > 0 {
		if err := a.store.SetProjectTags(created.ID, req.TagIDs); err != nil {
			writeStoreErr(w, err)
			return
		}
		created, err = a.store.GetProject(created.ID)
		if err != nil {
			writeStoreErr(w, err)
			return
		}
	}
	writeJSON(w, http.StatusCreated, created)
}

func (a *API) handleGetProject(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r, "id")
	if !ok {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	p, err := a.store.GetProject(id)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

type updateProjectRequest struct {
	Name   *string  `json:"name"`
	URL    *string  `json:"url"`
	InUse  *bool    `json:"in_use"`
	TagIDs *[]int64 `json:"tag_ids"`
}

func (a *API) handleUpdateProject(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r, "id")
	if !ok {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req updateProjectRequest
	if err := decode(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	p, err := a.store.UpdateProjectTx(id, store.ProjectUpdate{
		Name:  req.Name,
		URL:   req.URL,
		InUse: req.InUse,
	}, req.TagIDs)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (a *API) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r, "id")
	if !ok {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := a.store.DeleteProject(id); err != nil {
		writeStoreErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type setTagsRequest struct {
	TagIDs []int64 `json:"tag_ids"`
}

func (a *API) handleSetProjectTags(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r, "id")
	if !ok {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req setTagsRequest
	if err := decode(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if err := a.store.SetProjectTags(id, req.TagIDs); err != nil {
		writeStoreErr(w, err)
		return
	}
	p, err := a.store.GetProject(id)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

type addTagRequest struct {
	TagID int64  `json:"tag_id"`
	Name  string `json:"name"`
}

func (a *API) handleAddProjectTag(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r, "id")
	if !ok {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req addTagRequest
	if err := decode(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	tagID := req.TagID
	if tagID == 0 && strings.TrimSpace(req.Name) != "" {
		tag, err := a.store.EnsureTag(strings.TrimSpace(req.Name))
		if err != nil {
			writeStoreErr(w, err)
			return
		}
		tagID = tag.ID
	}
	if tagID == 0 {
		writeErr(w, http.StatusBadRequest, "tag_id or name is required")
		return
	}
	if err := a.store.AddProjectTag(id, tagID); err != nil {
		writeStoreErr(w, err)
		return
	}
	p, err := a.store.GetProject(id)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (a *API) handleRemoveProjectTag(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r, "id")
	if !ok {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	tagID, ok := parseID(r, "tagId")
	if !ok {
		writeErr(w, http.StatusBadRequest, "invalid tag id")
		return
	}
	if err := a.store.RemoveProjectTag(id, tagID); err != nil {
		writeStoreErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func boolOr(p *bool, def bool) bool {
	if p != nil {
		return *p
	}
	return def
}

// handleProjectReleases fetches releases live (through the cache) for one project.
func (a *API) handleProjectReleases(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r, "id")
	if !ok {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	p, err := a.store.GetProject(id)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	limit := 20
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	ref := refFromProject(p.SourceType, p.SourceIdentifier, p.SourceHost)
	releases, err := a.releases.Releases(r.Context(), ref, limit)
	if err != nil {
		if isNotPollable(err) {
			writeJSON(w, http.StatusOK, []any{})
			return
		}
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, releases)
}

// handleRefreshProject busts this project's cache so the next fetch is live.
func (a *API) handleRefreshProject(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r, "id")
	if !ok {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	p, err := a.store.GetProject(id)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	a.releases.Invalidate(refFromProject(p.SourceType, p.SourceIdentifier, p.SourceHost))
	w.WriteHeader(http.StatusNoContent)
}
