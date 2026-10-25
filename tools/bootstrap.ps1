#Requires -Version 7
<#
.SYNOPSIS
    MVGN Bootstrap — instala o actualiza MVGN desde GitHub Releases.
.DESCRIPTION
    Entry point de una línea para adoptantes:

        irm https://raw.githubusercontent.com/mvgn91/mvgnlabs-starter-kit/main/tools/bootstrap.ps1 | iex

    Interfaz CNX: marco, barra de carga real y confirmación ENTER al cerrar.
    Layout consolidado (LAYOUT/2.0): TODO el motor se instala bajo <proyecto>/.mvgn/.

    Comportamiento según estado del directorio destino:
      - Sin MVGN         -> instala el último release e inicializa.
      - Misma versión    -> se niega: nada que hacer.
      - Versión anterior -> muestra changelog delta, confirma, actualiza el motor
                            preservando blackboard, docs del usuario y .gitignore.
      - Legacy pre-5.x   -> aborta con instrucciones (no migra estados antiguos).
.PARAMETER Destination
    Directorio destino. Por defecto: directorio actual.
.PARAMETER Repo
    Repositorio GitHub dueño del releases. Por defecto: mvgn91/mvgnlabs-starter-kit.
.PARAMETER PayloadZip
    Ruta a zip de payload local (testing u offline): omite la consulta a GitHub.
.PARAMETER Force
    Confirmación automática en flujo de actualización.
.PARAMETER NoPause
    No espera ENTER al final (para invocación programática: NYX update, cnx).
.EXAMPLE
    irm https://raw.githubusercontent.com/mvgn91/mvgnlabs-starter-kit/main/tools/bootstrap.ps1 | iex
    .\tools\bootstrap.ps1 -Destination C:\proyectos\nuevo
#>

param(
    [string]$Destination = ".",
    [string]$Repo = "mvgn91/mvgnlabs-starter-kit",
    [string]$PayloadZip = "",
    [switch]$Force,
    [switch]$NoPause
)

$ErrorActionPreference = "Stop"

# ═════════════════════════════════════════════════════════════
#  CNX UI — colores, marcos y barra (self-contained)
# ═════════════════════════════════════════════════════════════

$red     = "`e[38;5;196m"
$redDark = "`e[38;5;124m"
$cyan    = "`e[38;5;51m"
$green   = "`e[38;5;82m"
$yellow  = "`e[38;5;220m"
$gray    = "`e[38;5;245m"
$white   = "`e[38;5;255m"
$reset   = "`e[0m"

function Remove-Ansi([string]$Text) { $Text -replace "\e\[[0-9;]*m", "" }

function Get-FrameWidth {
    $w = 100
    try { $w = $Host.UI.RawUI.WindowSize.Width } catch { }
    if ($w -gt 116) { $w = 116 }
    if ($w -lt 94) { $w = 94 }
    return $w - 2
}

function Write-Frame {
    param([string[]]$Lines = @(), [string]$Title = "", [string]$Tag = "", [string]$Color = $red)
    $width = Get-FrameWidth
    $inner = $width - 4
    $border = [string][char]0x2500
    $edgeL  = [string][char]0x2502
    Write-Host "$Color$edgeL$("$border" * ($width - 2))$edgeL$reset"
    if ($Title) {
        $row = " $Title"
        if ($Tag) {
            $padT = " " * [Math]::Max(1, $inner - (Remove-Ansi $row).Length - (Remove-Ansi $Tag).Length - 1)
            $row += "$padT$gray$Tag$reset"
        }
        $pad = " " * [Math]::Max(0, $inner - (Remove-Ansi $row).Length)
        Write-Host "$Color$edgeL$reset$row$pad $Color$edgeL$reset"
    }
    foreach ($l in $Lines) {
        $pad = " " * [Math]::Max(0, $inner - (Remove-Ansi $l).Length - 1)
        Write-Host "$Color$edgeL$reset $l$pad$Color$edgeL$reset"
    }
    Write-Host "$Color$edgeL$("$border" * ($width - 2))$edgeL$reset"
}

function Show-Banner {
    param([string]$Mode, [string]$Dest, [string]$RepoName)
    Write-Host ""
    Write-Frame -Lines @(
        "$gray DIVISION:$reset   $white Research & Creative Systems$reset",
        "$gray DESTINO:$reset   $cyan$Dest$reset",
        $(if ($RepoName) { "$gray FUENTE:$reset    $cyan$RepoName$reset" } else { "$gray FUENTE:$reset    $cyan zip local$reset" })
    ) -Title "$red M V G N$reset // ΠΥΛΗ ΕΓΚΑΤΑΣΤΑΣΗΣ" -Tag "CRIMSON NEXUS · NODE MX-03"
    Write-Host ""
}

function Write-Bar {
    param([double]$Pct, [string]$Label = "")
    $pct = [Math]::Min(100.0, [Math]::Max(0.0, $Pct))
    $w = 36
    $filled = [int][Math]::Floor($w * $pct / 100)
    $bar = ([string][char]0x2588) * $filled
    $empty = ([string][char]0x2591) * ($w - $filled)
    $text = "  ${cyan}[$bar$empty]$reset ${gray}$([int]$pct)%$reset ${gray}$Label$reset"
    Write-Host "`r$text`e[K" -NoNewline
}

function Complete-Bar { Write-Host "" }

function Show-Seal([bool]$Ok) {
    $verdict = if ($Ok) { "$green ΕΝΤΑΞΕΙ$reset" } else { "$red ΣΦΑΛΜΑ$reset" }
    Write-Host ""
    Write-Host "  $gray ΦΥΛΑΚΗ: AEON-PROTOCOL-v1 :: ΕΠΙΒΛΕΨΙΣ$reset"
    Write-Host "  $gray Veredicto:$reset $verdict"
    Write-Host ""
}

# ═════════════════════════════════════════════════════════════
#  Fin CNX UI
# ═════════════════════════════════════════════════════════════

try {

$destPath = (Resolve-Path $Destination -ErrorAction SilentlyContinue)?.Path ?? $Destination

Write-Host ""
Show-Banner -Mode "INSTALL" -Dest $destPath -RepoName $Repo

# ─── Utilidades ──────────────────────────────────────────────

function Get-SemVer([string]$v) {
    if (-not $v) { return $null }
    $parts = ($v.TrimStart("v") -split "\.") + @("0", "0")
    try { return [version] (($parts | Select-Object -First 3) -join ".") } catch { return $null }
}

function Get-ChangelogDelta([string]$ChangelogPath, [version]$LocalVersion) {
    if (-not (Test-Path $ChangelogPath)) { return @() }
    $sections = @()
    $current = $null
    foreach ($line in Get-Content $ChangelogPath) {
        if ($line -match '^## \[?v?([0-9]+(?:\.[0-9]+){1,2})') {
            $current = @{ Version = Get-SemVer $Matches[1]; Lines = [System.Collections.Generic.List[string]]::new() }
            $sections += $current
        }
        elseif ($current) { $current.Lines.Add($line) }
    }
    if (-not $LocalVersion) { return $sections }
    return @($sections | Where-Object { $_.Version -gt $LocalVersion })
}

function Show-ChangelogDelta([object[]]$Delta) {
    if ($Delta.Count -eq 0) {
        Write-Host "  Changelog: sin entradas detalladas para los cambios pendientes." -ForegroundColor Gray
        return
    }
    foreach ($sec in $Delta) {
        Write-Host "  ## [$($sec.Version)]" -ForegroundColor White
        $sec.Lines | Where-Object { $_ -match '\S' } | ForEach-Object { Write-Host "  $_" -ForegroundColor Gray }
        Write-Host ""
    }
}

# Descarga con barra de progreso real (bytes recibidos / total)
function Get-PayloadWithProgress {
    param([string]$Url, [string]$OutFile)
    $req = [System.Net.HttpWebRequest]::Create($Url)
    $req.UserAgent = "mvgn-bootstrap"
    $req.Timeout = 30000
    $req.ReadWriteTimeout = 300000
    $resp = $req.GetResponse()
    try {
        $total = $resp.ContentLength
        $stream = $resp.GetResponseStream()
        $file = [IO.File]::Create($OutFile)
        try {
            $buffer = [byte[]]::new(65536)
            $received = 0L
            while (($read = $stream.Read($buffer, 0, $buffer.Length)) -gt 0) {
                $file.Write($buffer, 0, $read)
                $received += $read
                if ($total -gt 0) { Write-Bar (100.0 * $received / $total) "descargando payload" }
            }
            if ($total -le 0) { Write-Bar 100 "descargando payload" }
            Write-Bar 100 "payload descargado"
            Complete-Bar
        } finally { $file.Dispose() }
    } finally { $resp.Close() }
}

# ─── 1/4 Evaluar instalación existente ──────────────────────
Write-Host "${white}  [1/4]${reset} ${gray}Evaluando instalación...${reset}"

$runtimeFile = Join-Path $destPath ".mvgn\mvgn-runtime.json"
$legacyRuntime = Join-Path $destPath "mvgn-runtime.json"   # layout pre-consolidación
$localVersionText = $null
if (Test-Path $runtimeFile) {
    try { $localVersionText = (Get-Content $runtimeFile -Raw | ConvertFrom-Json).runtime_version } catch {}
}
elseif (Test-Path $legacyRuntime) {
    Write-Frame -Lines @(
        "$yellow Layout pre-consolidación detectado (mvgn-runtime.json en raíz).$reset",
        "Desde LAYOUT/2.0 todo MVGN vive en .mvgn/. Migra manualmente:",
        "${gray}mover NYX.*, MVGN.md, mvgn-runtime.json, tools/, scripts/ hacia .mvgn/${reset}"
    ) -Title " ABORT · [LEGACY]" 
    Show-Seal -Ok $false
    exit 1
}
$localVer = Get-SemVer $localVersionText
$hasMvgn = Test-Path (Join-Path $destPath ".mvgn\NYX.md")

if ($hasMvgn -and -not $localVer) {
    Write-Frame -Lines @(
        "$yellow Existe MVGN sin .mvgn/mvgn-runtime.json legible.$reset",
        "Instalación corrupta."
    ) -Title " ABORT"
    Show-Seal -Ok $false
    exit 1
}

if (-not $hasMvgn) {
    Write-Host "${gray}        Sin MVGN previo. Instalación nueva.${reset}"
}

# ─── 2/4 Resolver payload (remoto o local) ──────────────────
Write-Host "${white}  [2/4]${reset} ${gray}Resolviendo payload...${reset}"

$tmp = Join-Path ([IO.Path]::GetTempPath()) "mvgn-bootstrap-$([guid]::NewGuid())"
New-Item -ItemType Directory -Path $tmp -Force | Out-Null
$isUpdate = $false
$releaseTag = ""

try {
    if ($PayloadZip -ne "") {
        # Modo testing/offline: zip local provisto
        if (-not (Test-Path $PayloadZip)) {
            Write-Frame -Lines @("${red}Zip no encontrado: $PayloadZip${reset}") -Title " ΣΦΑΛΜΑ"
            Show-Seal -Ok $false
            exit 1
        }
        $releaseTag = "local"
        Write-Bar 10 "leyendo zip local"
        Expand-Archive -Path $PayloadZip -DestinationPath "$tmp\extracted" -Force
        Write-Bar 70 "zip extraído"
        Complete-Bar
    }
    else {
        try {
            $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" `
                -Headers @{ "User-Agent" = "mvgn-bootstrap" } -TimeoutSec 30
        } catch {
            $status = $_.Exception.Response?.StatusCode
            if ($status -eq "NotFound") {
                Write-Frame -Lines @(
                    "${red}$Repo no tiene releases publicados.${reset}",
                    "Se habilitarán con tools/package-release.ps1 al sellar una versión."
                ) -Title " ΣΦΑΛΜΑ · 404"
            } else {
                Write-Frame -Lines @("${red}No se pudo consultar GitHub Releases:$reset", "$_") -Title " ΣΦΑΛΜΑ"
            }
            Show-Seal -Ok $false
            exit 1
        }

        $remoteVer = Get-SemVer $release.tag_name
        $releaseTag = $release.tag_name

        # Ramas de decisión ANTES de descargar nada
        if ($hasMvgn -and $localVer -eq $remoteVer) {
            Write-Frame -Lines @(
                "${green}MVGN $($release.tag_name) ya está instalado. Nada que hacer.$reset"
            ) -Title " OK"
            Show-Seal -Ok $true
            exit 0
        }
        if ($hasMvgn -and $localVer -gt $remoteVer) {
            Write-Frame -Lines @(
                "${yellow}Local ($localVersionText) es más reciente que el remoto ($($release.tag_name)).$reset",
                "Abortando."
            ) -Title " ABORT"
            Show-Seal -Ok $true
            exit 0
        }

        $asset = $release.assets | Where-Object { $_.name -like "mvgn-*.zip" } | Select-Object -First 1
        if (-not $asset) {
            Write-Frame -Lines @("${red}El release $($release.tag_name) no contiene mvgn-*.zip.${reset}") -Title " ΣΦΑΛΜΑ"
            Show-Seal -Ok $false
            exit 1
        }

        # UNA descarga, UNA extracción: sirve al changelog Y al despliegue (fix H-6)
        $zip = Join-Path $tmp $asset.name
        Get-PayloadWithProgress -Url $asset.browser_download_url -OutFile $zip
        Write-Bar 72 "extrayendo payload"
        Expand-Archive -Path $zip -DestinationPath "$tmp\extracted" -Force
        Write-Bar 78 "payload extraído"
        Complete-Bar
    }

    $srcRoot = (Get-ChildItem "$tmp\extracted" -Directory | Select-Object -First 1).FullName

    # ─── Flujo de actualización: changelog + confirmación ───
    if ($hasMvgn) {
        $isUpdate = $true
        Write-Host ""
        Write-Host "${yellow}  Actualización disponible:${reset} ${white}$localVersionText${reset} ${gray}->${reset} ${green}$releaseTag${reset}"
        Write-Host ""
        Show-ChangelogDelta (Get-ChangelogDelta (Join-Path $srcRoot "templates\docs\04_changelog.md") $localVer)

        if ($Force) {
            Write-Host "  [-Force] Confirmación automática." -ForegroundColor Gray
        } else {
            $answer = Read-Host "  ¿Actualizar ahora? (S/n)"
            if ($answer -match '^[nN]') {
                Write-Host "  Actualización cancelada." -ForegroundColor Gray
                Show-Seal -Ok $true
                exit 0
            }
        }
    }

    # ─── 3/4 Desplegar (todo bajo .mvgn/) ───────────────────
    Write-Host "${white}  [3/4]${reset} ${gray}Desplegando payload...${reset}"
    New-Item -ItemType Directory -Path (Join-Path $destPath ".mvgn") -Force | Out-Null

    if ($isUpdate) {
        # Motor: sobrescribe dentro de .mvgn/. Memoria del proyecto: se preserva.
        Copy-Item "$srcRoot\NYX.*"   "$destPath\.mvgn\" -Force
        Copy-Item "$srcRoot\MVGN.md" "$destPath\.mvgn\" -Force
        foreach ($dir in "tools", "scripts", "templates") {
            if (Test-Path "$srcRoot\$dir") {
                Copy-Item "$srcRoot\$dir" "$destPath\.mvgn\" -Recurse -Force
            }
        }
        Write-Host "${gray}        Motor actualizado. Blackboard, docs y .gitignore preservados.${reset}"
    } else {
        Copy-Item "$srcRoot\*" "$destPath\.mvgn\" -Recurse -Force
    }
    Write-Bar 88 "motor desplegado"
    Complete-Bar

    # ─── 4/4 Inicializar y sincronizar manifiesto ───────────
    Write-Host "${white}  [4/4]${reset} ${gray}Sincronizando e inicializando...${reset}"

    $initScript = Join-Path $destPath ".mvgn\tools\mvgn-init.ps1"
    $initOut = & (Get-Process -Id $PID).Path -NoProfile -File $initScript 2>&1
    if ($LASTEXITCODE -ne 0) {
        $initOut | ForEach-Object { Write-Host "  $_" }
        throw "mvgn-init.ps1 falló (exit $LASTEXITCODE)"
    }
    $initOut | Where-Object { $_ -match '^\s*(\[|\+|=)' } | ForEach-Object { Write-Host "${gray}  $_${reset}" }

    if (Get-SemVer $releaseTag) {
        # H-13/H-14: el manifiesto refleja SIEMPRE la versión del tag real,
        # tanto en instalación fresca como en actualización. En modo
        # -PayloadZip ("local") queda intacto para no romper el próximo bootstrap.
        $rt = Get-Content $runtimeFile -Raw | ConvertFrom-Json
        $rt.runtime_version = $releaseTag.TrimStart("v")
        $rt.schema_version  = $releaseTag.TrimStart("v")
        $rt | ConvertTo-Json -Depth 10 | Set-Content $runtimeFile -Encoding UTF8
    } else {
        Write-Host "${gray}        Modo local: versión del manifiesto sin modificar.${reset}"
    }

    Write-Bar 100 "instalación completa"
    Complete-Bar

    $action = if ($isUpdate) { "actualizado" } else { "instalado" }
    Write-Host ""
    Write-Frame -Lines @(
        "${green}MVGN $releaseTag $action.$reset",
        "",
        "${gray}Entry point:${reset} abre ${cyan}.mvgn/NYX.md${reset} en tu sesión de IA.",
        "${gray}Actualizar :${reset} ${cyan}.\.mvgn\tools\bootstrap.ps1${reset}"
    ) -Title " ΤΕΛΟΣ · INSTALACIÓN"
    Show-Seal -Ok $true
    Write-Host ""

} finally {
    Remove-Item $tmp -Recurse -Force -ErrorAction SilentlyContinue
}

} catch {
    Write-Host ""
    Write-Frame -Lines @("${red}$_${reset}") -Title " ΣΦΑΛΜΑ · INSTALACIÓN INTERRUMPIDA"
    Show-Seal -Ok $false
    exit 1
} finally {
    # El instalador SIEMPRE espera ENTER, instale MVGN o no (requisito UX).
    if (-not $NoPause -and -not [Console]::IsOutputRedirected) {
        Write-Host "  ${gray}ΤΕΛΟΣ${reset} ${white}- presiona ENTER para cerrar${reset}" -NoNewline
        try { [void](Read-Host " ") } catch { }
    }
}