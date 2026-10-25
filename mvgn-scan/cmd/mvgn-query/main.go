package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/mvgn/mvgn-scan/internal/display"
	"github.com/mvgn/mvgn-scan/internal/query"
)

func main() {
	// ─── FLAGS ──────────────────────────────────────────────────────────
	watch := flag.Bool("watch", false, "Enable live observation mode")
	interval := flag.Duration("interval", 5*time.Second, "Polling interval (e.g., 5s, 2s, 500ms)")
	project := flag.String("project", ".", "Root directory of the MVGN project")
	noColor := flag.Bool("no-color", false, "Disable ANSI colors")
	autoClose := flag.Bool("auto-close", false, "Exit immediately after snapshot (no pause)")
	staleThreshold := flag.Int64("stale", 300, "Stale threshold in seconds (default: 5min)")
	version := flag.Bool("version", false, "Show version information")
	flag.Parse()

	// ─── VERSION ────────────────────────────────────────────────────────
	if *version {
		fmt.Println("mvgn-query v1.0.0 — QUERY Dashboard")
		fmt.Println("MVGN v5.2 — Crimson Nexus Corporation")
		fmt.Println("Node MX-03")
		os.Exit(0)
	}

	// ─── SNAPSHOT MODE ─────────────────────────────────────────────────
	if !*watch {
		// Defer: siempre esperar tecla antes de salir (aunque haya panic)
		if !*autoClose {
			defer func() {
				fmt.Print("\n\nPresione una tecla para cerrar...")
				buf := make([]byte, 1)
				os.Stdin.Read(buf) // Espera cualquier tecla, no solo Enter
			}()
		}
		printSingleShot(*project, *noColor, *staleThreshold)
		return
	}

	// ─── WATCH MODE ─────────────────────────────────────────────────────
	runWatch(*project, *interval, *noColor, *staleThreshold)
}

// printSingleShot ejecuta una sola lectura y muestra el estado (modo snapshot).
func printSingleShot(projectPath string, noColor bool, staleThresholdSecs int64) {
	if !noColor {
		fmt.Printf(display.Banner, display.FG_CRIMSON, display.ANSI_RESET)
		fmt.Println()
	}

	source := query.NewFileStateSource(projectPath, staleThresholdSecs)
	ctx := context.Background()

	state, err := source.Current(ctx)
	if err != nil && state.Err == nil {
		state.Err = err
	}

	intervalStr := "5s"
	renderer := query.NewQueryRenderer(intervalStr)
	output := renderer.Render(state, noColor)
	fmt.Print(output)
}

// runWatch inicia el modo de observación continua.
func runWatch(projectPath string, interval time.Duration, noColor bool, staleThresholdSecs int64) {
	source := query.NewFileStateSource(projectPath, staleThresholdSecs)
	intervalStr := interval.String()
	renderer := query.NewQueryRenderer(intervalStr)

	watcher := query.NewWatcher(source, interval, noColor, renderer)

	fmt.Fprintf(os.Stderr, "MVGN QUERY Dashboard — Watching project: %s\n", projectPath)
	fmt.Fprintf(os.Stderr, "Polling interval: %s  ·  Stale threshold: %ds\n", interval, staleThresholdSecs)
	fmt.Fprintf(os.Stderr, "Press Ctrl+C to exit.\n\n")

	ctx := context.Background()
	if err := watcher.Run(ctx); err != nil {
		// Context.Canceled is expected on Ctrl+C
		if err != context.Canceled {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}
