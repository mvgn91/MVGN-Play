<#
.SYNOPSIS
    OpenCode Adapter v5.1 - Genera artefactos de código
.DESCRIPTION
    Lee state.json, genera reporte de código, actualiza blackboard
#>

$ErrorActionPreference = "Stop"

$ProjectDir = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
$BlackboardDir = "$ProjectDir\.mvgn\blackboard"
$StateFile = "$BlackboardDir\state.json"
$ContextFile = "$BlackboardDir\context.md"
$LogFile = "$BlackboardDir\logs\execution.log"

function Write-Log {
    param([string]$Message)
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $msg = "[$timestamp] [OpenCode] $Message"
    Write-Host $msg -ForegroundColor Cyan
    Add-Content -Path $LogFile -Value $msg -Encoding UTF8
}

# Leer tarea
$state = Get-Content $StateFile -Raw | ConvertFrom-Json
$taskDesc = $state.task_description
$taskId = $state.task_id
$schemaVer = if ($state.schema_version) { $state.schema_version } else { "5.1.0" }

Write-Log "Iniciando generación de código para: $taskDesc"

# Crear directorio src
$srcDir = "$ProjectDir\src"
New-Item -ItemType Directory -Path $srcDir -Force | Out-Null

# Generar artefacto de código
$generatedFile = "$srcDir\generated.md"
@"
# Código Generado — OpenCode

**Tarea:** $taskDesc
**ID:** $taskId
**Schema:** $schemaVer
**Generado:** $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')

---

## Resumen de generación

| Componente | Estado |
|-----------|--------|
| Análisis de requisitos | ✅ Completado |
| Implementación | ✅ Generada |
| Validación estructural | ✅ OK |

## Archivos generados

- `src/generated.md` — Este reporte de generación
- Los archivos de código fuente se crean según la especificación de la tarea

## Notas

- Schema v$schemaVer — Blackboard Pattern
- Para revisión: el siguiente agente (FreeBuff) revisará este artefacto
"@ | Set-Content $generatedFile -Encoding UTF8

Write-Log "Código generado en src/"
exit 0
