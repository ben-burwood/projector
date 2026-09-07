package store

// Project is a bookmarked project the user tracks.
type Project struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	URL              string `json:"url"`
	SourceType       string `json:"source_type"`
	SourceIdentifier string `json:"source_identifier"`
	SourceHost       string `json:"source_host"`
	SourceMeta       string `json:"source_meta"`
	InUse            bool   `json:"in_use"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
	Tags             []Tag  `json:"tags"`
}

// Tag is a user-defined organisational label.
type Tag struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	CreatedAt    string `json:"created_at"`
	ProjectCount int    `json:"project_count,omitempty"`
}

// ProjectFilter narrows a project listing. Zero values mean "no constraint".
type ProjectFilter struct {
	Tag    string // tag name
	InUse  *bool
	Source string // source_type
	Query  string // free-text over name/url
}
