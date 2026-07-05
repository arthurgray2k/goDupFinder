# goDupFinder Usage Guide

## CLI Reference

### scan

Find and display duplicate files.

```bash
goDupFinder scan [directories...] [flags]

Flags:
  --workers int     Number of concurrent hash workers (default 8)
  --algo string     Hash algorithm: sha256, sha1, md5, blake2 (default "sha256")
  --min-size string Minimum file size to consider (default "1")

Examples:
  goDupFinder scan D:\Photos
  goDupFinder scan D:\Photos D:\Backup --workers 16
  goDupFinder scan D:\Photos --algo blake2
```

### export

Scan and export results to a file.

```bash
goDupFinder export [directories...] [flags]

Flags:
  --format string   Output format: json, ndjson, csv (default "json")
  --output string   Output file path (default stdout)

Examples:
  goDupFinder export --format json --output dupes.json D:\Photos
  goDupFinder export --format csv --output dupes.csv D:\Photos D:\Backup
  goDupFinder export --format ndjson D:\Photos
```

### stats

Scan and print detailed statistics.

```bash
goDupFinder stats [directories...]

Examples:
  goDupFinder stats D:\Photos D:\Backup

Output:
  ─────────────────────────────────
   Scan Complete
  ─────────────────────────────────
   Files scanned:     128,456
   Dirs scanned:      3,241
   Files skipped:     12
   Bytes processed:   54.32 GB
   Duplicate groups:  892
   Duplicate files:   1,784
   Wasted space:      12.40 GB
   Elapsed:           2m14s
   Files/sec:         957
   Throughput:        412.00 MB/s
  ─────────────────────────────────
```

### delete

Delete duplicate files (keeps first in each group).

```bash
goDupFinder delete [directories...] [flags]

Flags:
  --dry-run         Preview what would be deleted (no changes made)
  --skip-confirm    Skip per-file confirmation prompts

Examples:
  # Preview (safe — no files are deleted)
  goDupFinder delete --dry-run D:\Photos

  # Delete with confirmation for each file
  goDupFinder delete D:\Photos

  # Delete all without prompting
  goDupFinder delete --skip-confirm D:\Photos
```

### move

Move duplicate files to a destination directory.

```bash
goDupFinder move [directories...] --dest <destination> [flags]

Flags:
  --dest string     Destination directory (required)
  --dry-run         Preview what would be moved
  --skip-confirm    Skip per-file confirmation prompts

Examples:
  goDupFinder move --dry-run --dest D:\Quarantine D:\Photos
  goDupFinder move --dest D:\Quarantine D:\Photos D:\Backup
```

### symlink

Replace duplicate files with symbolic links.

```bash
goDupFinder symlink [directories...] [flags]

Flags:
  --dry-run         Preview what would be symlinked
  --skip-confirm    Skip per-file confirmation prompts

Examples:
  goDupFinder symlink --dry-run D:\Photos
  goDupFinder symlink --skip-confirm D:\Photos D:\Backup
```

### hardlink

Replace duplicate files with hard links (same filesystem only).

```bash
goDupFinder hardlink [directories...] [flags]

Flags:
  --dry-run         Preview what would be hard-linked
  --skip-confirm    Skip per-file confirmation prompts

Examples:
  goDupFinder hardlink --dry-run D:\Photos
  goDupFinder hardlink D:\Photos
```

---

## Go Library

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/arthurgray2k/goDupFinder/pkg/dupfinder"
)

func main() {
    opts := dupfinder.Options{
        Workers:          8,
        Algorithm:        dupfinder.SHA256,
        VerifyContents:   false,
        MinSize:          1,
        MaxSize:          0, // no limit
        IncludeHidden:    false,
        IgnoreSystemDirs: true,
    }

    finder := dupfinder.New(opts)

    groups, err := finder.Scan(context.Background(), []string{
        "/photos",
        "/backup",
    })
    if err != nil {
        log.Fatal(err)
    }

    for _, g := range groups {
        fmt.Printf("Hash: %s\n", g.Hash)
        for _, f := range g.Files {
            fmt.Printf("  %s\n", f)
        }
    }
}
```

---

## Configuration Defaults

| Option           | Default  | Description                        |
|------------------|----------|------------------------------------|
| Workers          | 8        | Concurrent hash goroutines         |
| Algorithm        | sha256   | Hash algorithm                     |
| VerifyContents   | false    | Byte-for-byte verification         |
| MinSize          | 1 byte   | Ignore empty files by default      |
| MaxSize          | 0        | No upper size limit                |
| IncludeHidden    | false    | Skip dot-files by default          |
| IgnoreSystemDirs | true     | Skip node_modules, .git, vendor    |

---

## DuckDB Persistent Cache

When built with the `duckdb` build tag, goDupFinder caches computed file hashes in a local DuckDB database. On subsequent scans of the same directory, unchanged files are retrieved from the cache instead of being re-hashed — dramatically reducing scan time for large datasets.

```bash
go build -tags duckdb ./cmd/goDupFinder
```

The cache is automatically invalidated when a file's size or modification time changes.
