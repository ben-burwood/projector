package store

import (
	"database/sql"
	"errors"
	"strings"
)

const projectCols = `id, name, url, source_type, source_identifier,
	source_host, source_meta, in_use, created_at, updated_at`

func scanProject(sc interface{ Scan(...any) error }) (Project, error) {
	var p Project
	err := sc.Scan(&p.ID, &p.Name, &p.URL, &p.SourceType,
		&p.SourceIdentifier, &p.SourceHost, &p.SourceMeta, &p.InUse,
		&p.CreatedAt, &p.UpdatedAt)
	return p, err
}

// ListProjects returns projects matching the filter, ordered by name, each with
// its tags loaded.
func (s *Store) ListProjects(f ProjectFilter) ([]Project, error) {
	var where []string
	var args []any

	if f.Source != "" {
		where = append(where, "p.source_type = ?")
		args = append(args, f.Source)
	}
	if f.InUse != nil {
		where = append(where, "p.in_use = ?")
		args = append(args, boolToInt(*f.InUse))
	}
	if f.Query != "" {
		where = append(where, `(p.name LIKE ? ESCAPE '\' OR p.url LIKE ? ESCAPE '\')`)
		like := "%" + escapeLike(f.Query) + "%"
		args = append(args, like, like)
	}
	if f.Tag != "" {
		where = append(where, `p.id IN (
			SELECT pt.project_id FROM project_tags pt
			JOIN tags t ON t.id = pt.tag_id
			WHERE t.name = ? COLLATE NOCASE)`)
		args = append(args, f.Tag)
	}

	q := "SELECT " + projectCols + " FROM projects p"
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	q += " ORDER BY p.name COLLATE NOCASE"

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Project
	var ids []int64
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
		ids = append(ids, p.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	tagsByProject, err := s.loadTagsFor(ids)
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].Tags = tagsByProject[out[i].ID]
	}
	return out, nil
}

// GetProject returns a single project with its tags.
func (s *Store) GetProject(id int64) (Project, error) {
	row := s.db.QueryRow("SELECT "+projectCols+" FROM projects p WHERE id = ?", id)
	p, err := scanProject(row)
	if errors.Is(err, sql.ErrNoRows) {
		return p, ErrNotFound
	}
	if err != nil {
		return p, err
	}
	tagsByProject, err := s.loadTagsFor([]int64{id})
	if err != nil {
		return p, err
	}
	p.Tags = tagsByProject[id]
	return p, nil
}

// CreateProject inserts a new project and returns it (with tags loaded, empty).
func (s *Store) CreateProject(p Project) (Project, error) {
	now := nowUTC()
	if p.SourceMeta == "" {
		p.SourceMeta = "{}"
	}
	res, err := s.db.Exec(`
		INSERT INTO projects
			(name, url, source_type, source_identifier, source_host,
			 source_meta, in_use, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.Name, p.URL, p.SourceType, p.SourceIdentifier, p.SourceHost,
		p.SourceMeta, boolToInt(p.InUse), now, now)
	if err != nil {
		return Project{}, err
	}
	id, _ := res.LastInsertId()
	return s.GetProject(id)
}

// ProjectUpdate carries optional field updates; nil fields are left unchanged.
type ProjectUpdate struct {
	Name  *string
	URL   *string
	InUse *bool
}

// UpdateProject applies a partial field update and returns the updated project.
func (s *Store) UpdateProject(id int64, u ProjectUpdate) (Project, error) {
	set, args := buildProjectSet(u)
	if len(set) > 0 {
		set = append(set, "updated_at = ?")
		args = append(args, nowUTC(), id)
		res, err := s.db.Exec("UPDATE projects SET "+strings.Join(set, ", ")+" WHERE id = ?", args...)
		if err != nil {
			return Project{}, err
		}
		if err := requireAffected(res); err != nil {
			return Project{}, err
		}
	} else if _, err := s.GetProject(id); err != nil {
		return Project{}, err
	}
	return s.GetProject(id)
}

// UpdateProjectTx applies a partial field update and (optionally) replaces the
// project's tag set atomically in one transaction. It returns ErrNotFound if the
// project does not exist, without partially applying either change.
func (s *Store) UpdateProjectTx(id int64, u ProjectUpdate, tagIDs *[]int64) (Project, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return Project{}, err
	}
	defer tx.Rollback()

	// Existence check up front so a bad id is a clean 404, not an FK error.
	var exists int
	if err := tx.QueryRow(`SELECT 1 FROM projects WHERE id = ?`, id).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Project{}, ErrNotFound
		}
		return Project{}, err
	}

	set, args := buildProjectSet(u)
	if len(set) > 0 {
		set = append(set, "updated_at = ?")
		args = append(args, nowUTC(), id)
		if _, err := tx.Exec("UPDATE projects SET "+strings.Join(set, ", ")+" WHERE id = ?", args...); err != nil {
			return Project{}, err
		}
	}

	if tagIDs != nil {
		if _, err := tx.Exec(`DELETE FROM project_tags WHERE project_id = ?`, id); err != nil {
			return Project{}, err
		}
		stmt, err := tx.Prepare(`INSERT OR IGNORE INTO project_tags (project_id, tag_id) VALUES (?, ?)`)
		if err != nil {
			return Project{}, err
		}
		defer stmt.Close()
		for _, tid := range *tagIDs {
			if _, err := stmt.Exec(id, tid); err != nil {
				return Project{}, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return Project{}, err
	}
	return s.GetProject(id)
}

// buildProjectSet turns a ProjectUpdate into SET clauses and args (excluding
// updated_at, which callers append).
func buildProjectSet(u ProjectUpdate) ([]string, []any) {
	var set []string
	var args []any
	if u.Name != nil {
		set = append(set, "name = ?")
		args = append(args, *u.Name)
	}
	if u.URL != nil {
		set = append(set, "url = ?")
		args = append(args, *u.URL)
	}
	if u.InUse != nil {
		set = append(set, "in_use = ?")
		args = append(args, boolToInt(*u.InUse))
	}
	return set, args
}

// DeleteProject removes a project (and its tag associations via cascade).
func (s *Store) DeleteProject(id int64) error {
	res, err := s.db.Exec(`DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return requireAffected(res)
}

// SetProjectTags replaces a project's tag set with exactly the given tag ids.
func (s *Store) SetProjectTags(projectID int64, tagIDs []int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM project_tags WHERE project_id = ?`, projectID); err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT OR IGNORE INTO project_tags (project_id, tag_id) VALUES (?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, tid := range tagIDs {
		if _, err := stmt.Exec(projectID, tid); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// AddProjectTag attaches a single tag to a project (idempotent).
func (s *Store) AddProjectTag(projectID, tagID int64) error {
	_, err := s.db.Exec(`INSERT OR IGNORE INTO project_tags (project_id, tag_id) VALUES (?, ?)`, projectID, tagID)
	return err
}

// RemoveProjectTag detaches a tag from a project.
func (s *Store) RemoveProjectTag(projectID, tagID int64) error {
	_, err := s.db.Exec(`DELETE FROM project_tags WHERE project_id = ? AND tag_id = ?`, projectID, tagID)
	return err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ImportProjects inserts each project not already tracked (matched on
// source_type + source_identifier, case-insensitively) and attaches tagIDs to
// the new rows, all in one transaction. It returns how many were inserted.
func (s *Store) ImportProjects(ps []Project, tagIDs []int64) (int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	exists, err := tx.Prepare(`SELECT 1 FROM projects
		WHERE source_type = ? AND source_identifier = ? COLLATE NOCASE`)
	if err != nil {
		return 0, err
	}
	defer exists.Close()
	insert, err := tx.Prepare(`
		INSERT INTO projects
			(name, url, source_type, source_identifier, source_host,
			 source_meta, in_use, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, '{}', ?, ?, ?)`)
	if err != nil {
		return 0, err
	}
	defer insert.Close()
	tag, err := tx.Prepare(`INSERT OR IGNORE INTO project_tags (project_id, tag_id) VALUES (?, ?)`)
	if err != nil {
		return 0, err
	}
	defer tag.Close()

	now := nowUTC()
	n := 0
	for _, p := range ps {
		var one int
		err := exists.QueryRow(p.SourceType, p.SourceIdentifier).Scan(&one)
		if err == nil {
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return 0, err
		}
		res, err := insert.Exec(p.Name, p.URL, p.SourceType, p.SourceIdentifier,
			p.SourceHost, boolToInt(p.InUse), now, now)
		if err != nil {
			return 0, err
		}
		id, _ := res.LastInsertId()
		for _, tid := range tagIDs {
			if _, err := tag.Exec(id, tid); err != nil {
				return 0, err
			}
		}
		n++
	}
	return n, tx.Commit()
}
