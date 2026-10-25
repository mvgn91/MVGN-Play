package query

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ████████████████████████████████████████████████████████████████████████████
// Fixtures para pruebas (ATENEA criterio #20)
// ████████████████████████████████████████████████████████████████████████████

const (
	fixtureValid = `{
		"schema_version": "5.2.0",
		"task_id": "test-task-001",
		"task_description": "Test task",
		"status": "in_progress",
		"current_phase": 2,
		"agent_sequence": ["buffy", "opencode"],
		"completed_agents": ["buffy"],
		"artifacts_generated": ["docs/test.md"],
		"last_error": null,
		"timestamp": "2026-07-30T18:30:00Z"
	}`

	fixtureCorrupt = `{
		"schema_version": "5.2.0",
		"task_id": "test-task-001",
		"task_description": "Test task,`

	fixtureEmptySchema = `{
		"schema_version": "",
		"status": "idle",
		"timestamp": "2026-07-30T18:30:00Z"
	}`

	fixtureInvalidStatus = `{
		"schema_version": "5.2.0",
		"status": "hyperdrive",
		"timestamp": "2026-07-30T18:30:00Z"
	}`

	fixtureNoTimestamp = `{
		"schema_version": "5.2.0",
		"task_id": "test-task-002",
		"status": "completed",
		"timestamp": ""
	}`

	fixtureFailed = `{
		"schema_version": "5.2.0",
		"task_id": "test-task-003",
		"task_description": "Failed task",
		"status": "failed",
		"current_phase": 1,
		"agent_sequence": ["buffy"],
		"completed_agents": [],
		"artifacts_generated": [],
		"last_error": "Something went wrong",
		"timestamp": "2026-07-30T18:30:00Z"
	}`
)

// setupTestDir crea un directorio temporal con un state.json de prueba.
func setupTestDir(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	blackboardDir := filepath.Join(tmpDir, ".mvgn", "blackboard")
	if err := os.MkdirAll(blackboardDir, 0755); err != nil {
		t.Fatalf("Error creating blackboard dir: %v", err)
	}
	statePath := filepath.Join(blackboardDir, "state.json")
	if err := os.WriteFile(statePath, []byte(content), 0644); err != nil {
		t.Fatalf("Error writing state.json: %v", err)
	}
	return tmpDir
}

// ████████████████████████████████████████████████████████████████████████████
// Tests: Estado válido (P0-01, P0-03)
// ████████████████████████████████████████████████████████████████████████████

func TestFileStateSource_ValidState(t *testing.T) {
	projectPath := setupTestDir(t, fixtureValid)
	source := NewFileStateSource(projectPath, 3600) // 1h threshold

	state, err := source.Current(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if state.Err != nil {
		t.Fatalf("Unexpected state error: %v", state.Err)
	}
	if state.State != StateInProgress {
		t.Errorf("Expected IN_PROGRESS, got %s", state.State)
	}
	if state.SchemaVersion != "5.2.0" {
		t.Errorf("Expected schema 5.2.0, got %s", state.SchemaVersion)
	}
	if state.TaskID != "test-task-001" {
		t.Errorf("Expected task-id test-task-001, got %s", state.TaskID)
	}
	// Note: confidence and freshness depend on age vs threshold.
	// With threshold 3600 and test timestamp ~1.5h ago, state may be stale.
	// We only verify that state was read and parsed correctly.
	if state.Confidence == ConfidenceInvalid {
		t.Errorf("Should not be INVALID for valid input, got %s", state.Confidence)
	}
}

// ████████████████████████████████████████████████████████████████████████████
// Tests: Ausencia de blackboard (P0-07)
// ████████████████████████████████████████████████████████████████████████████

func TestFileStateSource_NoBlackboard(t *testing.T) {
	tmpDir := t.TempDir() // No .mvgn/blackboard/ dir at all
	source := NewFileStateSource(tmpDir, 3600)

	state, err := source.Current(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if state.State != StateUnknown {
		t.Errorf("Expected UNKNOWN state, got %s", state.State)
	}
	if state.Confidence != ConfidenceInvalid {
		t.Errorf("Expected INVALID confidence, got %s", state.Confidence)
	}
	if state.Err == nil {
		t.Fatal("Expected error for missing blackboard")
	}
}

// ████████████████████████████████████████████████████████████████████████████
// Tests: JSON corrupto (P0-08)
// ████████████████████████████████████████████████████████████████████████████

func TestFileStateSource_CorruptJSON(t *testing.T) {
	projectPath := setupTestDir(t, fixtureCorrupt)
	source := NewFileStateSource(projectPath, 3600)

	state, err := source.Current(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// Corrupt JSON causes INVALID confidence (no lastValidState to fall back to)
	if state.Confidence != ConfidenceInvalid {
		t.Errorf("Expected INVALID confidence after corrupt JSON (no prior valid state), got %s", state.Confidence)
	}
	if state.State != StateUnknown {
		t.Errorf("Expected UNKNOWN state for corrupt JSON, got %s", state.State)
	}
}

// ████████████████████████████████████████████████████████████████████████████
// Tests: Schema inválido — versión vacía (P0-02)
// ████████████████████████████████████████████████████████████████████████████

func TestFileStateSource_EmptySchema(t *testing.T) {
	projectPath := setupTestDir(t, fixtureEmptySchema)
	source := NewFileStateSource(projectPath, 3600)

	state, err := source.Current(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// Empty schema causes INVALID (no lastValidState to fall back to)
	if state.Confidence != ConfidenceInvalid {
		t.Errorf("Expected INVALID for empty schema, got %s", state.Confidence)
	}
}

// ████████████████████████████████████████████████████████████████████████████
// Tests: Status desconocido (P0-02)
// ████████████████████████████████████████████████████████████████████████████

func TestFileStateSource_InvalidStatus(t *testing.T) {
	projectPath := setupTestDir(t, fixtureInvalidStatus)
	source := NewFileStateSource(projectPath, 3600)

	state, err := source.Current(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if state.State != StateUnknown {
		t.Errorf("Expected UNKNOWN state for invalid status, got %s", state.State)
	}
	// Invalid status causes INVALID (no lastValidState to fall back to)
	if state.Confidence != ConfidenceInvalid {
		t.Errorf("Expected INVALID for invalid status, got %s", state.Confidence)
	}
}

// ████████████████████████████████████████████████████████████████████████████
// Tests: Estado stale (P0-10)
// ████████████████████████████████████████████████████████████████████████████

func TestFileStateSource_StaleState(t *testing.T) {
	// Usar un threshold bajo para forzar stale
	projectPath := setupTestDir(t, fixtureValid)
	source := NewFileStateSource(projectPath, 1) // 1 second threshold

	// Esperar a que pase el threshold
	time.Sleep(2 * time.Second)

	state, err := source.Current(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if state.Freshness != FreshnessStale {
		t.Errorf("Expected STALE, got %s", state.Freshness)
	}
	if state.Confidence != ConfidenceDegraded {
		t.Errorf("Expected DEGRADED for stale state, got %s", state.Confidence)
	}
}

// ████████████████████████████████████████████████████████████████████████████
// Tests: Estado failed preserva error (P0-09)
// ████████████████████████████████████████████████████████████████████████████

func TestFileStateSource_FailedState(t *testing.T) {
	projectPath := setupTestDir(t, fixtureFailed)
	source := NewFileStateSource(projectPath, 3600)

	state, err := source.Current(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if state.State != StateFailed {
		t.Errorf("Expected FAILED, got %s", state.State)
	}
	if state.LastError != "Something went wrong" {
		t.Errorf("Expected 'Something went wrong', got '%s'", state.LastError)
	}
	if state.Phase != 1 {
		t.Errorf("Expected phase 1, got %d", state.Phase)
	}
}

// ████████████████████████████████████████████████████████████████████████████
// Tests: Sin timestamp (P0-09 edge case)
// ████████████████████████████████████████████████████████████████████████████

func TestFileStateSource_NoTimestamp(t *testing.T) {
	projectPath := setupTestDir(t, fixtureNoTimestamp)
	source := NewFileStateSource(projectPath, 3600)

	state, err := source.Current(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// Empty timestamp fails RFC3339 validation → INVALID schema → UNKNOWN
	if state.State != StateUnknown {
		t.Errorf("Expected UNKNOWN for empty timestamp (schema validation fails), got %s", state.State)
	}
	if state.Confidence != ConfidenceInvalid {
		t.Errorf("Expected INVALID for invalid schema, got %s", state.Confidence)
	}
}
