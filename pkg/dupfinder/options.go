package dupfinder

import "errors"

// Algorithm defines the hashing algorithm to use.
type Algorithm string

const (
	SHA256 Algorithm = "sha256"
	SHA1   Algorithm = "sha1"
	MD5    Algorithm = "md5"
	BLAKE2 Algorithm = "blake2"
)

// Options configures the duplicate finder.
type Options struct {
	// Concurrency
	Workers int

	// Hashing
	Algorithm      Algorithm
	VerifyContents bool

	// Filters
	MinSize          int64
	MaxSize          int64
	MaxDepth         int
	IncludeExts      []string
	ExcludeExts      []string
	IncludeHidden    bool
	IgnoreSystemDirs bool

	// Hook receives live progress events during a scan.
	// Leave nil to disable progress tracking (a NoopHook is used internally).
	Hook ProgressHook
}

// DefaultOptions returns a sensible set of default options.
func DefaultOptions() Options {
	return Options{
		Workers:          8,
		Algorithm:        SHA256,
		VerifyContents:   false,
		MinSize:          1, // default ignore empty files
		MaxSize:          0, // 0 means unlimited
		IncludeHidden:    false,
		IgnoreSystemDirs: true,
	}
}

// Validate checks the options for validity.
func (o Options) Validate() error {
	if o.Workers <= 0 {
		return errors.New("workers must be at least 1")
	}
	if o.MinSize < 0 {
		return errors.New("min size cannot be negative")
	}
	if o.MaxSize > 0 && o.MinSize > o.MaxSize {
		return errors.New("min size cannot be greater than max size")
	}
	switch o.Algorithm {
	case SHA256, SHA1, MD5, BLAKE2:
		// valid
	default:
		return errors.New("unsupported algorithm")
	}
	return nil
}
