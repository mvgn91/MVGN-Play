package core

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mvgn/telemetry/internal/artifact"
	"github.com/mvgn/telemetry/internal/events"
	"github.com/mvgn/telemetry/internal/mvgn"
)

// App manages the lifecycle of the telemetry system: init → run → shutdown.
// Telemetry v2 — Session Trace Engine (ADR-011)
type App struct {
	config      *Config
	session     *Session
	ctxReader   *mvgn.ContextReader
	docsReader  *mvgn.DocsReader
	eventReader *mvgn.EventHistoryReader
	eventBus    *events.Bus
	ctx         context.Context
	cancel      context.CancelFunc
}

// New creates a new App with default state.
func New() *App {
	return &App{}
}

// Init loads config and initializes subsystems.
func (a *App) Init() error {
	fmt.Println("MVGN Telemetry v2 — Initializing...")

	// 1. Load config
	cfg, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	a.config = cfg

	fmt.Printf("  Config: project=%q\n", a.config.ProjectPath)

	// 2. Init Event Bus (ADR-003)
	eventsDir := cfg.ProjectPath + "/events"
	if _, err := os.Stat(eventsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(eventsDir, 0755); err != nil {
			fmt.Printf("  Warning: could not create events dir: %v\n", err)
		}
	}
	writer, err := events.NewWriter(eventsDir)
	if err != nil {
		return fmt.Errorf("init event writer: %w", err)
	}
	bus := events.NewBus(writer)
	a.eventBus = bus
	fmt.Println("  Event Bus: initialized.")

	// 3. Init cancellable context
	a.ctx, a.cancel = context.WithCancel(context.Background())

	// 4. Init MVGN readers
	a.ctxReader = mvgn.NewContextReader(cfg.ProjectPath)
	a.docsReader = mvgn.NewDocsReader(cfg.ProjectPath)
	a.eventReader = mvgn.NewEventReader(cfg.ProjectPath)
	fmt.Println("  MVGN Readers: initialized.")

	// 5. Load AI actors from runtime manifest
	if manifest, err := a.ctxReader.ReadRuntime(); err == nil {
		actors := make(map[string]events.Actor)
		for _, ra := range manifest.Actors {
			act := events.Actor{
				ID:   ra.ID,
				Type: ra.Type,
				Name: ra.Name,
			}
			if ra.Model != "" {
				act.Model = ra.Model
			}
			actors[ra.ID] = act
		}
		if len(actors) > 0 {
			bus.SetActors(actors)
			fmt.Printf("  Actors: %d registered.\n", len(actors))
		}
	}

	fmt.Println("MVGN Telemetry v2 — Initialized.")
	return nil
}

// Run captures events for the session duration.
// In headless mode (no TUI), it runs until interrupted.
func (a *App) Run() error {
	fmt.Println("MVGN Telemetry v2 — Capturing events...")

	// Create telemetry session
	sess := NewSession()
	a.session = sess

	// Set session ID on event bus
	a.eventBus.SetSessionID(sess.ID)

	// Emit session started
	if err := a.eventBus.EmitNew(events.EventSessionStarted,
		events.WithMetadata("project", a.config.ProjectPath),
		events.WithMetadata("session_id", sess.ID),
	); err != nil {
		fmt.Printf("  Warning: could not emit SESSION_STARTED: %v\n", err)
	}

	fmt.Printf("  Session: %s\n", sess.ID[:8])

	// Wait for interrupt or completion
	<-a.ctx.Done()

	return nil
}

// Shutdown performs a graceful close of all subsystems.
func (a *App) Shutdown() error {
	fmt.Println("MVGN Telemetry v2 — Shutting down...")

	// 1. Close session & emit SESSION_CLOSED
	if a.session != nil {
		a.session.Close()
		if err := a.eventBus.EmitNew(events.EventSessionClosed,
			events.WithMetadata("duration_secs", fmt.Sprintf("%d", a.session.Duration)),
			events.WithMetadata("session_id", a.session.ID),
		); err != nil {
			fmt.Printf("  Warning: could not emit SESSION_CLOSED: %v\n", err)
		}
		fmt.Printf("  Session closed: %s (duration: %ds)\n", a.session.ID[:8], a.session.Duration)
	}

	// 2. Close Event Bus
	if a.eventBus != nil {
		if err := a.eventBus.Close(); err != nil {
			fmt.Printf("  Warning: event bus close: %v\n", err)
		}
		fmt.Println("  Event Bus: closed.")
	}

	// 3. Generate session artifact automatically
	if a.session != nil {
		fmt.Println("  Generating session artifact...")
		engine := artifact.NewEngine("events", "reports")
		if err := engine.Generate(a.session.ID); err != nil {
			fmt.Printf("  Warning: artifact generation: %v\n", err)
		} else {
			fmt.Printf("  Artifact generated for session: %s\n", a.session.ID)
		}
	}

	fmt.Println("MVGN Telemetry v2 — Shutdown complete.")
	return nil
}

// Stop signals the app to shut down gracefully.
func (a *App) Stop() {
	if a.cancel != nil {
		a.cancel()
	}
}

// Session represents a telemetry session.
type Session struct {
	ID        string
	StartTime time.Time
	EndTime   time.Time
	Duration  int64
}

// NewSession creates a new session with a unique ID.
func NewSession() *Session {
	return &Session{
		ID:        generateSessionID(),
		StartTime: time.Now().UTC(),
	}
}

// Close marks the session as closed and calculates duration.
func (s *Session) Close() {
	s.EndTime = time.Now().UTC()
	s.Duration = int64(s.EndTime.Sub(s.StartTime).Seconds())
}

// generateSessionID creates a session ID in the format ses-XXXX.
func generateSessionID() string {
	return fmt.Sprintf("ses-%s", time.Now().UTC().Format("20060102-150405"))
}
