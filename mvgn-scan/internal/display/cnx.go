package display

import (
	"fmt"
	"strings"
	"time"
)

// ─── ANSI Color Constants (CNX Scientific Glass UI palette) ────
const (
	ANSI_RESET = "\033[0m"
	ANSI_CLEAR = "\033[2J"
	ANSI_HOME  = "\033[H"

	// Foreground
	FG_CRIMSON   = "\033[38;2;214;0;28m"
	FG_SIGNAL    = "\033[38;2;255;40;48m"
	FG_WHITE     = "\033[38;2;232;232;237m"
	FG_SECONDARY = "\033[38;2;152;152;168m"
	FG_MUTED     = "\033[38;2;92;92;110m"
	FG_GREEN     = "\033[38;2;80;200;120m"
	FG_YELLOW    = "\033[38;2;220;180;60m"

)

// ─── Section Meta Header ───────────────────────────────────────
// Patrón CNX-CANON-002: Toda sección comienza con ID tipo expediente

type SectionID struct {
	Code string // e.g. "MVGN"
	Type string // e.g. "SCN"
	Num  int    // e.g. 001
}

func (s SectionID) String() string {
	return fmt.Sprintf("%s-%s-%03d", s.Code, s.Type, s.Num)
}

// SectionHeader renderiza el meta header tipo CNX-CANON-002
// Ejemplo: "  MVGN-SCN-001 · OBSERVATION INTERFACE"
func SectionHeader(id SectionID, label string) string {
	return fmt.Sprintf("  %s%s\u2502%s %s%s%s",
		FG_CRIMSON, id.String(),
		ANSI_RESET,
		FG_SECONDARY, label, ANSI_RESET)
}

// ─── Box Drawing ───────────────────────────────────────────────

// BoxSimple renderiza un recuadro simple con título opcional
// ┌─ title ─────────────────────────────────┐
// │  content                                │
// └─────────────────────────────────────────┘
func BoxSimple(title string, width int, lines []string) string {
	var b strings.Builder

	// Top border
	if title != "" {
		b.WriteString(fmt.Sprintf("  %s\u250C\u2500 %s \u2500", FG_MUTED, title))
		remain := width - len(title) - 4
		if remain > 0 {
			b.WriteString(strings.Repeat("\u2500", remain))
		}
		b.WriteString(fmt.Sprintf("\u2510%s\n", ANSI_RESET))
	} else {
		b.WriteString(fmt.Sprintf("  %s\u250C%s\u2510%s\n",
			FG_MUTED, strings.Repeat("\u2500", width-2), ANSI_RESET))
	}

	// Content lines
	for _, line := range lines {
		b.WriteString(fmt.Sprintf("  %s\u2502%s%s%s\u2502%s\n",
			FG_MUTED, ANSI_RESET, line, FG_MUTED, ANSI_RESET))
	}

	// Bottom border
	b.WriteString(fmt.Sprintf("  %s\u2514%s\u2518%s\n",
		FG_MUTED, strings.Repeat("\u2500", width-2), ANSI_RESET))

	return b.String()
}

// StatusTag renderiza un tag de estado al estilo CNX
func StatusTag(label string, status string) string {
	var color string
	switch status {
	case "ok", "completed", "active", "yes":
		color = FG_GREEN
	case "warn", "warning", "in_progress", "blocked":
		color = FG_YELLOW
	case "error", "failed", "invalid", "no":
		color = FG_SIGNAL
	default:
		color = FG_SECONDARY
	}
	return fmt.Sprintf("%s\u25CF%s %s%s%s",
		color, ANSI_RESET, FG_WHITE, label, ANSI_RESET)
}

// MetricCard renderiza una métrica al estilo CNX
// ┌─────────────────────┐
// │  VALUE              │  ← rojo
// │  label              │  ← muted mono
// └─────────────────────┘
func MetricCard(value, label string) string {
	cardW := 22
	
	visibleLen := func(s string) int {
		n := 0
		in := false
		for _, c := range s {
			if c == '\033' { in = true; continue }
			if in { if c == 'm' { in = false }; continue }
			n++
		}
		return n
	}

	vLen := visibleLen(value)
	lLen := visibleLen(label)

	valPad := (cardW - 2 - vLen) / 2
	if valPad < 0 {
		valPad = 0
	}
	labelPad := (cardW - 2 - lLen) / 2
	if labelPad < 0 {
		labelPad = 0
	}

	valPadRight := cardW - 2 - valPad - vLen
	if valPadRight < 0 {
		valPadRight = 0
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s\u250C%s\u2510%s\n",
		FG_MUTED, strings.Repeat("\u2500", cardW-2), ANSI_RESET))
	b.WriteString(fmt.Sprintf("%s\u2502%s%s%s%s\u2502%s\n",
		FG_MUTED, strings.Repeat(" ", valPad), value,
		strings.Repeat(" ", valPadRight), ANSI_RESET))
	b.WriteString(fmt.Sprintf("%s\u2502%s%s%s%s\u2502%s\n",
		FG_MUTED, strings.Repeat(" ", labelPad), FG_MUTED, label,
		strings.Repeat(" ", cardW-2-labelPad-lLen), ANSI_RESET))
	b.WriteString(fmt.Sprintf("%s\u2514%s\u2518%s",
		FG_MUTED, strings.Repeat("\u2500", cardW-2), ANSI_RESET))
	return b.String()
}

// MetricRow renderiza una fila de 3 métricas
func MetricRow(v1, l1, v2, l2, v3, l3 string) string {
	m1 := MetricCard(v1, l1)
	m2 := MetricCard(v2, l2)
	m3 := MetricCard(v3, l3)

	lines1 := strings.Split(m1, "\n")
	lines2 := strings.Split(m2, "\n")
	lines3 := strings.Split(m3, "\n")

	var b strings.Builder
	for i := 0; i < len(lines1); i++ {
		line := lines1[i]
		if i < len(lines2) {
			line += "  " + lines2[i]
		}
		if i < len(lines3) {
			line += "  " + lines3[i]
		}
		b.WriteString(fmt.Sprintf("  %s\n", line))
	}
	return b.String()
}

// ─── Header Banner ─────────────────────────────────────────────

const Banner = `
%s  ╔═══════════════════════════════════════════════════════════════╗
  ║                                                                 ║
  ║   ███╗   ███╗██╗   ██╗ ██████╗ ███╗   ██╗                     ║
  ║   ████╗ ████║██║   ██║██╔════╝ ████╗  ██║                     ║
  ║   ██╔████╔██║██║   ██║██║  ███╗██╔██╗ ██║                     ║
  ║   ██║╚██╔╝██║╚██╗ ██╔╝██║   ██║██║╚██╗██║                     ║
  ║   ██║ ╚═╝ ██║ ╚████╔╝ ╚██████╔╝██║ ╚████║                     ║
  ║   ╚═╝     ╚═╝  ╚═══╝   ╚═════╝ ╚═╝  ╚═══╝                     ║
  ║                                                                 ║
  ╚══════════════════════════════════════════════════════════════════╝%s`

// Footer con línea de cierre CNX
func Footer(timestamp time.Time) string {
	return fmt.Sprintf("\n%s%s\n%s  Scan completed · %s%s\n%s  Crimson Nexus Corporation · MVGN Labs · Node MX-03%s\n%s%s\n",
		FG_SECONDARY, strings.Repeat("\u2500", 75),
		FG_MUTED, timestamp.Format("2006-01-02T15:04:05Z"), ANSI_RESET,
		FG_MUTED, ANSI_RESET,
		FG_SECONDARY, strings.Repeat("\u2500", 75))
}

// ─── Label:Value ───────────────────────────────────────────────
func Field(label, value string) string {
	return fmt.Sprintf("  %s%s:%s %s", FG_SECONDARY, label, ANSI_RESET, value)
}

// ─── Error styling ─────────────────────────────────────────────
func Error(msg string) string {
	return fmt.Sprintf("%s\u2716 %s%s", FG_SIGNAL, msg, ANSI_RESET)
}

func Warning(msg string) string {
	return fmt.Sprintf("%s\u26A0 %s%s", FG_YELLOW, msg, ANSI_RESET)
}

func Success(msg string) string {
	return fmt.Sprintf("%s\u2714 %s%s", FG_GREEN, msg, ANSI_RESET)
}
