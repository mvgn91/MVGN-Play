package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"sync"

	"github.com/mvgn/mvgn-scan/internal/display"
)

//go:embed web/*
var webFS embed.FS

type TelemetryEvent struct {
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Source    string `json:"source"`
	Data      string `json:"data"`
}

type Update struct {
	Version         string           `json:"version"`
	Status          string           `json:"status"`
	Health          int              `json:"health"`
	Phase           int              `json:"phase"`
	Project         string           `json:"project"`
	TaskID          string           `json:"task_id"`
	Description     string           `json:"description"`
	Agents          []string         `json:"agents"`
	Artifacts       int              `json:"artifacts"`
	NextAction      string           `json:"next_action"`
	Timestamp       string           `json:"timestamp"`
	AteneaScore     int              `json:"atenea_score"`
	Interval        int              `json:"interval"`
	TelemetryEvents []TelemetryEvent `json:"telemetry_events"`
}

type hub struct {
	mu      sync.RWMutex
	clients map[chan []byte]struct{}
}

func newHub() *hub {
	return &hub{clients: make(map[chan []byte]struct{})}
}

func (h *hub) subscribe() chan []byte {
	ch := make(chan []byte, 4)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *hub) unsubscribe(ch chan []byte) {
	h.mu.Lock()
	delete(h.clients, ch)
	close(ch)
	h.mu.Unlock()
}

func (h *hub) broadcast(data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients {
		select {
		case ch <- data:
		default:
		}
	}
}

type Server struct {
	hub      *hub
	port     int
	updates  chan *Update
	interval int
	project  string
}

func New(port, interval int, project string) *Server {
	return &Server{
		hub:      newHub(),
		port:     port,
		updates:  make(chan *Update, 8),
		interval: interval,
		project:  project,
	}
}

func (s *Server) Send(upd *Update) {
	s.updates <- upd
}

func (s *Server) Start() error {
	subFS, err := fs.Sub(webFS, "web")
	if err != nil {
		return fmt.Errorf("web fs: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(subFS)))
	mux.HandleFunc("/api/events", s.sseHandler)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("listen :%d: %w", s.port, err)
	}

	go s.broadcastLoop()

	url := fmt.Sprintf("http://localhost:%d", s.port)
	fmt.Printf("\n  %s✨ MVGN Observer running at:%s %s\n\n", display.FG_GREEN, display.ANSI_RESET, url)
	openBrowser(url)

	return http.Serve(listener, mux)
}

func (s *Server) StartBackground() error {
	subFS, err := fs.Sub(webFS, "web")
	if err != nil {
		return fmt.Errorf("web fs: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(subFS)))
	mux.HandleFunc("/api/events", s.sseHandler)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("listen :%d: %w", s.port, err)
	}

	go s.broadcastLoop()
	go http.Serve(listener, mux)

	return nil
}

func (s *Server) broadcastLoop() {
	for upd := range s.updates {
		data, err := json.Marshal(upd)
		if err != nil {
			log.Printf("json marshal: %v", err)
			continue
		}
		s.hub.broadcast(data)
	}
}

func (s *Server) sseHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := s.hub.subscribe()
	defer s.hub.unsubscribe(ch)

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

func openBrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}

	if err := exec.Command(cmd, args...).Start(); err != nil {
		log.Printf("auto-open: %v (open %s manually)", err, url)
	}
}
