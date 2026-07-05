package storage

import (
	"context"

	"github.com/arthurgray2k/goDupFinder/internal/pipeline"
)

// Store defines the interface for persisting hashes to speed up subsequent runs.
// The engine must remain independent of this implementation.
type Store interface {
	// Init initializes the storage engine
	Init(ctx context.Context) error

	// Close shuts down the storage engine
	Close() error

	// GetHash retrieves a cached hash for a given file and its metadata (size, modtime)
	GetHash(ctx context.Context, path string, size int64, modTime int64) (string, error)

	// SaveHash persists a hash for a given file
	SaveHash(ctx context.Context, path string, size int64, modTime int64, hash string) error

	// SaveDuplicates saves a discovered duplicate group
	SaveDuplicates(ctx context.Context, group pipeline.DuplicateGroup) error
}
