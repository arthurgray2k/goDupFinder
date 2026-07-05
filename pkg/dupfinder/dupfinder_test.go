package dupfinder

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFinder_Scan(t *testing.T) {
	tempDir := t.TempDir()

	// Create some dummy files
	files := map[string]string{
		"file1.txt":   "hello world",
		"file2.txt":   "hello world", // Duplicate of file1
		"file3.txt":   "foo bar",
		"file4.txt":   "foo bar", // Duplicate of file3
		"file5.txt":   "unique content",
		"empty1.txt":  "",
		"empty2.txt":  "",
		"ignore.log":  "hello world", // Duplicate, but different extension (if we test filters)
		"hidden/.txt": "hidden",
	}

	for name, content := range files {
		path := filepath.Join(tempDir, filepath.FromSlash(name))
		err := os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	opts := DefaultOptions()
	opts.Algorithm = SHA256
	
	finder := New(opts)
	duplicates, err := finder.Scan(context.Background(), []string{tempDir})
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// We expect empty files to be ignored by default (MinSize=1),
	// so "hello world" and "foo bar" and "unique" are left.
	// "hello world" appears 3 times. "foo bar" appears 2 times.
	
	if len(duplicates) != 2 {
		t.Errorf("Expected 2 duplicate groups, got %d", len(duplicates))
	}
}

func TestFinder_ScanValidation(t *testing.T) {
	opts := Options{Workers: -1} // Invalid options
	finder := New(opts)
	_, err := finder.Scan(context.Background(), []string{"."})
	if err == nil {
		t.Error("Expected error for invalid options, got nil")
	}
}

func TestFinder_ScanCanceled(t *testing.T) {
	tempDir := t.TempDir()
	os.WriteFile(filepath.Join(tempDir, "a.txt"), []byte("a"), 0644)

	opts := DefaultOptions()
	finder := New(opts)
	
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := finder.Scan(ctx, []string{tempDir})
	if err == nil || err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}
