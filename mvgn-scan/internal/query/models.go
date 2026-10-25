package query

import (
	"context"
	"fmt"
	"time"
)

// ████████████████████████████████████████████████████████████████████████████
// StateSource — Contrato de fuente de estado (desacopla origen de presentación)
// ████████████████████████████████████████████████████████████████████████████

// StateSource define el contrato para obtener el estado observado de un proyecto MVGN.
type StateSource interface {
	// Current obtiene el estado actual del proyecto.
	Current(ctx context.Context) (ObservedState, error)
}

// ████████████████████████████████████████████████████████████████████████████
// ObservedState — Modelo interno de estado observado normalizado
// ████████████████████████████████████████████████████████████████████████████

// ObservedState representa el estado normalizado de un proyecto MVGN.
type ObservedState struct {
	// State es el estado operativo del proyecto.
	State State `json:"state"`

	// Freshness indica si el estado está fresco o desactualizado.
	Freshness Freshness `json:"freshness"`

	// Confidence indica el nivel de confianza en el estado reportado.
	Confidence Confidence `json:"confidence"`

	// Source es la ruta del archivo del cual se leyó el estado.
	Source string `json:"source"`

	// UpdatedAt es el timestamp de la última actualización del estado.
	UpdatedAt time.Time `json:"updated_at"`

	// AgeSeconds es la antigüedad del estado en segundos.
	AgeSeconds int64 `json:"age_seconds"`

	// StaleThresholdSecs es el umbral en segundos para considerar el estado como stale.
	StaleThresholdSecs int64 `json:"stale_threshold_secs"`

	// TaskID es el identificador de la tarea activa o última tarea.
	TaskID string `json:"task_id"`

	// TaskDescription es la descripción de la tarea activa o última tarea.
	TaskDescription string `json:"task_description"`

	// Phase es la fase actual del proyecto.
	Phase int `json:"phase"`

	// Agents es la lista de agentes que han completado fases.
	Agents []string `json:"agents"`

	// Artifacts es la lista de artefactos generados.
	Artifacts []string `json:"artifacts"`

	// LastError es el último error registrado (puede estar vacío).
	LastError string `json:"last_error"`

	// SchemaVersion es la versión del schema de state.json.
	SchemaVersion string `json:"schema_version"`

	// Err contiene el error operativo si ocurrió uno durante la lectura.
	Err error `json:"-"`
}

// ████████████████████████████████████████████████████████████████████████████
// Enums
// ████████████████████████████████████████████████████████████████████████████

// State representa los posibles estados operativos de un proyecto MVGN.
type State string

const (
	StateIdle        State = "IDLE"
	StatePending     State = "PENDING"
	StateInProgress  State = "IN_PROGRESS"
	StateCompleted   State = "COMPLETED"
	StateFailed      State = "FAILED"
	StateUnknown     State = "UNKNOWN"
)

// Freshness indica si el estado está fresco o desactualizado.
type Freshness string

const (
	FreshnessFresh  Freshness = "FRESH"
	FreshnessStale  Freshness = "STALE"
)

// Confidence indica el nivel de confianza en el estado reportado.
type Confidence string

const (
	ConfidenceValid    Confidence = "VALID"
	ConfidenceDegraded Confidence = "DEGRADED"
	ConfidenceInvalid  Confidence = "INVALID"
)

// Semaphore representa el color del semáforo asociado a un estado.
type Semaphore string

const (
	SemaphoreIdle       Semaphore = "⚪"
	SemaphorePending    Semaphore = "🔵"
	SemaphoreInProgress Semaphore = "🟡"
	SemaphoreCompleted  Semaphore = "🟢"
	SemaphoreFailed     Semaphore = "🔴"
	SemaphoreUnknown    Semaphore = "❓"
)

// StringMap devuelve el estado como mapa para renderizado.
func (s ObservedState) StringMap() map[string]string {
	stateLabel := string(s.State)
	semaforo := string(SemaphoreForState(s.State))

	ageStr := "—"
	if s.AgeSeconds >= 0 {
		ageStr = formatAge(s.AgeSeconds)
	}

	lastUpdate := "—"
	if !s.UpdatedAt.IsZero() {
		lastUpdate = s.UpdatedAt.Format("2006-01-02T15:04:05Z")
	}

	errorStr := s.LastError
	if errorStr == "" && s.Err != nil {
		errorStr = s.Err.Error()
	}
	if errorStr == "" {
		errorStr = "—"
	}

	return map[string]string{
		"IDENTIFIER":   "MVGN-QRY-001 · QUERY DASHBOARD",
		"STATE":        semaforo + " " + stateLabel,
		"SEMAPHORE":    semaforo,
		"STATE_LABEL":  stateLabel,
		"FRESHNESS":    string(s.Freshness),
		"CONFIDENCE":   string(s.Confidence),
		"SOURCE":       s.Source,
		"LAST_UPDATE":  lastUpdate,
		"AGE":          ageStr,
		"TASK":         s.TaskDescription,
		"TASK_ID":      s.TaskID,
		"PHASE":        formatPhase(s.Phase),
		"AGENTS":       formatAgents(s.Agents),
		"ARTIFACTS":    formatCount(len(s.Artifacts)),
		"ERROR":        errorStr,
		"SCHEMA":       s.SchemaVersion,
	}
}

// ████████████████████████████████████████████████████████████████████████████
// Helpers
// ████████████████████████████████████████████████████████████████████████████

// SemaphoreForState devuelve el semáforo correspondiente a un estado.
func SemaphoreForState(s State) Semaphore {
	switch s {
	case StateIdle:
		return SemaphoreIdle
	case StatePending:
		return SemaphorePending
	case StateInProgress:
		return SemaphoreInProgress
	case StateCompleted:
		return SemaphoreCompleted
	case StateFailed:
		return SemaphoreFailed
	default:
		return SemaphoreUnknown
	}
}

func formatAge(seconds int64) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	} else if seconds < 3600 {
		return fmt.Sprintf("%dm", seconds/60)
	}
	return fmt.Sprintf("%dh", seconds/3600)
}

func formatPhase(phase int) string {
	if phase == 0 {
		return "—"
	}
	return fmt.Sprintf("Phase %d/3", phase)
}

func formatAgents(agents []string) string {
	if len(agents) == 0 {
		return "—"
	}
	return fmt.Sprintf("%d active", len(agents))
}

func formatCount(n int) string {
	if n == 0 {
		return "—"
	}
	return fmt.Sprintf("%d generated", n)
}
