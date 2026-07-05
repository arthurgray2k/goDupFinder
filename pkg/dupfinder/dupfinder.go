package dupfinder

import (
	"context"

	"github.com/arthurgray2k/goDupFinder/internal/matcher"
	"github.com/arthurgray2k/goDupFinder/internal/pipeline"
	"github.com/arthurgray2k/goDupFinder/internal/walker"
)

// DuplicateGroup represents a group of identical files.
type DuplicateGroup struct {
	Hash  string   `json:"hash"`
	Files []string `json:"files"`
}

// ProgressHook receives live progress callbacks during a scan.
// Implementations must be safe for concurrent use.
// Pass a nil value (or leave Options.Hook unset) for no-op progress tracking.
type ProgressHook interface {
	// OnFileFound is called for every file that passes the walker's filters.
	OnFileFound(path string, size int64)
	// OnFileHashed is called after a file's full hash has been computed.
	OnFileHashed(path string, bytes int64)
	// OnDirScanned is called when the walker enters a directory.
	OnDirScanned(dir string)
	// OnFileSkipped is called for every file that is excluded by filters.
	OnFileSkipped()
}

// Finder is the main engine for finding duplicate files.
type Finder struct {
	opts Options
}

// New creates a new duplicate finder with the given options.
func New(opts Options) *Finder {
	return &Finder{opts: opts}
}

// Scan traverses the provided directories and returns groups of duplicate files.
// If opts.Hook is set, it receives live progress callbacks throughout the scan.
func (f *Finder) Scan(ctx context.Context, directories []string) ([]DuplicateGroup, error) {
	if err := f.opts.Validate(); err != nil {
		return nil, err
	}

	// Resolve hook — use NoopHook if none supplied.
	var hook pipeline.Hook = pipeline.NoopHook{}
	if f.opts.Hook != nil {
		hook = f.opts.Hook
	}

	wOpts := walker.Config{
		MinSize:          f.opts.MinSize,
		MaxSize:          f.opts.MaxSize,
		MaxDepth:         f.opts.MaxDepth,
		IncludeExts:      f.opts.IncludeExts,
		ExcludeExts:      f.opts.ExcludeExts,
		IncludeHidden:    f.opts.IncludeHidden,
		IgnoreSystemDirs: f.opts.IgnoreSystemDirs,
		Hook:             hook,
	}
	w := walker.New(wOpts)
	fileChan, errChan := w.Walk(ctx, directories)

	mOpts := matcher.Config{
		Workers:        f.opts.Workers,
		Algorithm:      string(f.opts.Algorithm),
		VerifyContents: f.opts.VerifyContents,
		Hook:           hook,
	}
	m := matcher.New(mOpts)
	duplicatesInternal, err := m.Match(ctx, fileChan)
	if err != nil {
		return nil, err
	}

	// Convert internal duplicates to the public type.
	duplicates := make([]DuplicateGroup, 0, len(duplicatesInternal))
	for _, d := range duplicatesInternal {
		duplicates = append(duplicates, DuplicateGroup{
			Hash:  d.Hash,
			Files: d.Files,
		})
	}

	// Surface any fatal walker errors.
	if err := <-errChan; err != nil {
		return nil, err
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return duplicates, nil
}
