package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/mvgn/telemetry/internal/core"
	"github.com/mvgn/telemetry/internal/artifact"
)

func main() {
	// Subcommands
	reportCmd := flag.NewFlagSet("report", flag.ExitOnError)
	reportSession := reportCmd.String("session", "", "Session ID to generate report for")
	reportLatest := reportCmd.Bool("latest", false, "Generate report for latest session")
	reportForce := reportCmd.Bool("force", false, "Force regeneration even if reports exist")

	listCmd := flag.NewFlagSet("list", flag.ExitOnError)

	// Root flags
	help := flag.Bool("help", false, "Show help")
	flag.Parse()

	if *help {
		printUsage()
		os.Exit(0)
	}

	// Route to subcommand
	switch flag.Arg(0) {
	case "report":
		reportCmd.Parse(os.Args[2:])
		runReport(*reportSession, *reportLatest, *reportForce)
	case "list":
		listCmd.Parse(os.Args[2:])
		runList()
	default:
		runCapture()
	}
}

func printUsage() {
	fmt.Println("MVGN Telemetry v2 — Session Trace Engine")
	fmt.Println()
	fmt.Println("Uso:")
	fmt.Println("  mvgn-telemetry                  Capture events (default)")
	fmt.Println("  mvgn-telemetry report           Generate session report")
	fmt.Println("    --latest                        Use latest session")
	fmt.Println("    --session <id>                  Use specific session")
	fmt.Println("    --force                         Force regeneration")
	fmt.Println("  mvgn-telemetry list             List available sessions")
	fmt.Println()
	fmt.Println("Ejemplos:")
	fmt.Println("  mvgn-telemetry                  Start capturing events")
	fmt.Println("  mvgn-telemetry report --latest  Generate report for last session")
	fmt.Println("  mvgn-telemetry report --session ses-20260717-130000")
	fmt.Println("  mvgn-telemetry report --latest --force  Force regenerate latest")
	fmt.Println("  mvgn-telemetry list             List all sessions")
}

func runCapture() {
	app := core.New()
	if err := app.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Handle interrupt — only set up after Init succeeds (H-CR-30)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println("\nInterrupted. Shutting down...")
		app.Stop()
	}()

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if err := app.Shutdown(); err != nil {
		fmt.Fprintf(os.Stderr, "Error during shutdown: %v\n", err)
		os.Exit(1)
	}
}

// reportExists checks if all 3 artifact files exist for a session.
func reportExists(reportsDir, sessionID string) bool {
	exts := []string{".md", ".json", ".metrics.json"}
	for _, ext := range exts {
		path := filepath.Join(reportsDir, sessionID+ext)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func runReport(sessionID string, latest bool, force bool) {
	eventsDir := "events"
	reportsDir := "reports"

	if latest {
		sessionID = artifact.FindLatestSession(eventsDir)
		if sessionID == "" {
			fmt.Println("No sessions found.")
			os.Exit(1)
		}
	}

	if sessionID == "" {
		fmt.Println("Error: specify --session <id> or --latest")
		os.Exit(1)
	}

	// Check if all 3 artifacts already exist (M-CR-31)
	if !force && reportExists(reportsDir, sessionID) {
		fmt.Printf("Reports already exist for session: %s\n", sessionID)
		fmt.Println("Use --force to regenerate.")
		return
	}

	engine := artifact.NewEngine(eventsDir, reportsDir)
	if err := engine.Generate(sessionID); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating report: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Report generated for session: %s\n", sessionID)
}

func runList() {
	eventsDir := "events"
	sessions := artifact.ListSessions(eventsDir)

	if len(sessions) == 0 {
		fmt.Println("No sessions found.")
		return
	}

	fmt.Printf("Sessions (%d):\n", len(sessions))
	for _, s := range sessions {
		fmt.Printf("  %s\n", s)
	}
}
