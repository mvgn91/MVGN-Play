<#
.SYNOPSIS
    Instala mvgn-query.exe en el proyecto MVGN.
.DESCRIPTION
    Compila la QUERY Dashboard desde mvgn-scan/ y copia el binario a bin/.
    Luego de ejecutar, usa: .\bin\mvgn-query.exe --watch
.EXAMPLE
    .\tools\install-query.ps1
    .\tools\install-query.ps1 -OutputDir ".\bin"
#>

param(
    [string]$OutputDir = (Join-Path (Split-Path $PSScriptRoot -Parent) "bin")
)

Write-Host "MVGN QUERY Dashboard Installer" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host ""

# Resolve mvgn-scan directory
$ScanDir = Join-Path (Split-Path $PSScriptRoot -Parent) "mvgn-scan"

if (-not (Test-Path $ScanDir)) {
    Write-Host "ERROR: mvgn-scan directory not found at $ScanDir" -ForegroundColor Red
    Write-Host "Make sure you cloned the full MVGN starter kit repository." -ForegroundColor Yellow
    exit 1
}

# Create output directory
if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
}

Write-Host "Building mvgn-query.exe..." -ForegroundColor Yellow
Write-Host "  Source: $ScanDir\cmd\mvgn-query\" -ForegroundColor Gray

$OutputPath = Join-Path $OutputDir "mvgn-query.exe"

# Build
$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"

Push-Location $ScanDir
go build -ldflags="-s -w" -o $OutputPath ./cmd/mvgn-query/
$exitCode = $LASTEXITCODE
Pop-Location

if ($exitCode -eq 0) {
    $size = (Get-Item $OutputPath).Length / 1MB
    Write-Host ""
    Write-Host "  ✓ mvgn-query.exe built ($([math]::Round($size, 2)) MB)" -ForegroundColor Green
    Write-Host "  Location: $OutputPath" -ForegroundColor Green
    Write-Host ""
    Write-Host "Usage:" -ForegroundColor Cyan
    Write-Host "  .\bin\mvgn-query.exe --watch" -ForegroundColor White
    Write-Host "  .\bin\mvgn-query.exe --watch --interval 2s --project ." -ForegroundColor White
    Write-Host "  .\bin\mvgn-query.exe --version" -ForegroundColor White
    Write-Host ""
    Write-Host "MVGN QUERY Dashboard installed successfully." -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "  ✗ Build failed with exit code $exitCode" -ForegroundColor Red
    Write-Host "  Ensure Go 1.21+ is installed." -ForegroundColor Yellow
    exit 1
}
