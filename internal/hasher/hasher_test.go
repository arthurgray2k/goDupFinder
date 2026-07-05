package hasher

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHasher(t *testing.T) {
	tempDir := t.TempDir()

	file1 := filepath.Join(tempDir, "file1.txt")
	os.WriteFile(file1, []byte("hello world"), 0644)

	file2 := filepath.Join(tempDir, "file2.txt")
	os.WriteFile(file2, []byte("hello world"), 0644)

	file3 := filepath.Join(tempDir, "file3.txt")
	os.WriteFile(file3, []byte("hello there"), 0644)

	// Test FullHash
	hash1, err := FullHash(file1, "sha256")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hash2, _ := FullHash(file2, "sha256")
	hash3, _ := FullHash(file3, "sha256")

	if hash1 != hash2 {
		t.Errorf("expected hashes to match for identical files")
	}
	if hash1 == hash3 {
		t.Errorf("expected hashes to differ for different files")
	}

	// Test PartialHash (first 5 bytes)
	phash1, err := PartialHash(file1, "sha256", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	phash3, _ := PartialHash(file3, "sha256", 5) // "hello" for both

	if phash1 != phash3 {
		t.Errorf("expected partial hashes to match since the first 5 bytes are 'hello'")
	}

	// Test CompareFiles
	same, err := CompareFiles(file1, file2)
	if err != nil || !same {
		t.Errorf("expected CompareFiles to return true for identical files, err: %v", err)
	}

	same, err = CompareFiles(file1, file3)
	if err != nil || same {
		t.Errorf("expected CompareFiles to return false for different files, err: %v", err)
	}
}
