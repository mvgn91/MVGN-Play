package query

import (
	"fmt"
	"strings"

	"github.com/mvgn/mvgn-scan/internal/display"
)

// ████████████████████████████████████████████████████████████████████████████
// QueryRenderer — Renderiza el estado observado con Scientific Glass UI (CNX-CANON-002)
// ████████████████████████████████████████████████████████████████████████████

// QueryRenderer implementa Renderer para la QUERY Dashboard.
type QueryRenderer struct {
	PollInterval string // intervalo de polling para mostrar en footer
}

// NewQueryRenderer crea un nuevo renderer para QUERY Dashboard.
func NewQueryRenderer(pollInterval string) *QueryRenderer {
	if pollInterval == "" {
		pollInterval = "5s"
	}
	return &QueryRenderer{PollInterval: pollInterval}
}

// Render implementa Renderer.Render.
func (r *QueryRenderer) Render(state ObservedState, noColor bool) string {
	var b strings.Builder

	if !noColor {
		b.WriteString("\033[2J\033[H")
	}

	if noColor {
		b.WriteString(r.renderPlain(state))
	} else {
		b.WriteString(r.renderANSI(state))
	}

	return b.String()
}

// renderANSI renderiza con colores ANSI (Scientific Glass UI).
func (r *QueryRenderer) renderANSI(state ObservedState) string {
	m := state.StringMap()
	var b strings.Builder

	// Section-meta header (CNX-CANON-002)
	b.WriteString(fmt.Sprintf("  %s%s%s  %s%s%s\n",
		display.FG_CRIMSON, "┌─ MVGN-QRY-001 ──────── QUERY DASHBOARD ────────┐",
		display.ANSI_RESET,
		display.FG_SECONDARY, "LIVE OBSERVATION", display.ANSI_RESET))

	// Panel superior: métricas
	b.WriteString(fmt.Sprintf("  %s┌─────────────────────────────────────────────────────────────┐%s\n",
		display.FG_MUTED, display.ANSI_RESET))

	// VERSION + STATUS row
	b.WriteString(fmt.Sprintf("  %s│%s  %sVersion:%s v%-5s  %sStatus:%s %s %s%s\n",
		display.FG_MUTED, display.ANSI_RESET,
		display.FG_SECONDARY, display.ANSI_RESET, state.SchemaVersion,
		display.FG_SECONDARY, display.ANSI_RESET,
		r.semaColor(state.State), m["SEMAPHORE"], m["STATE_LABEL"]))

	// Task section
	b.WriteString(fmt.Sprintf("  %s│%s───────────────────────────────────────────────────────────────│%s\n",
		display.FG_MUTED, display.ANSI_RESET, display.FG_MUTED))
	b.WriteString(fmt.Sprintf("  %s│%s\n", display.FG_MUTED, display.ANSI_RESET))
	b.WriteString(fmt.Sprintf("  %s│%s  %s%s%s %s\n",
		display.FG_MUTED, display.ANSI_RESET,
		display.FG_SECONDARY, "TASK", display.ANSI_RESET,
		m["TASK"]))
	b.WriteString(fmt.Sprintf("  %s│%s  %s%s%s %s\n",
		display.FG_MUTED, display.ANSI_RESET,
		display.FG_MUTED, "ID:", display.ANSI_RESET,
		m["TASK_ID"]))
	b.WriteString(fmt.Sprintf("  %s│%s  %s%s%s %s  │  %s%s%s %s\n",
		display.FG_MUTED, display.ANSI_RESET,
		display.FG_MUTED, "PHASE:", display.ANSI_RESET, m["PHASE"],
		display.FG_MUTED, "AGENTS:", display.ANSI_RESET, m["AGENTS"]))

	// Metadata section
	b.WriteString(fmt.Sprintf("  %s│%s───────────────────────────────────────────────────────────────│%s\n",
		display.FG_MUTED, display.ANSI_RESET, display.FG_MUTED))
	b.WriteString(fmt.Sprintf("  %s│%s\n", display.FG_MUTED, display.ANSI_RESET))
	b.WriteString(fmt.Sprintf("  %s│%s  %s%s%s %s\n",
		display.FG_MUTED, display.ANSI_RESET,
		display.FG_MUTED, "LAST UPDATE:", display.ANSI_RESET,
		m["LAST_UPDATE"]))
	b.WriteString(fmt.Sprintf("  %s│%s  %s%s%s %s  │  %s%s%s %s\n",
		display.FG_MUTED, display.ANSI_RESET,
		display.FG_MUTED, "AGE:", display.ANSI_RESET, m["AGE"],
		display.FG_MUTED, "SOURCE:", display.ANSI_RESET, "file"))
	b.WriteString(fmt.Sprintf("  %s│%s  %s%s%s %s  │  %s%s%s %s\n",
		display.FG_MUTED, display.ANSI_RESET,
		display.FG_MUTED, "FRESHNESS:", display.ANSI_RESET, r.freshnessANSI(state.Freshness),
		display.FG_MUTED, "CONFIDENCE:", display.ANSI_RESET, r.confidenceANSI(state.Confidence)))

	// Error section
	if state.Err != nil || state.LastError != "" {
		errMsg := state.LastError
		if errMsg == "" && state.Err != nil {
			errMsg = state.Err.Error()
		}
		b.WriteString(fmt.Sprintf("  %s│%s\n", display.FG_MUTED, display.ANSI_RESET))
		b.WriteString(fmt.Sprintf("  %s│%s  %s%s%s %s\n",
			display.FG_MUTED, display.ANSI_RESET,
			display.FG_SIGNAL, "ERROR:", display.ANSI_RESET, errMsg))
		b.WriteString(fmt.Sprintf("  %s│%s  %s%s%s %s\n",
			display.FG_MUTED, display.ANSI_RESET,
			display.FG_YELLOW, "ACTION:", display.ANSI_RESET, ActionForError(state.Err)))
	}

	// Cierre
	b.WriteString(fmt.Sprintf("  %s└─────────────────────────────────────────────────────────────┘%s\n",
		display.FG_MUTED, display.ANSI_RESET))
	b.WriteString(fmt.Sprintf("  %s── Watching · Poll: %s · Ctrl+C to exit ──────────────%s\n",
		display.FG_MUTED, r.PollInterval, display.ANSI_RESET))
	b.WriteString(fmt.Sprintf("  %sCrimson Nexus Corporation · MVGN Labs · Node MX-03%s\n",
		display.FG_MUTED, display.ANSI_RESET))

	return b.String()
}

// renderPlain renderiza sin colores ANSI.
func (r *QueryRenderer) renderPlain(state ObservedState) string {
	m := state.StringMap()
	var b strings.Builder

	b.WriteString("┌─ MVGN-QRY-001 ──────── QUERY DASHBOARD ────────┐\n")
	b.WriteString(fmt.Sprintf("  Version: v%-5s  Status: %s %s\n",
		state.SchemaVersion, m["SEMAPHORE"], m["STATE_LABEL"]))
	b.WriteString(fmt.Sprintf("  Task:     %s (%s)\n", m["TASK"], m["TASK_ID"]))
	b.WriteString(fmt.Sprintf("  Phase:    %s  |  Agents: %s\n", m["PHASE"], m["AGENTS"]))
	b.WriteString(fmt.Sprintf("  Updated:  %s  |  Age: %s\n", m["LAST_UPDATE"], m["AGE"]))
	b.WriteString(fmt.Sprintf("  Freshness: %s  |  Confidence: %s\n", m["FRESHNESS"], m["CONFIDENCE"]))

	if state.Err != nil || state.LastError != "" {
		errMsg := state.LastError
		if errMsg == "" && state.Err != nil {
			errMsg = state.Err.Error()
		}
		b.WriteString(fmt.Sprintf("  ERROR:  %s\n", errMsg))
		b.WriteString(fmt.Sprintf("  ACTION: %s\n", ActionForError(state.Err)))
	}

	b.WriteString("└──────────────────────────────────────────────────┘\n")
	return b.String()
}

// ████████████████████████████████████████████████████████████████████████████
// Helpers ANSI
// ████████████████████████████████████████████████████████████████████████████

func (r *QueryRenderer) semaColor(s State) string {
	switch s {
	case StateIdle:
		return display.FG_SECONDARY
	case StatePending:
		return display.FG_WHITE
	case StateInProgress:
		return display.FG_YELLOW
	case StateCompleted:
		return display.FG_GREEN
	case StateFailed:
		return display.FG_SIGNAL
	default:
		return display.FG_MUTED
	}
}

func (r *QueryRenderer) freshnessANSI(f Freshness) string {
	switch f {
	case FreshnessFresh:
		return fmt.Sprintf("%sFRESH%s", display.FG_GREEN, display.ANSI_RESET)
	case FreshnessStale:
		return fmt.Sprintf("%sSTALE%s", display.FG_YELLOW, display.ANSI_RESET)
	default:
		return fmt.Sprintf("%sUNKNOWN%s", display.FG_MUTED, display.ANSI_RESET)
	}
}

func (r *QueryRenderer) confidenceANSI(c Confidence) string {
	switch c {
	case ConfidenceValid:
		return fmt.Sprintf("%sVALID%s", display.FG_GREEN, display.ANSI_RESET)
	case ConfidenceDegraded:
		return fmt.Sprintf("%sDEGRADED%s", display.FG_YELLOW, display.ANSI_RESET)
	case ConfidenceInvalid:
		return fmt.Sprintf("%sINVALID%s", display.FG_SIGNAL, display.ANSI_RESET)
	default:
		return fmt.Sprintf("%sUNKNOWN%s", display.FG_MUTED, display.ANSI_RESET)
	}
}

// padCenter centra un string con padding ANSI-aware.
// No se usa actualmente en el MVP; reservado para formato avanzado.
// TODO: Activar cuando se implemente modo grid con alineación.
func (r *QueryRenderer) padCenter(s string, width int) string {
	plain := stripANSI(s)
	padding := width - len(plain)
	if padding <= 0 {
		return s
	}
	left := padding / 2
	right := padding - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

// stripANSI elimina secuencias de escape ANSI de un string.
func stripANSI(s string) string {
	var result strings.Builder
	inEscape := false
	for _, ch := range s {
		if ch == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
				inEscape = false
			}
			continue
		}
		result.WriteRune(ch)
	}
	return result.String()
}
