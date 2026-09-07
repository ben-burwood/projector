-- Bookmarks only. Releases are fetched live and never persisted.

CREATE TABLE projects (
    id                INTEGER PRIMARY KEY,
    name              TEXT    NOT NULL,
    url               TEXT    NOT NULL,                 -- what the user pasted / homepage
    source_type       TEXT    NOT NULL,                 -- github|gitlab|gitea|dockerhub|pypi|crates|website
    source_identifier TEXT    NOT NULL DEFAULT '',      -- owner/repo | package | image | feed URL
    source_host       TEXT    NOT NULL DEFAULT '',      -- codeberg.org / self-hosted host
    source_meta       TEXT    NOT NULL DEFAULT '{}',    -- JSON escape hatch
    in_use            INTEGER NOT NULL DEFAULT 1,
    created_at        TEXT    NOT NULL,
    updated_at        TEXT    NOT NULL
);
CREATE INDEX idx_projects_source  ON projects(source_type);
CREATE INDEX idx_projects_in_use  ON projects(in_use);

CREATE TABLE tags (
    id         INTEGER PRIMARY KEY,
    name       TEXT    NOT NULL UNIQUE COLLATE NOCASE,
    created_at TEXT    NOT NULL
);

CREATE TABLE project_tags (
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    tag_id     INTEGER NOT NULL REFERENCES tags(id)     ON DELETE CASCADE,
    PRIMARY KEY (project_id, tag_id)
);
CREATE INDEX idx_project_tags_tag ON project_tags(tag_id);
