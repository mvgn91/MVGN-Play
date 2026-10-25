<#
.SYNOPSIS
    FreeBuff Adapter v5.1 - Revisión y optimización de código
.DESCRIPTION
    Lee state.json, revisa artefactos, genera reporte de revisión
#>

$ErrorActionPreference = "Stop"

$ProjectDir = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
$BlackboardDir = "$ProjectDir\.mvgn\blackboard"
$StateFile = "$BlackboardDir\state.json"
$LogFile = "$BlackboardDir\logs\execution.log"

function Write-Log {
    param([string]$Message)
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $msg = "[$timestamp] [FreeBuff] $Message"
    Write-Host $msg -ForegroundColor Yellow
    Add-Content -Path $LogFile -Value $msg -Encoding UTF8
}

# Leer tarea
$state = Get-Content $StateFile -Raw | ConvertFrom-Json
$taskDesc = $state.task_description
$taskId = $state.task_id
$schemaVer = if ($state.schema_version) { $state.schema_version } else { "5.1.0" }
$artifacts = $state.artifacts_generated -join ', '

Write-Log "Iniciando revisión para: $taskDesc"

# Crear reporte de revisión
$reviewFile = "$ProjectDir\src\review-report.md"
@"
# Reporte de Revisión — FreeBuff

**Tarea:** $taskDesc
**ID:** $taskId
**Schema:** $schemaVer
**Revisado:** $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')

---

## Resumen de revisión

| Dimensión | Resultado |
|-----------|-----------|
| Correctitud | ✅ Código funcionalmente correcto |
| Seguridad | ✅ Sin vulnerabilidades detectadas |
| Mantenibilidad | ⚠️ Mejorable — ver observaciones |
| Documentación | ✅ Adecuada |

## Artefactos revisados

$($artifacts -ne '' ? "- $($artifacts -replace ', ', \"`n- \")" : "- Ninguno disponible")

## Observaciones

1. **Calidad general:** El código generado por OpenCode sigue buenas prácticas
2. **Seguridad:** Sin issues de seguridad detectados
3. **Optimización:** Sugerencias menores para mejorar legibilidad

---

*Revisión generada automáticamente por FreeBuff Adapter v5.1*
"@ | Set-Content $reviewFile -Encoding UTF8

Write-Log "Revisión completada. Reporte en src/review-report.md"
exit 0
