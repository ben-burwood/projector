// Package store persists the user's bookmarks (projects, tags) in SQLite.
// It deliberately stores no release/version data — releases are fetched live.
package store

import (
	"database/sql"
	"embed"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	_ "modernc.org/sqlite" // pure-Go, CGO-free SQLite driver
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Store wraps the database handle and query methods.
type Store struct {
	db *sql.DB
}

// Open opens (creating if needed) the SQLite database at path, applies pragmas
// tuned for a single-user app, and runs any pending migrations.
func Open(path string) (*Store, error) {
	// Pragmas via DSN query params (modernc.org/sqlite supports _pragma).
	dsn := fmt.Sprintf("file:%s?%s", path, strings.Join([]string{
		"_pragma=busy_timeout(5000)",
		"_pragma=journal_mode(WAL)",
		"_pragma=foreign_keys(ON)",
		"_pragma=synchronous(NORMAL)",
		"_pragma=temp_store(MEMORY)",
	}, "&"))

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	// Single writer eliminates SQLITE_BUSY for our tiny single-user workload.
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

// Close closes the underlying database.
func (s *Store) Close() error { return s.db.Close() }

// migrate applies embedded migrations in order, tracking progress with the
// SQLite PRAGMA user_version counter. Forward-only.
func (s *Store) migrate() error {
	var current int
	if err := s.db.QueryRow("PRAGMA user_version").Scan(&current); err != nil {
		return fmt.Errorf("read user_version: %w", err)
	}

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		v, err := migrationVersion(name)
		if err != nil {
			return err
		}
		if v <= current {
			continue
		}
		sqlBytes, err := migrationsFS.ReadFile(filepath.ToSlash(filepath.Join("migrations", name)))
		if err != nil {
			return err
		}
		tx, err := s.db.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(string(sqlBytes)); err != nil {
			tx.Rollback()
			return fmt.Errorf("apply %s: %w", name, err)
		}
		// PRAGMA user_version doesn't accept placeholders; v is an int we parsed.
		if _, err := tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", v)); err != nil {
			tx.Rollback()
			return fmt.Errorf("bump user_version to %d: %w", v, err)
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		current = v
	}
	return nil
}

// migrationVersion parses the leading integer of a migration filename such as
// "0001_init.sql" -> 1.
func migrationVersion(name string) (int, error) {
	base := name
	if i := strings.IndexByte(base, '_'); i >= 0 {
		base = base[:i]
	}
	v, err := strconv.Atoi(base)
	if err != nil {
		return 0, fmt.Errorf("migration %q must start with a numeric version: %w", name, err)
	}
	return v, nil
}

// escapeLike escapes LIKE wildcards so user search text is treated literally.
// Use with `LIKE ? ESCAPE '\'`.
func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}
