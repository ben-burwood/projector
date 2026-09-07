package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrNotFound is returned when a row does not exist.
var ErrNotFound = errors.New("not found")

func nowUTC() string { return time.Now().UTC().Format(time.RFC3339) }

// ListTags returns all tags ordered by name, each with its project count.
func (s *Store) ListTags() ([]Tag, error) {
	rows, err := s.db.Query(`
		SELECT t.id, t.name, t.created_at, COUNT(pt.project_id)
		FROM tags t
		LEFT JOIN project_tags pt ON pt.tag_id = t.id
		GROUP BY t.id
		ORDER BY t.name COLLATE NOCASE`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.CreatedAt, &t.ProjectCount); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// GetTag returns a single tag by id.
func (s *Store) GetTag(id int64) (Tag, error) {
	var t Tag
	err := s.db.QueryRow(`SELECT id, name, created_at FROM tags WHERE id = ?`, id).
		Scan(&t.ID, &t.Name, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return t, ErrNotFound
	}
	return t, err
}

// CreateTag inserts a new tag, returning it. Names are unique (case-insensitive).
func (s *Store) CreateTag(name string) (Tag, error) {
	now := nowUTC()
	res, err := s.db.Exec(`INSERT INTO tags (name, created_at) VALUES (?, ?)`, name, now)
	if err != nil {
		return Tag{}, err
	}
	id, _ := res.LastInsertId()
	return Tag{ID: id, Name: name, CreatedAt: now}, nil
}

// EnsureTag returns the existing tag with this name (case-insensitive) or creates it.
func (s *Store) EnsureTag(name string) (Tag, error) {
	var t Tag
	err := s.db.QueryRow(`SELECT id, name, created_at FROM tags WHERE name = ? COLLATE NOCASE`, name).
		Scan(&t.ID, &t.Name, &t.CreatedAt)
	if err == nil {
		return t, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return t, err
	}
	return s.CreateTag(name)
}

// RenameTag updates a tag's name.
func (s *Store) RenameTag(id int64, name string) error {
	res, err := s.db.Exec(`UPDATE tags SET name = ? WHERE id = ?`, name, id)
	if err != nil {
		return err
	}
	return requireAffected(res)
}

// DeleteTag removes a tag (and its project associations via cascade).
func (s *Store) DeleteTag(id int64) error {
	res, err := s.db.Exec(`DELETE FROM tags WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return requireAffected(res)
}

func requireAffected(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// loadTagsFor returns the tags attached to a set of project ids, keyed by project id.
func (s *Store) loadTagsFor(ids []int64) (map[int64][]Tag, error) {
	out := make(map[int64][]Tag, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	// Build an IN clause with placeholders.
	ph := make([]byte, 0, len(ids)*2)
	args := make([]any, len(ids))
	for i, id := range ids {
		if i > 0 {
			ph = append(ph, ',')
		}
		ph = append(ph, '?')
		args[i] = id
	}
	q := fmt.Sprintf(`
		SELECT pt.project_id, t.id, t.name, t.created_at
		FROM project_tags pt
		JOIN tags t ON t.id = pt.tag_id
		WHERE pt.project_id IN (%s)
		ORDER BY t.name COLLATE NOCASE`, ph)
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var pid int64
		var t Tag
		if err := rows.Scan(&pid, &t.ID, &t.Name, &t.CreatedAt); err != nil {
			return nil, err
		}
		out[pid] = append(out[pid], t)
	}
	return out, rows.Err()
}
