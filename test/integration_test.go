// Package test contains end-to-end integration tests for goDupFinder.
// Tests run against temporary directories with a known file tree and assert
// that the scanner detects exactly the expected duplicate groups.
package test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/arthurgray2k/goDupFinder/pkg/dupfinder"
)

// createTree writes a tree of files to a temp directory.
// Spec: map[relativePath]content
func createTree(t *testing.T, spec map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for relPath, content := range spec {
		full := filepath.Join(root, relPath)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatalf("createTree: mkdir %s: %v", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatalf("createTree: write %s: %v", full, err)
		}
	}
	return root
}

// sortedGroups returns duplicate groups with file paths sorted, then groups sorted
// by their hash, making comparisons deterministic.
func sortedGroups(groups []dupfinder.DuplicateGroup) []dupfinder.DuplicateGroup {
	for i := range groups {
		sort.Strings(groups[i].Files)
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Hash < groups[j].Hash
	})
	return groups
}

// TestNoDuplicates asserts that a tree of unique files returns zero groups.
func TestNoDuplicates(t *testing.T) {
	root := createTree(t, map[string]string{
		"a.txt": "alpha",
		"b.txt": "beta",
		"c.txt": "gamma",
	})

	opts := dupfinder.DefaultOptions()
	groups, err := dupfinder.New(opts).Scan(context.Background(), []string{root})
	if err != nil {
		t.Fatalf("Scan error: %v", err)
	}
	if len(groups) != 0 {
		t.Errorf("expected 0 duplicate groups, got %d", len(groups))
	}
}

// TestOneDuplicateGroup asserts a single group of two identical files.
func TestOneDuplicateGroup(t *testing.T) {
	const content = "duplicate content"
	root := createTree(t, map[string]string{
		"original.txt": content,
		"copy.txt":     content,
		"unique.txt":   "something else entirely",
	})

	opts := dupfinder.DefaultOptions()
	groups, err := dupfinder.New(opts).Scan(context.Background(), []string{root})
	if err != nil {
		t.Fatalf("Scan error: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 duplicate group, got %d", len(groups))
	}
	if len(groups[0].Files) != 2 {
		t.Errorf("expected 2 files in group, got %d", len(groups[0].Files))
	}
}

// TestMultipleDuplicateGroups asserts that two independent groups are detected.
func TestMultipleDuplicateGroups(t *testing.T) {
	root := createTree(t, map[string]string{
		"img_a1.jpg": "image-data-A",
		"img_a2.jpg": "image-data-A",
		"doc_b1.pdf": "document-data-B",
		"doc_b2.pdf": "document-data-B",
		"doc_b3.pdf": "document-data-B",
		"solo.txt":   "i am alone",
	})

	opts := dupfinder.DefaultOptions()
	groups, err := dupfinder.New(opts).Scan(context.Background(), []string{root})
	if err != nil {
		t.Fatalf("Scan error: %v", err)
	}
	groups = sortedGroups(groups)

	if len(groups) != 2 {
		t.Fatalf("expected 2 duplicate groups, got %d", len(groups))
	}
	// Group A: 2 files
	// Group B: 3 files
	// The order after sorting by hash is not predictable, so check sizes.
	sizes := []int{len(groups[0].Files), len(groups[1].Files)}
	sort.Ints(sizes)
	if sizes[0] != 2 || sizes[1] != 3 {
		t.Errorf("expected group sizes [2,3], got %v", sizes)
	}
}

// TestSubdirectoryTraversal ensures duplicates are found across sub-directories.
func TestSubdirectoryTraversal(t *testing.T) {
	const content = "shared binary content"
	root := createTree(t, map[string]string{
		"photos/2024/img.jpg":  content,
		"backup/2024/img.jpg":  content,
		"archive/2024/img.jpg": content,
	})

	opts := dupfinder.DefaultOptions()
	groups, err := dupfinder.New(opts).Scan(context.Background(), []string{root})
	if err != nil {
		t.Fatalf("Scan error: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 duplicate group across subdirs, got %d", len(groups))
	}
	if len(groups[0].Files) != 3 {
		t.Errorf("expected 3 files in group, got %d", len(groups[0].Files))
	}
}

// TestMultipleRootDirectories scans two separate root directories.
func TestMultipleRootDirectories(t *testing.T) {
	const content = "cross-root duplicate"
	root1 := createTree(t, map[string]string{"file.txt": content})
	root2 := createTree(t, map[string]string{"file.txt": content})

	opts := dupfinder.DefaultOptions()
	groups, err := dupfinder.New(opts).Scan(context.Background(), []string{root1, root2})
	if err != nil {
		t.Fatalf("Scan error: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 duplicate group, got %d", len(groups))
	}
	if len(groups[0].Files) != 2 {
		t.Errorf("expected 2 files in group, got %d", len(groups[0].Files))
	}
}

// TestMinSizeFilter checks that files below MinSize are ignored.
func TestMinSizeFilter(t *testing.T) {
	root := createTree(t, map[string]string{
		"small1.txt": "hi", // 2 bytes — below MinSize
		"small2.txt": "hi", // 2 bytes — below MinSize
		"big1.txt":   "hello world big file content that is long enough",
		"big2.txt":   "hello world big file content that is long enough",
	})

	opts := dupfinder.DefaultOptions()
	opts.MinSize = 10 // only consider files >= 10 bytes
	groups, err := dupfinder.New(opts).Scan(context.Background(), []string{root})
	if err != nil {
		t.Fatalf("Scan error: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 group (small files excluded), got %d", len(groups))
	}
}

// TestExtensionFilter checks that only matching extensions are scanned.
func TestExtensionFilter(t *testing.T) {
	const content = "same content"
	root := createTree(t, map[string]string{
		"photo1.jpg": content,
		"photo2.jpg": content,
		"doc1.pdf":   content, // should be excluded when filtering by jpg
		"doc2.pdf":   content,
	})

	opts := dupfinder.DefaultOptions()
	opts.IncludeExts = []string{"jpg"}
	groups, err := dupfinder.New(opts).Scan(context.Background(), []string{root})
	if err != nil {
		t.Fatalf("Scan error: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 group (only jpg), got %d", len(groups))
	}
	for _, f := range groups[0].Files {
		if filepath.Ext(f) != ".jpg" {
			t.Errorf("found non-jpg file in results: %s", f)
		}
	}
}

// TestContextCancellation checks that cancelling the context stops the scan.
func TestContextCancellation(t *testing.T) {
	// Build a moderately large tree.
	root := t.TempDir()
	const content = "cancellation test data"
	for i := 0; i < 200; i++ {
		path := filepath.Join(root, fmt.Sprintf("file%04d.txt", i))
		os.WriteFile(path, []byte(fmt.Sprintf("%s-%d", content, i)), 0644)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	opts := dupfinder.DefaultOptions()
	// Should not panic or hang, even with cancelled context.
	_, err := dupfinder.New(opts).Scan(ctx, []string{root})
	_ = err // may or may not return an error; the key requirement is it returns promptly
}

// TestProgressHookCounts verifies that the hook counters match the actual file tree.
func TestProgressHookCounts(t *testing.T) {
	root := createTree(t, map[string]string{
		"a.txt": "alpha",
		"b.txt": "beta",
		"c.txt": "alpha", // duplicate of a.txt
	})

	var found, hashed, dirs int

	hook := &testHook{
		onFound:  func() { found++ },
		onHashed: func() { hashed++ },
		onDir:    func() { dirs++ },
	}

	opts := dupfinder.DefaultOptions()
	opts.Hook = hook
	_, err := dupfinder.New(opts).Scan(context.Background(), []string{root})
	if err != nil {
		t.Fatalf("Scan error: %v", err)
	}

	if found != 3 {
		t.Errorf("expected OnFileFound called 3 times, got %d", found)
	}
	// Only a.txt and c.txt share a size, so 2 files are hashed.
	if hashed != 2 {
		t.Errorf("expected OnFileHashed called 2 times, got %d", hashed)
	}
	// At minimum the root dir itself is scanned.
	if dirs < 1 {
		t.Errorf("expected OnDirScanned called at least once, got %d", dirs)
	}
}

// testHook is a simple hook for unit testing.
type testHook struct {
	onFound  func()
	onHashed func()
	onDir    func()
}

func (h *testHook) OnFileFound(_ string, _ int64)  { h.onFound() }
func (h *testHook) OnFileHashed(_ string, _ int64) { h.onHashed() }
func (h *testHook) OnDirScanned(_ string)          { h.onDir() }
func (h *testHook) OnFileSkipped()                 {}
