package progress

import (
	"bytes"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestFormatBytes(t *testing.T) {
	cases := []struct {
		input    int64
		contains string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.00 KB"},
		{1024 * 1024, "1.00 MB"},
		{1024 * 1024 * 1024, "1.00 GB"},
	}
	for _, c := range cases {
		got := formatBytes(c.input)
		if !strings.Contains(got, c.contains) {
			t.Errorf("formatBytes(%d) = %q, want it to contain %q", c.input, got, c.contains)
		}
	}
}

func TestStats(t *testing.T) {
	s := NewStats()
	s.FilesScanned.Add(100)
	s.BytesProcessed.Add(1024 * 1024)
	s.DuplicateGroups.Add(3)
	s.DuplicateFiles.Add(9)

	if s.FilesScanned.Load() != 100 {
		t.Errorf("expected 100 files scanned")
	}
	if s.FilesPerSecond() < 0 {
		t.Errorf("FilesPerSecond must not be negative")
	}
}

func TestStatsSummary(t *testing.T) {
	s := NewStats()
	s.FilesScanned.Add(50)
	s.BytesProcessed.Add(2048)

	var buf bytes.Buffer
	s.Summary(&buf)
	out := buf.String()

	for _, want := range []string{"Files scanned", "Elapsed", "Throughput"} {
		if !strings.Contains(out, want) {
			t.Errorf("summary missing %q", want)
		}
	}
}

func TestTrackerEnabled(t *testing.T) {
	s := NewStats()
	var buf bytes.Buffer
	tracker := NewTracker(s, &buf, true)

	tracker.SetCurrentFile("/some/file.jpg")
	tracker.SetCurrentDir("/some")
	tracker.Start()
	// Let it tick once
	s.FilesScanned.Add(5)
	tracker.Stop()
}

func TestTrackerDisabled(t *testing.T) {
	s := NewStats()
	var buf bytes.Buffer
	tracker := NewTracker(s, &buf, false)
	tracker.Start()
	tracker.SetCurrentFile("/should/not/write")
	tracker.Stop()

	if buf.Len() != 0 {
		t.Errorf("disabled tracker must not write to output")
	}
}

func TestTruncate(t *testing.T) {
	long := "/very/long/path/to/some/deep/nested/file.jpg"
	short := truncate(long, 20)
	if utf8.RuneCountInString(short) > 20 {
		t.Errorf("truncate result too long: %d runes", utf8.RuneCountInString(short))
	}
	// Short strings must be returned as-is
	if truncate("hi", 20) != "hi" {
		t.Errorf("short string must not be modified")
	}
}
