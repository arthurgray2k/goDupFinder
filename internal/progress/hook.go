package progress

import "github.com/arthurgray2k/goDupFinder/internal/pipeline"

// StatsHook adapts *Stats so it satisfies the pipeline.Hook interface.
// The CLI creates a StatsHook and passes it into the scan pipeline so that
// Stats counters are updated atomically from the walker and matcher goroutines.
type StatsHook struct {
	stats   *Stats
	tracker *Tracker
}

// NewStatsHook creates a StatsHook wrapping the given Stats and Tracker.
// tracker may be nil if no live display is required.
func NewStatsHook(s *Stats, t *Tracker) *StatsHook {
	return &StatsHook{stats: s, tracker: t}
}

// OnFileFound is called by the walker for every file that passes filters.
func (h *StatsHook) OnFileFound(path string, size int64) {
	h.stats.FilesScanned.Add(1)
	if h.tracker != nil {
		h.tracker.SetCurrentFile(path)
	}
}

// OnFileHashed is called by the matcher after each file is fully hashed.
func (h *StatsHook) OnFileHashed(_ string, bytes int64) {
	h.stats.BytesProcessed.Add(bytes)
}

// OnDirScanned is called by the walker when it enters a directory.
func (h *StatsHook) OnDirScanned(dir string) {
	h.stats.DirsScanned.Add(1)
	if h.tracker != nil {
		h.tracker.SetCurrentDir(dir)
	}
}

// OnFileSkipped is called by the walker for every file excluded by filters.
func (h *StatsHook) OnFileSkipped() {
	h.stats.FilesSkipped.Add(1)
}

// Ensure StatsHook satisfies the interface at compile time.
var _ pipeline.Hook = (*StatsHook)(nil)
