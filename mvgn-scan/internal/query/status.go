package query

// ████████████████████████████████████████████████████████████████████████████
// Status Aggregation — Regla de agregación multi-agente (ATENEA §4.3)
// ████████████████████████████████████████████████████████████████████████████

// AggregatedState representa el estado agregado de múltiples agentes o componentes.
type AggregatedState struct {
	// State es el estado resultante después de aplicar la regla de agregación.
	State State `json:"state"`

	// Semaphore es el símbolo del semáforo correspondiente.
	Semaphore Semaphore `json:"semaphore"`

	// Description es una descripción textual del estado agregado.
	Description string `json:"description"`
}

// AggregateStates aplica la regla de agregación multi-agente definida por ATENEA.
//
// Regla:
//  1. Si algún componente está FAILED       → FAILED 🔴
//  2. Si no hay FAILED pero hay IN_PROGRESS → IN_PROGRESS 🟡
//  3. Si no hay FAILED ni IN_PROGRESS pero hay PENDING → PENDING 🔵
//  4. Si todos están COMPLETED              → COMPLETED 🟢
//  5. Si no hay actividad                   → IDLE ⚪
//  6. Si no puede determinarse             → UNKNOWN ❓
func AggregateStates(states []State) AggregatedState {
	if len(states) == 0 {
		return AggregatedState{
			State:       StateIdle,
			Semaphore:   SemaphoreIdle,
			Description: "No active agents. System idle.",
		}
	}

	hasFailed := false
	hasInProgress := false
	hasPending := false
	allCompleted := true

	for _, s := range states {
		switch s {
		case StateFailed:
			hasFailed = true
			allCompleted = false
		case StateInProgress:
			hasInProgress = true
			allCompleted = false
		case StatePending:
			hasPending = true
			allCompleted = false
		case StateUnknown:
			allCompleted = false
		}
	}

	// Regla 1: FAILED domina
	if hasFailed {
		return AggregatedState{
			State:       StateFailed,
			Semaphore:   SemaphoreFailed,
			Description: "One or more agents have failed. Requires intervention.",
		}
	}

	// Regla 2: IN_PROGRESS
	if hasInProgress {
		return AggregatedState{
			State:       StateInProgress,
			Semaphore:   SemaphoreInProgress,
			Description: "One or more agents are actively working.",
		}
	}

	// Regla 3: PENDING
	if hasPending {
		return AggregatedState{
			State:       StatePending,
			Semaphore:   SemaphorePending,
			Description: "Tasks are pending execution.",
		}
	}

	// Regla 4: COMPLETED
	if allCompleted {
		return AggregatedState{
			State:       StateCompleted,
			Semaphore:   SemaphoreCompleted,
			Description: "All agents have completed their tasks.",
		}
	}

	// Regla 6: UNKNOWN
	return AggregatedState{
		State:       StateUnknown,
		Semaphore:   SemaphoreUnknown,
		Description: "State could not be determined.",
	}
}

// StateDescription devuelve una descripción textual para un estado individual.
func StateDescription(s State) string {
	switch s {
	case StateIdle:
		return "System idle. No active task."
	case StatePending:
		return "Task pending execution."
	case StateInProgress:
		return "Task in progress."
	case StateCompleted:
		return "Task completed successfully."
	case StateFailed:
		return "Task failed. Recovery required."
	case StateUnknown:
		return "State could not be determined."
	default:
		return "Unknown state."
	}
}

// IsTerminal indica si un estado es terminal (no se esperan más transiciones sin intervención).
func IsTerminal(s State) bool {
	return s == StateCompleted || s == StateFailed
}
