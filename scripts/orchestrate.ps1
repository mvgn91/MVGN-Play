# MVGN Orchestrator - Blackboard Pattern
# Ejecuta: OpenCode -> FreeBuff -> MiMoCode
# Costo: $0 | Resiliencia: Manejo de errores + resume

param(
    [string]$Task = "",
    [switch]$Resume,
    [switch]$Status,
    [switch]$AutoCommit,
    [string]$Handoff = ""
)

$BlackboardDir = ".mvgn\blackboard"
$StateFile = "$BlackboardDir\state.json"
$LogFile = "$BlackboardDir\logs\execution.log"
$AdaptersDir = "$PSScriptRoot\agents"

# ─── FUNCIONES DE LOG ────────────────────────────────────────

function Log($msg) {
    $ts = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    Write-Host "[$ts] $msg" -ForegroundColor Cyan
    "[$ts] $msg" | Out-File $LogFile -Append
}

function Success($msg) {
    $ts = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    Write-Host "[$ts] OK: $msg" -ForegroundColor Green
    "[$ts] OK: $msg" | Out-File $LogFile -Append
}

function Fail($msg) {
    $ts = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    Write-Host "[$ts] ERROR: $msg" -ForegroundColor Red
    "[$ts] ERROR: $msg" | Out-File $LogFile -Append
}

function Warn($msg) {
    $ts = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    Write-Host "[$ts] WARN: $msg" -ForegroundColor Yellow
    "[$ts] WARN: $msg" | Out-File $LogFile -Append
}

# ─── FUNCIONES DE ESTADO ─────────────────────────────────────

function Update-State($script) {
    $s = Get-Content $StateFile -Raw | ConvertFrom-Json
    $s = $script.Invoke($s)
    $s | ConvertTo-Json -Depth 10 | Set-Content $StateFile
}

function Get-State {
    if (-not (Test-Path $StateFile)) { return $null }
    return Get-Content $StateFile -Raw | ConvertFrom-Json
}

function Show-Status {
    $s = Get-State
    if (-not $s) {
        Write-Host "No hay tarea activa" -ForegroundColor Yellow
        return
    }
    Write-Host ""
    Write-Host "MVGN BLACKBOARD STATUS" -ForegroundColor Cyan
    Write-Host "======================" -ForegroundColor Cyan
    Write-Host "Tarea:      $($s.task_description)"
    Write-Host "ID:         $($s.task_id)"
    Write-Host "Estado:     $($s.status)"
    Write-Host "Fase:       $($s.current_phase)/3"
    Write-Host "Completado: $($s.completed_agents -join ', ')"
    Write-Host "Error:      $($s.last_error)"
    Write-Host ""
}

function Execute-Handoff {
    param([string]$NextAgent)
    
    $s = Get-State
    if (-not $s) {
        Fail "No hay estado para handoff"
        return
    }
    
    Write-Host ""
    Write-Host "MVGN HANDOFF" -ForegroundColor Cyan
    Write-Host "============" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Agente actual completó: $($s.completed_agents -join ', ')" -ForegroundColor Green
    Write-Host "Tarea: $($s.task_description)" -ForegroundColor Gray
    Write-Host ""
    
    # Guardar contexto del handoff
    $handoff = @{
        schema_version = if ($s.schema_version) { $s.schema_version } else { "5.1.0" }
        from_agent = $s.completed_agents[-1]
        to_agent = $NextAgent
        timestamp = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ssZ")
        task_id = $s.task_id
        task_description = $s.task_description
        phase = if ($s.current_phase) { $s.current_phase } else { 0 }
        artifacts = $s.artifacts_generated
        next_action = "Continúa con la siguiente fase. Si hay error, ejecuta recovery (docs/mvgn-recovery.md)"
    }
    
    $handoffFile = "$BlackboardDir\handoff-latest.json"
    $handoff | ConvertTo-Json -Depth 10 | Set-Content $handoffFile
    
    # Actualizar state.json para siguiente agente
    Update-State { param($s) 
        $s.status = "pending"
        $s.current_phase = 0
        $s.agent_sequence = @($NextAgent)
        return $s
    }
    
    # Actualizar JOIN.md con contexto del handoff
    $joinContent = @"
# JOIN — Handoff entre agentes MVGN

> **Contexto:** Handoff desde agente anterior.
> **Schema v5.1** — Blackboard Pattern

---

## Resumen del handoff

- **De:** $($handoff.from_agent)
- **Para:** $NextAgent
- **Tarea:** $($s.task_description)
- **Fase:** $($handoff.phase)/3
- **Artefactos:** $($handoff.artifacts -join ', ')

---

## Tu tarea

1. Lee `.mvgn/blackboard/state.json` para estado actual
2. Lee `docs/mvgn-kernel.md` para reglas de comportamiento
3. Revisa el trabajo del agente anterior
4. Continúa con la siguiente fase o corrige issues
5. Si hay error, ejecuta recovery: `docs/mvgn-recovery.md`
6. Actualiza `state.json` al terminar

---

## Estado esperado al finalizar

```json
{
  "completed_agents": ["$NextAgent"],
  "artifacts_generated": ["ruta/del/artefacto"]
}
```

---

**No preguntes. Solo ejecuta.**
"@
    
    $joinContent | Set-Content "$BlackboardDir\JOIN.md"
    
    Write-Host "Handoff preparado:" -ForegroundColor Green
    Write-Host "  De:   $($handoff.from_agent)" -ForegroundColor Gray
    Write-Host "  Para: $NextAgent" -ForegroundColor Gray
    Write-Host "  JOIN.md actualizado" -ForegroundColor Gray
    Write-Host ""
    Write-Host "Para activar al siguiente agente, dale:" -ForegroundColor Yellow
    Write-Host '  Lee .mvgn/blackboard/JOIN.md y ejecuta lo que dice.' -ForegroundColor White
    Write-Host ""
}

# ─── CREAR DIRECTORIOS ───────────────────────────────────────

New-Item -ItemType Directory -Path $BlackboardDir -Force | Out-Null
New-Item -ItemType Directory -Path "$BlackboardDir\logs" -Force | Out-Null

# ─── MODO STATUS ─────────────────────────────────────────────

if ($Status) {
    Show-Status
    exit
}

# ─── MODO HANDOFF ────────────────────────────────────────────

if ($Handoff -ne "") {
    Execute-Handoff -NextAgent $Handoff
    exit
}

# ─── INICIALIZAR O REANUDAR ──────────────────────────────────

if ($Task -ne "") {
    # Nueva tarea con schema v5.1
    @{
        schema_version = "5.1.0"
        task_id = "task-$(Get-Date -Format 'yyyyMMddHHmmss')"
        task_description = $Task
        status = "pending"
        current_phase = 0
        agent_sequence = @("opencode", "freebuff", "mimocode")
        completed_agents = @()
        artifacts_generated = @()
        recovered = $false
        last_error = $null
        timestamp = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ssZ")
    } | ConvertTo-Json -Depth 10 | Set-Content $StateFile
    Log "Tarea inicializada: $Task"
} elseif ($Resume) {
    # Reanudar tarea existente
    $s = Get-State
    if (-not $s) {
        Fail "No hay tarea que reanudar"
        exit 1
    }
    if ($s.status -eq "completed") {
        Warn "La tarea ya esta completada"
        Show-Status
        exit 0
    }
    if ($s.status -eq "failed") {
        Warn "La tarea fallo: $($s.last_error)"
        Write-Host "¿Reanudar de todos modos? (S/N)" -ForegroundColor Yellow
        $confirm = Read-Host
        if ($confirm -ne "S") { exit 0 }
        Update-State { param($s) $s.status = "pending"; $s.last_error = $null; return $s }
    }
    Log "Reanudando tarea: $($s.task_description)"
} else {
    Write-Host ""
    Write-Host "MVGN Orchestrator - Blackboard Pattern" -ForegroundColor Cyan
    Write-Host "======================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Uso:"
    Write-Host "  .\orchestrate.ps1 -Task 'descripcion'           # Nueva tarea"
    Write-Host "  .\orchestrate.ps1 -Resume                        # Reanudar tarea"
    Write-Host "  .\orchestrate.ps1 -Status                        # Ver estado"
    Write-Host "  .\orchestrate.ps1 -Handoff 'auditor'             # Handoff a siguiente agente"
    Write-Host "  .\orchestrate.ps1 -Task 'x' -AutoCommit          # Auto-commit al finalizar"
    Write-Host ""
    exit
}

# ─── EJECUTAR ORQUESTACION ───────────────────────────────────

Log "Iniciando orquestacion..."

$s = Get-State
$startPhase = if ($s.completed_agents -contains "opencode") { 2 }
              elseif ($s.completed_agents -contains "freebuff") { 3 }
              else { 1 }

# Phase 1: OpenCode
if ($startPhase -le 1) {
    Log "Fase 1: OpenCode..."
    Update-State { param($s) $s.status = "in_progress"; $s.current_phase = 1; return $s }
    
    try {
        & "$AdaptersDir\opencode-adapter.ps1"
        if ($LASTEXITCODE -ne 0) { throw "OpenCode fallo" }
        Update-State { param($s) $s.completed_agents += "opencode"; $s.status = "completed_phase_1"; return $s }
        Success "OpenCode completado"
    } catch {
        Update-State { param($s) $s.status = "failed"; $s.last_error = "OpenCode: $_"; return $s }
        Fail "OpenCode fallo: $_"
        exit 1
    }
    Start-Sleep -Seconds 2
}

# Phase 2: FreeBuff
if ($startPhase -le 2) {
    Log "Fase 2: FreeBuff..."
    Update-State { param($s) $s.current_phase = 2; return $s }
    
    try {
        & "$AdaptersDir\freebuff-adapter.ps1"
        if ($LASTEXITCODE -ne 0) { throw "FreeBuff fallo" }
        Update-State { param($s) $s.completed_agents += "freebuff"; $s.status = "completed_phase_2"; return $s }
        Success "FreeBuff completado"
    } catch {
        Update-State { param($s) $s.status = "failed"; $s.last_error = "FreeBuff: $_"; return $s }
        Fail "FreeBuff fallo: $_"
        exit 1
    }
    Start-Sleep -Seconds 2
}

# Phase 3: MiMoCode
if ($startPhase -le 3) {
    Log "Fase 3: MiMoCode..."
    Update-State { param($s) $s.current_phase = 3; return $s }
    
    try {
        & "$AdaptersDir\mimocode-adapter.ps1"
        if ($LASTEXITCODE -ne 0) { throw "MiMoCode fallo" }
        Update-State { param($s) $s.completed_agents += "mimocode"; $s.status = "completed"; return $s }
        Success "MiMoCode completado"
    } catch {
        Update-State { param($s) $s.status = "failed"; $s.last_error = "MiMoCode: $_"; return $s }
        Fail "MiMoCode fallo: $_"
        exit 1
    }
}

# ─── COMMIT ──────────────────────────────────────────────────

$desc = (Get-Content $StateFile -Raw | ConvertFrom-Json).task_description

if ($AutoCommit) {
    git add -A
    git commit -m "[MVGN-ORCH] $desc"
    Success "Commit automatico ejecutado"
} else {
    Write-Host ""
    Write-Host "¿Crear commit? (S/N)" -ForegroundColor Yellow
    $confirm = Read-Host
    if ($confirm -eq "S") {
        git add -A
        git commit -m "[MVGN-ORCH] $desc"
        Success "Commit creado"
    } else {
        Warn "Commit omitido - cambios sin commitear"
    }
}

Success "Pipeline completado"
Show-Status