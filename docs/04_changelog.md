# 04 — Changelog

> Historial de cambios del proyecto.

---

## [5.3.0] - 2026-08-26 - Separation: dev/impl separation + LAYOUT/2.0

### Added
- **Distribucion por GitHub Releases**: tools/bootstrap.ps1 (instalador una linea irm|iex), tools/package-release.ps1 (zip por manifiesto + validacion 1:1)
- **LAYOUT/2.0**: todo el motor MVGN consolidado en .mvgn/ - publicacion segura por construccion (.gitignore de 1 linea para adoptantes)
- **JOIN/2.0**: protocolo canonico inter-modelo con cosmogonia CNX (designadores AAAA-##, panteon, custodia AEON) + Ciclo de supervision estandar COLLAB
- **templates/**: docs base 00-04 como fuente unica, payload-manifest.json, gitignore.template, schemas/
- **CLI global mvgn**: tools/install-profile.ps1 - bloque idempotente con resolucion ascendente .mvgn/NYX.ps1
- **Audio ingestion GPU**: importado y probado (CUDA int8, batch queue, processed.json); punto unico tools/mvgn-audio.ps1; opcion en INIT y READY (-Audio)
- **Guards legacy**: bootstrap aborta ante prototipos pre-5.x y layout pre-consolidacion; cnx.ps1 etiqueta [LEGACY]
- **Interfaz CNX en bootstrap**: marco/banner estilo cnx.ps1, referencias griegas leves (ΠΥΛΗ, ΤΕΛΟΣ, ΕΝΤΑΞΕΙ/ΣΦΑΛΜΑ, sello AEON), barra de carga real por bytes en la descarga y ENTER obligatorio al cierre (-NoPause para invocacion programatica)

### Changed
- NYX.ps1 -Mode update delega en bootstrap (semver contra Releases, changelog delta; sin git history)
- mvgn-init.ps1 determinista: copia desde templates/, cero here-strings, ProjectRoot derivado de script
- cnx.ps1: flujo clone-based eliminado (3 fugas cerradas)

### Fixed
- H-6: flujo update de bootstrap (una sola descarga/extraccion sirve changelog+deploy)
- H-13: modo -PayloadZip no muta runtime_version
- INC-1 (supervision): motor revertido por test fuera de sandbox - restaurado con danio colateral incluido

---
## [0.5.0] — 2026-07-30 — Finalización: Observer + ATENEA + Update Checker

### Added
- **ATENEA Governance Scoring**: 9 condiciones de governance evaluadas en tiempo real (context, JOIN, handoff, agents, artifacts, timestamp, errors, phase, schema) — blend 20% en health score
- **AteneaConditions struct**: Nuevo modelo en `models.go` con 9 campos booleanos + integración en `BlackboardState` y `ProjectData`
- **healthScore refinado**: Baseline (status) + artifact/agent bonus + ATENEA blend — máximo 100
- **tools/install-observer.ps1**: Instalación one-command — build Go, copia a PATH, script startup, scheduled task opcional
- **NYX.ps1 modo OBSERVER**: `-Mode observer` con flag `-Web` para dashboard, auto-build si falta el binario
- **NYX.ps1 modo UPDATE**: `Check-Update` compara local vs remoto (runtime, git tag, commit), `Invoke-Update` ejecuta git pull
- **NYX.md**: Documentación completa de UPDATE y OBSERVER modes
- **MVGN.md**: Sección completa MVGN Observer con health score formula, ATENEA conditions, comandos
- **README.md**: Badges de Observer Console y ATENEA, feature list actualizada
- **Dashboard web**: Card ATENEA con score en tiempo real vía SSE

### Changed
- `state.json`: Schema extendido con `atenea_conditions` (9 campos)
- `observer.go`: `healthScore()` ahora acepta ateneaScore, `evaluateAtenea()` con 9 checks
- `server.go`: SSE Update incluye `atenea_score`
- `app.js` + `index.html`: Nueva métrica ATENEA en dashboard web
- `NYX.ps1`: ValidateSet extendido con `update` y `observer`, auto-check en READY

---

## [0.4.0] — 2026-07-30 — Observation Console (Web Dashboard + Telemetry Timeline)

### Added
- **Modo `--observe`**: Dashboard visual con servidor HTTP embebido, SSE hub, auto-open browser
- **Frontend web**: HTML + CSS + JS con anillo SVG de salud, barra de progreso, timeline de eventos
- **Telemetry timeline**: Últimos eventos desde `events/*.jsonl` (SESSION_STARTED, HEARTBEAT, etc.)
- **Modo unificado**: `--watch --observe` corre terminal + web simultáneamente
- **`tools/observe.bat`**: Atajo doble-click para modo unificado

### Changed
- `main.go`: Broadcast channel conecta observer → server SSE
- `server.go`: SSE endpoint `/api/events`, hub multiclient

---

## [0.3.0] — 2026-07-30 — Observation Console (Terminal ANSI)

### Added
- **Modo `--watch`**: Terminal ANSI con semáforo en vivo en `mvgn-scan` (polling state.json)
- **Modo `--observe`**: Dashboard visual con servidor HTTP embebido + SSE + auto-open browser
- **`internal/observer/`**: Package Go que lee `state.json` + `mvgn-runtime.json` en ciclo
- **`internal/server/`**: Package Go con SSE hub, rutas HTTP, y embed de frontend
- **Frontend embebido**: HTML + CSS + JS con tarjetas, anillo de salud, barra de progreso, timeline
- **Health Score**: Score compuesto (estado + artefactos + agentes) con anillo SVG
- **`tools/observe.bat`**: Script para doble click directo al dashboard visual
- **`BlackboardState`**: Struct en models.go para leer `.mvgn/blackboard/state.json`

### Changed
- `mvgn-scan/cmd/mvgn-scan/main.go`: Flags `--watch`, `--observe`, `--port`, `--interval`
- `mvgn-scan/internal/display/cnx.go`: `MetricCard` con soporte ANSI colors, nuevas constantes `ANSI_CLEAR`, `ANSI_HOME`
- `.mvgn/blackboard/state.json`: Nueva tarea IMP-OBS-CONSOLE-002, fase 1 completada

---

## [0.2.0] — 2026-07-30 — Audio Ingestion (v5.2)

### Added
- **Audio Ingestion Pipeline**: `scripts/audio/transcribe.ps1` — orquestador completo
- **Faster-Whisper support**: `scripts/audio/transcriber.py` — motor default
- **OpenAI Whisper support**: Motor alternativo vía `-Engine whisper`
- **Context Builder**: `scripts/audio/context-builder.py` — 12 secciones semánticas
- **AUDIO DISCOVERY**: Detección automática durante INIT en `tools/mvgn-init.ps1`
- **Nuevos estados**: `audio_discovery`, `transcribing`, `context_building`, `context_ready`
- **mvgn-runtime.json**: v5.2, capability `audio_ingestion: true`, config de audio
- **Documentación**: PRD (ATHN1), Arquitectura, Tasks, Changelog, Lessons, State Report

### Changed
- NYX.md: Modo AUDIO en tabla de modos, INIT→AUDIO en detección
- MVGN.md: Sección Audio Ingestion, estados extendidos en schema
- README.md: Badge audio, comandos, estructura actualizada

---

## [0.1.0] — 2026-07-30 — Inicialización

### Added
- MVGN v5.1 Blackboard Pattern inicializado
- Blackboard: .mvgn/blackboard/state.json
- Entry points: NYX.md + NYX.ps1
- Scripts de orquestación: orchestrate.ps1 + orchestrate.sh
- Adaptadores: OpenCode, FreeBuff, MiMoCode (PS1 + SH)

---

*Documento generado por MVGN — versión actual: v5.2*
