#Requires -Version 7
<#
.SYNOPSIS
    MVGN Init — Inicialización determinista desde templates/
.DESCRIPTION
    Copia la estructura canónica definida en templates/LAYOUT.md.
    Sin contenido embebido: toda pieza proviene de templates/ (fuente única).
    Dos ejecuciones sobre el mismo estado producen resultados idénticos (idempotente).
.PARAMETER Profile
    Perfil a usar: lite (único disponible)
.PARAMETER Collab
    Activar modo colaborativo multi-agente
.PARAMETER Role
    Rol de la otra IA (auditor, developer, reviewer, custom)
.PARAMETER WhatIf
    Simular sin ejecutar cambios
.EXAMPLE
    .\tools\mvgn-init.ps1
    .\tools\mvgn-init.ps1 -Collab -Role "auditor"
    .\tools\mvgn-init.ps1 -WhatIf
#>

param(
    [ValidateSet("lite")]
    [string]$Profile = "lite",
    [switch]$Collab,
    [string]$Role = "",
    [switch]$Audio,
    [switch]$WhatIf
)

# El proyecto dueño se deriva de la UBICACIÓN DEL SCRIPT (<proyecto>/.mvgn/tools/),
# nunca del cwd del llamador: bootstrap puede invocar desde cualquier directorio.
$ProjectRoot   = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path
$MvgnDir       = Join-Path $ProjectRoot ".mvgn"
$BlackboardDir = Join-Path $MvgnDir "blackboard"
$RuntimeFile   = Join-Path $MvgnDir "mvgn-runtime.json"
$TemplatesDir  = Join-Path $PSScriptRoot "..\templates"

# Pipeline de audio presente? (es opcional en el payload; fix H-8: no preguntar si no viaja)
$AudioPipeline = Join-Path $MvgnDir "scripts\audio\transcribe.ps1"

function Invoke-AudioIngest {
    if (Test-Path $AudioPipeline) {
        & (Join-Path $PSScriptRoot "mvgn-audio.ps1")
    } else {
        Write-Host "  [i] El pipeline de audio no está instalado en este proyecto." -ForegroundColor Gray
        Write-Host "      Es opcional en el payload (scripts/audio/). Sin él, mvgn-audio no está disponible." -ForegroundColor DarkGray
    }
}

Write-Host ""

function Write-Step($step, $msg) { Write-Host "[$step] $msg" -ForegroundColor Cyan }

# Copia idempotente: solo crea si el destino no existe
function Copy-IfMissing {
    param([string]$Src, [string]$Dst)
    if (-not (Test-Path $Src)) {
        Write-Host "  [WARN] Plantilla faltante: $Src" -ForegroundColor Yellow
        return
    }
    if (Test-Path $Dst) {
        Write-Host "  = $Dst (ya existe)" -ForegroundColor DarkGray
        return
    }
    if ($WhatIf) {
        Write-Host "  [WhatIf] Se crearía $Dst" -ForegroundColor Yellow
        return
    }
    $parent = Split-Path $Dst -Parent
    if ($parent) { New-Item -ItemType Directory -Path $parent -Force | Out-Null }
    Copy-Item $Src $Dst
    Write-Host "  + $Dst" -ForegroundColor Green
}

Write-Host ""
Write-Host "MVGN v5.3 - Inicializacion determinista" -ForegroundColor Cyan
Write-Host "=======================================" -ForegroundColor Cyan

$hasBlackboard = Test-Path (Join-Path $BlackboardDir "state.json")
$hasRuntime = Test-Path $RuntimeFile

if (-not (Test-Path "$TemplatesDir\LAYOUT.md")) {
    Write-Host "[FAIL] templates/ no encontrado en $TemplatesDir" -ForegroundColor Red
    Write-Host "       Instala MVGN con tools\bootstrap.ps1 o clona el payload completo." -ForegroundColor Gray
    exit 1
}

# ─── MODO INIT ───────────────────────────────────────────────

if (-not $hasBlackboard -and -not $hasRuntime) {

    Write-Step "1/6" "Creando estructura blackboard..."
    if (-not $WhatIf) {
        New-Item -ItemType Directory -Path (Join-Path $BlackboardDir "logs") -Force | Out-Null
    }

    Write-Step "2/6" "Generando state.json..."
    $state = @{
        schema_version    = "5.2.0"
        task_id           = ""
        task_description  = ""
        status            = "idle"
        current_phase     = 0
        agent_sequence    = @()
        completed_agents  = @()
        artifacts_generated = @()
        recovered         = $false
        last_error        = $null
        timestamp         = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
    }
    if (-not $WhatIf) {
        $state | ConvertTo-Json -Depth 10 | Set-Content (Join-Path $BlackboardDir "state.json") -Encoding UTF8
        Write-Host "  + $BlackboardDir\state.json" -ForegroundColor Green
    } else {
        Write-Host "  [WhatIf] Se crearía $BlackboardDir\state.json" -ForegroundColor Yellow
    }

    Write-Step "3/6" "Desplegando contrato y blackboard desde templates..."
    Copy-IfMissing "$TemplatesDir\LAYOUT.md"      (Join-Path $MvgnDir "LAYOUT.md")
    Copy-IfMissing "$TemplatesDir\JOIN.md"        (Join-Path $BlackboardDir "JOIN.md")
    Copy-IfMissing "$TemplatesDir\context.md"     (Join-Path $BlackboardDir "context.md")
    Copy-IfMissing "$TemplatesDir\gitignore.template" (Join-Path $ProjectRoot ".gitignore")

    Write-Step "4/6" "Desplegando documentación base..."
    foreach ($doc in Get-ChildItem "$TemplatesDir\docs" -File) {
        Copy-IfMissing $doc.FullName (Join-Path $ProjectRoot ("docs\" + $doc.Name))
    }

    Write-Step "5/6" "Creando mvgn-runtime.json..."
    if (-not (Test-Path $RuntimeFile)) {
        if (-not $WhatIf) {
            @{
                runtime_version = "5.3.0"
                schema_version  = "5.3.0"
                profile         = "lite"
                execution_mode  = "NORMAL"
                capabilities    = @{
                    blackboard      = $true
                    collaboration   = $true
                    handoff         = $true
                    dual_entry      = $true
                    archival        = $true
                    orchestration   = $true
                    audio_ingestion = $false
                }
                actors = @()
            } | ConvertTo-Json -Depth 10 | Set-Content $RuntimeFile -Encoding UTF8
            Write-Host "  + mvgn-runtime.json" -ForegroundColor Green
        } else {
            Write-Host "  [WhatIf] Se crearía mvgn-runtime.json" -ForegroundColor Yellow
        }
    } else {
        Write-Host "  = mvgn-runtime.json (ya existe)" -ForegroundColor DarkGray
    }

    Write-Step "6/6" "Estructura final según LAYOUT/2.0..."
    Write-Host ""
    Write-Host "Estructura creada:" -ForegroundColor Green
    Write-Host "  .mvgn/blackboard/{state.json, JOIN.md, context.md}" -ForegroundColor Gray
    Write-Host "  .mvgn/LAYOUT.md (contrato de estructura)" -ForegroundColor Gray
    Write-Host "  docs/ (documentación base desde templates)" -ForegroundColor Gray
    Write-Host "  .gitignore (exclusión de runtime)" -ForegroundColor Gray
    Write-Host "  mvgn-runtime.json" -ForegroundColor Gray
    Write-Host ""
    Write-Host "MVGN inicializado. Entry point: .mvgn/NYX.md" -ForegroundColor Cyan

    # ─── INGESTA DE AUDIO opcional al crear el proyecto (solo si el pipeline viaja — fix H-8) ───
    if (Test-Path $AudioPipeline) {
        $runAudio = $Audio
        if (-not $runAudio -and -not $WhatIf) {
            $answer = Read-Host "¿Hay audio para alimentar el contexto inicial? (s/N)"
            $runAudio = ($answer -eq "s" -or $answer -eq "S")
        }
        if ($runAudio) { Invoke-AudioIngest }
    }
}

# ─── MODO COLLAB ─────────────────────────────────────────────

if ($Collab -or $Role -ne "") {
    Write-Host "Modo: COLLAB (Colaboración Multi-Agente)" -ForegroundColor Yellow

    if ($Role -eq "") {
        Write-Host "¿Qué rol tendrá la otra IA?" -ForegroundColor Cyan
        Write-Host "  1. auditor    - Revisión y reportes (jurisdicción: ALETHEIA)" -ForegroundColor White
        Write-Host "  2. developer  - Generación de código" -ForegroundColor White
        Write-Host "  3. reviewer   - Code review / validación (jurisdicción: NIKE)" -ForegroundColor White
        Write-Host "  4. custom     - Personalizado" -ForegroundColor White
        $choice = Read-Host "Selecciona (1-4)"
        $Role = switch ($choice) { "1" { "auditor" } "2" { "developer" } "3" { "reviewer" } "4" { "custom" } default { "auditor" } }
    }

    # JOIN.md base siempre viene del canon; la misión se agrega como anexo
    Copy-IfMissing "$TemplatesDir\JOIN.md" (Join-Path $BlackboardDir "JOIN.md")

    $task = ""
    if (-not $WhatIf) {
        $task = Read-Host "Tarea para la otra IA (vacío para skip)"
    }

    # H-9: numerar el anexo después de la última sección existente del JOIN desplegado
    $joinPath = (Join-Path $BlackboardDir "JOIN.md")
    $maxSection = 0
    if (Test-Path $joinPath) {
        Get-Content $joinPath | ForEach-Object {
            if ($_ -match '^## ([IVX]+)\.') {
                $roman = $matches[1]
                $val = 0; $prev = 0
                foreach ($ch in $roman.ToCharArray()) {
                    $v = switch ($ch) { 'I' {1} 'V' {5} 'X' {10} default {0} }
                    if ($v -gt $prev) { $val += $v - 2 * $prev } else { $val += $v }
                    $prev = $v
                }
                if ($val -gt $maxSection) { $maxSection = $val }
            }
        }
    }
    function ConvertTo-Roman([int]$n) {
        $map = @(@(10,'X'), @(9,'IX'), @(5,'V'), @(4,'IV'), @(1,'I'))
        $out = ""
        foreach ($m in $map) { while ($n -ge $m[0]) { $out += $m[1]; $n -= $m[0] } }
        $out
    }
    $missionSection = ConvertTo-Roman ($maxSection + 1)

    $mission = @"


---

## $missionSection. Misión asignada

```
ROL:        $Role
TAREA:      $task
ASIGNADO:   $(Get-Date -Format 'yyyy-MM-dd')
```

**No preguntes qué hacer. La misión está arriba. Ejecuta. Reporta. Transfiere.**
"@

    if (-not $WhatIf) {
        Add-Content (Join-Path $BlackboardDir "JOIN.md") $mission -Encoding UTF8
        Write-Host "  + Misión agregada a $BlackboardDir\JOIN.md" -ForegroundColor Green

        if ($task -ne "") {
            @{
                schema_version    = "5.2.0"
                task_id           = "task-$(Get-Date -Format 'yyyyMMddHHmmss')"
                task_description  = $task
                status            = "pending"
                current_phase     = 0
                agent_sequence    = @($Role)
                completed_agents  = @()
                artifacts_generated = @()
                recovered         = $false
                last_error        = $null
                timestamp         = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
            } | ConvertTo-Json -Depth 10 | Set-Content (Join-Path $BlackboardDir "state.json") -Encoding UTF8
            Write-Host "  + Tarea registrada en state.json" -ForegroundColor Green
        }
    } else {
        Write-Host "  [WhatIf] Se agregaría misión ($Role) al JOIN.md" -ForegroundColor Yellow
    }

    Write-Host ""
    Write-Host "Sesión colaborativa activada. Instrucción para la otra IA:" -ForegroundColor Yellow
    Write-Host "  Lee .mvgn/blackboard/JOIN.md y ejecuta lo que dice." -ForegroundColor White
}

# ─── MODO READY ──────────────────────────────────────────────

if ($hasBlackboard -and -not $Collab -and $Role -eq "") {
    Write-Host "Modo: READY (MVGN activo)" -ForegroundColor Yellow
    $state = Get-Content (Join-Path $BlackboardDir "state.json") -Raw | ConvertFrom-Json
    Write-Host "Estado:      $($state.status)" -ForegroundColor Cyan
    Write-Host "Tarea:       $($state.task_description)" -ForegroundColor Gray
    Write-Host "Schema:      $($state.schema_version)" -ForegroundColor Gray
    Write-Host "Agentes:     $($state.completed_agents -join ', ')" -ForegroundColor Gray

    # ─── Ingesta de audio para ciclos de features/correcciones (T-118, fix H-8) ───
    if ($Audio) {
        Invoke-AudioIngest
    } elseif (Test-Path $AudioPipeline) {
        Write-Host ""
        Write-Host "Ingesta de audio (features, correcciones): mvgn-init.ps1 -Audio o .mvgn\tools\mvgn-audio.ps1" -ForegroundColor DarkGray
    }
}

Write-Host ""
Write-Host "mvgn-init.ps1 completado." -ForegroundColor Green
