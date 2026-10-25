<#
.SYNOPSIS
    Compila mvgn-scan y mvgn-query para múltiples plataformas.
.DESCRIPTION
    Genera binarios estáticos para Windows, Linux y macOS en el directorio bin/ raíz.
.EXAMPLE
    .\build.ps1
    .\build.ps1 -OutputDir "C:\dist"
#>

param(
    [Parameter(Mandatory = $false)]
    [string]$OutputDir = "..\bin"
)

Write-Host "MVGN Scanner Build" -ForegroundColor Cyan
Write-Host "==================" -ForegroundColor Cyan
Write-Host ""

# Crear directorio de salida
if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir | Out-Null
}

# Plataformas objetivo
$platforms = @(
    @{ GOOS = "windows"; GOARCH = "amd64"; Suffix = ".exe" },
    @{ GOOS = "linux"; GOARCH = "amd64"; Suffix = "" },
    @{ GOOS = "darwin"; GOARCH = "amd64"; Suffix = "" },
    @{ GOOS = "darwin"; GOARCH = "arm64"; Suffix = "" }
)

# Compilar para cada plataforma
$components = @(
    @{ Name = "mvgn-scan";    Path = "./cmd/mvgn-scan/" },
    @{ Name = "mvgn-query";   Path = "./cmd/mvgn-query/" }
)

foreach ($platform in $platforms) {
    $env:GOOS = $platform.GOOS
    $env:GOARCH = $platform.GOARCH
    
    foreach ($comp in $components) {
        $outputName = "$($comp.Name)-$($platform.GOOS)-$($platform.GOARCH)$($platform.Suffix)"
        $outputPath = Join-Path $OutputDir $outputName
        
        Write-Host "Compilando $($comp.Name) para $($platform.GOOS)/$($platform.GOARCH)..." -ForegroundColor Yellow
        
        $env:CGO_ENABLED = "0"
        go build -ldflags="-s -w" -o $outputPath $($comp.Path)
        
        if ($LASTEXITCODE -eq 0) {
            $size = (Get-Item $outputPath).Length / 1MB
            Write-Host "  ✓ $outputName ($([math]::Round($size, 2)) MB)" -ForegroundColor Green

            # En Windows, también copiar como nombre simple para el repo
            if ($platform.GOOS -eq "windows" -and $platform.GOARCH -eq "amd64") {
                $simplePath = Join-Path $OutputDir "$($comp.Name).exe"
                Copy-Item $outputPath $simplePath -Force
                Write-Host "    → $($comp.Name).exe (simplified)" -ForegroundColor Gray
            }
        } else {
            Write-Host "  ✗ Error compilando $outputName" -ForegroundColor Red
        }
    }
}

Write-Host ""
Write-Host "Build completado" -ForegroundColor Green
Write-Host "Binarios en: $OutputDir"