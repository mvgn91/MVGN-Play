package consolidator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mvgn/mvgn-scan/internal/models"
	"github.com/mvgn/mvgn-scan/internal/scanner"
	"github.com/mvgn/mvgn-scan/internal/validator"
)

// removeBOM elimina el BOM UTF-8 (0xEF 0xBB 0xBF) del inicio de los datos
func removeBOM(data []byte) []byte {
	return bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
}

// Consolidator maneja la consolidación de datos de proyectos
type Consolidator struct {
	OutputDir string
}

// NewConsolidator crea un nuevo consolidador
func NewConsolidator(outputDir string) *Consolidator {
	return &Consolidator{
		OutputDir: outputDir,
	}
}

// ConsolidateProjects consolida datos de múltiples proyectos
func (c *Consolidator) ConsolidateProjects(projects []scanner.ScanResult) (*models.DashboardData, error) {
	dashboard := &models.DashboardData{
		ScanTimestamp: time.Now(),
		ProjectCount:  len(projects),
		Projects:      []models.ProjectData{},
	}

	for _, project := range projects {
		projectData, err := c.extractProjectData(project)
		if err != nil {
			// Log error pero continuar con otros proyectos
			fmt.Printf("Error procesando %s: %v\n", project.Path, err)
			continue
		}
		dashboard.Projects = append(dashboard.Projects, *projectData)
	}

	return dashboard, nil
}

// extractProjectData extrae datos de un proyecto individual
func (c *Consolidator) extractProjectData(project scanner.ScanResult) (*models.ProjectData, error) {
	data := &models.ProjectData{
		ProjectPath:   project.Path,
		ScanTimestamp: time.Now(),
	}

	// Extraer metadatos estáticos de mvgn-runtime.json
	runtimePath := filepath.Join(project.Path, "mvgn-runtime.json")
	if project.HasRuntime {
		runtimeData, err := c.readRuntimeFile(runtimePath)
		if err != nil {
			return nil, fmt.Errorf("error leyendo runtime: %v", err)
		}
		data.ProjectName = extractProjectName(project.Path)
		data.RuntimeVersion = runtimeData.RuntimeVersion
		data.SchemaVersion = runtimeData.SchemaVersion
		data.Profile = runtimeData.Profile
		data.Actors = runtimeData.Actors
		data.Capabilities = runtimeData.Capabilities
	}

	// Extraer estado dinámico de .mvgn-context.json
	contextPath := filepath.Join(project.Path, ".mvgn-context.json")
	if project.HasContext {
		contextData, err := c.readContextFile(contextPath)
		if err != nil {
			return nil, fmt.Errorf("error leyendo context: %v", err)
		}
		data.ContextSchemaVersion = contextData.SchemaVersion
		data.ProjectID = contextData.ProjectID
		data.Progress = &contextData.Progress
	}

	return data, nil
}

// readRuntimeFile lee y parsea mvgn-runtime.json
func (c *Consolidator) readRuntimeFile(path string) (*models.RuntimeManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	data = removeBOM(data)
	var runtime models.RuntimeManifest
	if err := json.Unmarshal(data, &runtime); err != nil {
		return nil, err
	}

	return &runtime, nil
}

// readContextFile lee y parsea .mvgn-context.json
func (c *Consolidator) readContextFile(path string) (*models.ContextFile, error) {
	// Validar primero
	result := validator.ValidateContextFile(path)
	if !result.Valid {
		return nil, fmt.Errorf("validación fallida: %v", result.Errors)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	data = removeBOM(data)
	var context models.ContextFile
	if err := json.Unmarshal(data, &context); err != nil {
		return nil, err
	}

	return &context, nil
}

// readBridgeFile lee y parsea smg-bridge.json
// extractProjectName extrae el nombre del proyecto de la ruta
func extractProjectName(path string) string {
	return filepath.Base(path)
}

// SaveDashboard guarda el dashboard en disco
func (c *Consolidator) SaveDashboard(dashboard *models.DashboardData) error {
	// Crear directorio si no existe
	if err := os.MkdirAll(c.OutputDir, 0755); err != nil {
		return fmt.Errorf("error creando directorio de salida: %v", err)
	}

	outputPath := filepath.Join(c.OutputDir, "dashboard-data.json")

	data, err := json.MarshalIndent(dashboard, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializando dashboard: %v", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("error escribiendo archivo: %v", err)
	}

	return nil
}