# MVGN Entry Point — PowerShell
# Versión: v5.3.0 — Sincronizado con NYX.md
# PowerShell ONLY (Decisor: Armand)
# Uso: .\NYX.ps1 [-Mode init|ready|upgrade|update|detect|collab|observer] [-Web] [-Port N] [-Force]

param(
    [ValidateSet("detect", "init", "ready", "upgrade", "update", "collab", "observer")]
    [string]$Mode = "detect",
    [string]$Task = "",
    [string]$CollabRole = "",
    [string]$CollabTask = "",
    [switch]$Web,
    [int]$Port = 8080,
    [switch]$Force
)

$BlackboardDir = ".mvgn\blackboard"
$StateFile = "$BlackboardDir\state.json"
$ContextFile = "$BlackboardDir\context.md"
$JoinFile = "$BlackboardDir\JOIN.md"
$LogFile = "$BlackboardDir\logs\execution.log"
$RuntimeFile = "mvgn-runtime.json"

# ─── CONSTANTES DE VERSIÓN ───────────────────────────────────
$MVGN_REPO = "https://github.com/mvgn91/mvgnlabs-starter-kit"
$MVGN_RAW = "https://raw.githubusercontent.com/mvgn91/mvgnlabs-starter-kit/main"

# ─── FUNCIONES ───────────────────────────────────────────────

function Log($msg) {
    $ts = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    Write-Host "[$ts] $msg" -ForegroundColor Cyan
}

function Success($msg) {
    $ts = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    Write-Host "[$ts] OK: $msg" -ForegroundColor Green
}

function Warn($msg) {
    $ts = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    Write-Host "[$ts] WARN: $msg" -ForegroundColor Yellow
}

function Fail($msg) {
    $ts = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    Write-Host "[$ts] ERROR: $msg" -ForegroundColor Red
    exit 1
}

# ─── VERSIÓN ──────────────────────────────────────────────────

function Parse-Version($v) {
    if (-not $v) { return @{major=0; minor=0; patch=0; raw="0.0.0"} }
    $clean = $v.TrimStart('v')
    $parts = $clean.Split('-')[0].Split('.')
    return @{
        major = if ($parts[0]) { [int]$parts[0] } else { 0 }
        minor = if ($parts[1]) { [int]$parts[1] } else { 0 }
        patch = if ($parts[2]) { [int]$parts[2] } else { 0 }
        raw   = $clean
    }
}

function Compare-Version($a, $b) {
    $pa = Parse-Version $a
    $pb = Parse-Version $b
    if ($pa.major -ne $pb.major) { return ($pa.major - $pb.major) }
    if ($pa.minor -ne $pb.minor) { return ($pa.minor - $pb.minor) }
    return ($pa.patch - $pb.patch)
}

function Format-Version($v, $color) {
    $p = Parse-Version $v
    return "v$($p.major).$($p.minor).$($p.patch)"
}

# ─── UPDATE CHECK ─────────────────────────────────────────────

function Check-Update {
    Write-Host ""
    Write-Host "MVGN VERSION CHECK" -ForegroundColor Cyan
    Write-Host "==================" -ForegroundColor Cyan
    Write-Host ""

    # ── Leer versión local ──────────────────────────────────
    $localVersion = $null
    if (Test-Path $RuntimeFile) {
        try {
            $r = Get-Content $RuntimeFile -Raw | ConvertFrom-Json
            $localVersion = $r.runtime_version
        } catch { }
    }
    $localTag = $null
    $localCommit = $null
    $gitAvailable = $null -ne (Get-Command git -ErrorAction SilentlyContinue)
    if ($gitAvailable) {
        $localTagRaw = @(git tag --list 2>$null | Sort-Object -Descending)
        if ($localTagRaw.Count -gt 0) { $localTag = $localTagRaw[0] }
        $localCommit = git rev-parse --short HEAD 2>$null
    }

    $localDisplay = if ($localVersion) { "v$localVersion" } else { "—" }
    $tagDisplay = if ($localTag) { $localTag } else { "—" }
    $commitDisplay = if ($localCommit) { $localCommit } else { "—" }

    Write-Host "  LOCAL" -ForegroundColor White
    Write-Host "  ├─ Runtime:    $localDisplay"
    Write-Host "  ├─ Git tag:    $tagDisplay"
    Write-Host "  └─ Commit:     $commitDisplay"
    Write-Host ""

    # ── Fetch remoto ────────────────────────────────────────
    Log "Consultando remoto: $MVGN_REPO"
    Write-Host ""

    $remoteVersion = $null
    $remoteTag = $null
    $remoteCommit = $null
    $fetchOk = $true

    if ($gitAvailable) {
        $fetchOut = git fetch --tags origin 2>&1
        if ($LASTEXITCODE -eq 0) {
            $allTags = @(git tag --list 2>$null | Sort-Object -Descending)
            if ($allTags.Count -gt 0) { $remoteTag = $allTags[0] }
            $remoteCommit = git rev-parse --short origin/main 2>$null
            if (-not $remoteCommit) { $remoteCommit = git rev-parse --short origin/master 2>$null }
        } else {
            $fetchOk = $false
        }
    }

    # Intentar fetch de mvgn-runtime.json desde raw
    try {
        $rawUrl = "$MVGN_RAW/mvgn-runtime.json"
        $remoteJson = Invoke-WebRequest -Uri $rawUrl -UseBasicParsing -TimeoutSec 5 2>$null
        if ($remoteJson) {
            $remote = $remoteJson.Content | ConvertFrom-Json
            $remoteVersion = $remote.runtime_version
        }
    } catch {
        if (-not $remoteTag) { $fetchOk = $false }
    }

    if (-not $fetchOk) {
        Warn "No se pudo contactar el remoto ($MVGN_REPO)"
        Write-Host "  Verifica tu conexión a internet o si el repo es accesible." -ForegroundColor Gray
        return $null
    }

    $remoteDisplayVersion = if ($remoteVersion) { "v$remoteVersion" } else { "—" }
    $remoteDisplayTag = if ($remoteTag) { $remoteTag } else { "—" }
    $remoteDisplayCommit = if ($remoteCommit) { $remoteCommit } else { "—" }

    Write-Host "  REMOTO ($MVGN_REPO)" -ForegroundColor White
    Write-Host "  ├─ Runtime:    $remoteDisplayVersion"
    Write-Host "  ├─ Git tag:    $remoteDisplayTag"
    Write-Host "  └─ Commit:     $remoteDisplayCommit"
    Write-Host ""

    # ── Comparar ────────────────────────────────────────────
    $diff = @()

    if ($localVersion -and $remoteVersion) {
        $cmp = Compare-Version $localVersion $remoteVersion
        if ($cmp -lt 0) {
            $diff += "Runtime: $localDisplay → $remoteDisplayVersion"
        } elseif ($cmp -gt 0) {
            $diff += "Runtime: $localDisplay (versión local más nueva que remoto)"
        }
    }

    if ($localTag -and $remoteTag -and $localTag -ne $remoteTag) {
        $cmp = Compare-Version $localTag $remoteTag
        if ($cmp -lt 0) {
            $diff += "Git tag: $localTag → $remoteTag"
        } elseif ($cmp -gt 0) {
            $diff += "Git tag: $localTag (tag local más nuevo que remoto)"
        }
    }

    if ($localCommit -and $remoteCommit -and $localCommit -ne $remoteCommit) {
        $diff += "Commit: $localCommit → $remoteCommit"
    }

    if ($diff.Count -eq 0) {
        Success "MVGN está actualizado (v$localVersion)"
        return @{
            updateAvailable = $false
            localVersion    = $localVersion
            remoteVersion   = $remoteVersion
        }
    }

    Warn "Hay diferencias entre local y remoto:"
    foreach ($d in $diff) {
        Write-Host "    • $d" -ForegroundColor Yellow
    }
    Write-Host ""

    return @{
        updateAvailable = $true
        localVersion    = $localVersion
        remoteVersion   = $remoteVersion
        localTag        = $localTag
        remoteTag       = $remoteTag
        localCommit     = $localCommit
        remoteCommit    = $remoteCommit
        diff            = $diff
    }
}

function Invoke-Update {
    # T-112: actualización vía GitHub Releases.
    # No depende del historial git: los proyectos adoptados no clonan el lab.
    # Delega en bootstrap.ps1 (comparación de versiones, changelog delta,
    # preservación del blackboard).
    $bootstrap = Join-Path $PSScriptRoot "tools\bootstrap.ps1"
    if (-not (Test-Path $bootstrap)) {
        Fail "tools/bootstrap.ps1 no encontrado. Actualiza manualmente desde $MVGN_REPO"
        return
    }

    Log "Verificando actualización vía GitHub Releases..."
    Write-Host ""

    $pwshExe = (Get-Process -Id $PID).Path
    $args_ = @("-NoProfile", "-File", $bootstrap, "-Destination", (Get-Location).Path, "-NoPause")
    if ($Force) { $args_ += "-Force" }

    & $pwshExe @args_
}

# ─── DETECCIÓN DE ENTORNO ────────────────────────────────────

# ─── DETECCIÓN DE ENTORNO ────────────────────────────────────

function Test-Environment {
    $hasBlackboard = Test-Path $StateFile
    $hasRuntime = Test-Path $RuntimeFile

    if (-not $hasBlackboard -and -not $hasRuntime) {
        return "INIT"
    } elseif ($hasBlackboard) {
        return "READY"
    } elseif ($hasRuntime -and -not $hasBlackboard) {
        return "UPGRADE"
    }
    return "READY"
}

# ─── MODO INIT ───────────────────────────────────────────────

function Invoke-Init {
    Log "Inicializando proyecto MVGN v5.1..."

    # Crear estructura de directorios
    New-Item -ItemType Directory -Path "$BlackboardDir\logs" -Force | Out-Null
    New-Item -ItemType Directory -Path "docs" -Force | Out-Null

    # Crear state.json inicial
    $initState = @{
        task_id = "init-$(Get-Date -Format 'yyyyMMddHHmmss')"
        task_description = "Proyecto inicializado"
        status = "pending"
        current_phase = 0
        agent_sequence = @("buffy")
        completed_agents = @()
        artifacts_generated = @()
        last_error = $null
        timestamp = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ssZ")
    }
    $initState | ConvertTo-Json -Depth 10 | Set-Content $StateFile
    Success "state.json creado"

    # Crear mvgn-runtime.json
    $runtime = @{
        runtime_version = "5.1.0"
        schema_version = "3.0.0"
        profile = "lite"
        execution_mode = "NORMAL"
    }
    $runtime | ConvertTo-Json -Depth 10 | Set-Content $RuntimeFile
    Success "mvgn-runtime.json creado"

    Log "MVGN v5.1 inicializado. ¿Activar modo COLLAB?"
    Log "Ejecuta: .\NYX.ps1 -Mode collab"
}

# ─── MODO READY ──────────────────────────────────────────────

function Invoke-Ready {
    if (-not (Test-Path $StateFile)) {
        Fail "state.json no encontrado"
    }

    try {
        $s = Get-Content $StateFile -Raw -ErrorAction Stop | ConvertFrom-Json -ErrorAction Stop
    } catch {
        Fail "state.json corrupto: $_"
    }

    Write-Host ""
    Write-Host "MVGN BLACKBOARD STATUS" -ForegroundColor Cyan
    Write-Host "======================" -ForegroundColor Cyan
    Write-Host "Tarea:      $($s.task_description)"
    Write-Host "ID:         $($s.task_id)"
    Write-Host "Estado:     $($s.status)"
    Write-Host "Fase:       $($s.current_phase)/3"
    Write-Host "Completado: $($s.completed_agents -join ', ')"
    Write-Host "Artefactos: $($s.artifacts_generated -join ', ')"
    Write-Host "Error:      $($s.last_error)"
    Write-Host ""

    switch ($s.status) {
        "pending" {
            Log "Tarea pendiente. Ejecuta: .\scripts\orchestrate.ps1 -Resume"
        }
        "in_progress" {
            Log "Tarea en progreso. Fase $($s.current_phase). Para continuar: .\scripts\orchestrate.ps1 -Resume"
        }
        "completed" {
            Log "Tarea completada. Crea una nueva con: .\scripts\orchestrate.ps1 -Task 'descripcion'"
        }
        "failed" {
            Warn "Tarea falló: $($s.last_error)"
            Log "Para reanudar: .\scripts\orchestrate.ps1 -Resume"
            Log "Para recovery: Lee docs/mvgn-recovery.md"
        }
    }
}

# ─── MODO UPGRADE ────────────────────────────────────────────

function Invoke-Upgrade {
    Log "Migrando a Blackboard v5.1..."
    Invoke-Init
    Success "MVGN migrado a Blackboard Pattern v5.1"
}

# ─── MODO OBSERVER ───────────────────────────────────────────

function Invoke-Observer {
    param([switch]$EnableWeb, [int]$WebPort)

    $observerExe = Join-Path $PSScriptRoot "mvgn-scan\mvgn-scan.exe"

    if (-not (Test-Path $observerExe)) {
        Log "mvgn-scan.exe no encontrado. Construyendo..."
        $buildDir = Join-Path $PSScriptRoot "mvgn-scan"
        $build = & go build -o "$buildDir\mvgn-scan.exe" ./cmd/mvgn-scan/ 2>&1
        if ($LASTEXITCODE -ne 0) {
            Fail "Build falló: $build"
        }
        Success "mvgn-scan.exe construido"
    }

    $args = @()

    if ($EnableWeb) {
        $args += "-observe", "-port", $WebPort
    } else {
        $args += "-watch"
    }
    $args += "-dir", $PSScriptRoot

    Log "Iniciando MVGN Observer..."
    if ($EnableWeb) {
        Log "Dashboard: http://localhost:$WebPort"
    }
    Write-Host ""
    Write-Host "  Presiona Ctrl+C para cerrar." -ForegroundColor Yellow
    Write-Host ""

    & $observerExe $args
}

# ─── MODO COLLAB ─────────────────────────────────────────────

function Invoke-Collab {
    param([string]$Role, [string]$TaskDesc)

    if ([string]::IsNullOrEmpty($Role)) {
        Write-Host ""
        Write-Host "Roles disponibles para colaboración:" -ForegroundColor Cyan
        Write-Host "  1. Auditor      → Revisión y reportes"
        Write-Host "  2. Desarrollador → Generación de código"
        Write-Host "  3. Revisor      → Code review"
        Write-Host "  4. Personalizado → Describir el rol"
        Write-Host ""
        $Role = Read-Host "Selecciona un rol (1-4)"
        switch ($Role) {
            "1" { $Role = "Auditor" }
            "2" { $Role = "Desarrollador" }
            "3" { $Role = "Revisor" }
            "4" { $Role = Read-Host "Describe el rol personalizado" }
        }
    }

    if ([string]::IsNullOrEmpty($TaskDesc)) {
        $TaskDesc = Read-Host "¿Tarea para la otra IA? (dejar vacío si no hay)"
    }

    # Generar JOIN.md
    $taskSection = if ($TaskDesc -ne "") { "**Descripción:** $TaskDesc`r`n`r`n" } else { "" }
    
    $roleSection = switch ($Role) {
        "Auditor" { "- Revisa los archivos del proyecto`r`n  - Genera reportes en docs/`r`n  - Actualiza state.json con hallazgos" }
        "Desarrollador" { "- Genera código en src/`r`n  - Sigue la arquitectura existente`r`n  - Actualiza state.json con artefactos" }
        "Revisor" { "- Revisa el código existente`r`n  - Reporta issues y sugerencias`r`n  - Actualiza state.json con resultados" }
        default { "- Ejecuta la tarea según tu criterio`r`n  - Documenta los resultados`r`n  - Actualiza state.json" }
    }
    
    $joinContent = @"
# JOIN — Colaboración Multi-Agente

> **Rol:** $Role
> **Generado:** $(Get-Date -Format 'yyyy-MM-ddTHH:mm:ssZ')
> **Por:** MVGN NYX.ps1 v5.1.0

---

## Tu tarea

$taskSection
1. Lee `.mvgn/blackboard/state.json` para entender el contexto
2. Ejecuta la tarea según tu rol
3. Actualiza `state.json` al terminar

## Rol: $Role

$roleSection

---

**No preguntes. Solo ejecuta.**
"@

    $joinContent | Set-Content $JoinFile
    Success "JOIN.md generado para rol: $Role"

    if (-not [string]::IsNullOrEmpty($TaskDesc)) {
        # Actualizar state.json con la nueva tarea
        $s = Get-Content $StateFile -Raw | ConvertFrom-Json
        $s.task_id = "collab-$(Get-Date -Format 'yyyyMMddHHmmss')"
        $s.task_description = $TaskDesc
        $s.status = "pending"
        $s.current_phase = 0
        $s.completed_agents = @()
        $s.artifacts_generated = @()
        $s.last_error = $null
        $s.timestamp = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ssZ")
        $s | ConvertTo-Json -Depth 10 | Set-Content $StateFile
        Success "state.json actualizado con nueva tarea"
    }

    Write-Host ""
    Log "Sesión colaborativa activada."
    Log "Para que la otra IA se una, dale esta instrucción:"
    Write-Host ""
    Write-Host "  Lee .mvgn/blackboard/JOIN.md y ejecuta lo que dice." -ForegroundColor Yellow
    Write-Host ""
}

# ─── MAIN ────────────────────────────────────────────────────

Write-Host ""
Write-Host "MVGN v5.1 — Entry Point" -ForegroundColor Cyan
Write-Host "========================" -ForegroundColor Cyan
Write-Host ""

$detectedMode = if ($Mode -eq "detect") { Test-Environment } else { $Mode.ToUpper() }

# Soporte: si el usuario pasa "observer" sin -Mode
if ($detectedMode -ne "OBSERVER" -and $Mode -eq "observer") {
    $detectedMode = "OBSERVER"
}

switch ($detectedMode) {
    "INIT" {
        Log "Modo detectado: INIT"
        Invoke-Init
    }
    "READY" {
        Log "Modo detectado: READY"
        Invoke-Ready
        Write-Host ""
        $updateCheck = Check-Update
        if ($updateCheck -and $updateCheck.updateAvailable) {
            Write-Host ""
            $r = Read-Host "  ¿Actualizar MVGN ahora? (s/N)"
            if ($r -eq "s" -or $r -eq "S") {
                Invoke-Update
            }
        }
    }
    "UPGRADE" {
        Log "Modo detectado: UPGRADE"
        Invoke-Upgrade
    }
    "UPDATE" {
        Log "Modo: UPDATE"
        Invoke-Update
    }
    "OBSERVER" {
        Log "Modo: OBSERVER"
        Invoke-Observer -EnableWeb:$Web -WebPort $Port
    }
    "collab" {
        Log "Modo: COLLAB"
        Invoke-Collab -Role $CollabRole -TaskDesc $CollabTask
    }
}

Write-Host ""
Success "NYX.ps1 completado"
