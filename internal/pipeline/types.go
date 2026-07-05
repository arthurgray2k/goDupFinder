package pipeline

// FileItem represents a discovered file in the scanning pipeline.
type FileItem struct {
	Path string
	Size int64
}

// Group represents a slice of FileItems that share common traits (like size or partial hash).
type Group struct {
	Items []FileItem
}

// DuplicateGroup represents a group of identical files internal to the pipeline.
type DuplicateGroup struct {
	Hash  string
	Files []string
}
