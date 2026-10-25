package scanner

import (
	"os"
	"path/filepath"
	"strings"
)

// IgnoredDirs contiene directorios que deben ser ignorados
var IgnoredDirs = map[string]bool{
	"node_modules": true,
	"vendor":       true,
	".git":         true,
	"dist":         true,
	"build":        true,
	".dart_tool":   true,
	"tools":        true,
}

// SecretPatterns contiene patrones de archivos que deben ser ignorados
var SecretPatterns = []string{
	".env",
	".pem",
	".key",
}

// ScanResult representa el resultado del escaneo de un directorio
type ScanResult struct {
	HasRuntime bool
	HasContext bool
	Path       string
}

// ScanDirectorio escanea un directorio raíz y retorna proyectos válidos
func ScanDirectorio(rootDir string) ([]ScanResult, error) {
	var results []ScanResult

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Ignorar errores y continuar
		}

		// Ignorar directorios (excepto el raíz, para evitar que "." o ".." se salten todo el árbol)
		if info.IsDir() {
			if path != rootDir && shouldIgnoreDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		// Ignorar patrones de secretos
		if isSecretFile(info.Name()) {
			return nil
		}

		// Detectar pares de archivos MVGN
		if info.Name() == "mvgn-runtime.json" || info.Name() == ".mvgn-context.json" {
			dir := filepath.Dir(path)
			result := checkProjectFiles(dir)
			if result.HasRuntime && result.HasContext {
				// Evitar duplicados
				if !containsResult(results, result.Path) {
					results = append(results, result)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

// shouldIgnoreDir determina si un directorio debe ser ignorado
func shouldIgnoreDir(name string) bool {
	// Ignorar directorios en la lista
	if IgnoredDirs[name] {
		return true
	}

	// Ignorar directorios que empiezan con . (excepto .mvgn que es directorio del framework)
	if strings.HasPrefix(name, ".") && name != ".mvgn" {
		return true
	}

	return false
}

// isSecretFile determina si un archivo es un patrón de secreto
func isSecretFile(name string) bool {
	for _, pattern := range SecretPatterns {
		// Coincidencia exacta o sufijo: .env, file.env, cert.pem, private.key
		if strings.HasSuffix(name, pattern) || strings.HasSuffix(name, pattern[1:]) {
			return true
		}
		// Para patrones con punto, detectar variantes como .env.local, .env.production
		// sin falsos positivos como ".environment"
		if strings.HasPrefix(pattern, ".") &&
			strings.HasPrefix(name, pattern) &&
			(len(name) == len(pattern) || name[len(pattern)] == '.') {
			return true
		}
	}
	return false
}

// checkProjectFiles verifica si un directorio contiene archivos MVGN
func checkProjectFiles(dir string) ScanResult {
	result := ScanResult{
		Path: dir,
	}

	// Verificar mvgn-runtime.json
	runtimePath := filepath.Join(dir, "mvgn-runtime.json")
	if _, err := os.Stat(runtimePath); err == nil {
		result.HasRuntime = true
	}

	// Verificar .mvgn-context.json
	contextPath := filepath.Join(dir, ".mvgn-context.json")
	if _, err := os.Stat(contextPath); err == nil {
		result.HasContext = true
	}

	return result
}

// containsResult verifica si un resultado ya existe
func containsResult(results []ScanResult, path string) bool {
	for _, r := range results {
		if r.Path == path {
			return true
		}
	}
	return false
}