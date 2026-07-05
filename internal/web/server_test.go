package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleScan(t *testing.T) {
	server := NewServer(":0")
	
	reqBody := ScanRequest{
		Directories: []string{"."},
		Workers:     1,
		MinSize:     1,
		Algorithm:   "md5",
	}
	buf, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/scan", bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleScan(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", res.Status)
	}

	var resp ScanResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
}

func TestHandleScanMethodNotAllowed(t *testing.T) {
	server := NewServer(":0")
	req := httptest.NewRequest(http.MethodGet, "/api/scan", nil)
	w := httptest.NewRecorder()

	server.handleScan(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status MethodNotAllowed, got %v", res.Status)
	}
}
