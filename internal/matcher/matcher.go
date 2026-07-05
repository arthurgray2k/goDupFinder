package matcher

import (
	"context"
	"os"
	"sync"

	"github.com/arthurgray2k/goDupFinder/internal/hasher"
	"github.com/arthurgray2k/goDupFinder/internal/pipeline"
)

// Config controls matcher behaviour.
type Config struct {
	Workers        int
	Algorithm      string
	VerifyContents bool
	// Hook receives progress callbacks. If nil, a NoopHook is used.
	Hook pipeline.Hook
}

// Matcher identifies groups of duplicate files.
type Matcher struct {
	opts Config
	hook pipeline.Hook
}

// New creates a new Matcher.
func New(opts Config) *Matcher {
	h := opts.Hook
	if h == nil {
		h = pipeline.NoopHook{}
	}
	return &Matcher{opts: opts, hook: h}
}

// Match consumes FileItems from in, groups them by size then by full hash,
// and returns groups with two or more identical files.
func (m *Matcher) Match(ctx context.Context, in <-chan pipeline.FileItem) ([]pipeline.DuplicateGroup, error) {
	// Phase 1: Group by size. Files with a unique size cannot be duplicates.
	sizeGroups := make(map[int64][]string)
	for item := range in {
		sizeGroups[item.Size] = append(sizeGroups[item.Size], item.Path)
	}

	// Collect only the size groups that have at least two files.
	var candidates [][]string
	var smallCandidates [][]string // size <= 4096
	const partialHashLimit = 4096

	for size, paths := range sizeGroups {
		if len(paths) > 1 {
			if size <= partialHashLimit {
				smallCandidates = append(smallCandidates, paths)
			} else {
				candidates = append(candidates, paths)
			}
		}
	}

	// Phase 2: Partial Hashing for files > 4KB
	// Group candidate paths by partial hash.
	var postPartialCandidates [][]string
	if len(candidates) > 0 {
		type partialResult struct {
			path string
			hash string
			err  error
		}

		sem := make(chan struct{}, m.opts.Workers)
		for _, group := range candidates {
			var wg sync.WaitGroup
			results := make(chan partialResult, len(group))

			for _, path := range group {
				wg.Add(1)
				go func(p string) {
					defer wg.Done()
					select {
					case <-ctx.Done():
						return
					default:
					}

					sem <- struct{}{}
					defer func() { <-sem }()

					h, err := hasher.PartialHash(p, m.opts.Algorithm, partialHashLimit)
					results <- partialResult{path: p, hash: h, err: err}
				}(path)
			}

			wg.Wait()
			close(results)

			// Group by partial hash
			pGroups := make(map[string][]string)
			for res := range results {
				if res.err == nil {
					pGroups[res.hash] = append(pGroups[res.hash], res.path)
				}
			}

			for _, paths := range pGroups {
				if len(paths) > 1 {
					postPartialCandidates = append(postPartialCandidates, paths)
				}
			}
		}
	}

	// Merge all files that need full hashing
	allCandidates := append(smallCandidates, postPartialCandidates...)

	if len(allCandidates) == 0 {
		return nil, nil
	}

	// Phase 3: Full hash — bounded worker pool via semaphore.
	type hashResult struct {
		path string
		hash string
		size int64
		err  error
	}

	hashGroups := make(map[string][]string)
	var mu sync.Mutex

	sem := make(chan struct{}, m.opts.Workers)

	for _, group := range allCandidates {
		var wg sync.WaitGroup
		results := make(chan hashResult, len(group))

		for _, path := range group {
			wg.Add(1)
			go func(p string) {
				defer wg.Done()

				select {
				case <-ctx.Done():
					return
				default:
				}

				sem <- struct{}{}
				defer func() { <-sem }()

				var size int64
				if fi, err := os.Stat(p); err == nil {
					size = fi.Size()
				}

				h, err := hasher.FullHash(p, m.opts.Algorithm)
				results <- hashResult{path: p, hash: h, size: size, err: err}
			}(path)
		}

		wg.Wait()
		close(results)

		for res := range results {
			if res.err == nil {
				m.hook.OnFileHashed(res.path, res.size)
				mu.Lock()
				hashGroups[res.hash] = append(hashGroups[res.hash], res.path)
				mu.Unlock()
			}
		}
	}

	// Phase 4: Emit groups with 2+ files and optionally verify byte-for-byte.
	var duplicates []pipeline.DuplicateGroup
	for h, paths := range hashGroups {
		if len(paths) > 1 {
			if m.opts.VerifyContents {
				verifiedSubGroups := verifyGroup(paths)
				for _, subPaths := range verifiedSubGroups {
					if len(subPaths) > 1 {
						duplicates = append(duplicates, pipeline.DuplicateGroup{
							Hash:  h,
							Files: subPaths,
						})
					}
				}
			} else {
				duplicates = append(duplicates, pipeline.DuplicateGroup{
					Hash:  h,
					Files: paths,
				})
			}
		}
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return duplicates, nil
}

// verifyGroup splits a duplicate group into verified subgroups using byte-for-byte check.
func verifyGroup(group []string) [][]string {
	if len(group) <= 1 {
		return [][]string{group}
	}
	var subGroups [][]string
	for _, file := range group {
		matched := false
		for i, sg := range subGroups {
			eq, err := hasher.CompareFiles(sg[0], file)
			if err == nil && eq {
				subGroups[i] = append(subGroups[i], file)
				matched = true
				break
			}
		}
		if !matched {
			subGroups = append(subGroups, []string{file})
		}
	}
	return subGroups
}
