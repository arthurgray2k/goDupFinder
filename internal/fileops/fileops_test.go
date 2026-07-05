package fileops

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDelete(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	os.WriteFile(a, []byte("data"), 0644)
	os.WriteFile(b, []byte("data"), 0644)

	op := New(Options{DryRun: false, KeepFirst: true, SkipConfirm: true}, nil)
	results := op.Delete([]string{a, b})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Err != nil {
		t.Fatalf("unexpected error: %v", results[0].Err)
	}
	if _, err := os.Stat(b); !os.IsNotExist(err) {
		t.Errorf("expected b.txt to be deleted")
	}
	if _, err := os.Stat(a); err != nil {
		t.Errorf("expected a.txt to be kept: %v", err)
	}
}

func TestDeleteDryRun(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	os.WriteFile(a, []byte("data"), 0644)
	os.WriteFile(b, []byte("data"), 0644)

	op := New(Options{DryRun: true, KeepFirst: true, SkipConfirm: true}, nil)
	results := op.Delete([]string{a, b})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].DryRun {
		t.Errorf("expected DryRun result")
	}
	// Both files must still exist
	for _, f := range []string{a, b} {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("dry-run must not delete %s", f)
		}
	}
}

func TestMove(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "moved")
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	os.WriteFile(a, []byte("data"), 0644)
	os.WriteFile(b, []byte("data"), 0644)

	op := New(Options{DryRun: false, KeepFirst: true, SkipConfirm: true, MoveDestination: dest}, nil)
	results := op.Move([]string{a, b})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Err != nil {
		t.Fatalf("unexpected error: %v", results[0].Err)
	}
	// b should now be in dest
	if _, err := os.Stat(filepath.Join(dest, "b.txt")); err != nil {
		t.Errorf("expected b.txt in dest: %v", err)
	}
}

func TestHardLink(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	os.WriteFile(a, []byte("data"), 0644)
	os.WriteFile(b, []byte("data"), 0644)

	op := New(Options{DryRun: false, SkipConfirm: true}, nil)
	results := op.HardLink([]string{a, b})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Err != nil {
		t.Fatalf("unexpected hard link error: %v", results[0].Err)
	}
}

func TestConfirmSkip(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	os.WriteFile(a, []byte("data"), 0644)
	os.WriteFile(b, []byte("data"), 0644)

	// Confirm function that always says NO
	neverConfirm := func(op, path string) bool { return false }
	op := New(Options{DryRun: false, KeepFirst: true}, neverConfirm)
	results := op.Delete([]string{a, b})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Skipped {
		t.Errorf("expected result to be Skipped")
	}
	// b.txt must NOT have been deleted
	if _, err := os.Stat(b); err != nil {
		t.Errorf("file must remain when user declines: %v", err)
	}
}
