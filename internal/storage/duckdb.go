//go:build duckdb

package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/arthurgray2k/goDupFinder/internal/pipeline"
	_ "github.com/marcboeker/go-duckdb"
)

// DuckDBStore implements the Store interface using DuckDB.
// This provides fast analytical capabilities and caching for massive datasets.
type DuckDBStore struct {
	dbPath string
	db     *sql.DB
}

// NewDuckDBStore creates a new DuckDB store.
func NewDuckDBStore(dbPath string) *DuckDBStore {
	return &DuckDBStore{dbPath: dbPath}
}

func (s *DuckDBStore) Init(ctx context.Context) error {
	db, err := sql.Open("duckdb", s.dbPath)
	if err != nil {
		return err
	}
	s.db = db

	// Create tables if not exists
	const schema = `
		CREATE TABLE IF NOT EXISTS file_hashes (
			path VARCHAR PRIMARY KEY,
			size BIGINT,
			mod_time BIGINT,
			hash VARCHAR
		);
		CREATE TABLE IF NOT EXISTS duplicates (
			group_hash VARCHAR,
			path VARCHAR
		);
	`
	_, err = s.db.ExecContext(ctx, schema)
	return err
}

func (s *DuckDBStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *DuckDBStore) GetHash(ctx context.Context, path string, size int64, modTime int64) (string, error) {
	var hash string
	var dbSize, dbModTime int64

	err := s.db.QueryRowContext(ctx, "SELECT size, mod_time, hash FROM file_hashes WHERE path = ?", path).
		Scan(&dbSize, &dbModTime, &hash)

	if err == sql.ErrNoRows {
		return "", nil // cache miss
	}
	if err != nil {
		return "", err
	}

	// Cache invalidation: if size or modTime changed, return empty
	if dbSize != size || dbModTime != modTime {
		return "", nil
	}

	return hash, nil
}

func (s *DuckDBStore) SaveHash(ctx context.Context, path string, size int64, modTime int64, hash string) error {
	query := `
		INSERT INTO file_hashes (path, size, mod_time, hash) 
		VALUES (?, ?, ?, ?)
		ON CONFLICT (path) DO UPDATE SET 
			size = EXCLUDED.size, 
			mod_time = EXCLUDED.mod_time, 
			hash = EXCLUDED.hash
	`
	_, err := s.db.ExecContext(ctx, query, path, size, modTime, hash)
	return err
}

func (s *DuckDBStore) SaveDuplicates(ctx context.Context, group pipeline.DuplicateGroup) error {
	// Optional analytical export to DB
	for _, file := range group.Files {
		query := `INSERT INTO duplicates (group_hash, path) VALUES (?, ?)`
		if _, err := s.db.ExecContext(ctx, query, group.Hash, file); err != nil {
			return fmt.Errorf("failed to save duplicate %s: %w", file, err)
		}
	}
	return nil
}
