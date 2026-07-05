# goDupFinder

A fast, scalable, and memory-efficient duplicate file finder written in Go.

## Features

- 🚀 **Fast** — multi-stage pipeline: size grouping → hashing → optional byte-for-byte verification
- 📦 **Reusable library** — clean `pkg/dupfinder` public API usable by other Go applications
- 🔎 **Rich filters** — min/max size, extension include/exclude, hidden files, system dirs
- 🔑 **Multiple algorithms** — SHA256 (default), SHA1, MD5, BLAKE2
- 📤 **Multiple output formats** — JSON, NDJSON, CSV
- 🗑️ **Safe file operations** — delete, move, symlink, hardlink with dry-run and confirmation prompts
- 📊 **Live progress** — files/sec, throughput, ETA — rendered to stderr
- 📈 **Statistics** — wasted space, duplicate groups, files scanned, elapsed time
- 💾 **Optional DuckDB storage** — persistent hash cache for incremental scans (compile with `-tags duckdb`)
- ♻️ **Concurrent** — worker pools with context cancellation throughout

## Installation

```bash
go install github.com/arthurgray2k/goDupFinder/cmd/goDupFinder@latest
```

Or build from source:

```bash
git clone https://github.com/arthurgray2k/goDupFinder
cd goDupFinder
go build ./cmd/goDupFinder
```

## Quick Start

```bash
# Find duplicates in a directory
goDupFinder scan D:\Photos

# Scan multiple directories
goDupFinder scan D:\Photos D:\Backup

# Show statistics
goDupFinder stats D:\Photos D:\Backup

# Export results to JSON
goDupFinder export --format json --output duplicates.json D:\Photos

# Preview what would be deleted (never deletes automatically)
goDupFinder delete --dry-run D:\Photos

# Move duplicates to a quarantine folder
goDupFinder move --dest D:\Duplicates D:\Photos

# Replace duplicates with symlinks (saves disk space)
goDupFinder symlink --dry-run D:\Photos
```

## Project Structure

```text
goDupFinder/
├── cmd/goDupFinder/       # CLI entry point (Cobra commands)
│   ├── main.go
│   ├── scan.go
│   ├── export.go
│   ├── ops.go             # delete, move, symlink, hardlink
│   └── stats.go
├── internal/
│   ├── fileops/           # Safe file operations engine
│   ├── hasher/            # Streaming SHA256/SHA1/MD5/BLAKE2
│   ├── matcher/           # Size-first duplicate detection
│   ├── exporter/          # JSON / NDJSON / CSV exporters
│   ├── pipeline/          # Shared pipeline types
│   ├── progress/          # Live terminal progress + statistics
│   ├── storage/           # Storage interface + DuckDB backend (optional)
│   └── walker/            # Concurrent directory traversal
├── pkg/dupfinder/         # Public library API
├── .specs/                # Architecture & design specs
├── go.mod
├── README.md
└── USAGE.md
```

## Library Usage

```go
import "github.com/arthurgray2k/goDupFinder/pkg/dupfinder"

opts := dupfinder.Options{
    Workers:          8,
    Algorithm:        dupfinder.SHA256,
    VerifyContents:   false,
    MinSize:          1,
    IncludeHidden:    false,
    IgnoreSystemDirs: true,
}

finder := dupfinder.New(opts)
groups, err := finder.Scan(ctx, []string{"/photos", "/backup"})
```

## Build

```bash
go fmt ./...
go vet ./...
go test ./...
go build ./...
```

With DuckDB persistent cache (requires CGO / C++ compiler):

```bash
go build -tags duckdb ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Update `.specs/` before implementing significant changes
4. Ensure `go test ./...` passes with >80% coverage
5. Submit a pull request

## License

MIT
