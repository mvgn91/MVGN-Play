# tools/validate-entry-points.ps1
# Valida que NYX.md y NYX.ps1 compartan los mismos modos de operación
# Se ejecuta automáticamente antes de commits que toquen entry points
# Versión: v1.0 — MVGN v5.1

param(
    [switch]$Fix,
    [switch]$Quiet
)

$nyxMd = "NYX.md"
$nyxPs1 = "NYX.ps1"
$errors = @()
$warnings = @()

function Log($msg) {
    if (-not $Quiet) { Write-Host $msg -ForegroundColor Gray }
}

function Ok($msg) {
    if (-not $Quiet) { Write-Host "[OK] $msg" -ForegroundColor Green }
}

function Err($msg) {
    if (-not $Quiet) { Write-Host "[FAIL] $msg" -ForegroundColor Red }
}

function Warn($msg) {
    if (-not $Quiet) { Write-Host "[WARN] $msg" -ForegroundColor Yellow }
}

# ─── VERIFICAR EXISTENCIA ────────────────────────────────────

Log "Verificando existencia de entry points..."

if (-not (Test-Path $nyxMd)) { $errors += "$nyxMd no encontrado" }
if (-not (Test-Path $nyxPs1)) { $errors += "$nyxPs1 no encontrado" }

if ($errors.Count -gt 0) {
    $errors | ForEach-Object { Err $_ }
    exit 1
}

Ok "$nyxMd encontrado"
Ok "$nyxPs1 encontrado"

# ─── VERIFICAR MODOS ─────────────────────────────────────────

Log "Verificando modos de operación..."
$md = Get-Content $nyxMd -Raw
$ps1 = Get-Content $nyxPs1 -Raw

$modos = @("INIT", "READY", "UPGRADE", "COLLAB")
$modosOk = 0

foreach ($modo in $modos) {
    $inMd = $md -match "\b$modo\b"
    $inPs1 = $ps1 -match "\b$modo\b"
    
    if (-not $inMd) {
        $errors += "Falta modo '$modo' en $nyxMd"
    } elseif (-not $inPs1) {
        $errors += "Falta modo '$modo' en $nyxPs1"
    } else {
        $modosOk++
        Ok "Modo '$modo' presente en ambos entry points"
    }
}

# ─── VERIFICAR VERSIÓN ───────────────────────────────────────

Log "Verificando versiones..."
$versionPattern = 'v\d+\.\d+\.\d+'

$mdVersions = [regex]::Matches($md, $versionPattern) | ForEach-Object { $_.Value }
$ps1Versions = [regex]::Matches($ps1, $versionPattern) | ForEach-Object { $_.Value }

if ($mdVersions.Count -gt 0 -and $ps1Versions.Count -gt 0) {
    $mdVersion = $mdVersions[0]
    $ps1Version = $ps1Versions[0]
    if ($mdVersion -eq $ps1Version) {
        Ok "Versiones sincronizadas: $mdVersion"
    } else {
        $errors += "Versiones diferentes: NYX.md=$mdVersion vs NYX.ps1=$ps1Version"
    }
} else {
    $warnings += "No se pudo determinar versión en ambos entry points"
}

# ─── VERIFICAR ESTRUCTURA DE MODO ────────────────────────────

Log "Verificando estructura de Modos..."

# NYX.md debe mencionar tabla de modos
if ($md -match '\|.*INIT.*\|') {
    Ok "NYX.md tiene tabla de modos"
} else {
    $warnings += "NYX.md podría no tener tabla de modos explícita"
}

# NYX.ps1 debe tener ValidateSet con modos
if ($ps1 -match 'ValidateSet\("detect",\s*"init",\s*"ready",\s*"upgrade",\s*"update",\s*"collab",\s*"observer"\)') {
    Ok "NYX.ps1 tiene ValidateSet de modos"
} else {
    $errors += "NYX.ps1 no tiene ValidateSet correcto de modos"
}

# ─── VERIFICAR SINCRONIZACIÓN DE MODO COLLAB ─────────────────

# Ambos deben mencionar roles de COLLAB
$mdHasRoles = $md -match "Auditor" -and $md -match "Desarrollador" -and $md -match "Revisor"
$ps1HasRoles = $ps1 -match "Auditor" -and $ps1 -match "Desarrollador" -and $ps1 -match "Revisor"

if ($mdHasRoles) { Ok "NYX.md menciona roles COLLAB" }
else { $warnings += "NYX.md podría no mencionar roles COLLAB" }

if ($ps1HasRoles) { Ok "NYX.ps1 implementa roles COLLAB" }
else { $errors += "NYX.ps1 no implementa roles COLLAB" }

# ─── REPORTE FINAL ───────────────────────────────────────────

Write-Host ""
Write-Host "═══════════════════════════════════════════════" -ForegroundColor Cyan

if ($errors.Count -eq 0 -and $warnings.Count -eq 0) {
    Ok "NYX.md y NYX.ps1 están COMPLETAMENTE sincronizados."
    Write-Host ""
    Log "Modos verificados: $modosOk/$($modos.Count)"
    exit 0
}

if ($errors.Count -eq 0) {
    Warn "NYX.md y NYX.ps1 sincronizados con advertencias:"
    $warnings | ForEach-Object { Warn "  - $_" }
    Write-Host ""
    Log "Modos verificados: $modosOk/$($modos.Count)"
    exit 0
}

Err "Entry points DESINCRONIZADOS:"
$errors | ForEach-Object { Err "  - $_" }
if ($warnings.Count -gt 0) {
    $warnings | ForEach-Object { Warn "  - $_" }
}

Write-Host ""
Log "Modos verificados: $modosOk/$($modos.Count)"
exit 1
