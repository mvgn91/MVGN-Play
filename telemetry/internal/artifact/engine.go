// Package artifact implements the Session Artifact Engine (ADR-011).
// It reads JSONL events, queries Git, and generates session reports.
package artifact

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mvgn/telemetry/internal/events"
)

// Engine generates session artifacts from JSONL events.
type Engine struct {
	eventsDir  string
	reportsDir string
}

// NewEngine creates a new Artifact Engine.
func NewEngine(eventsDir, reportsDir string) *Engine {
	return &Engine{
		eventsDir:  eventsDir,
		reportsDir: reportsDir,
	}
}

// Generate creates all artifacts for a given session.
func (e *Engine) Generate(sessionID string) error {
	// Load events for this session
	sessionEvents, err := e.loadSessionEvents(sessionID)
	if err != nil {
		return fmt.Errorf("load events: %w", err)
	}

	if len(sessionEvents) == 0 {
		return fmt.Errorf("no events found for session %s", sessionID)
	}

	// Create reports directory
	if err := os.MkdirAll(e.reportsDir, 0755); err != nil {
		return fmt.Errorf("create reports dir: %w", err)
	}

	// Get Git info
	gitInfo := e.getGitInfo()

	// Calculate metrics
	metrics := calculateMetrics(sessionEvents, gitInfo)

	// Generate Markdown report
	if err := e.generateMarkdown(sessionID, sessionEvents, metrics, gitInfo); err != nil {
		return fmt.Errorf("generate markdown: %w", err)
	}

	// Generate JSON report
	if err := e.generateJSON(sessionID, sessionEvents, metrics, gitInfo); err != nil {
		return fmt.Errorf("generate json: %w", err)
	}

	// Generate Metrics JSON
	if err := e.generateMetricsJSON(sessionID, metrics); err != nil {
		return fmt.Errorf("generate metrics json: %w", err)
	}

	return nil
}

// loadSessionEvents reads all JSONL files and filters events for the given session.
func (e *Engine) loadSessionEvents(sessionID string) ([]events.Event, error) {
	var result []events.Event

	// Read all JSONL files
	files, err := filepath.Glob(filepath.Join(e.eventsDir, "*.jsonl"))
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		evts, err := readJSONL(file, sessionID)
		if err != nil {
			continue // Skip corrupted files
		}
		result = append(result, evts...)
	}

	// Sort by timestamp
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp < result[j].Timestamp
	})

	return result, nil
}

// readJSONL reads a JSONL file and returns events matching the session ID.
func readJSONL(path, sessionID string) ([]events.Event, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var result []events.Event
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var evt events.Event
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			continue // Skip malformed lines
		}

		// Filter by session ID if provided
		if sessionID != "" && evt.SessionID != sessionID {
			continue
		}

		result = append(result, evt)
	}

	return result, scanner.Err()
}

// GitInfo contains Git repository information.
type GitInfo struct {
	SHA      string
	Branch   string
	Modified []string
	Added    []string
	Deleted  []string
}

// getGitInfo retrieves Git repository information.
func (e *Engine) getGitInfo() GitInfo {
	info := GitInfo{}

	// Get current SHA
	if out, err := exec.Command("git", "rev-parse", "HEAD").Output(); err == nil {
		info.SHA = strings.TrimSpace(string(out))
		if len(info.SHA) > 7 {
			info.SHA = info.SHA[:7]
		}
	}

	// Get current branch
	if out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
		info.Branch = strings.TrimSpace(string(out))
	}

	// Get changed files — try diff against parent, fallback to git show for first commit
	out, err := exec.Command("git", "diff", "--name-status", "HEAD~1").Output()
	if err != nil {
		// First commit (no parent) — use git show --name-status
		out, err = exec.Command("git", "show", "--name-status", "--format=", "HEAD").Output()
		if err != nil {
			return info
		}
	}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 3 {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) < 2 {
			continue
		}
		status := parts[0]
		file := parts[1]
		switch status {
		case "A":
			info.Added = append(info.Added, file)
		case "D":
			info.Deleted = append(info.Deleted, file)
		case "M":
			info.Modified = append(info.Modified, file)
		}
	}

	return info
}

// SessionMetrics contains calculated metrics for a session.
type SessionMetrics struct {
	Duration         int64   `json:"duration_secs"`
	EventCount       int     `json:"event_count"`
	FilesCreated     int     `json:"files_created"`
	FilesModified    int     `json:"files_modified"`
	FilesDeleted     int     `json:"files_deleted"`
	Commits          int     `json:"commits"`
	GatesApproved    int     `json:"gates_approved"`
	GatesRejected    int     `json:"gates_rejected"`
	GatesTotal       int     `json:"gates_total"`
	Warnings         int     `json:"warnings"`
	Blockers         int     `json:"blockers"`
	HealthScore      float64 `json:"health_score"`
	StartTime        string  `json:"start_time"`
	EndTime          string  `json:"end_time"`
	FirstEvent       string  `json:"first_event"`
	LastEvent        string  `json:"last_event"`
	GitSHA           string  `json:"git_sha,omitempty"`
	GitBranch        string  `json:"git_branch,omitempty"`
	GitFilesModified int     `json:"git_files_modified,omitempty"`
	GitFilesAdded    int     `json:"git_files_added,omitempty"`
	GitFilesDeleted  int     `json:"git_files_deleted,omitempty"`
}

// calculateMetrics calculates metrics from a list of events.
func calculateMetrics(evts []events.Event, git GitInfo) SessionMetrics {
	if len(evts) == 0 {
		return SessionMetrics{}
	}

	m := SessionMetrics{
		EventCount: len(evts),
		FirstEvent: evts[0].Type,
		LastEvent:  evts[len(evts)-1].Type,
		GitSHA:     git.SHA,
		GitBranch:  git.Branch,
	}

	// Parse start/end timestamps for duration
	startTime := evts[0].Timestamp
	endTime := evts[len(evts)-1].Timestamp

	if t, err := time.Parse(time.RFC3339, startTime); err == nil {
		m.StartTime = t.Format("15:04:05")
	}
	if t, err := time.Parse(time.RFC3339, endTime); err == nil {
		m.EndTime = t.Format("15:04:05")
		if start, err := time.Parse(time.RFC3339, startTime); err == nil {
			m.Duration = int64(t.Sub(start).Seconds())
		}
	}

	// Count event types
	for _, evt := range evts {
		switch evt.Type {
		case "FILE_CREATED":
			m.FilesCreated++
		case "FILE_MODIFIED":
			m.FilesModified++
		case "FILE_DELETED":
			m.FilesDeleted++
		case "COMMIT":
			m.Commits++
		case "GATE_APPROVED":
			m.GatesApproved++
		case "GATE_FAILED":
			m.GatesRejected++
		case "WARNING":
			m.Warnings++
		case "TASK_BLOCKED":
			m.Blockers++
		}
	}

	// Add Git stats
	m.GitFilesAdded = len(git.Added)
	m.GitFilesModified = len(git.Modified)
	m.GitFilesDeleted = len(git.Deleted)

	// Calculate health score based on gates
	m.GatesTotal = m.GatesApproved + m.GatesRejected
	if m.GatesTotal > 0 {
		m.HealthScore = float64(m.GatesApproved) / float64(m.GatesTotal) * 100
	}
	// No gates = health score stays 0 (no assessment)

	return m
}

// JSONReport is the typed structure for session JSON reports.
type JSONReport struct {
	SessionID   string              `json:"session_id"`
	Metrics     SessionMetrics      `json:"metrics"`
	Events      []events.Event      `json:"events"`
	Git         JSONGitInfo         `json:"git"`
	Fingerprint JSONFingerprint     `json:"fingerprint"`
}

// JSONGitInfo holds Git file change info for the JSON report.
type JSONGitInfo struct {
	SHA           string   `json:"sha"`
	Branch        string   `json:"branch"`
	FilesAdded    []string `json:"files_added"`
	FilesModified []string `json:"files_modified"`
	FilesDeleted  []string `json:"files_deleted"`
}

// JSONFingerprint holds report identification metadata for reproducibility.
type JSONFingerprint struct {
	TelemetryVersion string `json:"telemetry_version"`
	RuntimeVersion   string `json:"runtime_version,omitempty"`
	MVGNVersion      string `json:"mvgn_version,omitempty"`
	GitSHA           string `json:"git_sha,omitempty"`
	GitBranch        string `json:"git_branch,omitempty"`
	GeneratedAt      string `json:"generated_at"`
}

// generateMarkdown creates the Markdown report.
func (e *Engine) generateMarkdown(sessionID string, evts []events.Event, m SessionMetrics, git GitInfo) error {
	path := filepath.Join(e.reportsDir, sessionID+".md")

	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("# Session Report: %s\n\n", sessionID))
	sb.WriteString("| Campo | Valor |\n")
	sb.WriteString("|-------|-------|\n")
	sb.WriteString(fmt.Sprintf("| **Session ID** | %s |\n", sessionID))
	sb.WriteString(fmt.Sprintf("| **Inicio** | %s |\n", m.StartTime))
	sb.WriteString(fmt.Sprintf("| **Fin** | %s |\n", m.EndTime))
	sb.WriteString(fmt.Sprintf("| **Duración** | %ds |\n", m.Duration))
	sb.WriteString(fmt.Sprintf("| **Eventos** | %d |\n", m.EventCount))
	if git.SHA != "" {
		sb.WriteString(fmt.Sprintf("| **Git SHA** | %s |\n", git.SHA))
		sb.WriteString(fmt.Sprintf("| **Branch** | %s |\n", git.Branch))
	}
	sb.WriteString("\n")

	// Timeline
	sb.WriteString("## Timeline\n\n")
	for _, evt := range evts {
		t, _ := time.Parse(time.RFC3339, evt.Timestamp)
		sb.WriteString(fmt.Sprintf("`%s` %s\n", t.Format("15:04"), evt.Type))
	}
	sb.WriteString("\n")

	// Files changed
	if len(git.Added) > 0 || len(git.Modified) > 0 || len(git.Deleted) > 0 {
		sb.WriteString("## Archivos\n\n")
		if len(git.Added) > 0 {
			sb.WriteString("**Creados:**\n")
			for _, f := range git.Added {
				sb.WriteString(fmt.Sprintf("- %s\n", f))
			}
			sb.WriteString("\n")
		}
		if len(git.Modified) > 0 {
			sb.WriteString("**Modificados:**\n")
			for _, f := range git.Modified {
				sb.WriteString(fmt.Sprintf("- %s\n", f))
			}
			sb.WriteString("\n")
		}
		if len(git.Deleted) > 0 {
			sb.WriteString("**Eliminados:**\n")
			for _, f := range git.Deleted {
				sb.WriteString(fmt.Sprintf("- %s\n", f))
			}
			sb.WriteString("\n")
		}
	}

	// Metrics
	sb.WriteString("## Métricas\n\n")
	sb.WriteString("| Métrica | Valor |\n")
	sb.WriteString("|---------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Gates totales | %d |\n", m.GatesTotal))
	sb.WriteString(fmt.Sprintf("| Gates aprobados | %d |\n", m.GatesApproved))
	sb.WriteString(fmt.Sprintf("| Gates rechazados | %d |\n", m.GatesRejected))
	if m.GatesTotal > 0 {
		sb.WriteString(fmt.Sprintf("| Health Score | %.1f%% |\n", m.HealthScore))
	} else {
		sb.WriteString("| Health Score | N/A (sin gates) |\n")
	}
	sb.WriteString(fmt.Sprintf("| Archivos creados | %d |\n", m.FilesCreated))
	sb.WriteString(fmt.Sprintf("| Archivos modificados | %d |\n", m.FilesModified))
	sb.WriteString(fmt.Sprintf("| Archivos eliminados | %d |\n", m.FilesDeleted))
	sb.WriteString(fmt.Sprintf("| Commits | %d |\n", m.Commits))
	sb.WriteString(fmt.Sprintf("| Warnings | %d |\n", m.Warnings))
	sb.WriteString(fmt.Sprintf("| Bloqueos | %d |\n", m.Blockers))
	sb.WriteString("\n")

	// Fingerprint
	sb.WriteString("## Fingerprint\n\n")
	sb.WriteString("| Campo | Valor |\n")
	sb.WriteString("|-------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Generado | %s |\n", time.Now().UTC().Format(time.RFC3339)))
	sb.WriteString("| Telemetry Version | v2.0.0 |\n")
	sb.WriteString("| Runtime Version | 4.1.0 |\n")
	sb.WriteString("| MVGN Version | 4.1 |\n")
	if git.SHA != "" {
		sb.WriteString(fmt.Sprintf("| Git SHA | %s |\n", git.SHA))
	}
	if git.Branch != "" {
		sb.WriteString(fmt.Sprintf("| Branch | %s |\n", git.Branch))
	}
	sb.WriteString("\n")

	return os.WriteFile(path, []byte(sb.String()), 0644)
}

// generateJSON creates the JSON report.
func (e *Engine) generateJSON(sessionID string, evts []events.Event, m SessionMetrics, git GitInfo) error {
	path := filepath.Join(e.reportsDir, sessionID+".json")

	report := JSONReport{
		SessionID: sessionID,
		Metrics:   m,
		Events:    evts,
		Git: JSONGitInfo{
			SHA:           git.SHA,
			Branch:        git.Branch,
			FilesAdded:    git.Added,
			FilesModified: git.Modified,
			FilesDeleted:  git.Deleted,
		},
		Fingerprint: JSONFingerprint{
			TelemetryVersion: "v2.0.0",
			RuntimeVersion:   "4.1.0",
			MVGNVersion:      "4.1",
			GitSHA:           git.SHA,
			GitBranch:        git.Branch,
			GeneratedAt:      time.Now().UTC().Format(time.RFC3339),
		},
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// generateMetricsJSON creates the metrics-only JSON report.
func (e *Engine) generateMetricsJSON(sessionID string, m SessionMetrics) error {
	path := filepath.Join(e.reportsDir, sessionID+".metrics.json")

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// FindLatestSession finds the most recent session ID from JSONL files.
func FindLatestSession(eventsDir string) string {
	files, err := filepath.Glob(filepath.Join(eventsDir, "*.jsonl"))
	if err != nil || len(files) == 0 {
		return ""
	}

	// Sort by modification time (newest first)
	sort.Slice(files, func(i, j int) bool {
		iInfo, _ := os.Stat(files[i])
		jInfo, _ := os.Stat(files[j])
		if iInfo == nil || jInfo == nil {
			return false
		}
		return iInfo.ModTime().After(jInfo.ModTime())
	})

	// Scan all events in newest file to find latest session
	for _, filePath := range files {
		file, err := os.Open(filePath)
		if err != nil {
			continue
		}

		var lastSession string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			var evt events.Event
			if err := json.Unmarshal([]byte(line), &evt); err != nil {
				continue
			}

			if evt.SessionID != "" {
				lastSession = evt.SessionID
			}
		}
		file.Close()

		if lastSession != "" {
			return lastSession
		}
	}

	return ""
}

// ListSessions lists all unique session IDs from JSONL files.
func ListSessions(eventsDir string) []string {
	sessions := make(map[string]bool)

	files, err := filepath.Glob(filepath.Join(eventsDir, "*.jsonl"))
	if err != nil {
		return nil
	}

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			var evt events.Event
			if err := json.Unmarshal([]byte(line), &evt); err != nil {
				continue
			}

			if evt.SessionID != "" {
				sessions[evt.SessionID] = true
			}
		}
		f.Close()
	}

	// Convert to sorted slice
	var result []string
	for s := range sessions {
		result = append(result, s)
	}
	sort.Strings(result)

	return result
}
