<#
.SYNOPSIS
    Instala el MVGN Observer en el sistema — un solo comando.
.DESCRIPTION
    Construye el binario, lo copia a un directorio en PATH, agrega un
    script de auto-arranque y registra una tarea programada opcional.
.EXAMPLE
    .\tools\install-observer.ps1
    .\tools\install-observer.ps1 -InstallDir "C:\tools\mvgn" -RegisterTask
#>

param(
    [string]$InstallDir = "",
    [switch]$RegisterTask,
    [switch]$SkipBuild
)

$MVGN_CRIMSON = "214;0;28"

function Log($msg) { Write-Host "[$(Get-Date -Format 'HH:mm:ss')] $msg" -ForegroundColor Cyan }
function Success($msg) { Write-Host "  OK $msg" -ForegroundColor Green }
function Warn($msg) { Write-Host "  WARN $msg" -ForegroundColor Yellow }
function Fail($msg) { Write-Host "  ERROR $msg" -ForegroundColor Red; exit 1 }

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Split-Path -Parent $scriptDir
$scanDir = Join-Path $projectRoot "mvgn-scan"

Write-Host ""
Write-Host "MVGN Observer — Installer" -ForegroundColor Cyan
Write-Host "==========================" -ForegroundColor Cyan
Write-Host ""

# ─── Build ────────────────────────────────────────────────────
if (-not $SkipBuild) {
    $binExist = Test-Path (Join-Path $scanDir "mvgn-scan.exe")
    if (-not $binExist) {
        Log "Construyendo mvgn-scan.exe desde Go..."
        $build = & go build -o (Join-Path $scanDir "mvgn-scan.exe") ./cmd/mvgn-scan/ 2>&1
        if ($LASTEXITCODE -ne 0) {
            Fail "Build falló: $build"
        }
        Success "mvgn-scan.exe construido"
    } else {
        Log "mvgn-scan.exe ya existe — usa -SkipBuild para omitir"
    }
}

$sourceExe = Join-Path $scanDir "mvgn-scan.exe"
if (-not (Test-Path $sourceExe)) {
    Fail "mvgn-scan.exe no encontrado en $sourceExe"
}

# ─── Install dir ──────────────────────────────────────────────
if (-not $InstallDir) {
    $installDir = Join-Path $env:USERPROFILE "mvgn-scan"
}
New-Item -ItemType Directory -Force -Path $installDir | Out-Null

Copy-Item -Path $sourceExe -Destination (Join-Path $installDir "mvgn-scan.exe") -Force
Success "mvgn-scan.exe copiado a $installDir"

# ─── PATH ─────────────────────────────────────────────────────
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    $newPath = if ($userPath) { "$userPath;$installDir" } else { $installDir }
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    $env:Path = "$env:Path;$installDir"
    Success "Directorio agregado al PATH del usuario"
} else {
    Log "Directorio ya estaba en PATH"
}

# ─── Startup script ───────────────────────────────────────────
$startupDir = Join-Path $env:APPDATA "Microsoft\Windows\Start Menu\Programs\Startup"
$startupScript = Join-Path $startupDir "mvgn-observer.ps1"

$startupContent = @"
# MVGN Observer — Auto-start
Start-Process -WindowStyle Hidden -FilePath "$installDir\mvgn-scan.exe" -ArgumentList "-watch","-dir","$projectRoot"
"@

try {
    $startupContent | Set-Content $startupScript -Force
    Success "Startup script creado en $startupScript"
} catch {
    Warn "No se pudo crear startup script: $_"
}

# ─── Scheduled task (opcional) ────────────────────────────────
if ($RegisterTask) {
    $taskName = "MVGN-Observer"
    $action = New-ScheduledTaskAction -Execute "$installDir\mvgn-scan.exe" -Argument "-watch -dir `"$projectRoot`" -interval 10"
    $trigger = New-ScheduledTaskTrigger -AtStartup
    $principal = New-ScheduledTaskPrincipal -UserId "SYSTEM" -LogonType ServiceAccount -RunLevel Highest
    try {
        Register-ScheduledTask -TaskName $taskName -Action $action -Trigger $trigger -Principal $principal -Force | Out-Null
        Success "Tarea programada '$taskName' registrada (inicio con Windows)"
    } catch {
        Warn "No se pudo registrar tarea programada: $_ (ejecuta como Administrador)"
    }
}

# ─── Info ─────────────────────────────────────────────────────
Write-Host ""
Write-Host "MVGN Observer instalado!" -ForegroundColor Green
Write-Host ""
Write-Host "Comandos disponibles:" -ForegroundColor Cyan
Write-Host "  mvgn-scan.exe -watch                      # Terminal observer" -ForegroundColor White
Write-Host "  mvgn-scan.exe -observe -port 8080         # Dashboard web" -ForegroundColor White
Write-Host "  mvgn-scan.exe -watch -observe             # Terminal + web" -ForegroundColor White
Write-Host ""
Write-Host "O desde este proyecto:" -ForegroundColor Cyan
Write-Host "  .\tools\observe.bat                        # Double-click" -ForegroundColor White
Write-Host ""
