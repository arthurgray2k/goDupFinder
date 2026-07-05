package pipeline

// Hook is a set of callbacks invoked during a scan to report live progress.
// Implementations must be safe for concurrent use by multiple goroutines.
type Hook interface {
	// OnFileFound is called when a file is discovered by the walker.
	OnFileFound(path string, size int64)

	// OnFileHashed is called after a file's hash has been computed.
	OnFileHashed(path string, bytes int64)

	// OnDirScanned is called when the walker enters a directory.
	OnDirScanned(dir string)

	// OnFileSkipped is called when a file is excluded by filters.
	OnFileSkipped()
}

// NoopHook is a Hook that does nothing. Use it when no progress tracking is needed.
type NoopHook struct{}

func (NoopHook) OnFileFound(string, int64)  {}
func (NoopHook) OnFileHashed(string, int64) {}
func (NoopHook) OnDirScanned(string)        {}
func (NoopHook) OnFileSkipped()             {}
