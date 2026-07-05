package walker

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWalker(t *testing.T) {
	// Create a temp directory structure
	tempDir := t.TempDir()

	file1 := filepath.Join(tempDir, "file1.txt")
	os.WriteFile(file1, []byte("hello"), 0644)

	file2 := filepath.Join(tempDir, "file2.jpg")
	os.WriteFile(file2, []byte("image"), 0644)

	hiddenFile := filepath.Join(tempDir, ".hidden")
	os.WriteFile(hiddenFile, []byte("secret"), 0644)

	opts := Config{
		MinSize:          1,
		IncludeExts:      []string{"txt", "jpg"},
		IgnoreSystemDirs: true,
	}
	w := New(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	out, errs := w.Walk(ctx, []string{tempDir})

	var files []string
	for item := range out {
		files = append(files, item.Path)
	}

	select {
	case err := <-errs:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	default:
	}

	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}
}
