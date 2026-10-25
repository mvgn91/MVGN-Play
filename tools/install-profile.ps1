#Requires -Version 7
<#
.SYNOPSIS
    Instala la integración de MVGN en el perfil de PowerShell del usuario.
.DESCRIPTION
    Agrega un bloque idempotente (# >>> MVGN BLOCK >>>) al perfil
    CurrentUserAllHosts con:
      - Función global mvgn: resuelve el NYX.ps1 más cercano hacia arriba
        desde el directorio actual (como git resuelve .git).
      - Sin proyecto MVGN cerca: ofrece bootstrap o informa cómo inicializar.
      - Tab-completion para los modos de NYX.ps1.
    No duplica contenido si ya existe el bloque.
.EXAMPLE
    .\tools\install-profile.ps1
    .\tools\install-profile.ps1 -WhatIf
#>

param([switch]$WhatIf)

$ErrorActionPreference = "Stop"

$MarkerStart = "# >>> MVGN BLOCK v2 >>>"
$MarkerEnd   = "# <<< MVGN BLOCK v2 <<<"

$Block = @'

# >>> MVGN BLOCK v2 >>>
# MVGN - Multi-Agent Governance Framework (generado por tools/install-profile.ps1)
# NO editar manualmente: reinstalar con .\tools\install-profile.ps1

function mvgn {
    <#
    .SYNOPSIS
        MVGN CLI - ejecuta NYX.ps1 del proyecto MVGN mas cercano hacia arriba.
    .EXAMPLE
        mvgn detect
        mvgn update
        mvgn observer -Web -Port 8080
    #>
    $dir = (Get-Location).Path
    $nyx = $null
    while ($dir -and $dir -ne [IO.Path]::GetPathRoot($dir)) {
        # LAYOUT/2.0: motor consolidado en .mvgn/
        $candidate = Join-Path $dir ".mvgn\NYX.ps1"
        if (Test-Path $candidate) { $nyx = $candidate; break }
        # Layout legacy pre-consolidacion (v5.3-): NYX.ps1 en raiz
        $candidate = Join-Path $dir "NYX.ps1"
        if (Test-Path $candidate) { $nyx = $candidate; break }
        $dir = Split-Path -Parent $dir
    }

    if (-not $nyx) {
        Write-Host "  [MVGN] No hay un proyecto MVGN aqui ni en directorios padres." -ForegroundColor Yellow
        Write-Host ""
        Write-Host "  Instalar MVGN en este directorio:" -ForegroundColor Cyan
        Write-Host '    irm https://raw.githubusercontent.com/mvgn91/mvgnlabs-starter-kit/main/tools/bootstrap.ps1 | iex' -ForegroundColor White
        return
    }

    & (Get-Process -Id $PID).Path -NoProfile -File $nyx @args
}

# Tab-completion de modos NYX
Register-ArgumentCompleter -CommandName mvgn -Native -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)
    $modes = @('detect', 'init', 'ready', 'upgrade', 'update', 'observer', 'collab')
    $modes | Where-Object { $_ -like "$wordToComplete*" } |
        ForEach-Object { [System.Management.Automation.CompletionResult]::new($_) }
}

# <<< MVGN BLOCK v2 <<<
'@

$profilePath = $PROFILE.CurrentUserAllHosts
$existing = if (Test-Path $profilePath) { Get-Content $profilePath -Raw } else { "" }

Write-Host ""
Write-Host "MVGN Profile Integration" -ForegroundColor Cyan
Write-Host "========================" -ForegroundColor Cyan

if ($existing -match [regex]::Escape($MarkerStart)) {
    if ($existing -match [regex]::Escape($Block)) {
        Write-Host "[OK] El bloque MVGN ya está instalado y actualizado." -ForegroundColor Green
        Write-Host "     Perfil: $profilePath" -ForegroundColor Gray
        exit 0
    }
    # Versión anterior del bloque: reemplazar
    if ($WhatIf) {
        Write-Host "[WhatIf] Se reemplazaría el bloque MVGN existente en $profilePath" -ForegroundColor Yellow
        exit 0
    }
    $pattern = [regex]::Escape($MarkerStart) + "[\s\S]*?" + [regex]::Escape($MarkerEnd)
    $updated = $existing -replace $pattern, $Block.TrimEnd()
    Set-Content $profilePath $updated -Encoding UTF8
    Write-Host "[OK] Bloque MVGN actualizado." -ForegroundColor Green
} else {
    if ($WhatIf) {
        Write-Host "[WhatIf] Se agregaría el bloque MVGN a $profilePath" -ForegroundColor Yellow
        exit 0
    }
    $parent = Split-Path $profilePath -Parent
    if (-not (Test-Path $parent)) { New-Item -ItemType Directory -Path $parent -Force | Out-Null }
    Add-Content $profilePath "`r`n$Block" -Encoding UTF8
    Write-Host "[OK] Bloque MVGN agregado al perfil." -ForegroundColor Green
}

Write-Host "     Perfil: $profilePath" -ForegroundColor Gray
Write-Host ""
Write-Host "Recargue su sesion o ejecute:" -ForegroundColor White
Write-Host "  . `$PROFILE" -ForegroundColor Yellow
Write-Host ""
Write-Host "Uso:" -ForegroundColor Cyan
Write-Host "  mvgn detect      # dentro de un proyecto MVGN" -ForegroundColor Gray
Write-Host "  mvgn update      # actualizar vía GitHub Releases" -ForegroundColor Gray
Write-Host "  mvgn             # fuera de proyecto: muestra cómo instalar" -ForegroundColor Gray
