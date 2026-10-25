#Requires -Version 7
<#
.SYNOPSIS
    MVGN Package Release — arma el zip de payload según el manifiesto.
.DESCRIPTION
    Lee templates/payload-manifest.json, copia EXACTAMENTE los archivos listados
    a un staging limpio y genera mvgn-v<VERSION>.zip.
    El desarrollo del lab (archive/, reports/, events/, telemetría, .git) NUNCA
    entra al release: el zip se arma por manifiesto, no copiando el repo.
    Al terminar valida que el contenido del zip coincida 1:1 con el manifiesto.
.PARAMETER Version
    Versión del release (tag). Por defecto: runtime_version de mvgn-runtime.json.
.PARAMETER OutDir
    Directorio de salida del zip. Por defecto: ./release
.EXAMPLE
    .\tools\package-release.ps1 -Version v5.3.0
#>

param(
    [string]$Version = "",
    [string]$OutDir = "./release"
)

$ErrorActionPreference = "Stop"
$RepoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")

if (-not $Version) {
    $rt = Get-Content (Join-Path $RepoRoot "mvgn-runtime.json") -Raw | ConvertFrom-Json
    $Version = "v$($rt.runtime_version)"
}
$Version = $Version.TrimStart("v")
$Tag = "v$Version"

Write-Host ""
Write-Host "MVGN Package Release - $Tag" -ForegroundColor Cyan
Write-Host "============================" -ForegroundColor Cyan

# ─── 1/4 Cargar manifiesto ───────────────────────────────────
Write-Host "[1/4] Cargando manifiesto..." -ForegroundColor Cyan

$manifestPath = Join-Path $RepoRoot "templates\payload-manifest.json"
$manifest = Get-Content $manifestPath -Raw | ConvertFrom-Json

# ─── 2/4 Staging por manifiesto ──────────────────────────────
Write-Host "[2/4] Armando staging..." -ForegroundColor Cyan

$staging = Join-Path ([IO.Path]::GetTempPath()) "mvgn-release-$([guid]::NewGuid())"
# Carpeta raíz dentro del zip: bootstrap.ps1 la normaliza al extraer
$payloadRoot = Join-Path $staging "mvgn-$Version"
New-Item -ItemType Directory -Path $payloadRoot -Force | Out-Null
$missing = @()

foreach ($item in $manifest.payload) {
    if ($item.pending) { continue }
    $src = Join-Path $RepoRoot $item.path

    if ($item.recursive -or (Test-Path $src -PathType Container)) {
        if (-not (Test-Path $src)) { $missing += $item.path; continue }
        $dst = Join-Path $payloadRoot $item.dest
        New-Item -ItemType Directory -Path $dst -Force | Out-Null
        Copy-Item "$src\*" $dst -Recurse -Force
    }
    else {
        if (-not (Test-Path $src)) { $missing += $item.path; continue }
        $dst = Join-Path $payloadRoot $item.dest
        New-Item -ItemType Directory -Path (Split-Path $dst -Parent) -Force | Out-Null
        Copy-Item $src $dst -Force
    }
}

if ($missing.Count -gt 0) {
    Write-Host "[FAIL] Archivos del manifiesto no encontrados:" -ForegroundColor Red
    $missing | ForEach-Object { Write-Host "       - $_" -ForegroundColor Yellow }
    Remove-Item $staging -Recurse -Force -ErrorAction SilentlyContinue
    exit 1
}

# ─── 3/4 Empaquetar ──────────────────────────────────────────
Write-Host "[3/4] Empaquetando..." -ForegroundColor Cyan

New-Item -ItemType Directory -Path $OutDir -Force | Out-Null
$zipPath = Join-Path $OutDir "mvgn-$Version.zip"
if (Test-Path $zipPath) { Remove-Item $zipPath -Force }

# Carpeta raíz mvgn-<versión>/ dentro del zip
Compress-Archive -Path $payloadRoot -DestinationPath $zipPath -CompressionLevel Optimal
$zipSize = "{0:N1} KB" -f ((Get-Item $zipPath).Length / 1KB)

# ─── 4/4 Validación: zip == manifiesto ──────────────────────
Write-Host "[4/4] Validando contenido contra manifiesto..." -ForegroundColor Cyan

Add-Type -AssemblyName System.IO.Compression.FileSystem
$zip = [IO.Compression.ZipFile]::OpenRead($zipPath)
try {
    $inZip = $zip.Entries |
        Where-Object { $_.Name -ne "" } |
        ForEach-Object { $_.FullName -replace "^mvgn-$Version/", "" } |
        Sort-Object

    $expected = @(
        foreach ($item in $manifest.payload) {
            if ($item.pending) { continue }
            $src = (Join-Path $RepoRoot $item.path).TrimEnd('\', '/')
            if ($item.recursive -or (Test-Path $src -PathType Container)) {
                Get-ChildItem $src -Recurse -File | ForEach-Object {
                    ($item.dest.TrimEnd('/') + '/' + $_.FullName.Substring($src.Length + 1)) -replace "\\", "/"
                }
            } else { $item.dest -replace "\\", "/" }
        }
    ) | Sort-Object | Select-Object -Unique

    $diff = Compare-Object $inZip ($expected | Select-Object -Unique)
    if ($diff) {
        Write-Host "[FAIL] El zip NO coincide con el manifiesto:" -ForegroundColor Red
        $diff | ForEach-Object { Write-Host "       $($_.SideIndicator) $($_.InputObject)" -ForegroundColor Yellow }
        exit 1
    }
} finally { $zip.Dispose() }

Remove-Item $staging -Recurse -Force -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "[OK] $($zipPath) ($zipSize)" -ForegroundColor Green
Write-Host "     $($inZip.Count) archivos - coincide 1:1 con payload-manifest.json" -ForegroundColor Gray
Write-Host ""
Write-Host "Publicar:" -ForegroundColor Cyan
Write-Host "  gh release create $Tag '$zipPath' --title 'MVGN $Tag' --notes-file <notas>" -ForegroundColor White
Write-Host ""
