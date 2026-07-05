package walker

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/arthurgray2k/goDupFinder/internal/pipeline"
)

// Config controls walker behaviour.
type Config struct {
	MinSize          int64
	MaxSize          int64
	MaxDepth         int
	IncludeExts      []string
	ExcludeExts      []string
	IncludeHidden    bool
	IgnoreSystemDirs bool
	// Hook receives progress callbacks. If nil, a NoopHook is used.
	Hook pipeline.Hook
}

// Walker is responsible for recursively scanning directories and yielding files.
type Walker struct {
	opts Config
	hook pipeline.Hook
}

// New creates a new Walker with the given options.
func New(opts Config) *Walker {
	h := opts.Hook
	if h == nil {
		h = pipeline.NoopHook{}
	}
	return &Walker{opts: opts, hook: h}
}

// Walk traverses the given directories concurrently and sends discovered files
// to the returned channel. It returns an error channel that receives any fatal errors.
func (w *Walker) Walk(ctx context.Context, directories []string) (<-chan pipeline.FileItem, <-chan error) {
	out := make(chan pipeline.FileItem, 1000)
	errs := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errs)

		var wg sync.WaitGroup

		for _, dir := range directories {
			wg.Add(1)
			go func(root string) {
				defer wg.Done()

				err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						// Skip files/dirs we can't access
						return nil
					}

					// Check context cancellation
					select {
					case <-ctx.Done():
						return ctx.Err()
					default:
					}

					if w.opts.MaxDepth > 0 {
						depth := 0
						if path != root {
							rel, err := filepath.Rel(root, path)
							if err == nil {
								depth = strings.Count(rel, string(os.PathSeparator)) + 1
							}
						}
						if depth > w.opts.MaxDepth {
							if d.IsDir() {
								return filepath.SkipDir
							}
							return nil
						}
					}

					if d.IsDir() {
						if w.shouldSkipDir(d.Name()) {
							return filepath.SkipDir
						}
						w.hook.OnDirScanned(path)
						return nil
					}

					if !d.Type().IsRegular() {
						return nil
					}

					info, err := d.Info()
					if err != nil {
						return nil
					}

					if w.shouldSkipFile(d.Name(), info.Size()) {
						w.hook.OnFileSkipped()
						return nil
					}

					w.hook.OnFileFound(path, info.Size())

					select {
					case out <- pipeline.FileItem{Path: path, Size: info.Size()}:
					case <-ctx.Done():
						return ctx.Err()
					}

					return nil
				})

				if err != nil && err != context.Canceled {
					select {
					case errs <- err:
					default:
					}
				}
			}(dir)
		}

		wg.Wait()
	}()

	return out, errs
}

func (w *Walker) shouldSkipDir(name string) bool {
	if name == "" || name == "." {
		return false
	}
	if !w.opts.IncludeHidden && strings.HasPrefix(name, ".") {
		return true
	}
	if w.opts.IgnoreSystemDirs {
		switch name {
		case "node_modules", ".git", ".svn", "vendor":
			return true
		}
	}
	return false
}

func (w *Walker) shouldSkipFile(name string, size int64) bool {
	if !w.opts.IncludeHidden && strings.HasPrefix(name, ".") {
		return true
	}
	if size < w.opts.MinSize {
		return true
	}
	if w.opts.MaxSize > 0 && size > w.opts.MaxSize {
		return true
	}
	if len(w.opts.IncludeExts) > 0 {
		ext := strings.TrimPrefix(filepath.Ext(name), ".")
		found := false
		for _, e := range w.opts.IncludeExts {
			if strings.EqualFold(e, ext) {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}
	if len(w.opts.ExcludeExts) > 0 {
		ext := strings.TrimPrefix(filepath.Ext(name), ".")
		for _, e := range w.opts.ExcludeExts {
			if strings.EqualFold(e, ext) {
				return true
			}
		}
	}
	return false
}
