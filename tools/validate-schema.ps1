<#
.SYNOPSIS
    Valida .mvgn-context.json contra Schema v3.0.
.DESCRIPTION
    Verifica que el archivo cumpla con el Schema v3.0 estricto.
    Detecta campos prohibidos como time_invested_hours.
.PARAMETER Path
    Ruta al archivo .mvgn-context.json (default: .mvgn-context.json en directorio actual).
.EXAMPLE
    .\validate-schema.ps1
    .\validate-schema.ps1 -Path "C:\proyecto\.mvgn-context.json"
#>

param (
    [string]$Path = ".mvgn-context.json"
)

Write-Host ""
Write-Host "MVGN Schema v3.0 Validator" -ForegroundColor Cyan
Write-Host "==========================" -ForegroundColor Cyan
Write-Host ""

# Verificar que el archivo existe
if (-not (Test-Path $Path)) {
    Write-Host "ERROR: Archivo no encontrado: $Path" -ForegroundColor Red
    exit 1
}

Write-Host "Validando: $Path" -ForegroundColor Cyan
Write-Host ""

# Leer y parsear JSON
try {
    $json = Get-Content $Path -Raw | ConvertFrom-Json
    Write-Host "  OK: JSON valido (sintaxis correcta)" -ForegroundColor Green
} catch {
    Write-Host "  ERROR: JSON invalido: $_" -ForegroundColor Red
    exit 1
}

# Array de errores
$errors = @()
$warnings = @()

# === VALIDACIONES OBLIGATORIAS ===

# 1. schema_version
if (-not $json.schema_version) {
    $errors += "Campo 'schema_version' requerido"
} elseif ($json.schema_version -ne "3.0.0") {
    $errors += "Campo 'schema_version' debe ser '3.0.0' (actual: $($json.schema_version))"
} else {
    Write-Host "  OK: schema_version = 3.0.0" -ForegroundColor Green
}

# 2. project_id
if (-not $json.project_id) {
    $errors += "Campo 'project_id' requerido"
} elseif ($json.project_id -notmatch "^PRD-.+") {
    $warnings += "project_id no tiene formato esperado 'PRD-XXX-XXX' (actual: $($json.project_id))"
    Write-Host "  WARN: project_id formato inusual" -ForegroundColor Yellow
} else {
    Write-Host "  OK: project_id = $($json.project_id)" -ForegroundColor Green
}

# 3. session
if (-not $json.session) {
    $errors += "Campo 'session' requerido"
} else {
    $sessionFields = @("id", "agent", "model", "provider", "started_at")
    foreach ($field in $sessionFields) {
        if (-not $json.session.$field) {
            $errors += "Campo 'session.$field' requerido"
        }
    }
    if ($json.session.started_at) {
        try {
            $null = [DateTime]::Parse($json.session.started_at)
            Write-Host "  OK: session.started_at formato ISO8601 valido" -ForegroundColor Green
        } catch {
            $errors += "session.started_at no es formato ISO8601 valido"
        }
    }
    if ($json.session.ended_at -and $json.session.ended_at -ne "null") {
        try {
            $null = [DateTime]::Parse($json.session.ended_at)
            Write-Host "  OK: session.ended_at formato ISO8601 valido" -ForegroundColor Green
        } catch {
            $errors += "session.ended_at no es formato ISO8601 valido"
        }
    }
}

# 4. progress
if (-not $json.progress) {
    $errors += "Campo 'progress' requerido"
} else {
    $progressFields = @("total_tasks", "completed_tasks")
    foreach ($field in $progressFields) {
        if ($null -eq $json.progress.$field) {
            $errors += "Campo 'progress.$field' requerido"
        } elseif ($json.progress.$field -isnot [int]) {
            $errors += "Campo 'progress.$field' debe ser numero entero"
        }
    }
    if ($null -ne $json.progress.total_tasks -and $null -ne $json.progress.completed_tasks) {
        if ($json.progress.completed_tasks -gt $json.progress.total_tasks) {
            $warnings += "progress.completed_tasks ($($json.progress.completed_tasks)) > total_tasks ($($json.progress.total_tasks))"
        }
    }
    Write-Host "  OK: progress = $($json.progress.completed_tasks)/$($json.progress.total_tasks) tareas" -ForegroundColor Green
}

# 5. tasks (array)
if (-not $json.tasks) {
    $errors += "Campo 'tasks' requerido (array)"
} elseif ($json.tasks -isnot [System.Array]) {
    $errors += "Campo 'tasks' debe ser un array"
} else {
    $validStatuses = @("pending", "in_progress", "completed", "blocked")
    foreach ($task in $json.tasks) {
        if (-not $task.id) { $errors += "Tarea sin 'id'" }
        if (-not $task.title) { $errors += "Tarea '$($task.id)' sin 'title'" }
        if (-not $task.status) {
            $errors += "Tarea '$($task.id)' sin 'status'"
        } elseif ($task.status -notin $validStatuses) {
            $errors += "Tarea '$($task.id)' status invalido: $($task.status)"
        }
    }
    Write-Host "  OK: tasks = $($json.tasks.Count) tareas definidas" -ForegroundColor Green
}

# 6. evaluation
if (-not $json.evaluation) {
    $errors += "Campo 'evaluation' requerido"
} else {
    if ($null -ne $json.evaluation.score) {
        if ($json.evaluation.score -lt 0 -or $json.evaluation.score -gt 100) {
            $warnings += "evaluation.score fuera de rango 0-100: $($json.evaluation.score)"
        }
    }
    Write-Host "  OK: evaluation definido" -ForegroundColor Green
}

# === CAMPOS PROHIBIDOS ===

$prohibitedFields = @("time_invested_hours", "estimated_hours", "duration")
foreach ($field in $prohibitedFields) {
    if ($json.PSObject.Properties[$field]) {
        $errors += "Campo PROHIBIDO detectado: '$field' - este campo no existe en Schema v3.0"
    }
}

# Verificar en tasks tambien
if ($json.tasks) {
    foreach ($task in $json.tasks) {
        foreach ($field in $prohibitedFields) {
            if ($task.PSObject.Properties[$field]) {
                $errors += "Campo PROHIBIDO en tarea '$($task.id)': '$field'"
            }
        }
    }
}

# === RESULTADO ===

Write-Host ""
Write-Host "==========================" -ForegroundColor Gray

if ($errors.Count -gt 0) {
    Write-Host ""
    Write-Host "ERRORES ($($errors.Count)):" -ForegroundColor Red
    foreach ($err in $errors) {
        Write-Host "  X $err" -ForegroundColor Red
    }
    Write-Host ""
    Write-Host "VALIDACION: FALLIDA" -ForegroundColor Red
    exit 1
} else {
    if ($warnings.Count -gt 0) {
        Write-Host ""
        Write-Host "ADVERTENCIAS ($($warnings.Count)):" -ForegroundColor Yellow
        foreach ($warning in $warnings) {
            Write-Host "  ! $warning" -ForegroundColor Yellow
        }
    }
    Write-Host ""
    Write-Host "VALIDACION: EXITOSA" -ForegroundColor Green
    exit 0
}