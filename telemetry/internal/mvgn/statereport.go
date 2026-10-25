package mvgn

import (
	"bufio"
	"os"
	"strings"
)

// StateReport holds parsed header fields from docs/06_state_report.md.
type StateReport struct {
	State      string
	Progress   string
	LastAction string
	NextAction string
	Blockers   string
}

// ReadStateReport parses the header table of docs/06_state_report.md.
// Returns defaults on missing or unparseable file.
func (r *ContextReader) ReadStateReport() (*StateReport, error) {
	if cached, ok := r.getCache("state_report"); ok {
		return cached.(*StateReport), nil
	}

	f, err := os.Open(r.projectPath + "/docs/06_state_report.md")
	if err != nil {
		return &StateReport{State: "unknown"}, nil
	}
	defer f.Close()

	sr := &StateReport{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case strings.HasPrefix(line, "| **Estado"):
			sr.State = extractValue(line)
		case strings.HasPrefix(line, "| **Progreso"):
			sr.Progress = extractValue(line)
		case strings.HasPrefix(line, "| **Última acción"):
			sr.LastAction = extractValue(line)
		case strings.HasPrefix(line, "| **Siguiente acción"):
			sr.NextAction = extractValue(line)
		case strings.HasPrefix(line, "| **Bloqueos activos"):
			sr.Blockers = extractValue(line)
		}
	}

	r.setCache("state_report", sr)
	return sr, nil
}

// extractValue pulls the third cell from a markdown table row.
// Row format: | **Field** | Value |
func extractValue(line string) string {
	parts := strings.Split(line, "|")
	if len(parts) >= 3 {
		return strings.TrimSpace(parts[2])
	}
	return ""
}