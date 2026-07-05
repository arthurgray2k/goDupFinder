// Package progress provides a real-time terminal progress display and statistics tracker.
// It is designed to be safe for concurrent use across multiple goroutines.
package progress

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

// Stats holds cumulative scan statistics.
// All integer fields are updated atomically for safe concurrent access.
type Stats struct {
	FilesScanned    atomic.Int64
	DirsScanned     atomic.Int64
	BytesProcessed  atomic.Int64
	DuplicateGroups atomic.Int64
	DuplicateFiles  atomic.Int64
	WastedBytes     atomic.Int64
	FilesSkipped    atomic.Int64

	startTime time.Time
}

// NewStats creates a Stats tracker, recording the current time as the start.
func NewStats() *Stats {
	s := &Stats{}
	s.startTime = time.Now()
	return s
}

// Elapsed returns wall-clock time since scanning started.
func (s *Stats) Elapsed() time.Duration {
	return time.Since(s.startTime)
}

// FilesPerSecond returns the average scanning throughput.
func (s *Stats) FilesPerSecond() float64 {
	elapsed := s.Elapsed().Seconds()
	if elapsed == 0 {
		return 0
	}
	return float64(s.FilesScanned.Load()) / elapsed
}

// Throughput returns bytes processed per second.
func (s *Stats) Throughput() float64 {
	elapsed := s.Elapsed().Seconds()
	if elapsed == 0 {
		return 0
	}
	return float64(s.BytesProcessed.Load()) / elapsed
}

// Summary prints a final stats report to the given writer.
func (s *Stats) Summary(w io.Writer) {
	elapsed := s.Elapsed()
	fmt.Fprintln(w, "\n─────────────────────────────────")
	fmt.Fprintln(w, " Scan Complete")
	fmt.Fprintln(w, "─────────────────────────────────")
	fmt.Fprintf(w, " Files scanned:     %d\n", s.FilesScanned.Load())
	fmt.Fprintf(w, " Dirs scanned:      %d\n", s.DirsScanned.Load())
	fmt.Fprintf(w, " Files skipped:     %d\n", s.FilesSkipped.Load())
	fmt.Fprintf(w, " Bytes processed:   %s\n", formatBytes(s.BytesProcessed.Load()))
	fmt.Fprintf(w, " Duplicate groups:  %d\n", s.DuplicateGroups.Load())
	fmt.Fprintf(w, " Duplicate files:   %d\n", s.DuplicateFiles.Load())
	fmt.Fprintf(w, " Wasted space:      %s\n", formatBytes(s.WastedBytes.Load()))
	fmt.Fprintf(w, " Elapsed:           %s\n", elapsed.Round(time.Millisecond))
	fmt.Fprintf(w, " Files/sec:         %.0f\n", s.FilesPerSecond())
	fmt.Fprintf(w, " Throughput:        %s/s\n", formatBytes(int64(s.Throughput())))
	fmt.Fprintln(w, "─────────────────────────────────")
}

// Tracker renders a live progress bar to a terminal.
type Tracker struct {
	stats    *Stats
	out      io.Writer
	mu       sync.Mutex
	lastFile string
	lastDir  string
	ticker   *time.Ticker
	done     chan struct{}
	width    int
	enabled  bool
}

// NewTracker creates a Tracker that writes progress updates to out (usually os.Stderr).
// Set enabled=false to suppress all terminal output (e.g. when writing JSON to stdout).
func NewTracker(stats *Stats, out io.Writer, enabled bool) *Tracker {
	if out == nil {
		out = os.Stderr
	}
	return &Tracker{
		stats:   stats,
		out:     out,
		done:    make(chan struct{}),
		width:   80,
		enabled: enabled,
	}
}

// SetCurrentFile records the path being hashed right now.
func (t *Tracker) SetCurrentFile(path string) {
	if !t.enabled {
		return
	}
	t.mu.Lock()
	t.lastFile = path
	t.mu.Unlock()
}

// SetCurrentDir records the directory being walked right now.
func (t *Tracker) SetCurrentDir(dir string) {
	if !t.enabled {
		return
	}
	t.mu.Lock()
	t.lastDir = dir
	t.mu.Unlock()
}

// Start begins the background render loop.
func (t *Tracker) Start() {
	if !t.enabled {
		return
	}
	t.ticker = time.NewTicker(200 * time.Millisecond)
	go func() {
		for {
			select {
			case <-t.ticker.C:
				t.render()
			case <-t.done:
				return
			}
		}
	}()
}

// Stop halts the render loop and clears the progress line.
func (t *Tracker) Stop() {
	if !t.enabled {
		return
	}
	if t.ticker != nil {
		t.ticker.Stop()
	}
	close(t.done)
	// Clear the progress line
	fmt.Fprintf(t.out, "\r%s\r", strings.Repeat(" ", t.width))
}

// render writes a single progress line to the terminal.
func (t *Tracker) render() {
	t.mu.Lock()
	lastFile := t.lastFile
	t.mu.Unlock()

	s := t.stats
	files := s.FilesScanned.Load()
	bytes := s.BytesProcessed.Load()
	fps := s.FilesPerSecond()
	throughput := s.Throughput()
	elapsed := s.Elapsed().Round(time.Second)

	shortFile := truncate(lastFile, 32)

	line := fmt.Sprintf(
		"\r[%s] Files: %6d | %s | %.0f files/s | %s/s | %s",
		elapsed,
		files,
		formatBytes(bytes),
		fps,
		formatBytes(int64(throughput)),
		shortFile,
	)

	// Pad to width and trim if too long
	if len(line) < t.width {
		line += strings.Repeat(" ", t.width-len(line))
	} else if len(line) > t.width {
		line = line[:t.width]
	}

	fmt.Fprint(t.out, line)
}

// formatBytes converts a byte count to a human-readable string.
func formatBytes(b int64) string {
	if b == 0 {
		return "0 B"
	}
	units := []string{"B", "KB", "MB", "GB", "TB"}
	exp := int(math.Log(float64(b)) / math.Log(1024))
	if exp >= len(units) {
		exp = len(units) - 1
	}
	val := float64(b) / math.Pow(1024, float64(exp))
	if exp == 0 {
		return fmt.Sprintf("%d %s", b, units[exp])
	}
	return fmt.Sprintf("%.2f %s", val, units[exp])
}

// truncate shortens a string to at most maxLen runes, adding "…" prefix if truncated.
func truncate(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	runes := []rune(s)
	// Reserve 1 rune for the ellipsis
	start := len(runes) - (maxLen - 1)
	if start < 0 {
		start = 0
	}
	return "…" + string(runes[start:])
}
