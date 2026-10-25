package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mvgn/mvgn-scan/internal/consolidator"
	"github.com/mvgn/mvgn-scan/internal/display"
	"github.com/mvgn/mvgn-scan/internal/observer"
	"github.com/mvgn/mvgn-scan/internal/scanner"
	"github.com/mvgn/mvgn-scan/internal/server"
	"github.com/mvgn/mvgn-scan/internal/validator"
)

func main() {
	// Flags
	rootDir := flag.String("dir", ".", "Directorio raíz para escanear")
	outputDir := flag.String("output", filepath.Join(os.Getenv("USERPROFILE"), ".mvgn"), "Directorio de salida para dashboard-data.json")
	verbose := flag.Bool("verbose", false, "Mostrar información detallada")
	watch := flag.Bool("watch", false, "Modo observación en terminal con semáforo")
	observe := flag.Bool("observe", false, "Modo dashboard visual (abre navegador)")
	port := flag.Int("port", 8080, "Puerto para el dashboard visual (--observe)")
	interval := flag.Int("interval", 5, "Intervalo de polling en segundos")
	flag.Parse()

	// ─── UNIFIED: WATCH MODE (Terminal + Server) ─────────────
	if *watch || *observe {
		obs := observer.New(*rootDir, time.Duration(*interval)*time.Second)
		srv := server.New(*port, *interval, *rootDir)

		updates := make(chan *observer.StateSnapshot, 8)
		obs.Broadcast = updates

		go func() {
			for snap := range updates {
				events := make([]server.TelemetryEvent, len(snap.TelemetryEvents))
				for i, e := range snap.TelemetryEvents {
					events[i] = server.TelemetryEvent{
						Timestamp: e.Timestamp,
						Type:      e.Type,
						Source:    e.Source,
						Data:      e.Data,
					}
				}
				srv.Send(&server.Update{
					Version:         snap.Version,
					Status:          snap.Status,
					Health:          snap.Health,
					Phase:           snap.Phase,
					Project:         snap.Project,
					TaskID:          snap.TaskID,
					Description:     snap.Description,
					Agents:          snap.Agents,
					Artifacts:       snap.Artifacts,
					NextAction:      snap.NextAction,
					Timestamp:       snap.Timestamp,
					AteneaScore:     snap.AteneaScore,
					Interval:        snap.Interval,
					TelemetryEvents: events,
				})
			}
		}()

		go func() {
			watchPort := *port
			if !*watch {
				watchPort = 0
			}
			if err := obs.Watch(watchPort); err != nil {
				fmt.Fprintf(os.Stderr, "Observer error: %v\n", err)
				os.Exit(1)
			}
		}()

		if *observe {
			if err := srv.Start(); err != nil {
				fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
				os.Exit(1)
			}
		} else {
			if err := srv.StartBackground(); err != nil {
				fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
				os.Exit(1)
			}
			// Observer.Watch already started above — it renders + broadcasts
			select {}
		}
		return
	}

	// ─── BANNER ────────────────────────────────────────────────
	fmt.Printf(display.Banner, display.FG_CRIMSON, display.ANSI_RESET)
	fmt.Println()
	fmt.Println(display.SectionHeader(
		display.SectionID{"MVGN", "SCN", 1},
		"OBSERVATION INTERFACE"))
	fmt.Println()

	// ─── SCAN CONFIG ───────────────────────────────────────────
	fmt.Println(display.Field("Scan Target", filepath.Clean(*rootDir)))

	// Validar directorio de entrada
	if _, err := os.Stat(*rootDir); os.IsNotExist(err) {
		fmt.Println(display.Error(fmt.Sprintf("Directory not found: %s", *rootDir)))
		os.Exit(1)
	}

	// ─── SCAN PHASE ────────────────────────────────────────────
	fmt.Println()
	fmt.Println(display.SectionHeader(
		display.SectionID{"MVGN", "SCN", 2},
		"SCANNING"))

	startTime := time.Now()
	projects, err := scanner.ScanDirectorio(*rootDir)
	if err != nil {
		fmt.Println(display.Error(fmt.Sprintf("Scan failed: %v", err)))
		os.Exit(1)
	}
	scanDuration := time.Since(startTime)

	if len(projects) == 0 {
		fmt.Println()
		box := display.BoxSimple("STATUS", 72, []string{
			fmt.Sprintf("  %s  %s", display.FG_YELLOW+"!"+display.ANSI_RESET, "No valid MVGN projects found."),
			"",
			"  A valid project requires:",
			"    · mvgn-runtime.json",
			"    · .mvgn-context.json",
		})
		fmt.Println(box)
		fmt.Println()
		fmt.Println(display.Footer(time.Now()))
		os.Exit(0)
	}

	// ─── METRIC ROW ────────────────────────────────────────────
	fmt.Println()
	fmt.Println(display.MetricRow(
		fmt.Sprintf("%d", len(projects)), "SCANNED",
		"-", "VALID",
		fmt.Sprintf("%.2fs", scanDuration.Seconds()), "DURATION"))

	// ─── VALIDATION PHASE ──────────────────────────────────────
	fmt.Println()
	fmt.Println(display.SectionHeader(
		display.SectionID{"MVGN", "SCN", 3},
		"VALIDATION"))

	validProjects := []scanner.ScanResult{}
	for _, project := range projects {
		pName := filepath.Base(project.Path)

		if *verbose {
			fmt.Println(display.Field("Project", pName))
		}

		valid := true
		warnings := []string{}

		// Validar mvgn-runtime.json
		if project.HasRuntime {
			runtimePath := filepath.Join(project.Path, "mvgn-runtime.json")
			result := validator.ValidateRuntimeFile(runtimePath)
			if !result.Valid {
				if *verbose {
					fmt.Printf("    %s\n", display.Error(fmt.Sprintf("Runtime invalid: %v", result.Errors)))
				}
				valid = false
			}
			warnings = append(warnings, result.Warnings...)
		}

		// Validar .mvgn-context.json
		if project.HasContext {
			contextPath := filepath.Join(project.Path, ".mvgn-context.json")
			result := validator.ValidateContextFile(contextPath)
			if !result.Valid {
				if *verbose {
					fmt.Printf("    %s\n", display.Error(fmt.Sprintf("Context invalid: %v", result.Errors)))
				}
				valid = false
			}
			warnings = append(warnings, result.Warnings...)
		}

		if !valid {
			fmt.Printf("  %s %s\n", display.Error(""), pName)
			continue
		}

		// Mostrar warnings si verbose
		if *verbose && len(warnings) > 0 {
			for _, w := range warnings {
				fmt.Printf("    %s\n", display.Warning(w))
			}
		}

		fmt.Printf("  %s\n", display.Success(pName))
		validProjects = append(validProjects, project)
	}

	// Recalcular métrica de válidos
	vCount := len(validProjects)
	fmt.Println()
	fmt.Println(display.MetricRow(
		fmt.Sprintf("%d", len(projects)), "SCANNED",
		fmt.Sprintf("%d", vCount), "VALID",
		fmt.Sprintf("%.2fs", scanDuration.Seconds()), "DURATION"))

	if vCount == 0 {
		fmt.Println()
		fmt.Println(display.Footer(time.Now()))
		os.Exit(0)
	}

	// ─── CONSOLIDATION ─────────────────────────────────────────
	fmt.Println()
	fmt.Println(display.SectionHeader(
		display.SectionID{"MVGN", "SCN", 4},
		"CONSOLIDATION"))

	cons := consolidator.NewConsolidator(*outputDir)
	dashboard, err := cons.ConsolidateProjects(validProjects)
	if err != nil {
		fmt.Println(display.Error(fmt.Sprintf("Consolidation failed: %v", err)))
		os.Exit(1)
	}

	if err := cons.SaveDashboard(dashboard); err != nil {
		fmt.Println(display.Error(fmt.Sprintf("Dashboard save failed: %v", err)))
		os.Exit(1)
	}

	outputPath := filepath.Join(*outputDir, "dashboard-data.json")
	fmt.Printf("  %s\n", display.Success(fmt.Sprintf("Dashboard: %s", outputPath)))

	// ─── PROJECT DETAILS ──────────────────────────────────────
	if vCount > 0 && *verbose {
		fmt.Println()
		fmt.Println(display.SectionHeader(
			display.SectionID{"MVGN", "SCN", 5},
			"PROJECT DETAILS"))

		for _, p := range validProjects {
			pName := filepath.Base(p.Path)
			content := []string{
				fmt.Sprintf("  Path:    %s", p.Path),
				fmt.Sprintf("  Runtime: %s", boolToYesNo(p.HasRuntime)),
				fmt.Sprintf("  Context: %s", boolToYesNo(p.HasContext)),
			}
			fmt.Println(display.BoxSimple(pName, 72, content))
		}
	}

	// ─── SUMMARY BOX ──────────────────────────────────────────
	fmt.Println()
	summaryLines := []string{
		fmt.Sprintf("  %s%d%s projects scanned  |  %s%d%s valid  |  %s%.2fs%s",
			display.FG_CRIMSON, len(projects), display.ANSI_RESET,
			display.FG_GREEN, vCount, display.ANSI_RESET,
			display.FG_CRIMSON, scanDuration.Seconds(), display.ANSI_RESET),
		fmt.Sprintf("  %s", display.Field("Dashboard", outputPath)),
	}
	if vCount > 0 && dashboard.Projects[0].ProjectID != "" {
		summaryLines = append(summaryLines,
			fmt.Sprintf("  %s", display.Field("Last Project", dashboard.Projects[0].ProjectID)))
	}
	fmt.Println(display.BoxSimple("SUMMARY", 72, summaryLines))

	// ─── FOOTER ────────────────────────────────────────────────
	fmt.Println(display.Footer(time.Now()))
}

// boolToYesNo es un helper inline
func boolToYesNo(b bool) string {
	if b {
		return display.FG_GREEN + "yes" + display.ANSI_RESET
	}
	return display.FG_MUTED + "no" + display.ANSI_RESET
}
