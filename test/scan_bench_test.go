package test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/arthurgray2k/goDupFinder/pkg/dupfinder"
)

// buildBenchTree creates a directory tree for scanning benchmarks.
// It writes fileCount unique files and dupeCount duplicates of a shared template.
func buildBenchTree(b *testing.B, fileCount, dupeCount int) string {
	b.Helper()
	root := b.TempDir()

	// Unique files
	for i := 0; i < fileCount; i++ {
		path := filepath.Join(root, fmt.Sprintf("unique_%06d.txt", i))
		content := fmt.Sprintf("unique content %d — some padding to make files non-trivially sized xxxxxxxxxx", i)
		os.WriteFile(path, []byte(content), 0644)
	}

	// Duplicate files (all share the same content)
	const dupeContent = "this is the duplicate content shared across all duplicate files in this benchmark"
	for i := 0; i < dupeCount; i++ {
		path := filepath.Join(root, fmt.Sprintf("dupe_%06d.txt", i))
		os.WriteFile(path, []byte(dupeContent), 0644)
	}

	return root
}

// BenchmarkScan measures end-to-end scan performance.
func BenchmarkScan(b *testing.B) {
	cases := []struct {
		unique int
		dupes  int
	}{
		{100, 10},
		{500, 50},
		{1000, 100},
	}

	for _, tc := range cases {
		tc := tc
		name := fmt.Sprintf("unique=%d/dupes=%d", tc.unique, tc.dupes)
		b.Run(name, func(b *testing.B) {
			root := buildBenchTree(b, tc.unique, tc.dupes)
			opts := dupfinder.DefaultOptions()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				groups, err := dupfinder.New(opts).Scan(context.Background(), []string{root})
				if err != nil {
					b.Fatalf("Scan: %v", err)
				}
				if len(groups) == 0 && tc.dupes > 0 {
					b.Fatal("expected at least one duplicate group")
				}
			}
		})
	}
}

// BenchmarkScanWorkers compares scan speed at different worker counts.
func BenchmarkScanWorkers(b *testing.B) {
	root := buildBenchTree(b, 500, 50)

	for _, workers := range []int{1, 2, 4, 8, 16} {
		workers := workers
		b.Run(fmt.Sprintf("workers=%d", workers), func(b *testing.B) {
			opts := dupfinder.DefaultOptions()
			opts.Workers = workers

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dupfinder.New(opts).Scan(context.Background(), []string{root}) //nolint:errcheck
			}
		})
	}
}
