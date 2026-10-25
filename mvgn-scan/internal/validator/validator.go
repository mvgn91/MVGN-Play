package validator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mvgn/mvgn-scan/internal/models"
)

// removeBOM elimina el BOM UTF-8 (0xEF 0xBB 0xBF) del inicio de los datos
func removeBOM(data []byte) []byte {
	return bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
}

// ValidationResult representa el resultado de una validación
type ValidationResult struct {
	Valid   bool
	Errors  []string
	Warnings []string
}

// ValidateContextFile valida .mvgn-context.json contra Schema v3.0
func ValidateContextFile(path string) ValidationResult {
	result := ValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Leer archivo
	data, err := os.ReadFile(path)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("No se pudo leer el archivo: %v", err))
		return result
	}

	// Remover BOM UTF-8 y parsear JSON
	data = removeBOM(data)
	var context models.ContextFile
	if err := json.Unmarshal(data, &context); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("JSON inválido: %v", err))
		return result
	}

	// Validar schema_version
	if context.SchemaVersion == "" {
		result.Errors = append(result.Errors, "Campo 'schema_version' requerido")
		result.Valid = false
	} else if context.SchemaVersion != "3.0.0" {
		result.Warnings = append(result.Warnings, fmt.Sprintf("schema_version inesperado: %s", context.SchemaVersion))
	}

	// Validar project_id
	if context.ProjectID == "" {
		result.Errors = append(result.Errors, "Campo 'project_id' requerido")
		result.Valid = false
	} else if !strings.HasPrefix(context.ProjectID, "PRD-") {
		result.Warnings = append(result.Warnings, fmt.Sprintf("project_id formato inusual: %s", context.ProjectID))
	}

	// Validar session
	if context.Session.ID == "" {
		result.Errors = append(result.Errors, "Campo 'session.id' requerido")
		result.Valid = false
	}
	if context.Session.Agent == "" {
		result.Errors = append(result.Errors, "Campo 'session.agent' requerido")
		result.Valid = false
	}

	// Validar progress
	if context.Progress.TotalTasks < 0 {
		result.Errors = append(result.Errors, "progress.total_tasks no puede ser negativo")
		result.Valid = false
	}
	if context.Progress.CompletedTasks > context.Progress.TotalTasks {
		result.Warnings = append(result.Warnings, "progress.completed_tasks > total_tasks")
	}

	// Validar tasks
	if len(context.Tasks) == 0 {
		result.Warnings = append(result.Warnings, "No hay tareas definidas")
	}

	// Verificar campos prohibidos (usar data limpio de BOM)
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err == nil {
		prohibited := []string{"time_invested_hours", "estimated_hours", "duration"}
		for _, field := range prohibited {
			if _, exists := raw[field]; exists {
				result.Errors = append(result.Errors, fmt.Sprintf("Campo PROHIBIDO detectado: %s", field))
				result.Valid = false
			}
		}
	}

	return result
}

// ValidateRuntimeFile valida mvgn-runtime.json
func ValidateRuntimeFile(path string) ValidationResult {
	result := ValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Leer archivo
	data, err := os.ReadFile(path)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("No se pudo leer el archivo: %v", err))
		return result
	}

	// Remover BOM UTF-8 y parsear JSON
	data = removeBOM(data)
	var runtime models.RuntimeManifest
	if err := json.Unmarshal(data, &runtime); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("JSON inválido: %v", err))
		return result
	}

	// Validar campos requeridos
	if runtime.RuntimeVersion == "" {
		result.Errors = append(result.Errors, "Campo 'runtime_version' requerido")
		result.Valid = false
	}
	if runtime.SchemaVersion == "" {
		result.Errors = append(result.Errors, "Campo 'schema_version' requerido")
		result.Valid = false
	}

	// Validar schema_version
	if runtime.SchemaVersion != "" {
		major, _ := strconv.Atoi(strings.Split(runtime.SchemaVersion, ".")[0])
		if major < 3 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("schema_version < 3.0.0: %s", runtime.SchemaVersion))
		}
	}

	return result
}