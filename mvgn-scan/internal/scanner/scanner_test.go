package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestShouldIgnoreDir(t *testing.T) {
	tests := []struct {
		name     string
		dirName  string
		expected bool
	}{
		{"Ignorar node_modules", "node_modules", true},
		{"Ignorar vendor", "vendor", true},
		{"Ignorar .git", ".git", true},
		{"Ignorar dist", "dist", true},
		{"Ignorar build", "build", true},
		{"Ignorar .dart_tool", ".dart_tool", true},
		{"Ignorar tools", "tools", true},
		{"No ignorar src", "src", false},
		{"No ignorar .mvgn", ".mvgn", false},
		{"Ignorar otro punto", ".env", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldIgnoreDir(tt.dirName)
			if result != tt.expected {
				t.Errorf("shouldIgnoreDir(%s) = %v, want %v", tt.dirName, result, tt.expected)
			}
		})
	}
}

func TestIsSecretFile(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		expected bool
	}{
		{"Detectar .env", ".env", true},
		{"Detectar .env.local", ".env.local", true},
		{"Detectar key.pem", "key.pem", true},
		{"Detectar cert.pem", "cert.pem", true},
		{"Detectar private.key", "private.key", true},
		{"No detectar main.go", "main.go", false},
		{"No detectar README.md", "README.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSecretFile(tt.fileName)
			if result != tt.expected {
				t.Errorf("isSecretFile(%s) = %v, want %v", tt.fileName, result, tt.expected)
			}
		})
	}
}

func TestScanDirectorio(t *testing.T) {
	// Crear directorio temporal
	tmpDir, err := os.MkdirTemp("", "mvgn-scan-test")
	if err != nil {
		t.Fatalf("Error creando directorio temporal: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Crear proyecto válido
	projectDir := filepath.Join(tmpDir, "test-project")
	os.MkdirAll(projectDir, 0755)

	// Crear archivos de proyecto
	os.WriteFile(filepath.Join(projectDir, "mvgn-runtime.json"), []byte(`{"runtime_version": "4.2.0", "schema_version": "3.0.0"}`), 0644)
	os.WriteFile(filepath.Join(projectDir, ".mvgn-context.json"), []byte(`{"schema_version": "3.0.0", "project_id": "PRD-TEST"}`), 0644)

	// Crear directorio ignorado
	os.MkdirAll(filepath.Join(tmpDir, "node_modules", "package"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "node_modules", "package", "index.js"), []byte(`console.log("test")`), 0644)

	// Ejecutar escaneo
	results, err := ScanDirectorio(tmpDir)
	if err != nil {
		t.Fatalf("Error escaneando: %v", err)
	}

	// Verificar resultados
	if len(results) != 1 {
		t.Errorf("Esperaba 1 proyecto, obtuve %d", len(results))
	}

	if len(results) > 0 && results[0].Path != projectDir {
		t.Errorf("Proyecto incorrecto: %s", results[0].Path)
	}
}

func TestContainsResult(t *testing.T) {
	results := []ScanResult{
		{Path: "/path/to/project1"},
		{Path: "/path/to/project2"},
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"Encontrar project1", "/path/to/project1", true},
		{"Encontrar project2", "/path/to/project2", true},
		{"No encontrar project3", "/path/to/project3", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsResult(results, tt.path)
			if result != tt.expected {
				t.Errorf("containsResult(%s) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}