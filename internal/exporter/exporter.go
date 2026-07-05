package exporter

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"os"

	"github.com/arthurgray2k/goDupFinder/pkg/dupfinder"
)

// Exporter defines how duplicates are exported.
type Exporter interface {
	Export(groups []dupfinder.DuplicateGroup) error
	Close() error
}

// JSONExporter exports to standard JSON.
type JSONExporter struct {
	w io.Writer
}

func NewJSONExporter(w io.Writer) *JSONExporter {
	return &JSONExporter{w: w}
}

func (e *JSONExporter) Export(groups []dupfinder.DuplicateGroup) error {
	encoder := json.NewEncoder(e.w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(groups)
}

func (e *JSONExporter) Close() error {
	if c, ok := e.w.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// NDJSONExporter exports to Newline Delimited JSON.
type NDJSONExporter struct {
	w io.Writer
}

func NewNDJSONExporter(w io.Writer) *NDJSONExporter {
	return &NDJSONExporter{w: w}
}

func (e *NDJSONExporter) Export(groups []dupfinder.DuplicateGroup) error {
	encoder := json.NewEncoder(e.w)
	for _, g := range groups {
		if err := encoder.Encode(g); err != nil {
			return err
		}
	}
	return nil
}

func (e *NDJSONExporter) Close() error {
	if c, ok := e.w.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// CSVExporter exports to CSV.
type CSVExporter struct {
	w io.Writer
}

func NewCSVExporter(w io.Writer) *CSVExporter {
	return &CSVExporter{w: w}
}

func (e *CSVExporter) Export(groups []dupfinder.DuplicateGroup) error {
	w := csv.NewWriter(e.w)
	defer w.Flush()

	// Header
	if err := w.Write([]string{"GroupHash", "FilePath"}); err != nil {
		return err
	}

	for _, g := range groups {
		for _, file := range g.Files {
			if err := w.Write([]string{g.Hash, file}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *CSVExporter) Close() error {
	if c, ok := e.w.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// Factory to create exporters
func CreateExporter(format string, out string) (Exporter, error) {
	var w io.Writer
	if out == "" || out == "-" {
		w = os.Stdout
	} else {
		f, err := os.Create(out)
		if err != nil {
			return nil, err
		}
		w = f
	}

	switch format {
	case "json":
		return NewJSONExporter(w), nil
	case "ndjson":
		return NewNDJSONExporter(w), nil
	case "csv":
		return NewCSVExporter(w), nil
	default:
		// Fallback to JSON if unknown
		return NewJSONExporter(w), nil
	}
}
