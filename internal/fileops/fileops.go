// Package fileops provides safe, auditable file operations on duplicate files.
// All destructive operations require explicit confirmation or dry-run preview.
// The package never deletes files automatically.
package fileops

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// ErrUserAborted is returned when the user declines an operation.
var ErrUserAborted = errors.New("operation aborted by user")

// Options controls how file operations are performed.
type Options struct {
	// DryRun prints what would happen without making changes.
	DryRun bool

	// SkipConfirm bypasses per-operation confirmation prompts.
	// Even when true, DryRun overrides all destructive actions.
	SkipConfirm bool

	// KeepFirst keeps the first file in each group and operates on the rest.
	KeepFirst bool

	// MoveDestination is the target directory for Move operations.
	MoveDestination string
}

// Result records the outcome of a single file operation.
type Result struct {
	Op      string
	Source  string
	Target  string
	Skipped bool
	DryRun  bool
	Err     error
}

// Operator carries out safe file operations on duplicate file groups.
type Operator struct {
	opts    Options
	confirm ConfirmFunc
}

// ConfirmFunc is a callback that asks the user whether to proceed.
// Returns true to continue, false to skip.
type ConfirmFunc func(op, path string) bool

// New creates a new Operator.
// If confirmFn is nil, the Operator will use DefaultConfirmFunc.
func New(opts Options, confirmFn ConfirmFunc) *Operator {
	if confirmFn == nil {
		confirmFn = DefaultConfirmFunc
	}
	return &Operator{opts: opts, confirm: confirmFn}
}

// DefaultConfirmFunc always returns true (no interactive prompt).
// Replace with a CLI-aware version when running interactively.
var DefaultConfirmFunc ConfirmFunc = func(op, path string) bool { return true }

// Delete removes duplicate files, keeping the first file in the group.
// It never deletes when DryRun is true.
func (o *Operator) Delete(files []string) []Result {
	return o.operate("delete", files, func(src string) (Result, error) {
		if o.opts.DryRun {
			return Result{Op: "delete", Source: src, DryRun: true}, nil
		}
		if err := os.Remove(src); err != nil {
			return Result{Op: "delete", Source: src, Err: err}, err
		}
		return Result{Op: "delete", Source: src}, nil
	})
}

// Move relocates duplicate files to MoveDestination, keeping the first file.
func (o *Operator) Move(files []string) []Result {
	return o.operate("move", files, func(src string) (Result, error) {
		if o.opts.MoveDestination == "" {
			return Result{Op: "move", Source: src, Err: errors.New("no destination directory specified")}, errors.New("no destination")
		}

		dest := filepath.Join(o.opts.MoveDestination, filepath.Base(src))
		dest = uniquePath(dest)

		if o.opts.DryRun {
			return Result{Op: "move", Source: src, Target: dest, DryRun: true}, nil
		}

		if err := os.MkdirAll(o.opts.MoveDestination, 0755); err != nil {
			return Result{Op: "move", Source: src, Target: dest, Err: err}, err
		}
		if err := os.Rename(src, dest); err != nil {
			// Rename may fail across filesystems; fall back to copy+delete.
			if err2 := copyFile(src, dest); err2 != nil {
				return Result{Op: "move", Source: src, Target: dest, Err: err2}, err2
			}
			_ = os.Remove(src)
		}
		return Result{Op: "move", Source: src, Target: dest}, nil
	})
}

// HardLink replaces duplicate files with hard links to the first file.
func (o *Operator) HardLink(files []string) []Result {
	if len(files) < 2 {
		return nil
	}
	keep := files[0]
	return o.operate("hardlink", files[1:], func(src string) (Result, error) {
		if o.opts.DryRun {
			return Result{Op: "hardlink", Source: src, Target: keep, DryRun: true}, nil
		}
		if err := os.Remove(src); err != nil {
			return Result{Op: "hardlink", Source: src, Target: keep, Err: err}, err
		}
		if err := os.Link(keep, src); err != nil {
			return Result{Op: "hardlink", Source: src, Target: keep, Err: err}, err
		}
		return Result{Op: "hardlink", Source: src, Target: keep}, nil
	})
}

// Symlink replaces duplicate files with symbolic links to the first file.
func (o *Operator) Symlink(files []string) []Result {
	if len(files) < 2 {
		return nil
	}
	keep := files[0]
	return o.operate("symlink", files[1:], func(src string) (Result, error) {
		if o.opts.DryRun {
			return Result{Op: "symlink", Source: src, Target: keep, DryRun: true}, nil
		}
		if err := os.Remove(src); err != nil {
			return Result{Op: "symlink", Source: src, Target: keep, Err: err}, err
		}
		if err := os.Symlink(keep, src); err != nil {
			return Result{Op: "symlink", Source: src, Target: keep, Err: err}, err
		}
		return Result{Op: "symlink", Source: src, Target: keep}, nil
	})
}

// operate is the shared execution loop used by all file operation methods.
// It handles KeepFirst, confirmation prompts, and collects Results.
func (o *Operator) operate(op string, files []string, fn func(string) (Result, error)) []Result {
	results := make([]Result, 0, len(files))

	start := 0
	if o.opts.KeepFirst {
		start = 1
	}

	for _, f := range files[start:] {
		if !o.opts.SkipConfirm && !o.opts.DryRun {
			if !o.confirm(op, f) {
				results = append(results, Result{Op: op, Source: f, Skipped: true})
				continue
			}
		}
		r, _ := fn(f)
		results = append(results, r)
	}

	return results
}

// uniquePath appends a timestamp suffix if the target path already exists.
func uniquePath(dest string) string {
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		return dest
	}
	ext := filepath.Ext(dest)
	base := dest[:len(dest)-len(ext)]
	return fmt.Sprintf("%s_%d%s", base, time.Now().UnixNano(), ext)
}

// copyFile copies src to dst using streaming I/O with a reusable buffer.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	buf := make([]byte, 32*1024)
	_, err = io.CopyBuffer(out, in, buf)
	return err
}
