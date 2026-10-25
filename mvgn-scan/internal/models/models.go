package models

import "time"

// RuntimeManifest representa mvgn-runtime.json
type RuntimeManifest struct {
	RuntimeVersion string            `json:"runtime_version"`
	SchemaVersion  string            `json:"schema_version"`
	Profile        string            `json:"profile"`
	Capabilities   map[string]bool   `json:"capabilities"`
	Actors         []Actor           `json:"actors"`
	Skills         []interface{}     `json:"skills"`
	Policies       map[string]bool   `json:"policies"`
	Adapters       []interface{}     `json:"adapters"`
}

// Actor representa un agente registrado
type Actor struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Model    string `json:"model"`
	Provider string `json:"provider"`
}

// ContextFile representa .mvgn-context.json (Schema v3.0)
type ContextFile struct {
	SchemaVersion string      `json:"schema_version"`
	ProjectID     string      `json:"project_id"`
	Session       Session     `json:"session"`
	Progress      Progress    `json:"progress"`
	Tasks         []Task      `json:"tasks"`
	Evaluation    Evaluation  `json:"evaluation"`
}

// Session representa la sesión activa
type Session struct {
	ID        string     `json:"id"`
	Agent     string     `json:"agent"`
	Model     string     `json:"model"`
	Provider  string     `json:"provider"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
}

// Progress representa el progreso del proyecto
type Progress struct {
	TotalTasks     int    `json:"total_tasks"`
	CompletedTasks int    `json:"completed_tasks"`
	CurrentTask    string `json:"current_task,omitempty"`
}

// Task representa una tarea
type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Status      string     `json:"status"`
	AssignedTo  string     `json:"assigned_to"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	EndedAt     *time.Time `json:"ended_at,omitempty"`
}

// Evaluation representa la evaluación del proyecto
type Evaluation struct {
	Score    *int   `json:"score,omitempty"`
	Reviewer string `json:"reviewer,omitempty"`
	Notes    string `json:"notes"`
}

// ProjectData representa los datos consolidados de un proyecto
type ProjectData struct {
	// Metadatos estáticos (de mvgn-runtime.json)
	ProjectName    string            `json:"project_name"`
	Stack          string            `json:"stack,omitempty"`
	RuntimeVersion string            `json:"runtime_version"`
	SchemaVersion  string            `json:"schema_version"`
	Profile        string            `json:"profile"`
	Actors         []Actor           `json:"actors"`
	Capabilities   map[string]bool   `json:"capabilities"`

	// Estado dinámico (de .mvgn-context.json)
	ContextSchemaVersion string    `json:"context_schema_version,omitempty"`
	ProjectID            string    `json:"project_id,omitempty"`
	State                string    `json:"state,omitempty"`
	Progress             *Progress `json:"progress,omitempty"`

	// ATENEA governance conditions
	AteneaConditions *AteneaConditions `json:"atenea_conditions,omitempty"`

	// Metadata de escaneo
	ScanTimestamp time.Time `json:"scan_timestamp"`
	ProjectPath   string    `json:"project_path"`
}

// AteneaConditions representa las condiciones evaluadas por ATENEA Governance & QA
type AteneaConditions struct {
	ContextExists    bool `json:"context_exists"`
	JoinExists       bool `json:"join_exists"`
	HandoffExists    bool `json:"handoff_exists"`
	TimestampFresh   bool `json:"timestamp_fresh"`
	ArtifactsExist   bool `json:"artifacts_exist"`
	HasAgentSequence bool `json:"has_agent_sequence"`
	SchemaCurrent    bool `json:"schema_current"`
	NoErrors         bool `json:"no_errors"`
	PhaseDefined     bool `json:"phase_defined"`
}

// BlackboardState representa .mvgn/blackboard/state.json
type BlackboardState struct {
	SchemaVersion    string           `json:"schema_version"`
	TaskID           string           `json:"task_id"`
	TaskDescription  string           `json:"task_description"`
	Status           string           `json:"status"`
	CurrentPhase     int              `json:"current_phase"`
	AgentSequence    []string         `json:"agent_sequence,omitempty"`
	CompletedAgents  []string         `json:"completed_agents"`
	Artifacts        []string         `json:"artifacts_generated"`
	LastError        *string          `json:"last_error"`
	Timestamp        string           `json:"timestamp"`
	NextAction       string           `json:"next_action"`
	AteneaConditions *AteneaConditions `json:"atenea_conditions,omitempty"`
}

// DashboardData representa el archivo de salida consolidado
type DashboardData struct {
	ScanTimestamp time.Time     `json:"scan_timestamp"`
	ProjectCount  int           `json:"project_count"`
	Projects      []ProjectData `json:"projects"`
}