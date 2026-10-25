package query

import "testing"

// ████████████████████████████████████████████████████████████████████████████
// Tests: SemaphoreForState
// ████████████████████████████████████████████████████████████████████████████

func TestSemaphoreForState(t *testing.T) {
	tests := []struct {
		state    State
		expected Semaphore
	}{
		{StateIdle, SemaphoreIdle},
		{StatePending, SemaphorePending},
		{StateInProgress, SemaphoreInProgress},
		{StateCompleted, SemaphoreCompleted},
		{StateFailed, SemaphoreFailed},
		{StateUnknown, SemaphoreUnknown},
		{State("WEIRDO"), SemaphoreUnknown},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			got := SemaphoreForState(tt.state)
			if got != tt.expected {
				t.Errorf("SemaphoreForState(%s) = %s, want %s", tt.state, got, tt.expected)
			}
		})
	}
}

// ████████████████████████████████████████████████████████████████████████████
// Tests: AggregateStates — Regla de agregación multi-agente (ATENEA §4.3)
// ████████████████████████████████████████████████████████████████████████████

func TestAggregateStates_Empty(t *testing.T) {
	result := AggregateStates(nil)
	if result.State != StateIdle {
		t.Errorf("Expected IDLE for empty, got %s", result.State)
	}
}

func TestAggregateStates_AllCompleted(t *testing.T) {
	result := AggregateStates([]State{StateCompleted, StateCompleted})
	if result.State != StateCompleted {
		t.Errorf("Expected COMPLETED, got %s", result.State)
	}
}

func TestAggregateStates_FailedDominates(t *testing.T) {
	result := AggregateStates([]State{
		StateCompleted, StateInProgress, StateFailed,
	})
	if result.State != StateFailed {
		t.Errorf("Expected FAILED to dominate, got %s", result.State)
	}
}

func TestAggregateStates_InProgressDominatesPending(t *testing.T) {
	result := AggregateStates([]State{
		StateCompleted, StatePending, StateInProgress,
	})
	if result.State != StateInProgress {
		t.Errorf("Expected IN_PROGRESS to dominate PENDING, got %s", result.State)
	}
}

func TestAggregateStates_PendingWhenNoProgress(t *testing.T) {
	result := AggregateStates([]State{StatePending, StateIdle})
	if result.State != StatePending {
		t.Errorf("Expected PENDING, got %s", result.State)
	}
}

func TestAggregateStates_FreshnessChain(t *testing.T) {
	// Regla 1: FAILED > IN_PROGRESS > PENDING > COMPLETED > IDLE
	tests := []struct {
		name     string
		states   []State
		expected State
	}{
		{"failed+in_progress", []State{StateFailed, StateInProgress}, StateFailed},
		{"failed+completed", []State{StateFailed, StateCompleted}, StateFailed},
		{"in_progress+pending", []State{StateInProgress, StatePending}, StateInProgress},
		{"in_progress+completed", []State{StateInProgress, StateCompleted}, StateInProgress},
		{"pending+completed", []State{StatePending, StateCompleted}, StatePending},
		{"pending+idle", []State{StatePending, StateIdle}, StatePending},
		{"completed+idle", []State{StateCompleted, StateIdle}, StateCompleted},
		{"unknown+completed", []State{StateUnknown, StateCompleted}, StateUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AggregateStates(tt.states)
			if result.State != tt.expected {
				t.Errorf("AggregateStates(%v) = %s, want %s", tt.states, result.State, tt.expected)
			}
		})
	}
}

// ████████████████████████████████████████████████████████████████████████████
// Tests: IsTerminal
// ████████████████████████████████████████████████████████████████████████████

func TestIsTerminal(t *testing.T) {
	tests := []struct {
		state    State
		terminal bool
	}{
		{StateIdle, false},
		{StatePending, false},
		{StateInProgress, false},
		{StateCompleted, true},
		{StateFailed, true},
		{StateUnknown, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			got := IsTerminal(tt.state)
			if got != tt.terminal {
				t.Errorf("IsTerminal(%s) = %v, want %v", tt.state, got, tt.terminal)
			}
		})
	}
}

// ████████████████████████████████████████████████████████████████████████████
// Tests: StateDescription
// ████████████████████████████████████████████████████████████████████████████

func TestStateDescription(t *testing.T) {
	tests := []struct {
		state State
		want  string
	}{
		{StateIdle, "System idle"},
		{StatePending, "Task pending"},
		{StateInProgress, "Task in progress"},
		{StateCompleted, "Task completed"},
		{StateFailed, "Task failed"},
		{StateUnknown, "could not be determined"},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			got := StateDescription(tt.state)
			if len(got) == 0 {
				t.Errorf("StateDescription(%s) returned empty string", tt.state)
			}
		})
	}
}
