package exporter

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/arthurgray2k/goDupFinder/pkg/dupfinder"
)

func TestJSONExporter(t *testing.T) {
	var buf bytes.Buffer
	exp := &JSONExporter{w: &buf}
	groups := []dupfinder.DuplicateGroup{
		{Hash: "abc", Files: []string{"file1", "file2"}},
	}
	err := exp.Export(groups)
	if err != nil {
		t.Fatal(err)
	}

	var parsed []dupfinder.DuplicateGroup
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatal(err)
	}
	if len(parsed) != 1 || parsed[0].Hash != "abc" {
		t.Fatal("Invalid JSON output")
	}
}

func TestNDJSONExporter(t *testing.T) {
	var buf bytes.Buffer
	exp := &NDJSONExporter{w: &buf}
	groups := []dupfinder.DuplicateGroup{
		{Hash: "abc", Files: []string{"file1", "file2"}},
	}
	err := exp.Export(groups)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(buf.Bytes(), []byte("abc")) {
		t.Fatal("Invalid NDJSON output")
	}
}

func TestCSVExporter(t *testing.T) {
	var buf bytes.Buffer
	exp := &CSVExporter{w: &buf}
	groups := []dupfinder.DuplicateGroup{
		{Hash: "abc", Files: []string{"file1", "file2"}},
	}
	err := exp.Export(groups)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(buf.Bytes(), []byte("file1")) {
		t.Fatal("Invalid CSV output")
	}
}

func TestCreateExporter(t *testing.T) {
	_, err := CreateExporter("json", "")
	if err != nil {
		t.Fatal(err)
	}

	tempFile := t.TempDir() + "/out.json"
	exp, err := CreateExporter("ndjson", tempFile)
	if err != nil {
		t.Fatal(err)
	}
	exp.Close()
	defer os.Remove(tempFile)

	// Fallback to JSON is expected for unknown formats
	exp, err = CreateExporter("invalid", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := exp.(*JSONExporter); !ok {
		t.Fatal("Expected JSONExporter fallback")
	}
}
