<#
.SYNOPSIS
    MVGN Audio Ingestion — Detect, transcribe, and build context from audio files.
.DESCRIPTION
    Extiende el modo INIT de MVGN con capacidad de ingesta de audio.
    1. Detecta archivos de audio (.mp3, .wav, .m4a, .flac, .ogg)
    2. Filtra los que ya fueron extraídos (cola de pendientes)
    3. Transcribe con Faster-Whisper (uno o toda la fila con -Batch)
    4. Genera context.md estructurado
    5. Prepara el pipeline para generación de PRD

    Pipeline: AUDIO_DISCOVERY → TRANSCRIBING → CONTEXT_BUILDING → CONTEXT_READY

    MODO FILA (-Batch): detecta todos los audios sin extraer, los lista y los
    transcribe todos en secuencia con UNA sola ejecución. El registro de lo ya
    extraído se mantiene en $AudioDir\processed.json (y por output completo en
    $AudioDir\<base>\). Se puede sembrar processed.json con bases ya procesadas
    para que la fila arranque solo con los pendientes.
.PARAMETER Path
    Ruta del proyecto (default: directorio actual)
.PARAMETER AudioFile
    Ruta específica a un archivo de audio (skip detection)
.PARAMETER Model
    Modelo (tiny, base, small, medium, large-v3, turbo)
.PARAMETER Engine
    Motor de transcripción (faster-whisper, whisper)
.PARAMETER Device
    Dispositivo (auto, cpu, cuda)
.PARAMETER ComputeType
    Tipo de cómputo ("" = auto: float16 en cuda, int8 en cpu)
.PARAMETER InstallDeps
    Instalar dependencias Python automáticamente
.PARAMETER Batch
    Procesar en fila TODOS los audios detectados que aún no fueron extraídos
.PARAMETER WhatIf
    Simular sin ejecutar cambios
.EXAMPLE
    .\scripts\audio\transcribe.ps1
    .\scripts\audio\transcribe.ps1 -Batch
    .\scripts\audio\transcribe.ps1 -Batch -Model medium
    .\scripts\audio\transcribe.ps1 -AudioFile "grabacion.mp3" -Model turbo -Engine whisper
    .\scripts\audio\transcribe.ps1 -InstallDeps
#>

param(
    [string]$Path = ".",
    [string]$AudioFile = "",
    [string]$Model = "small",
    [ValidateSet("faster-whisper", "whisper")]
    [string]$Engine = "faster-whisper",
    [string]$Device = "auto",
    [string]$ComputeType = "",
    [switch]$InstallDeps,
    [switch]$Batch,
    [switch]$Force,
    [switch]$WhatIf
)

$ErrorActionPreference = "Stop"

# ─── CUDA DLLs (Windows) ────────────────────────────────────
# CTranslate2 necesita cublas64_12.dll / cudnn64_9.dll en runtime.
# Los wheels nvidia-*-cu12 los instalan en site-packages\nvidia\<lib>\bin.
# Sin esto, el modelo carga pero transcribir falla con
# "Library cublas64_12.dll is not found or cannot be loaded".
# Nota: usar purelib (venv\Lib\site-packages); getsitepackages()[0]
# devuelve la raíz del venv y la ruta resultante no existe.
$nvPureLib = python -c "import sysconfig; print(sysconfig.get_paths()['purelib'])" 2>$null
if ($nvPureLib) {
    foreach ($nvLib in @("cublas", "cudnn")) {
        $nvBin = Join-Path $nvPureLib "nvidia\$nvLib\bin"
        if (Test-Path $nvBin) {
            $env:PATH = "$nvBin;$env:PATH"
        }
    }
}

# Fuerza UTF-8 en stdout/stderr de Python (Windows usa charmap cp1252 por
# defecto y los prints con ✓/á/ñ rompen con 'charmap' codec error)
$env:PYTHONIOENCODING = "utf-8"
$env:PYTHONUTF8 = "1"

# ─── COLORES Y LOGS ──────────────────────────────────────────

function Write-Step($step, $msg) {
    Write-Host "[$step] $msg" -ForegroundColor Cyan
}

function Write-OK($msg) {
    Write-Host "  ✓ $msg" -ForegroundColor Green
}

function Write-Warn($msg) {
    Write-Host "  ⚠ $msg" -ForegroundColor Yellow
}

function Write-Err($msg) {
    Write-Host "  ✗ $msg" -ForegroundColor Red
}

# ─── CONFIGURACIÓN ───────────────────────────────────────────

$ProjectDir = (Resolve-Path $Path).Path
$AudioDir = "$ProjectDir\.mvgn\audio"
$LogFile = "$AudioDir\pipeline.log"
$StateDir = "$ProjectDir\.mvgn\blackboard"
$StateFile = "$StateDir\state.json"
$ProcessedFile = "$AudioDir\processed.json"

$AudioExtensions = @(".mp3", ".wav", ".m4a", ".flac", ".ogg")

# ─── FUNCIONES ───────────────────────────────────────────────

function Log($msg) {
    $ts = Get-Date -Format "yyyy-MM-dd HH:mm:ss zzz"
    "$ts $msg" | Out-File $LogFile -Append
}

function Notify-Complete {
    <#
    .SYNOPSIS
        Notificación de fin de pipeline: sonido + toast de Windows (BurntToast si está instalado).
    .PARAMETER Title
        Título de la notificación.
    .PARAMETER Message
        Mensaje de la notificación.
    .PARAMETER Success
        true = completado (sonido Asterisk), false = fallo (sonido Hand).
    #>
    param([string]$Title, [string]$Message, [bool]$Success = $true)

    if ($Success) {
        [System.Media.SystemSounds]::Asterisk.Play()
    } else {
        [System.Media.SystemSounds]::Hand.Play()
    }

    if (Get-Command New-BurntToastNotification -ErrorAction SilentlyContinue) {
        try {
            New-BurntToastNotification -Text $Title, $Message -ErrorAction Stop | Out-Null
            return
        } catch {
            # Sin toast disponible: el sonido ya avisó
        }
    }
}

function Invoke-PythonStreaming {
    <#
    .SYNOPSIS
        Ejecuta un script Python mostrando la salida EN VIVO y convirtiendo
        las líneas [MVGN-AUDIO][PROGRESS] en una barra de progreso real.
    .PARAMETER Activity
        Texto de la barra de progreso.
    .PARAMETER Arguments
        Argumentos del script Python.
    #>
    param([string]$Activity, [string[]]$Arguments)

    python @Arguments 2>&1 | ForEach-Object {
        $line = "$_"
        if ($line -match '\[PROGRESS\] ([\d.]+)\|') {
            $pct = [double]$Matches[1]
            $status = "$('{0:N0}' -f $pct)% completado"
            if ($line -match '\|(\d+)\|(\d+)$') {
                $status = "$('{0:N0}' -f $pct)% ($($Matches[1])s / $($Matches[2])s)"
            }
            Write-Progress -Activity $Activity -Status $status -PercentComplete $pct
        } else {
            Write-Host "  $line" -ForegroundColor Gray
        }
    }

    return $LASTEXITCODE -eq 0
}

function Detect-AudioFiles {
    <#
    .SYNOPSIS
        Escanea el proyecto en busca de archivos de audio compatibles.
    #>
    Write-Step "1/4" "Escaneando archivos de audio..."

    $audioFiles = @()

    # Root del proyecto real: subir por ancestros buscando docs\audio.
    # NO usar git rev-parse: mvgnlabs-starter-kit tiene su propio .git y
    # devolvería el starter-kit en lugar del proyecto.
    $projectRoots = @($ProjectDir)
    $ancestor = Split-Path $ProjectDir -Parent
    for ($i = 0; $i -lt 4 -and $ancestor; $i++) {
        $projectRoots += $ancestor
        $ancestor = Split-Path $ancestor -Parent
    }

    $audioRoots = @($projectRoots | Where-Object { $_ -and (Test-Path "$_\docs\audio") })
    if ($audioRoots.Count -eq 0) {
        $audioRoots = @($projectRoots[0])
    }

    # Directorios a escanear: los del proyecto y los de cada root de audio
    $dirsToScan = @()
    foreach ($root in $audioRoots) {
        $dirsToScan += @($root, "$root\docs", "$root\src", "$root\audio", "$root\docs\audio")
    }
    $dirsToScan = $dirsToScan | Sort-Object -Unique

    foreach ($dir in $dirsToScan) {
        if (Test-Path $dir) {
            foreach ($ext in $AudioExtensions) {
                $files = Get-ChildItem -Path $dir -Filter "*$ext" -File -ErrorAction SilentlyContinue
                foreach ($f in $files) {
                    if ($f.FullName -notin $audioFiles) {
                        $audioFiles += $f.FullName
                    }
                }
            }
        }
    }

    # Buscar recursivo en carpetas de audio (docs\audio incluye subcarpetas
    # como "nuevos contextos 18 agosto")
    $searchDirs = @()
    foreach ($root in $audioRoots) {
        $searchDirs += @("$root\audio", "$root\docs\audio")
    }
    $searchDirs = $searchDirs | Sort-Object -Unique

    foreach ($dir in $searchDirs) {
        if (Test-Path $dir) {
            foreach ($ext in $AudioExtensions) {
                $files = Get-ChildItem -Path $dir -Filter "*$ext" -Recurse -File -ErrorAction SilentlyContinue
                foreach ($f in $files) {
                    if ($f.FullName -notin $audioFiles) {
                        $audioFiles += $f.FullName
                    }
                }
            }
        }
    }

    return $audioFiles | Sort-Object
}

function Get-ProcessedBases {
    <#
    .SYNOPSIS
        Lee el registro persistente de audios ya extraídos (processed.json).
    .RETURNS
        Array de nombres base (sin extensión) ya procesados.
    #>
    if (Test-Path $ProcessedFile) {
        try {
            $bases = @(Get-Content $ProcessedFile -Raw | ConvertFrom-Json)
            return @($bases | Where-Object { $_ })
        } catch {
            return @()
        }
    }
    return @()
}

function Save-ProcessedBase {
    <#
    .SYNOPSIS
        Registra un nombre base como extraído en processed.json.
    .PARAMETER Base
        Nombre base del audio (sin extensión).
    #>
    param([string]$Base)

    $null = New-Item -ItemType Directory -Path $AudioDir -Force
    $bases = @(Get-ProcessedBases)
    if ($Base -notin $bases) {
        $bases += $Base
    }
    $bases | Sort-Object -Unique | ConvertTo-Json | Set-Content $ProcessedFile
}

function Test-AudioProcessed {
    <#
    .SYNOPSIS
        Determina si un audio ya fue extraído: está en processed.json o tiene
        output completo (context.md + metadata.json) en $AudioDir\<base>.
    .PARAMETER AudioPath
        Ruta completa al archivo de audio.
    #>
    param([string]$AudioPath)

    $base = [System.IO.Path]::GetFileNameWithoutExtension($AudioPath)
    if ($base -in @(Get-ProcessedBases)) {
        return $true
    }

    $outDir = Join-Path $AudioDir $base
    return (Test-Path (Join-Path $outDir "context.md")) -and (Test-Path (Join-Path $outDir "metadata.json"))
}

function Get-PendingAudioFiles {
    <#
    .SYNOPSIS
        Detecta los audios que aún no fueron extraídos (cola pendiente).
    .RETURNS
        Array de rutas de audio pendientes, ordenadas.
    #>
    $all = @(Detect-AudioFiles)
    return @($all | Where-Object { -not (Test-AudioProcessed $_) })
}

function Select-AudioFile {
    <#
    .SYNOPSIS
        Permite al usuario seleccionar un archivo de audio.
        Opción "T" = procesar TODOS los pendientes en fila.
    #>
    param([string[]]$AudioFiles, [string]$SpecifiedFile)

    if ($SpecifiedFile -ne "") {
        if (Test-Path $SpecifiedFile) {
            return $SpecifiedFile
        }
        Write-Err "Archivo especificado no encontrado: $SpecifiedFile"
        return $null
    }

    if ($AudioFiles.Count -eq 0) {
        return $null
    }

    $pending = @($AudioFiles | Where-Object { -not (Test-AudioProcessed $_) })

    if ($AudioFiles.Count -eq 1) {
        if ($pending.Count -eq 1) {
            Write-OK "Audio detectado: $(Split-Path $AudioFiles[0] -Leaf)"
            return $AudioFiles[0]
        }
        Write-OK "Audio ya extraído: $(Split-Path $AudioFiles[0] -Leaf) — omitido"
        return "skip"
    }

    # Múltiples archivos — preguntar
    Write-Host ""
    Write-Host "Múltiples archivos de audio detectados:" -ForegroundColor Cyan
    Write-Host ""

    for ($i = 0; $i -lt $AudioFiles.Count; $i++) {
        $name = Split-Path $AudioFiles[$i] -Leaf
        $size = (Get-Item $AudioFiles[$i]).Length / 1MB
        $state = if (Test-AudioProcessed $AudioFiles[$i]) { "  [ya extraído]" } else { "" }
        Write-Host "  $($i + 1). $name ($('{0:N1}' -f $size) MB)$state" -ForegroundColor White
    }

    Write-Host "  T. Procesar TODOS los pendientes en fila ($($pending.Count))" -ForegroundColor Green
    Write-Host "  0. Cancelar (omitir audio)" -ForegroundColor Gray
    Write-Host ""
    $choice = Read-Host "Selecciona un archivo (0-$($AudioFiles.Count), T = todos)"

    if ($choice -eq "0" -or $choice -eq "") {
        return "skip"
    }

    if ($choice -eq "T" -or $choice -eq "t") {
        return "all"
    }

    $idx = [int]$choice - 1
    if ($idx -ge 0 -and $idx -lt $AudioFiles.Count) {
        return $AudioFiles[$idx]
    }

    Write-Err "Opción inválida"
    return $null
}

function Install-Dependencies {
    <#
    .SYNOPSIS
        Instala dependencias Python para transcripción.
    .PARAMETER Engine
        Motor de transcripción (faster-whisper, whisper)
    #>
    param([string]$Engine = "faster-whisper")

    Write-Step "-" "Instalando dependencias Python..."

    # Verificar Python
    try {
        $pyVersion = python --version 2>&1
        Write-OK "Python: $pyVersion"
    } catch {
        Write-Err "Python no está instalado. Descarga desde: https://python.org"
        return $false
    }

    $pipPkg = if ($Engine -eq "whisper") { "openai-whisper" } else { "faster-whisper" }
    $engineName = if ($Engine -eq "whisper") { "OpenAI Whisper" } else { "Faster-Whisper" }

    Write-Host "  Instalando $engineName ($pipPkg)..."
    $result = pip install $pipPkg 2>&1

    if ($LASTEXITCODE -eq 0) {
        Write-OK "$engineName instalado"
        return $true
    } else {
        Write-Warn "Error instalando $engineName. Intenta manualmente: pip install $pipPkg"
        return $false
    }
}

function Test-Dependencies {
    <#
    .SYNOPSIS
        Verifica que las dependencias necesarias estén instaladas.
    #>
    param([string]$Engine = "faster-whisper")

    $importCmd = if ($Engine -eq "faster-whisper") {
        "import faster_whisper; print(faster_whisper.__version__)"
    } else {
        "import whisper; print(whisper.__version__)"
    }

    try {
        $null = python -c $importCmd 2>&1
        return $true
    } catch {
        return $false
    }
}

function Update-State {
    <#
    .SYNOPSIS
        Actualiza state.json con el progreso del pipeline de audio.
    #>
    param(
        [string]$Status,
        [string]$AudioFile = "",
        [string]$Artifact = ""
    )

    $scriptPath = "$PSScriptRoot\transcribe.ps1"

    if (-not (Test-Path $StateFile)) {
        Write-Warn "state.json no encontrado. Creando..."
        @{
            schema_version = "5.1.0"
            task_id = "audio-$(Get-Date -Format 'yyyyMMddHHmmss')"
            task_description = "Ingesta de audio: $(Split-Path $AudioFile -Leaf -ErrorAction SilentlyContinue)"
            status = $Status
            current_phase = 0
            agent_sequence = @("audio-ingest")
            completed_agents = @()
            artifacts_generated = @()
            recovered = $false
            last_error = $null
            timestamp = (Get-Date).ToString("yyyy-MM-ddTHH:mm:sszzz")
            audio = @{
                status = $Status
                file = $AudioFile
            }
        } | ConvertTo-Json -Depth 10 | Set-Content $StateFile
        return
    }

    try {
        $s = Get-Content $StateFile -Raw | ConvertFrom-Json
    } catch {
        Write-Warn "state.json corrupto, recreando..."
        $s = @{
            schema_version = "5.1.0"
            task_id = ""
            task_description = ""
            status = "idle"
            current_phase = 0
            agent_sequence = @()
            completed_agents = @()
            artifacts_generated = @()
            recovered = $false
            last_error = $null
            timestamp = (Get-Date).ToString("yyyy-MM-ddTHH:mm:sszzz")
        }
    }

    $s.status = $Status
    $s.timestamp = (Get-Date).ToString("yyyy-MM-ddTHH:mm:sszzz")

    if ($AudioFile -ne "") {
        $s | Add-Member -NotePropertyName "audio" -NotePropertyValue @{
            status = $Status
            file = $AudioFile
        } -Force
    }

    if ($Artifact -ne "" -and $Artifact -notin $s.artifacts_generated) {
        $s.artifacts_generated += $Artifact
    }

    $s | ConvertTo-Json -Depth 10 | Set-Content $StateFile
}

function Invoke-AudioPipeline {
    <#
    .SYNOPSIS
        Ejecuta el pipeline completo para UN audio: transcripción + contexto.
    .PARAMETER SelectedFile
        Ruta del archivo de audio a procesar.
    .RETURNS
        $true si transcripción y contexto terminaron OK.
    #>
    param(
        [string]$SelectedFile,
        [string]$Model,
        [string]$Device,
        [string]$Engine,
        [string]$ComputeType,
        [switch]$WhatIf
    )

    $fileName = Split-Path $SelectedFile -Leaf
    $fileBase = [System.IO.Path]::GetFileNameWithoutExtension($fileName)

    Write-Host ""
    Write-Step "2/4" "Iniciando pipeline de transcripción..."
    Write-Host "  Archivo: $SelectedFile" -ForegroundColor White

    # Preparar directorio de salida
    $outputDir = "$AudioDir\$fileBase"
    if (-not $WhatIf) {
        New-Item -ItemType Directory -Path $outputDir -Force | Out-Null
    }

    # Verificar dependencias
    $depsReady = Test-Dependencies -Engine $Engine

    $engineName = if ($Engine -eq "faster-whisper") { "Faster-Whisper" } else { "OpenAI Whisper" }
    $pipPkg = if ($Engine -eq "faster-whisper") { "faster-whisper" } else { "openai-whisper" }

    if (-not $depsReady) {
        Write-Warn "$engineName no está instalado."
        Write-Host ""
        Write-Host "Para instalar las dependencias:" -ForegroundColor Cyan
        Write-Host "  pip install $pipPkg" -ForegroundColor White
        Write-Host "  O ejecuta: .\scripts\audio\transcribe.ps1 -InstallDeps" -ForegroundColor White
        Write-Host ""
        $install = Read-Host "¿Instalar ahora? (S/n)"

        if ($install -eq "" -or $install -eq "S" -or $install -eq "s") {
            $depsOk = Install-Dependencies -Engine $Engine
            if (-not $depsOk) {
                Write-Err "No se pudieron instalar las dependencias."
                return $false
            }
            $depsReady = $true
        } else {
            Write-Warn "Dependencias no instaladas. No se puede transcribir."
            return $false
        }
    }

    # Actualizar estado
    if (-not $WhatIf) {
        Update-State -Status "transcribing" -AudioFile $SelectedFile
    }

    # Fase 2: Transcribir
    Write-Step "3/4" "Ejecutando transcripción ($engineName, modelo: $Model, dispositivo: $Device)..."
    Write-Host "  Esto puede tomar varios minutos dependiendo del modelo y duración." -ForegroundColor Gray

    $transcriberPath = "$PSScriptRoot\transcriber.py"

    if ($WhatIf) {
        Write-Host "  [WhatIf] python $transcriberPath `"$SelectedFile`" `"$outputDir`" $Model $Device $Engine `"$ComputeType`"" -ForegroundColor Yellow
    } else {
        Log "Iniciando transcripción: $SelectedFile (motor: $Engine, modelo: $Model, dispositivo: $Device, cómputo: '$ComputeType')"

        try {
            $transcribeOk = Invoke-PythonStreaming `
                -Activity "Transcribiendo: $fileName ($engineName)" `
                -Arguments @($transcriberPath, $SelectedFile, $outputDir, $Model, $Device, $Engine, $ComputeType)

            if (-not $transcribeOk) {
                throw "Transcripción falló con código $LASTEXITCODE"
            }

            Write-Progress -Activity "Transcribiendo: $fileName ($engineName)" -Completed
            Write-OK "Transcripción completada"

        } catch {
            Write-Progress -Activity "Transcribiendo: $fileName ($engineName)" -Completed
            Write-Err "Error durante transcripción: $_"
            Notify-Complete -Title "MVGN Audio" -Message "FALLO en transcripción: $fileName" -Success $false
            Update-State -Status "failed" -AudioFile $SelectedFile
            return $false
        }
    }

    # Fase 3: Construir contexto
    Write-Step "4/4" "Construyendo contexto estructurado..."

    $contextBuilderPath = "$PSScriptRoot\context-builder.py"

    if ($WhatIf) {
        Write-Host "  [WhatIf] python $contextBuilderPath `"$outputDir`" `"$outputDir`"" -ForegroundColor Yellow
    } else {
        try {
            $contextOk = Invoke-PythonStreaming `
                -Activity "Construyendo contexto estructurado" `
                -Arguments @($contextBuilderPath, $outputDir, $outputDir)

            if (-not $contextOk) {
                throw "Context builder falló con código $LASTEXITCODE"
            }

            Write-Progress -Activity "Construyendo contexto estructurado" -Completed
            Write-OK "Contexto estructurado generado"

        } catch {
            Write-Progress -Activity "Construyendo contexto estructurado" -Completed
            Write-Err "Error durante construcción de contexto: $_"
            Notify-Complete -Title "MVGN Audio" -Message "FALLO en contexto: $fileName" -Success $false
            Update-State -Status "failed" -AudioFile $SelectedFile
            return $false
        }
    }

    # Pipeline completado
    if (-not $WhatIf) {
        Update-State -Status "context_ready" -AudioFile $SelectedFile -Artifact "$outputDir\context.md"
        Save-ProcessedBase -Base $fileBase
    }

    # Reporte de artefactos
    $artifacts = @(
        "$outputDir\transcript.md",
        "$outputDir\metadata.json",
        "$outputDir\segments.json",
        "$outputDir\context.md",
        "$outputDir\context-summary.json"
    )

    foreach ($a in $artifacts) {
        if (Test-Path $a) {
            Write-Host "    ✓ $a" -ForegroundColor Green
        }
    }

    return $true
}

function Invoke-PendingQueue {
    <#
    .SYNOPSIS
        Procesa en fila TODOS los audios pendientes (sin extraer) con una sola ejecución.
    .PARAMETER Model
        Modelo (tiny, base, small, medium, large-v3, turbo)
    .PARAMETER Device
        Dispositivo (auto, cpu, cuda)
    .PARAMETER Engine
        Motor de transcripción (faster-whisper, whisper)
    .PARAMETER ComputeType
        Tipo de cómputo ("" = auto)
    #>
    param(
        [string]$Model,
        [string]$Device,
        [string]$Engine,
        [string]$ComputeType,
        [switch]$Force,
        [switch]$WhatIf
    )

    $pending = @(Get-PendingAudioFiles)

    if ($pending.Count -eq 0) {
        Write-Host ""
        Write-OK "No hay audios pendientes por extraer. Fila vacía."
        Write-Host ""
        exit 0
    }

    Write-Host ""
    Write-Host "════════════════════════════════════════════════════" -ForegroundColor Cyan
    Write-Host "  FILA DE EXTRACCIÓN ($($pending.Count) audios pendientes)" -ForegroundColor Green
    Write-Host "════════════════════════════════════════════════════" -ForegroundColor Cyan
    Write-Host ""

    $n = 0
    foreach ($f in $pending) {
        $n++
        $size = (Get-Item $f).Length / 1MB
        Write-Host "  $n. $(Split-Path $f -Leaf) ($('{0:N1}' -f $size) MB)" -ForegroundColor White
    }

    Write-Host ""
    if (-not $Force) {
        $confirm = Read-Host "Procesar TODOS en fila? (S/n)"

        if ($confirm -ne "" -and $confirm -ne "S" -and $confirm -ne "s") {
            Write-Host "Fila cancelada." -ForegroundColor Yellow
            exit 0
        }
    }

    $okCount = 0
    $failCount = 0
    $startTime = Get-Date

    for ($i = 0; $i -lt $pending.Count; $i++) {
        Write-Host ""
        Write-Host "━━━ [$($i + 1)/$($pending.Count)] $(Split-Path $pending[$i] -Leaf) ━━━" -ForegroundColor Magenta

        $ok = Invoke-AudioPipeline `
            -SelectedFile $pending[$i] `
            -Model $Model `
            -Device $Device `
            -Engine $Engine `
            -ComputeType $ComputeType `
            -WhatIf:$WhatIf

        if ($ok) {
            $okCount++
            Write-Host "  ✓ [$($i + 1)/$($pending.Count)] completado: $(Split-Path $pending[$i] -Leaf)" -ForegroundColor Green
        } else {
            $failCount++
            Write-Host "  ✗ [$($i + 1)/$($pending.Count)] FALLÓ: $(Split-Path $pending[$i] -Leaf)" -ForegroundColor Red
        }
    }

    $duracionTotal = [math]::Round(((Get-Date) - $startTime).TotalSeconds)

    # Resumen final de la fila
    Write-Host ""
    Write-Host "════════════════════════════════════════════════════" -ForegroundColor Cyan
    Write-Host "  FILA COMPLETADA: $okCount extraídos, $failCount fallidos (${duracionTotal}s)" -ForegroundColor $(if ($failCount -eq 0) { "Green" } else { "Yellow" })
    Write-Host "════════════════════════════════════════════════════" -ForegroundColor Cyan
    Write-Host ""

    $remaining = @(Get-PendingAudioFiles)
    if ($remaining.Count -gt 0) {
        Write-Warn "Pendientes restantes: $($remaining.Count) — vuelve a ejecutar -Batch para reintentarlos"
        foreach ($r in $remaining) {
            Write-Host "    - $(Split-Path $r -Leaf)" -ForegroundColor Gray
        }
    } else {
        Write-OK "Todos los audios detectados fueron extraídos."
    }

    Notify-Complete -Title "MVGN Audio" -Message "Fila completada: $okCount extraídos, $failCount fallidos (${duracionTotal}s)" -Success ($failCount -eq 0)
}

# ─── MAIN ────────────────────────────────────────────────────

$script:inicio = Get-Date

Write-Host ""
Write-Host "MVGN Audio Ingestion v1.2 (modo fila)" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""

# Fase 0: Verificar/instalar dependencias
if ($InstallDeps) {
    $depsOk = Install-Dependencies
    if (-not $depsOk) {
        Write-Err "Instalación de dependencias fallida."
        exit 1
    }
}

# Fase 1: Detectar audio
if ($AudioFile -eq "") {
    $foundFiles = Detect-AudioFiles

    if ($foundFiles.Count -eq 0) {
        Write-Warn "No se detectaron archivos de audio compatibles."
        Write-Host "Formatos soportados: $($AudioExtensions -join ', ')" -ForegroundColor Gray
        Write-Host ""
        Write-Host "Para transcribir audio manualmente:" -ForegroundColor Cyan
        Write-Host "  1. Coloca el archivo en ./audio/ o ./docs/audio/" -ForegroundColor White
        Write-Host "  2. Vuelve a ejecutar este script" -ForegroundColor White
        Write-Host "  3. O usa: .\scripts\audio\transcribe.ps1 -AudioFile \"ruta\" -InstallDeps" -ForegroundColor White
        exit 0
    }
} else {
    $foundFiles = @($AudioFile)
}

# ─── MODO FILA (-Batch): procesar todos los pendientes en secuencia ───
if ($Batch -and $AudioFile -eq "") {
    Invoke-PendingQueue -Model $Model -Device $Device -Engine $Engine -ComputeType $ComputeType -Force:$Force -WhatIf:$WhatIf
    exit 0
}

$selectedFile = Select-AudioFile -AudioFiles $foundFiles -SpecifiedFile $AudioFile

if ($selectedFile -eq $null) {
    Write-Err "No se pudo seleccionar archivo de audio."
    exit 1
}

if ($selectedFile -eq "skip") {
    Write-Host "Audio omitido. Continuando con flujo normal MVGN." -ForegroundColor Yellow
    exit 0
}

# Opción "T": procesar todos los pendientes en fila
if ($selectedFile -eq "all") {
    Invoke-PendingQueue -Model $Model -Device $Device -Engine $Engine -ComputeType $ComputeType -Force:$Force -WhatIf:$WhatIf
    exit 0
}

# Pipeline para un solo archivo
$pipelineOk = Invoke-AudioPipeline `
    -SelectedFile $selectedFile `
    -Model $Model `
    -Device $Device `
    -Engine $Engine `
    -ComputeType $ComputeType `
    -WhatIf:$WhatIf

if (-not $pipelineOk) {
    exit 1
}

$fileName = Split-Path $selectedFile -Leaf

# Reporte final
Write-Host ""
Write-Host "════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "  PIPELINE DE AUDIO COMPLETADO" -ForegroundColor Green
Write-Host "════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""
Write-Host "  Estado:    CONTEXT_READY" -ForegroundColor Green
Write-Host "  Archivo:   $fileName" -ForegroundColor White
Write-Host "  Modelo:    $Model" -ForegroundColor Gray
Write-Host "  Artefactos generados:" -ForegroundColor Gray

$outputDir = "$AudioDir\$([System.IO.Path]::GetFileNameWithoutExtension($fileName))"

$artifacts = @(
    "$outputDir\transcript.md",
    "$outputDir\metadata.json",
    "$outputDir\segments.json",
    "$outputDir\context.md",
    "$outputDir\context-summary.json"
)

foreach ($a in $artifacts) {
    if (Test-Path $a) {
        Write-Host "    ✓ $a" -ForegroundColor Green
    }
}

Write-Host ""
Write-Host "  Próximo paso:" -ForegroundColor Cyan
Write-Host "  El context.md está listo para generar el PRD." -ForegroundColor White
Write-Host "  ⚠ REGLA: El PRD DEBE generarse exclusivamente desde context.md." -ForegroundColor Yellow
Write-Host "  ⚠ NO debe leerse transcript.md directamente para generar el PRD." -ForegroundColor Yellow
Write-Host "  Para generar PRD desde context.md: .\tools\mvgn-init.ps1" -ForegroundColor White
Write-Host ""
Write-Host "════════════════════════════════════════════════════" -ForegroundColor Cyan

# Notificación de fin de pipeline (sonido + toast de Windows)
$duracionTotal = [math]::Round(((Get-Date) - $script:inicio).TotalSeconds)
Notify-Complete -Title "MVGN Audio" -Message "Pipeline completado: $fileName (${duracionTotal}s)" -Success $true
