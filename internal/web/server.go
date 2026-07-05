package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/arthurgray2k/goDupFinder/pkg/dupfinder"
)

//go:embed static/*
var staticFS embed.FS

type Server struct {
	addr string
}

func NewServer(addr string) *Server {
	return &Server{addr: addr}
}

type ScanRequest struct {
	Directories []string `json:"directories"`
	MinSize     int64    `json:"min_size"`
	MaxDepth    int      `json:"max_depth"`
	Workers     int      `json:"workers"`
	Algorithm   string   `json:"algorithm"`
}

type ScanResponse struct {
	Error      string                     `json:"error,omitempty"`
	Duplicates []dupfinder.DuplicateGroup `json:"duplicates,omitempty"`
	Elapsed    string                     `json:"elapsed,omitempty"`
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Serve static files
	staticDir, err := fs.Sub(staticFS, "static")
	if err != nil {
		return err
	}
	mux.Handle("/", http.FileServer(http.FS(staticDir)))

	// API endpoints
	mux.HandleFunc("/api/scan", s.handleScan)

	fmt.Printf("Web dashboard running at http://%s/\n", s.addr)
	return http.ListenAndServe(s.addr, mux)
}

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	opts := dupfinder.DefaultOptions()
	if req.Workers > 0 {
		opts.Workers = req.Workers
	}
	opts.MinSize = req.MinSize
	opts.MaxDepth = req.MaxDepth
	switch req.Algorithm {
	case "md5":
		opts.Algorithm = dupfinder.MD5
	case "sha1":
		opts.Algorithm = dupfinder.SHA1
	case "blake2":
		opts.Algorithm = dupfinder.BLAKE2
	default:
		opts.Algorithm = dupfinder.SHA256
	}

	finder := dupfinder.New(opts)
	
	start := time.Now()
	// No progress hook attached to the web request for now, just wait for completion.
	// We use the request context so if the user closes the browser, the scan cancels.
	duplicates, err := finder.Scan(r.Context(), req.Directories)
	elapsed := time.Since(start).String()

	resp := ScanResponse{Elapsed: elapsed}
	if err != nil {
		if err == context.Canceled {
			resp.Error = "Scan canceled"
		} else {
			resp.Error = err.Error()
		}
	} else {
		resp.Duplicates = duplicates
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
