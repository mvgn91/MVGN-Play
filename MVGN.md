# MVGN — Multi-Agent Governance Framework v5.3

> **v5.3 Separation**: LAYOUT/2.0 (todo el motor en .mvgn/) - distribucion por GitHub Releases - JOIN/2.0 - audio GPU

> **Documentación del framework de gobierno multi-agente de Crimson Nexus Corp.**
> Este documento describe el **Blackboard Pattern**, una de las capacidades de MVGN: una capa de abstracción para la creación, gestión y documentación de proyectos desde la idea hasta su publicación.
> Pipeline: OpenCode → FreeBuff → MiMoCode.
> Costo: $0 | Dependencias: PowerShell 5.1+, Bash, Git.

---

## MVGN y el Blackboard Pattern

**MVGN** es el framework de gobierno multi-agente de Crimson Nexus Corp. Proporciona estándares, trazabilidad, session management y gobernanza para proyectos con IA.

El **Blackboard Pattern** es una de sus capacidades: una capa de abstracción que coordina múltiples agentes de IA secuencialmente usando un estado compartido en `state.json`. Cada agente ejecuta una fase, deja artefactos, y actualiza el estado para el siguiente.

### 📌 Ecosistema MVGN

MVGN es el framework de gobierno multi-agente de **Crimson Nexus Corp.** El Blackboard Pattern es una de sus capacidades. El ecosistema completo incluye:

| Componente | Rol |
|:-----------|:----|
| **MVGN Framework** | Framework de gobierno multi-agente. Incluye estándares (MVGN-STD-001 OER, MVGN-STD-002 FBCR), session management y trazabilidad. |
| **Blackboard Pattern** | Capa de abstracción de MVGN para creación, gestión y documentación de proyectos (este starter kit). |
| **KRONOS Runtime** | Motor de estado operacional en Go para vaults Obsidian. CLI con sync, status, context, query y session management. |
| **ATENEA** | Activo Cognitivo de Governance & QA. Elabora OER y audita sesiones. |
| **CARONTE** | Mapa de tránsito del dominio MVGN. Orientación contextual para agentes entrantes. |
| **CNX-CANON** | Canon de Crimson Nexus: fundamentos, protocolos y gobernanza. |
| **AEON Protocol** | Protocolo de preservación referencial y sellado de integridad (CNX-DOCTRINE-002). |

> **Nota:** Este starter kit implementa el **Blackboard Pattern** de MVGN, no MVGN completo. KRONOS es una capa opt-in para vaults Obsidian. Para el ecosistema completo, consulta el vault oficial de Crimson Nexus.

---


## Adopción (v5.3 — distribución por Releases)

```powershell
# Instalación desde cero, cualquier máquina con PowerShell 7:
irm https://raw.githubusercontent.com/mvgn91/mvgnlabs-starter-kit/main/tools/bootstrap.ps1 | iex
```

- Todo el motor vive en `<proyecto>/.mvgn/` (LAYOUT/2.0) — publicación segura por construcción.
- Actualización dentro del proyecto: `mvgn update` o `.mvgn\NYX.ps1 -Mode update`
  (semver contra Releases + changelog delta; blackboard preservado).
- Comando global opcional: `.mvgn\tools\install-profile.ps1` instala `mvgn` en el perfil PowerShell.

## MVGN Observer — Monitoreo en Tiempo Real

MVGN incluye un **Observation Console** implementado en Go dentro del módulo `mvgn-scan`. Ofrece dos modos visuales para monitorear el estado del proyecto en vivo:

| Modo | Comando | Descripción |
|:-----|:--------|:------------|
| **Terminal** | `-watch` | Dashboard ANSI con semáforo, health score, progreso y timeline |
| **Web** | `-observe` | Servidor HTTP con dashboard visual, ring SVG, SSE en tiempo real |
| **Unificado** | `-watch -observe` | Ambos simultáneamente |

### Terminal Dashboard (`--watch`)

```
╔══════════════════════════════════════════════════════════════════════════╗
║                        MVGN OBSERVER v5.2                              ║
║  Watching: C:\proyecto  |  Status: IN_PROGRESS                          ║
╠══════════════════════════════════════════════════════════════════════════╣
║  ┌────────────────────┐  ┌────────────────────┐  ┌────────────────────┐ ║
║  │      5.2           │  │      🟡             │  │     PHASE 2        │ ║
║  │      VERSION       │  │      STATUS         │  │     PROGRESS       │ ║
║  └────────────────────┘  └────────────────────┘  └────────────────────┘ ║
║  ┌────────────────────┐  ┌────────────────────┐                          ║
║  │      🟢60%         │  │     42/80          │                          ║
║  │      HEALTH        │  │     ATENEA         │                          ║
║  └────────────────────┘  └────────────────────┘                          ║
╠══════════════════════════════════════════════════════════════════════════╣
║  CURRENT TASK  task-abc123                                              ║
║    Implementar autenticación OAuth                                      ║
║    PROGRESS  🟢████████████████████░░░░░░░░░░░░░░░░░░  60%              ║
║    Agents   🟢 ● opencode                                               ║
║    Artifacts  3 generated                                               ║
╠══════════════════════════════════════════════════════════════════════════╣
║  EVENT TIMELINE                                                         ║
║    14:32:15  HEARTBEAT       ♥ Heartbeat — system alive                 ║
║    14:30:00  STATE_CHANGED   State changed: pending → in_progress       ║
║    14:25:00  SESSION_STARTED Session started · project: mi-proyecto     ║
╠══════════════════════════════════════════════════════════════════════════╣
║  NEXT ACTION                                                             ║
║    Continúa con la fase 2: revisión de código por FreeBuff              ║
╚══════════════════════════════════════════════════════════════════════════╝
```

### Health Score con ATENEA

El health score combina dos factores:
- **Baseline** (0-100): Según el estado (`completed`=95, `in_progress`=60, etc.) + artifact/agent bonus
- **ATENEA Score** (0-80): 9 condiciones de governance evaluadas:
  - context.md presente, JOIN.md disponible, handoff activo, agentes definidos
  - artifacts en disco, timestamp fresco, sin errores, fase definida, schema actual

ATENEA contribuye con un 20% al health final.

### Cómo usar el Observer

```powershell
# Desde la raíz del proyecto:
.\mvgn-scan\mvgn-scan.exe -watch                          # Terminal
.\mvgn-scan\mvgn-scan.exe -observe -port 8080             # Web dashboard
.\mvgn-scan\mvgn-scan.exe -watch -observe                 # Ambos

# Atajo directo:
.\tools\observe.bat

# Instalación permanente:
.\tools\install-observer.ps1                              # PATH + startup
.\tools\install-observer.ps1 -RegisterTask                # + inicio con Windows
```

### ATENEA Conditions Model

El observer evalúa automáticamente 9 condiciones de governance y las persiste en `state.json`:

```json
{
  "atenea_conditions": {
    "context_exists": true,
    "join_exists": true,
    "handoff_exists": true,
    "timestamp_fresh": true,
    "artifacts_exist": true,
    "has_agent_sequence": true,
    "schema_current": true,
    "no_errors": true,
    "phase_defined": true
  }
}
```

---

## Pipeline del Blackboard Pattern

```
  ┌──────────┐     ┌──────────┐     ┌──────────┐
  │ OpenCode │ ──→ │ FreeBuff │ ──→ │ MiMoCode │
  │ Fase 1   │     │ Fase 2   │     │ Fase 3   │
  └────┬─────┘     └────┬─────┘     └────┬─────┘
       │                │                │
       └────────────────┴────────────────┘
                        │
                 ┌──────▼──────┐
                 │  state.json │ ← Fuente de verdad única
                 │  Blackboard │
                 └─────────────┘
```

---

## Modo UPDATE — Verificación de Versión

MVGN incluye un **update checker** integrado en `NYX.ps1` que compara la versión local contra el remoto de GitHub.

### Check-Update

Al ejecutar `.\NYX.ps1 -Mode update` (o automáticamente al final de `READY`):

1. Lee `mvgn-runtime.json` → `runtime_version` local
2. Obtiene el último git tag local + commit HEAD
3. Fetch remoto: `git fetch --tags origin`
4. Obtiene último tag remoto + commit de origin/main
5. Intenta obtener `mvgn-runtime.json` remoto desde `raw.githubusercontent.com`
6. Compara semver y commits
7. Muestra tabla:
   ```
     LOCAL                           REMOTO
     ├─ Runtime:    v5.2.0           ├─ Runtime:    v5.2.0
     ├─ Git tag:    v4.2.0           ├─ Git tag:    v4.2.0
     └─ Commit:     dc1e74b          └─ Commit:     dc1e74b
   ```
8. Si hay diferencias: ofrece `git pull`, `git pull --rebase`, o cancelar

### Comandos

```powershell
.\NYX.ps1 -Mode update                     # Verificar + preguntar
.\NYX.ps1 -Mode update -Force              # Actualizar sin preguntar
```

El modo **READY** ejecuta auto-check al final y pregunta al usuario.

---

## Entry Points

| Archivo | Propósito | Compatibilidad |
|---------|-----------|---------------|
| `NYX.md` | Entry point universal — Markdown puro | **Cualquier** modelo de IA |
| `NYX.ps1` | Entry point imperativo — PowerShell | PowerShell 5.1+ |
| `tools/validate-entry-points.ps1` | Valida sincronización NYX.md ↔ NYX.ps1 | PowerShell 5.1+ |
| `tools/install-observer.ps1` | Instalación one-command del observer | PowerShell 5.1+ |
| `tools/observe.bat` | Atajo doble-click para observer unificado | Windows |
| `scripts/audio/transcribe.ps1` | Pipeline de ingesta de audio → contexto → PRD | PowerShell + Python |

---

## 🎙️ Audio Ingestion (Nuevo en v5.2)

MVGN extiende el modo **INIT** con capacidad de **ingesta de audio**, permitiendo transcribir notas de voz, reuniones o sesiones de lluvia de ideas directamente a documentos estructurados.

### Pipeline de Audio

```
AUDIO_DISCOVERY → TRANSCRIBING → CONTEXT_BUILDING → CONTEXT_READY → PRD
```

### Cómo funciona

1. Durante INIT, MVGN escanea el proyecto en busca de archivos `.mp3`, `.wav`, `.m4a`, `.flac`, `.ogg`
2. Si encuentra audio, pregunta al usuario si desea transcribirlo
3. Ejecuta **Faster-Whisper** (local, sin cloud) para transcribir
4. Genera `context.md` estructurado con:
   - Resumen ejecutivo
   - Objetivos detectados
   - Problemas identificados
   - Requisitos funcionales y no funcionales
   - Tecnologías mencionadas
   - Términos clave
5. El `context.md` alimenta el generador de PRD

### Artefactos generados

```
.mvgn/audio/{archivo}/
├── transcript.md         # Transcripción completa con timestamps
├── metadata.json         # Metadatos (modelo, idioma, duración)
├── segments.json         # Segmentos temporizados
├── context.md            # Contexto estructurado → PRD
└── context-summary.json  # Resumen para state.json
```

### Dependencias

```bash
pip install faster-whisper
```

O ejecutar el script con flag `-InstallDeps`:
```powershell
.\scripts\audio\transcribe.ps1 -InstallDeps
```

### NYX.md — Modos de operación

| Modo | Cuándo se activa |
|------|------------------|
| **INIT** | Proyecto nuevo: no existe `.mvgn/blackboard/` |
| **READY** | Proyecto existente: `state.json` válido |
| **UPGRADE** | Migración desde v4.x: existe `mvgn-runtime.json` sin blackboard |
| **COLLAB** | El usuario solicita colaboración multi-agente |

---

## Arquitectura

### Blackboard (estado compartido)

```
.mvgn/blackboard/
├── state.json              # ← Fuente de verdad única
├── context.md              # Instrucciones para agentes
├── JOIN.md                 # Onboarding de otras IAs
├── handoff-latest.json     # Contexto de handoff entre agentes
└── logs/
    └── execution.log       # Historial de ejecución
```

### state.json — Schema v5.1

```json
{
  "schema_version": "5.1.0",
  "task_id": "task-20260725120000",
  "task_description": "Implementar feature X",
  "status": "in_progress",
  "current_phase": 2,
  "agent_sequence": ["opencode", "freebuff", "mimocode"],
  "completed_agents": ["opencode"],
  "artifacts_generated": ["src/generated.md"],
  "recovered": false,
  "last_error": null,
  "timestamp": "2026-07-25T12:00:00Z"
}
```

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `schema_version` | string | Versión del schema (5.1.0) |
| `task_id` | string | ID único de la tarea |
| `task_description` | string | Descripción de la tarea |
| `status` | string | `pending`, `in_progress`, `completed_phase_1/2`, `completed`, `failed` <br> **Audio:** `audio_discovery`, `transcribing`, `context_building`, `context_ready` |
| `current_phase` | int | Fase actual (0-3) |
| `agent_sequence` | array | Orden de ejecución de agentes |
| `completed_agents` | array | Agentes que ya completaron |
| `artifacts_generated` | array | Artefactos producidos |
| `recovered` | bool | Si la tarea fue recuperada tras fallo |
| `last_error` | string/null | Último error registrado |
| `timestamp` | string | Última actualización |

---

## Handoff entre Agentes

El handoff es el mecanismo por el cual un agente pasa contexto al siguiente.

### Flujo de handoff

```
1. Agente completa su fase
2. Actualiza state.json: agrega su nombre a completed_agents
3. Se genera handoff-latest.json con:
   - from_agent, to_agent, task_id, artifacts
   - next_action recomendada
4. Se actualiza JOIN.md con instrucciones para el siguiente agente
5. El siguiente agente lee JOIN.md y continúa
```

### handoff-latest.json

```json
{
  "schema_version": "5.1.0",
  "from_agent": "opencode",
  "to_agent": "freebuff",
  "task_id": "task-20260725120000",
  "task_description": "Implementar feature X",
  "phase": 1,
  "artifacts": ["src/generated.md"],
  "next_action": "Continúa con la siguiente fase. Si hay error, ejecuta recovery."
}
```

### JOIN.md

Es el punto de entrada para otras IAs que se unen al proyecto:
1. La IA lee `JOIN.md`
2. Obtiene el rol y la tarea asignada
3. Lee `state.json` para estado actual
4. Ejecuta la tarea según su rol
5. Actualiza `state.json` al terminar

---

## Adaptadores CLI

Los adaptadores son scripts que cada agente IA ejecuta para integrarse con el Blackboard.

| Adaptador | PS1 | SH | Función |
|-----------|-----|----|---------|
| **OpenCode** | ✅ | ✅ | Genera código en `src/` |
| **FreeBuff** | ✅ | ✅ | Revisa y documenta en `src/review-report.md` |
| **MiMoCode** | ✅ | ✅ | Tests en `tests/` + docs en `docs/` |

### Ejecución

```powershell
# Nueva tarea
.\scripts\orchestrate.ps1 -Task "Implementar autenticación"

# Reanudar tarea fallida
.\scripts\orchestrate.ps1 -Resume

# Ver estado
.\scripts\orchestrate.ps1 -Status

# Handoff manual a otro agente
.\scripts\orchestrate.ps1 -Handoff "auditor"

# Auto-commit al finalizar
.\scripts\orchestrate.ps1 -Task "Feature" -AutoCommit
```

---

## Estructura del Proyecto

```
proyecto/
├── .mvgn/blackboard/          # Estado compartido (Blackboard)
│   ├── state.json             # Fuente de verdad
│   ├── context.md             # Instrucciones
│   ├── JOIN.md                # Onboarding IAs
│   ├── handoff-latest.json    # Handoff
│   └── logs/
├── archive/                   # Material legacy preservado
├── docs/                      # Documentación activa
│   ├── mvgn-kernel.md         # Comportamiento IA
│   ├── mvgn-recovery.md       # Protocolo de recuperación
│   └── mvgn-finalization.md   # Protocolo de finalización
├── scripts/                   # Orquestación
│   ├── orchestrate.ps1        # Motor principal
│   ├── orchestrate.sh         # Motor Bash
│   └── agents/                # Adaptadores por agente
├── tools/                     # Herramientas
│   ├── mvgn-init.ps1          # Init de proyecto
│   └── validate-entry-points.ps1 # Validador
├── src/                       # Código fuente
├── tests/                     # Tests
├── mvgn-runtime.json          # Manifiesto runtime
├── NYX.md                     # Entry point universal
├── NYX.ps1                    # Entry point PowerShell
└── README.md                  # Documentación raíz
```

---

## Manejo de Errores y Recovery

| Problema | Solución |
|----------|----------|
| Fase falla | `orchestrate.ps1 -Resume` después de corregir |
| state.json corrupto | Seguir `docs/mvgn-recovery.md` |
| Handoff perdido | Revisar `handoff-latest.json` + `JOIN.md` |
| Error de schema | Validar con `tools/validate-entry-points.ps1` |

---

## Changelog

| Versión | Fecha | Cambios |
|---------|-------|---------|
| v5.3.0 | 2026-08-26 | Separation: LAYOUT/2.0 consolidado (.mvgn/), bootstrap por Releases, JOIN/2.0, audio GPU, CLI global mvgn |
| v5.2.0 | 2026-07-30 | Observation Console (terminal --watch + web --observe + SSE), ATENEA Governance Scoring (9 condiciones), Update Checker en NYX.ps1, install-observer.ps1, Release v5.2.0 sellada |
| v5.1.0 | 2026-07-30 | Dual entry point, archive/ preservation, handoff formalizado, schema versionado |
| v5.0 | 2026-07-25 | Blackboard Pattern, COLLAB mode, JOIN.md |
| v4.2 | 2026-07-15 | Telemetry v2, SMG Bridge |

---

*MVGN v5.2 — Multi-Agent Governance Framework — Sellado 2026-07-30*
*Costo: $0 | Sin dependencias de pago | Release v5.2.0 completada*

---

*Parte del ecosistema Crimson Nexus Corp. — Node MX-03*
