package observer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mvgn/mvgn-scan/internal/display"
	"github.com/mvgn/mvgn-scan/internal/models"
)

type Observer struct {
	ProjectPath string
	Interval    time.Duration
	Broadcast   chan<- *StateSnapshot
	stop        chan struct{}
}

type TelemetryEvent struct {
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Source    string `json:"source"`
	Data      string `json:"data"`
}

type StateSnapshot struct {
	Version         string
	Status          string
	Health          int
	Phase           int
	Project         string
	TaskID          string
	Description     string
	Agents          []string
	Artifacts       int
	NextAction      string
	Timestamp       string
	TelemetryEvents []TelemetryEvent
	Interval        int
	AteneaScore     int
}

func New(projectPath string, interval time.Duration) *Observer {
	return &Observer{
		ProjectPath: projectPath,
		Interval:    interval,
		stop:        make(chan struct{}),
	}
}

func semaforo(status string) (string, string) {
	switch status {
	case "idle":
		return "\u26AA", display.FG_SECONDARY
	case "pending":
		return "\U0001F535", display.FG_WHITE
	case "in_progress":
		return "\U0001F7E1", display.FG_YELLOW
	case "completed":
		return "\U0001F7E2", display.FG_GREEN
	case "failed":
		return "\U0001F534", display.FG_SIGNAL
	default:
		return "\u26AA", display.FG_MUTED
	}
}

// evaluateAtenea evalúa 9 condiciones de governance y devuelve puntuación + condiciones
func evaluateAtenea(state *models.BlackboardState) (int, *models.AteneaConditions) {
	cond := &models.AteneaConditions{}
	checks := []struct {
		pass   *bool
		weight int
		eval   func(*models.BlackboardState) bool
	}{
		{&cond.ContextExists, 15, func(s *models.BlackboardState) bool { return true }}, // state loaded = context exists
		{&cond.JoinExists, 10, func(s *models.BlackboardState) bool { return true }},    // JOIN assumed when state exists
		{&cond.HandoffExists, 8, func(s *models.BlackboardState) bool { return len(s.CompletedAgents) > 0 }},
		{&cond.HasAgentSequence, 7, func(s *models.BlackboardState) bool { return len(s.AgentSequence) > 0 }},
		{&cond.ArtifactsExist, 10, func(s *models.BlackboardState) bool { return len(s.Artifacts) > 0 }},
		{&cond.TimestampFresh, 8, func(s *models.BlackboardState) bool { return s.Timestamp != "" }},
		{&cond.NoErrors, 10, func(s *models.BlackboardState) bool { return s.LastError == nil || *s.LastError == "" }},
		{&cond.PhaseDefined, 7, func(s *models.BlackboardState) bool { return s.CurrentPhase > 0 || s.Status != "pending" }},
		{&cond.SchemaCurrent, 5, func(s *models.BlackboardState) bool { return s.SchemaVersion != "" }},
	}
	score := 0
	for _, c := range checks {
		*c.pass = c.eval(state)
		if *c.pass {
			score += c.weight
		}
	}
	return score, cond
}

func healthScore(status string, artifactCount int, agentCount int, ateneaScore int) (int, string) {
	base := map[string]int{
		"completed": 95, "in_progress": 60,
		"pending": 35, "failed": 5, "idle": 15,
	}
	score := base[status]
	if score == 0 {
		score = 15
	}
	score += artifactCount * 2
	if agentCount > 0 {
		score += agentCount * 2
	}
	// ATENEA score (0-80) blended at 20%
	score += ateneaScore * 20 / 100
	if score > 100 {
		score = 100
	}
	color := display.FG_GREEN
	if score < 30 {
		color = display.FG_SIGNAL
	} else if score < 60 {
		color = display.FG_YELLOW
	}
	return score, color
}

func progressBar(pct int, width int) string {
	filled := pct * width / 100
	bar := strings.Repeat("█", filled)
	if filled < width {
		bar += strings.Repeat("░", width-filled)
	}
	return bar
}

func (o *Observer) readBlackboard() (*models.BlackboardState, error) {
	path := filepath.Join(o.ProjectPath, ".mvgn", "blackboard", "state.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read state.json: %w", err)
	}
	var state models.BlackboardState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("invalid state.json: %w", err)
	}
	return &state, nil
}

func (o *Observer) readRuntime() (*models.RuntimeManifest, error) {
	path := filepath.Join(o.ProjectPath, "mvgn-runtime.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read mvgn-runtime.json: %w", err)
	}
	var runtime models.RuntimeManifest
	if err := json.Unmarshal(data, &runtime); err != nil {
		return nil, fmt.Errorf("invalid mvgn-runtime.json: %w", err)
	}
	return &runtime, nil
}

func (o *Observer) readTelemetry(maxEvents int) []TelemetryEvent {
	eventsDir := filepath.Join(o.ProjectPath, "events")
	pattern := filepath.Join(eventsDir, "*.jsonl")
	files, err := filepath.Glob(pattern)
	if err != nil || len(files) == 0 {
		return nil
	}

	sort.Sort(sort.Reverse(sort.StringSlice(files)))

	var events []TelemetryEvent
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
			var raw struct {
				ID        string `json:"id"`
				Type      string `json:"type"`
				Timestamp string `json:"timestamp"`
				Actor     struct {
					Name string `json:"name"`
				} `json:"actor"`
				Metadata map[string]string `json:"metadata,omitempty"`
				State    *struct {
					From string `json:"from"`
					To   string `json:"to"`
				} `json:"state,omitempty"`
			}
			if err := json.Unmarshal([]byte(line), &raw); err != nil {
				continue
			}

			var st *struct{ From, To string }
			if raw.State != nil {
				st = &struct{ From, To string }{raw.State.From, raw.State.To}
			}
			data := describeEvent(raw.Type, raw.Metadata, st)
			events = append(events, TelemetryEvent{
				Timestamp: raw.Timestamp,
				Type:      raw.Type,
				Source:    raw.Actor.Name,
				Data:      data,
			})
		}
		file.Close()
		if len(events) >= maxEvents {
			events = events[:maxEvents]
			break
		}
	}
	return events
}

func describeEvent(typ string, meta map[string]string, state *struct{ From, To string }) string {
	switch typ {
	case "SESSION_STARTED":
		proj := meta["project"]
		return fmt.Sprintf("Session started · project: %s", proj)
	case "SESSION_CLOSED":
		return fmt.Sprintf("Session closed · duration: %ss", meta["duration_secs"])
	case "HEARTBEAT":
		return "♥ Heartbeat — system alive"
	case "HEALTH_CHECK":
		return fmt.Sprintf("Health check · score: %s", meta["score"])
	case "STATE_CHANGED":
		if state != nil {
			return fmt.Sprintf("State changed: %s → %s", state.From, state.To)
		}
		return "State changed"
	default:
		return fmt.Sprintf("%s event", typ)
	}
}

func topBorder(width int) string {
	return display.FG_MUTED + "╔" + strings.Repeat("═", width-2) + "╗" + display.ANSI_RESET
}
func bottomBorder(width int) string {
	return display.FG_MUTED + "╚" + strings.Repeat("═", width-2) + "╝" + display.ANSI_RESET
}
func sepBorder(width int) string {
	return display.FG_MUTED + "╠" + strings.Repeat("═", width-2) + "╣" + display.ANSI_RESET
}
func sideLine(content string, width int) string {
	c := content
	if len(c) > width-4 {
		c = c[:width-7] + "..."
	}
	pad := width - 4 - len(c)
	if pad < 0 {
		pad = 0
	}
	return display.FG_MUTED + "║ " + display.ANSI_RESET + c +
		strings.Repeat(" ", pad) + display.FG_MUTED + " ║" + display.ANSI_RESET
}
func panelHeader(title string, width int) string {
	inner := " " + title + " "
	remain := width - 4 - len(inner)
	if remain < 0 {
		remain = 0
	}
	return display.FG_MUTED + "║" + display.ANSI_RESET +
		display.FG_CRIMSON + inner + display.ANSI_RESET +
		strings.Repeat(" ", remain) +
		display.FG_MUTED + "║" + display.ANSI_RESET
}

func (o *Observer) render(state *models.BlackboardState, runtime *models.RuntimeManifest, events []TelemetryEvent, port int) {
	const width = 78

	sym, _ := semaforo(state.Status)
	statusLabel := strings.ToUpper(state.Status)

	version := "?"
	if runtime != nil {
		version = runtime.RuntimeVersion
	}

	ateneaScore, atenea := evaluateAtenea(state)
	state.AteneaConditions = atenea
	health, healthColor := healthScore(state.Status, len(state.Artifacts), len(state.CompletedAgents), ateneaScore)
	phaseLabel := fmt.Sprintf("PHASE %d", state.CurrentPhase+1)

	fmt.Print(display.ANSI_CLEAR + display.ANSI_HOME)

	// ─── TOP BANNER BAR ────────────────────────────────────────
	fmt.Println(topBorder(width))
	if port > 0 {
		fmt.Println(panelHeader(fmt.Sprintf("MVGN OBSERVER v%s  |  🌐 http://localhost:%d", version, port), width))
	} else {
		fmt.Println(panelHeader("MVGN OBSERVER v"+version, width))
	}
	fmt.Println(sideLine("Watching: "+o.ProjectPath+"  |  Status: "+statusLabel, width))
	fmt.Println(sepBorder(width))

	// ─── METRIC CARDS ROW ──────────────────────────────────────
	m1 := display.MetricCard(version, "VERSION")
	m2 := display.MetricCard(sym+" "+statusLabel, "STATUS")
	m3 := display.MetricCard(phaseLabel, "PROGRESS")

	healthStr := fmt.Sprintf("%s%d%%%s", healthColor, health, display.ANSI_RESET)
	healthCard := display.MetricCard(healthStr, "HEALTH")

	ateneaStr := fmt.Sprintf("%d/80", ateneaScore)
	ateneaCard := display.MetricCard(ateneaStr, "ATENEA")

	lines1 := strings.Split(m1, "\n")
	lines2 := strings.Split(m2, "\n")
	lines3 := strings.Split(m3, "\n")
	lines4 := strings.Split(healthCard, "\n")
	lines5 := strings.Split(ateneaCard, "\n")

	fmt.Println(sideLine("", width))
	for i := 0; i < len(lines1); i++ {
		a, b, c, d, e := "", "", "", "", ""
		if i < len(lines1) {
			a = lines1[i]
		}
		if i < len(lines2) {
			b = lines2[i]
		}
		if i < len(lines3) {
			c = lines3[i]
		}
		if i < len(lines4) {
			d = lines4[i]
		}
		if i < len(lines5) {
			e = lines5[i]
		}
		line := fmt.Sprintf("  %s  %s  %s  %s  %s", a, b, c, d, e)
		fmt.Println(sideLine(line, width))
	}
	fmt.Println(sepBorder(width))

	// ─── TASK PANEL ────────────────────────────────────────────
	fmt.Println(panelHeader("CURRENT TASK  "+display.FG_MUTED+state.TaskID+display.ANSI_RESET, width))
	fmt.Println(sideLine("", width))
	desc := state.TaskDescription
	if len(desc) > 60 {
		desc = desc[:57] + "..."
	}
	fmt.Println(sideLine("  "+display.FG_WHITE+desc+display.ANSI_RESET, width))
	pct := health
	bar := progressBar(pct, 40)
	fmt.Println(sideLine(fmt.Sprintf("  PROGRESS  %s%s %s%d%%%s",
		healthColor, bar, display.ANSI_RESET, healthColor, pct, display.ANSI_RESET), width))
	agents := "● " + strings.Join(state.CompletedAgents, "  ● ")
	fmt.Println(sideLine(fmt.Sprintf("  Agents   %s%s%s", display.FG_GREEN, agents, display.ANSI_RESET), width))
	fmt.Println(sideLine(fmt.Sprintf("  Artifacts  %s%d generated%s",
		display.FG_SECONDARY, len(state.Artifacts), display.ANSI_RESET), width))
	fmt.Println(sepBorder(width))

	// ─── TELEMETRY TIMELINE ────────────────────────────────────
	fmt.Println(panelHeader("EVENT TIMELINE", width))
	fmt.Println(sideLine("", width))
	if len(events) > 0 {
		max := 5
		if len(events) < max {
			max = len(events)
		}
		for i := 0; i < max; i++ {
			e := events[i]
			t := e.Timestamp
			if len(t) > 8 {
				t = t[len(t)-8:]
			}
			evType := e.Type
			if len(evType) > 16 {
				evType = evType[:16]
			}
			line := fmt.Sprintf("  %s%s%s %s%s  %s",
				display.FG_MUTED, t, display.ANSI_RESET,
				display.FG_SECONDARY, evType, display.ANSI_RESET)
			if len(e.Data) > 40 {
				line += e.Data[:37] + "..."
			} else {
				line += e.Data
			}
			fmt.Println(sideLine(line, width))
		}
	} else {
		fmt.Println(sideLine(fmt.Sprintf("  %sNo telemetry events yet%s",
			display.FG_MUTED, display.ANSI_RESET), width))
	}
	fmt.Println(sepBorder(width))

	// ─── NEXT ACTION ───────────────────────────────────────────
	fmt.Println(panelHeader("NEXT ACTION", width))
	fmt.Println(sideLine("", width))
	if state.NextAction != "" {
		na := state.NextAction
		for len(na) > 0 {
			chunk := na
			if len(chunk) > 68 {
				chunk = chunk[:68]
			}
			fmt.Println(sideLine("  "+chunk, width))
			if len(na) > 68 {
				na = na[68:]
			} else {
				na = ""
			}
		}
	} else {
		fmt.Println(sideLine("  No pending actions.", width))
	}
	fmt.Println(bottomBorder(width))

	// ─── FOOTER ────────────────────────────────────────────────
	fmt.Println()
	fmt.Printf("  %sPolling every %.0fs · Last update: %s · Press Ctrl+C to exit%s\n",
		display.FG_MUTED, o.Interval.Seconds(),
		time.Now().Format("15:04:05"), display.ANSI_RESET)
	if port > 0 {
		fmt.Printf("  %s🌐 Dashboard: http://localhost:%d%s\n",
			display.FG_GREEN, port, display.ANSI_RESET)
	}
	fmt.Printf("  %sCrimson Nexus Corporation · MVGN Labs · Node MX-03%s\n",
		display.FG_MUTED, display.ANSI_RESET)
}

func (o *Observer) snapshot(state *models.BlackboardState, runtime *models.RuntimeManifest, events []TelemetryEvent) *StateSnapshot {
	version := "?"
	if runtime != nil {
		version = runtime.RuntimeVersion
	}
	ateneaScore, _ := evaluateAtenea(state)
	health, _ := healthScore(state.Status, len(state.Artifacts), len(state.CompletedAgents), ateneaScore)
	return &StateSnapshot{
		Version:         version,
		Status:          state.Status,
		Health:          health,
		Phase:           state.CurrentPhase + 1,
		Project:         o.ProjectPath,
		TaskID:          state.TaskID,
		Description:     state.TaskDescription,
		Agents:          state.CompletedAgents,
		Artifacts:       len(state.Artifacts),
		NextAction:      state.NextAction,
		Timestamp:       time.Now().Format("15:04:05"),
		TelemetryEvents: events,
		Interval:        int(o.Interval.Seconds()),
		AteneaScore:     ateneaScore,
	}
}

func (o *Observer) Watch(port int) error {
	for {
		select {
		case <-o.stop:
			return nil
		default:
			state, stateErr := o.readBlackboard()
			runtime, runtimeErr := o.readRuntime()
			events := o.readTelemetry(10)

			if stateErr != nil {
				fmt.Print(display.ANSI_CLEAR + display.ANSI_HOME)
				fmt.Println(topBorder(78))
				fmt.Println(panelHeader("OBSERVER ERROR", 78))
				fmt.Println(sepBorder(78))
				fmt.Println(sideLine(display.Error(fmt.Sprintf("  %v", stateErr)), 78))
				if runtimeErr != nil {
					fmt.Println(sideLine(display.Error(fmt.Sprintf("  %v", runtimeErr)), 78))
				}
				fmt.Println(sideLine(fmt.Sprintf("  %sRetrying in %.0fs...%s",
					display.FG_MUTED, o.Interval.Seconds(), display.ANSI_RESET), 78))
				fmt.Println(bottomBorder(78))
			} else {
				o.render(state, runtime, events, port)

				if o.Broadcast != nil {
					snap := o.snapshot(state, runtime, events)
					select {
					case o.Broadcast <- snap:
					default:
					}
				}
			}
			time.Sleep(o.Interval)
		}
	}
}

func (o *Observer) Stop() {
	close(o.stop)
}
