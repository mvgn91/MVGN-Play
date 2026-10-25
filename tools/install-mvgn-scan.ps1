<#
.SYNOPSIS
    Instala mvgn-scan en el PATH del sistema para acceso desde cualquier terminal.
.DESCRIPTION
    Copia mvgn-scan.exe a un directorio en el PATH o agrega el directorio actual al PATH.
.EXAMPLE
    .\tools\install-mvgn-scan.ps1
    .\tools\install-mvgn-scan.ps1 -InstallDir "C:\tools\mvgn"
#>

param(
    [string]$InstallDir = "",
    [switch]$SkipTest
)

Write-Host ""
Write-Host "MVGN Scanner — Installer" -ForegroundColor Cyan
Write-Host "=========================" -ForegroundColor Cyan
Write-Host ""

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Split-Path -Parent $scriptDir
$sourceExe = Join-Path $projectRoot "mvgn-scan" "mvgn-scan.exe"

# Verificar que el binario existe
if (-not (Test-Path $sourceExe)) {
    Write-Host "ERROR: mvgn-scan.exe no encontrado en $sourceExe" -ForegroundColor Red
    Write-Host "Construyelo primero: cd mvgn-scan && go build -o mvgn-scan.exe ./cmd/mvgn-scan/" -ForegroundColor Yellow
    exit 1
}

Write-Host "Found: $sourceExe" -ForegroundColor Gray

# Determinar directorio de instalación
if (-not $InstallDir) {
    $installDir = Join-Path $env:USERPROFILE "mvgn-scan"
}
Write-Host "Target: $installDir" -ForegroundColor Yellow

# Crear directorio
New-Item -ItemType Directory -Force -Path $installDir | Out-Null

# Copiar binario
Copy-Item -Path $sourceExe -Destination (Join-Path $installDir "mvgn-scan.exe") -Force
Write-Host "  OK mvgn-scan.exe copiado" -ForegroundColor Green

# Agregar al PATH del usuario actual
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")

if ($userPath -notlike "*$installDir*") {
    if ($userPath) {
        $newPath = "$userPath;$installDir"
    } else {
        $newPath = $installDir
    }
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Host "  OK Directorio agregado al PATH del usuario" -ForegroundColor Green
    
    # Actualizar PATH de la sesión actual también
    $env:Path = "$env:Path;$installDir"
} else {
    Write-Host "  OK Directorio ya estaba en PATH" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Instalacion completada!" -ForegroundColor Green
Write-Host ""
Write-Host "Ahora puedes ejecutar desde cualquier terminal:" -ForegroundColor Cyan
Write-Host "  mvgn-scan.exe -dir ""C:\ruta\a\proyectos""" -ForegroundColor White
Write-Host ""
Write-Host "Ejemplo:" -ForegroundColor Cyan
Write-Host "  mvgn-scan.exe -dir ""C:\Users\$env:USERNAME\Downloads\MVGN Labs""" -ForegroundColor Gray
Write-Host ""
