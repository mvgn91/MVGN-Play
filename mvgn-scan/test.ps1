<#
.SYNOPSIS
    Ejecuta pruebas unitarias de mvgn-scan.
.DESCRIPTION
    Ejecuta todos los tests del proyecto.
.EXAMPLE
    .\test.ps1
#>

Write-Host "MVGN Scanner Tests" -ForegroundColor Cyan
Write-Host "==================" -ForegroundColor Cyan
Write-Host ""

# Ejecutar tests
Write-Host "Ejecutando tests..." -ForegroundColor Yellow
go test ./... -v

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "Todos los tests pasaron" -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "Algunos tests fallaron" -ForegroundColor Red
    exit 1
}