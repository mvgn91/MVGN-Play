package query

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ████████████████████████████████████████████████████████████████████████████
// FileStateSource — Implementación de StateSource que lee state.json del blackboard
// ████████████████████████████████████████████████████████████████████████████

// FileStateSource lee el estado desde .mvgn/blackboard/state.json de un proyecto MVGN.
type FileStateSource struct {
	// BlackboardPath es la ruta al directorio .mvgn/blackboard/
	BlackboardPath string

	// StaleThresholdSecs define después de cuántos segundos el estado se considera stale.
	StaleThresholdSecs int64

	// lastValidState preserva el último estado válido para cuando ocurran errores transitorios.
	lastValidState *ObservedState
}

// blackboardState representa la estructura de state.json en el blackboard.
type blackboardState struct {
	SchemaVersion   string   `json:"schema_version"`
	TaskID          string   `json:"task_id"`
	TaskDescription string   `json:"task_description"`
	Status          string   `json:"status"`
	CurrentPhase    int      `json:"current_phase"`
	CompletedAgents []string `json:"completed_agents"`
	Artifacts       []string `json:"artifacts_generated"`
	LastError       *string  `json:"last_error"`
	Timestamp       string   `json:"timestamp"`
}

// NewFileStateSource crea un nuevo FileStateSource para un proyecto.
func NewFileStateSource(projectPath string, staleThresholdSecs int64) *FileStateSource {
	return &FileStateSource{
		BlackboardPath:     filepath.Join(projectPath, ".mvgn", "blackboard"),
		StaleThresholdSecs: staleThresholdSecs,
	}
}

// Current implementa StateSource.Current.
// Lee state.json, valida el esquema mínimo, calcula frescura y retorna ObservedState normalizado.
func (f *FileStateSource) Current(ctx context.Context) (ObservedState, error) {
	statePath := filepath.Join(f.BlackboardPath, "state.json")

	// P0-07: Manejo de ausencia de blackboard
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		return f.degradedState(
			statePath,
			fmt.Errorf("%w: %s", ErrBlackboardNotFound, statePath),
		), nil
	}

	// Leer archivo
	data, err := os.ReadFile(statePath)
	if err != nil {
		return f.degradedState(
			statePath,
			fmt.Errorf("%w: %s: %v", ErrReadFailed, statePath, err),
		), nil
	}

	// P0-08: Manejo de JSON corrupto
	var raw blackboardState
	if err := json.Unmarshal(data, &raw); err != nil {
		return f.degradedState(
			statePath,
			fmt.Errorf("%w: %s: %v", ErrCorruptJSON, statePath, err),
		), nil
	}

	// P0-02: Validación mínima de esquema
	var validationErrors []string
	if raw.SchemaVersion == "" {
		validationErrors = append(validationErrors, "schema_version vacío")
	}
	if raw.Status == "" {
		validationErrors = append(validationErrors, "status vacío")
	}
	if raw.Timestamp == "" {
		validationErrors = append(validationErrors, "timestamp vacío")
	} else if _, err := time.Parse(time.RFC3339, raw.Timestamp); err != nil {
		validationErrors = append(validationErrors, "timestamp no es RFC3339 válido")
	}
	// Verificar que status sea un valor conocido
	knownStatuses := map[string]bool{"idle": true, "pending": true, "in_progress": true, "completed": true, "failed": true}
	if raw.Status != "" && !knownStatuses[raw.Status] {
		validationErrors = append(validationErrors, "status desconocido: "+raw.Status)
	}
	if len(validationErrors) > 0 {
		errMsg := ""
		for i, ve := range validationErrors {
			if i > 0 {
				errMsg += "; "
			}
			errMsg += ve
		}
		return f.degradedState(
			statePath,
			fmt.Errorf("%w: %s en %s: %s", ErrInvalidSchema, errMsg, statePath, string(data)),
		), nil
	}

	// Mapear estado raw a modelo normalizado
	state := f.normalize(raw, statePath)

	// P0-10: Detección de estado stale
	state = f.checkStaleness(state)

	// Preservar como último estado válido
	f.lastValidState = &state

	return state, nil
}

// normalize mapea el estado crudo del blackboard al modelo interno ObservedState.
func (f *FileStateSource) normalize(raw blackboardState, sourcePath string) ObservedState {
	// P0-03: Modelo interno de estado observado
	state := ObservedState{
		Source:             sourcePath,
		SchemaVersion:      raw.SchemaVersion,
		TaskID:             raw.TaskID,
		TaskDescription:    raw.TaskDescription,
		Phase:              raw.CurrentPhase,
		Agents:             raw.CompletedAgents,
		Artifacts:          raw.Artifacts,
		StaleThresholdSecs: f.StaleThresholdSecs,
		Confidence:         ConfidenceValid,
		Freshness:          FreshnessFresh,
	}

	// Mapear status
	state.State = f.mapStatus(raw.Status)

	// Parsear timestamp
	if raw.Timestamp != "" {
		if t, err := time.Parse(time.RFC3339, raw.Timestamp); err == nil {
			state.UpdatedAt = t
			state.AgeSeconds = int64(time.Since(t).Seconds())
		}
	}

	// Mapear error
	if raw.LastError != nil && *raw.LastError != "" {
		state.LastError = *raw.LastError
	}

	return state
}

// mapStatus mapea el string de estado del blackboard al tipo State normalizado.
func (f *FileStateSource) mapStatus(status string) State {
	switch status {
	case "idle":
		return StateIdle
	case "pending":
		return StatePending
	case "in_progress":
		return StateInProgress
	case "completed":
		return StateCompleted
	case "failed":
		return StateFailed
	default:
		return StateUnknown
	}
}

// checkStaleness determina si el estado está stale basado en su antigüedad.
func (f *FileStateSource) checkStaleness(state ObservedState) ObservedState {
	if f.StaleThresholdSecs <= 0 {
		return state
	}
	if state.AgeSeconds > f.StaleThresholdSecs {
		state.Freshness = FreshnessStale
		state.Confidence = ConfidenceDegraded
	}
	return state
}

// degradedState construye un ObservedState degradado a partir de un error.
// Si existe un lastValidState, lo usa como base. De lo contrario, crea uno desde cero.
func (f *FileStateSource) degradedState(sourcePath string, err error) ObservedState {
	if f.lastValidState != nil {
		// Preservar último estado válido, pero marcar como degradado
		degraded := *f.lastValidState
		degraded.Freshness = FreshnessStale
		degraded.Confidence = ConfidenceDegraded
		degraded.Err = err
		degraded.Source = sourcePath
		degraded.AgeSeconds = int64(time.Since(degraded.UpdatedAt).Seconds())
		return degraded
	}

	return ObservedState{
		State:      StateUnknown,
		Freshness:  FreshnessStale,
		Confidence: ConfidenceInvalid,
		Source:     sourcePath,
		Err:        err,
	}
}
