package mvgn

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// EventEntry represents a single event parsed from a .jsonl events file.
// Data contains a human-readable description (not raw JSON).
type EventEntry struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Source    string `json:"source"`
	Data      string `json:"data"`       // human-readable description
	SessionID string `json:"session_id"`
}

// Describe returns a human-readable summary of the event.
// Falls back to the event Type if Data is empty.
func (e EventEntry) Describe() string {
	if e.Data != "" {
		return e.Data
	}
	return e.Type
}

// internal types for JSONL parsing (matches templates/schemas/event.schema.json)
type rawEvent struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Timestamp string            `json:"timestamp"`
	Actor     rawActor          `json:"actor"`
	SessionID string            `json:"session_id"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	State     *rawState         `json:"state,omitempty"`
}

type rawActor struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

type rawState struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// describeEvent generates a human-readable one-line summary from a parsed event.
func describeEvent(raw rawEvent) string {
	switch raw.Type {
	case "SESSION_STARTED":
		proj := raw.Metadata["project"]
		model := raw.Metadata["model"]
		if model == "" || model == "unknown" {
			model = "—"
		}
		return fmt.Sprintf("Session started  ·  project: %s  ·  model: %s", proj, model)

	case "SESSION_CLOSED":
		dur := raw.Metadata["duration_secs"]
		return fmt.Sprintf("Session closed  ·  duration: %ss", dur)

	case "HEARTBEAT":
		return "♥ Heartbeat — system alive"

	case "HEALTH_CHECK":
		score := raw.Metadata["score"]
		docs := raw.Metadata["doc_coverage"]
		return fmt.Sprintf("Health check  ·  score: %s  ·  docs: %s", score, docs)

	case "STATE_CHANGED":
		if raw.State != nil {
			return fmt.Sprintf("State changed: %s → %s", raw.State.From, raw.State.To)
		}
		return "State changed"

	case "SKILL_LOADED":
		plugin := raw.Metadata["plugin"]
		return fmt.Sprintf("Plugin loaded: %s", plugin)

	case "SKILL_UNLOADED":
		plugin := raw.Metadata["plugin"]
		return fmt.Sprintf("Plugin unloaded: %s", plugin)

	default:
		// For unknown event types, show type + any metadata we can find
		if len(raw.Metadata) > 0 {
			var parts []string
			for k, v := range raw.Metadata {
				parts = append(parts, fmt.Sprintf("%s: %s", k, v))
			}
			return fmt.Sprintf("%s  ·  %s", raw.Type, strings.Join(parts, "  ·  "))
		}
		return fmt.Sprintf("%s event", raw.Type)
	}
}

// EventHistoryReader reads MVGN project events from events/*.jsonl files
// with TTL-based caching (default 5s). All methods degrade gracefully on
// missing directories or files.
type EventHistoryReader struct {
	projectPath string
	cache       map[string]cacheEntry
	ttl         time.Duration
}

// NewEventReader creates an EventHistoryReader rooted at projectPath.
func NewEventReader(projectPath string) *EventHistoryReader {
	return &EventHistoryReader{
		projectPath: projectPath,
		cache:       make(map[string]cacheEntry),
		ttl:         5 * time.Second,
	}
}

// ReadRecentEvents returns up to maxEvents entries from the project's
// events/*.jsonl files, reading from newest files first. Each entry
// includes parsed timestamp, type, source, and a human-readable Data field.
func (r *EventHistoryReader) ReadRecentEvents(maxEvents int) ([]EventEntry, error) {
	eventsDir := r.projectPath + "/events"
	files, err := filepath.Glob(eventsDir + "/*.jsonl")
	if err != nil {
		return nil, nil
	}

	if len(files) == 0 {
		return nil, nil
	}

	// Sort files descending by name (newest first, for daily rotation).
	sort.Sort(sort.Reverse(sort.StringSlice(files)))

	var events []EventEntry
	for _, f := range files {
		file, err := os.Open(f)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			var raw rawEvent
			if err := json.Unmarshal([]byte(line), &raw); err != nil {
				// If we can't parse the JSON, fall back gracefully
				events = append(events, EventEntry{
					Timestamp: extractFallback(line, "timestamp"),
					Type:      extractFallback(line, "type"),
					Source:    extractFallback(line, "source"),
					Data:      line, // show raw line only as last resort
				})
				continue
			}

			actorName := raw.Actor.Name
			if actorName == "" {
				actorName = raw.Actor.Type
			}

			events = append(events, EventEntry{
				ID:        raw.ID,
				Timestamp: raw.Timestamp,
				Type:      raw.Type,
				Source:    actorName,
				Data:      describeEvent(raw),
				SessionID: raw.SessionID,
			})
		}
		file.Close()

		if len(events) >= maxEvents {
			events = events[:maxEvents]
			break
		}
	}

	return events, nil
}

// extractFallback pulls a string field from a JSONL line using simple string search.
// Only used as a fallback when json.Unmarshal fails.
func extractFallback(line, field string) string {
	prefix := fmt.Sprintf(`"%s":"`, field)
	start := strings.Index(line, prefix)
	if start == -1 {
		return "unknown"
	}
	start += len(prefix)
	end := strings.Index(line[start:], `"`)
	if end == -1 {
		return "unknown"
	}
	return line[start : start+end]
}
