package matcher

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/arthurgray2k/goDupFinder/internal/pipeline"
)

func TestMatcherPipeline(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Files with unique sizes (should be skipped)
	f1 := filepath.Join(tempDir, "unique1.txt")
	os.WriteFile(f1, []byte("only one file has this content"), 0644)

	// 2. Duplicate files (same content, large size to trigger partial hashing)
	// Let's write content > 4096 bytes
	largeContent := make([]byte, 5000)
	for i := range largeContent {
		largeContent[i] = 'A'
	}
	f2 := filepath.Join(tempDir, "large1.txt")
	f3 := filepath.Join(tempDir, "large2.txt")
	os.WriteFile(f2, largeContent, 0644)
	os.WriteFile(f3, largeContent, 0644)

	// 3. Files with same partial hash, different full content
	// File 4 and 5 both start with 4096 'B's, but end differently
	colContent1 := make([]byte, 5000)
	colContent2 := make([]byte, 5000)
	for i := 0; i < 4096; i++ {
		colContent1[i] = 'B'
		colContent2[i] = 'B'
	}
	colContent1[4999] = 'C'
	colContent2[4999] = 'D'

	f4 := filepath.Join(tempDir, "col1.txt")
	f5 := filepath.Join(tempDir, "col2.txt")
	os.WriteFile(f4, colContent1, 0644)
	os.WriteFile(f5, colContent2, 0644)

	// Test Matcher without verify contents (should separate col1 and col2 because full hash differs, but find large1/large2)
	cfg := Config{
		Workers:        2,
		Algorithm:      "sha256",
		VerifyContents: false,
		Hook:           nil,
	}
	m := New(cfg)

	inChan := make(chan pipeline.FileItem, 10)
	inChan <- pipeline.FileItem{Path: f1, Size: int64(len("only one file has this content"))}
	inChan <- pipeline.FileItem{Path: f2, Size: 5000}
	inChan <- pipeline.FileItem{Path: f3, Size: 5000}
	inChan <- pipeline.FileItem{Path: f4, Size: 5000}
	inChan <- pipeline.FileItem{Path: f5, Size: 5000}
	close(inChan)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	groups, err := m.Match(ctx, inChan)
	if err != nil {
		t.Fatalf("unexpected Match error: %v", err)
	}

	// We expect 1 duplicate group: large1.txt and large2.txt.
	// col1.txt and col2.txt should NOT be grouped because full hashing detects they are different.
	if len(groups) != 1 {
		t.Fatalf("expected 1 duplicate group, got %d", len(groups))
	}

	g := groups[0]
	if len(g.Files) != 2 {
		t.Errorf("expected 2 files in group, got %d", len(g.Files))
	}
}

func TestMatcherVerifyContents(t *testing.T) {
	tempDir := t.TempDir()

	// Create two files with same content to group them
	f1 := filepath.Join(tempDir, "f1.txt")
	f2 := filepath.Join(tempDir, "f2.txt")
	os.WriteFile(f1, []byte("identical content"), 0644)
	os.WriteFile(f2, []byte("identical content"), 0644)

	cfg := Config{
		Workers:        2,
		Algorithm:      "sha256",
		VerifyContents: true,
		Hook:           nil,
	}
	m := New(cfg)

	inChan := make(chan pipeline.FileItem, 2)
	inChan <- pipeline.FileItem{Path: f1, Size: int64(len("identical content"))}
	inChan <- pipeline.FileItem{Path: f2, Size: int64(len("identical content"))}
	close(inChan)

	groups, err := m.Match(context.Background(), inChan)
	if err != nil {
		t.Fatalf("unexpected Match error: %v", err)
	}

	if len(groups) != 1 {
		t.Fatalf("expected 1 duplicate group, got %d", len(groups))
	}
}
