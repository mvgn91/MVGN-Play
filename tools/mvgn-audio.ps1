#Requires -Version 7
<#
.SYNOPSIS
    MVGN Audio Ingestion — punto de entrada único.
.DESCRIPTION
    Envoltorio delgado sobre scripts/audio/transcribe.ps1 (pipeline probado:
    Faster-Whisper GPU/CPU, cola batch, registro processed.json, progreso vivo).

    Disponible EN CUALQUIER momento del ciclo de vida del proyecto:
      - INIT:      alimenta el context.md que genera el PRD inicial
      - READY:     alimenta ciclos de features, correcciones y refactors
                   (los contextos nuevos quedan en .mvgn/audio/ y se registran
                    como artefactos del blackboard)

    Uso diario:
        mvgn-audio                # menú interactivo de audios pendientes
        mvgn-audio -Batch         # procesa TODA la fila pendiente
        mvgn-audio -AudioFile x.mp3 -Model turbo
.EXAMPLE
    .\tools\mvgn-audio.ps1 -Batch -Device cuda -ComputeType int8
#>

param(
    [string]$AudioFile = "",
    [string]$Model = "medium",
    [ValidateSet("faster-whisper", "whisper")]
    [string]$Engine = "faster-whisper",
    # O-6: por defecto se detecta GPU (nvidia-smi); sin GPU cae a CPU sin fallar
    [string]$Device = "$(if (Get-Command nvidia-smi -ErrorAction SilentlyContinue) { 'cuda' } else { 'cpu' })",
    [string]$ComputeType = "",
    [switch]$Batch,
    [switch]$Force,
    [switch]$WhatIf
)

# compute_type auto: int8 funciona bien en cuda (poca VRAM) y cpu
if (-not $ComputeType) { $ComputeType = "int8" }

$ErrorActionPreference = "Stop"

$pipeline = Join-Path $PSScriptRoot "..\scripts\audio\transcribe.ps1"
if (-not (Test-Path $pipeline)) {
    Write-Host "[FAIL] No se encontró scripts/audio/transcribe.ps1" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "MVGN Audio Ingestion" -ForegroundColor Cyan
Write-Host "--------------------"
Write-Host "  Motor: $Engine | Modelo: $Model | Device: $Device/$ComputeType" -ForegroundColor Gray
Write-Host ""

& $pipeline `
    -Path (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path `
    -AudioFile $AudioFile `
    -Model $Model `
    -Engine $Engine `
    -Device $Device `
    -ComputeType $ComputeType `
    -Batch:$Batch `
    -Force:$Force `
    -WhatIf:$WhatIf

exit $LASTEXITCODE
