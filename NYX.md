# MVGN — Entry Point Universal

> Carga este archivo en tu sesión de IA para activar MVGN.
> Este entry point activa el **Blackboard Pattern**, una de las capacidades del framework MVGN.
> MVGN es el framework de gobierno multi-agente de Crimson Nexus Corp.
> Funciona en **cualquier modelo**: ChatGPT, Claude, Gemini, DeepSeek, etc.
> Versión: v5.3.0 — Sincronizado con NYX.ps1
> Layout: LAYOUT/2.0 consolidado (todo MVGN en .mvgn/) · Instalación: bootstrap vía GitHub Releases

---

## Modos de operación

| Modo | Cuándo se activa |
|------|------------------|
| **INIT** | Proyecto nuevo: no existe `.mvgn/blackboard/state.json` |
| **AUDIO** | Sub-fase de INIT: detecta y transcribe audio para generar contexto |
| **READY** | Proyecto existente: existe `blackboard/` con `state.json` válido |
| **UPGRADE** | Migración: existe `mvgn-runtime.json` pero no `blackboard/` |
| **UPDATE** | El usuario solicita verificar/actualizar versión de MVGN |
| **COLLAB** | Colaboración: el usuario solicita activar modo multi-agente |

---

## Paso 1: Detectar entorno

Revisa si existe el archivo `.mvgn/blackboard/state.json`.

- Si **SÍ** → determina modo READY o COLLAB
- Si **NO** → revisa si existe `mvgn-runtime.json` para modo UPGRADE
- Si no existe nada → modo INIT

---

## Paso 2: Determinar modo

| Condición | Modo | Acción |
|-----------|------|--------|
| No existe `.mvgn/blackboard/` | **INIT** | Inicializar proyecto nuevo |
| INIT + archivos de audio (.mp3/.wav/.m4a/.flac/.ogg) | **INIT→AUDIO** | Sub-fase: transcribir audio → context.md → PRD |
| Existe `.mvgn/blackboard/state.json` | **READY** | MVGN activo — operación normal |
| Existe `mvgn-runtime.json` pero sin blackboard | **UPGRADE** | Migrar a Blackboard |
| Usuario solicita colaboración | **COLLAB** | Activar sesión colaborativa |

---

## Paso 3: Ejecutar según modo

### Modo INIT — Proyecto nuevo

1. Crear estructura MVGN:
   ```
   .mvgn/blackboard/
   ├── state.json        # Estado inicial (pending, sin tarea)
   ├── context.md        # Contexto general
   ├── JOIN.md           # Para colaboración multi-agente
   └── logs/
       └── execution.log
   ```
2. Crear `.mvgn/mvgn-runtime.json` con versión v5.3
3. Crear `docs/` directorio
4. **AUDIO DISCOVERY** — Escanear archivos de audio:
   - Busca archivos `.mp3`, `.wav`, `.m4a`, `.flac`, `.ogg`
   - Si encuentra uno: preguntar al usuario si desea transcribirlo
   - Si acepta: ejecutar pipeline `scripts/audio/transcribe.ps1`
     - `AUDIO_DISCOVERY` → `TRANSCRIBING` → `CONTEXT_BUILDING` → `CONTEXT_READY`
   - El `context.md` generado alimenta el generador de PRD
5. Preguntar al usuario: ¿Quieres activar modo COLLAB (colaboración multi-agente)?
   - **NO** → Continuar en modo normal
   - **SÍ** → Preguntar rol de la otra IA (Auditor, Desarrollador, Revisor, Personalizado) → Generar JOIN.md
6. Reportar: "MVGN v5.3 inicializado. Lee .mvgn/NYX.md para comenzar."

### Modo READY — Operación normal

1. Leer `.mvgn/blackboard/state.json`
2. Evaluar estado:
   - `status == "pending"` → Mostrar tarea pendiente, preguntar si ejecutar
   - `status == "in_progress"` → Mostrar progreso, ofrecer continuar
   - `status == "completed"` → Mostrar resumen de la última sesión
   - `status == "failed"` → Mostrar error, ofrecer resume o recovery
3. Preguntar al usuario qué acción desea
4. Si el usuario solicita colaboración → ir a modo COLLAB

### Modo UPGRADE — Migración a Blackboard

1. Detectar versión anterior del runtime
2. Crear estructura blackboard:
   - Mapear estado existente a `state.json`
   - Generar `context.md` desde documentación existente
   - Crear `JOIN.md` con información básica
3. Actualizar `mvgn-runtime.json` a v5.1
4. Reportar: "MVGN migrado a Blackboard Pattern v5.1"

### Modo UPDATE — Verificación y actualización de versión

Cuando el usuario solicita actualizar MVGN o verificar la versión:

1. **Check-Update** compara:
   - **Local**: `runtime_version` de `mvgn-runtime.json` + último git tag + commit HEAD
   - **Remoto**: mismo `mvgn-runtime.json` vía `raw.githubusercontent.com` + último git tag remoto
2. Muestra tabla comparativa:
   ```
     LOCAL                           REMOTO
     ├─ Runtime:    v5.2.0           ├─ Runtime:    v5.2.0
     ├─ Git tag:    v4.2.0           ├─ Git tag:    v4.2.0
     └─ Commit:     a1b2c3d          └─ Commit:     e5f6g7h
   ```
3. Si hay diferencias, pregunta si aplicar git pull
4. Opciones: `git pull`, `git pull --rebase`, o solo mostrar diferencias

```powershell
.\NYX.ps1 -Mode update                     # Verificar y preguntar
.\NYX.ps1 -Mode update -Force              # Actualizar sin preguntar
```

El **modo READY** también ejecuta un check automático al final y pregunta si querés actualizar.

### Modo OBSERVER — Monitoreo visual en vivo

Cuando el usuario solicita ver el estado visual del proyecto, activar el MVGN Observer:

1. Verificar que existe `mvgn-scan/mvgn-scan.exe`:
   - Si no existe: `cd mvgn-scan && go build -o mvgn-scan.exe ./cmd/mvgn-scan/`
2. Ejecutar observer:
   ```powershell
   .\mvgn-scan\mvgn-scan.exe -watch                     # Terminal ANSI
   .\mvgn-scan\mvgn-scan.exe -observe -port 8080        # Dashboard web
   .\mvgn-scan\mvgn-scan.exe -watch -observe            # Ambos
   ```
3. O desde el atajo: `.\tools\observe.bat`
4. El observer muestra en tiempo real:
   - **Semáforo** de estado (⚪ pending, 🟡 in_progress, 🟢 completed, 🔴 failed)
   - **Health score** combinado (baseline + ATENEA governance)
   - **Progreso** con barra visual
   - **Event timeline** desde telemetry events
   - **Próxima acción** desde blackboard
   - **ATENEA score** — 9 condiciones de governance
5. Instalación permanente opcional:
   ```powershell
   .\tools\install-observer.ps1                          # PATH + startup
   .\tools\install-observer.ps1 -RegisterTask            + tarea programada
   ```

### Modo COLLAB — Sesión colaborativa

1. Preguntar: ¿Qué rol tendrá la otra IA?
   - **Auditor** → Revisión y reportes
   - **Desarrollador** → Generación de código
   - **Revisor** → Code review
   - **Personalizado** → Describir el rol
2. Actualizar/crear `.mvgn/blackboard/JOIN.md` con:
   - Rol seleccionado
   - Instrucciones específicas
   - Criterios de aceptación
3. Preguntar: ¿Tarea para la otra IA?
   - **Sí** → Actualizar `state.json` con nueva tarea
   - **No** → Dejar JOIN.md listo para cuando se necesite
4. Reportar: "Sesión colaborativa activada. Otra IA puede unirse vía JOIN.md"

---

## Handoff entre agentes

Cuando un agente completa su tarea, debe:

1. Actualizar `state.json`: agregar su nombre a `completed_agents`
2. Actualizar `handoff-latest.json` con contexto de handoff
3. Actualizar `JOIN.md` con instrucciones para el siguiente agente

Para que otra IA se una, dale esta instrucción:
```
Lee .mvgn/blackboard/JOIN.md y ejecuta lo que dice.
```

---

## Estructura del proyecto (LAYOUT/2.0 — consolidada)

Todo lo de MVGN vive en `.mvgn/`. El proyecto solo tiene lo suyo.

```
proyecto/
├── .mvgn/                        # ═══ TODO lo de MVGN ═══
│   ├── NYX.md                    # Este archivo — entry point universal
│   ├── NYX.ps1                   # Entry point imperativo
│   ├── MVGN.md                   # Documentación del framework
│   ├── mvgn-runtime.json         # Manifiesto del runtime
│   ├── LAYOUT.md                 # Contrato de estructura (LAYOUT/2.0)
│   ├── blackboard/
│   │   ├── state.json            # Estado actual (fuente de verdad)
│   │   ├── context.md            # Instrucciones generales
│   │   └── JOIN.md               # Punto de entrada para otras IAs (JOIN/2.0)
│   ├── templates/                # Fuente canónica de plantillas
│   ├── scripts/                  # Orquestador + audio pipeline [opcional]
│   ├── tools/                    # init, bootstrap, validadores
│   ├── audio/                    # Audio artifacts                     [runtime]
│   └── logs/execution.log                                                 [runtime]
├── docs/                         # ═══ SOLO documentación DEL PROYECTO ═══
├── src/                          # Código del proyecto
└── .gitignore                    # Ignora .mvgn/ → publicación segura por construcción
```

Instalación desde cero: `irm .../tools/bootstrap.ps1 | iex` (ver `.mvgn/tools/bootstrap.ps1`).

---

## Versión y changelog

| Versión | Fecha | Cambios |
|---------|-------|---------|
| v5.3.0 | 2026-08-26 | Separation: LAYOUT/2.0 consolidado (.mvgn/), distribución por GitHub Releases (bootstrap), JOIN/2.0 con cosmogonía CNX, audio GPU, CLI global mvgn
| v5.2.0 | 2026-07-30 | Audio Ingestion: Faster-Whisper/Whisper, context.md, pipeline AUDIO_DISCOVERY → CONTEXT_READY |
| v5.1.0 | 2026-07-30 | Refactorizado a Markdown puro universal. Extraída lógica imperativa a NYX.ps1 |
| v5.0 | 2026-07-25 | Blackboard Pattern. Modo COLLAB. JOIN.md. |
| v4.2 | 2026-07-15 | Telemetry v2. Session Trace Engine. SMG. |

---

*MVGN v5.3 — Multi-Agent Governance Framework*
*Costo: $0 | Sin dependencias de pago*

---

## Ecosistema MVGN

MVGN es el framework de gobierno multi-agente de Crimson Nexus Corp. Este entry point activa el **Blackboard Pattern**, una de sus capacidades.

| Sistema | Propósito |
|:--------|:----------|
| **MVGN Framework** | Framework de gobierno multi-agente. Estándares, trazabilidad, sesiones. |
| **Blackboard Pattern** | ⬅️ Capacidad activada por este entry point. |
| **KRONOS Runtime** | Motor de estado Go para vaults Obsidian. |
| **ATENEA** | Governance & QA — Auditoría y OER. |
| **CARONTE** | Mapa de tránsito del dominio MVGN. |
| **AEON Protocol** | Preservación referencial y sellado. |
| **CNX-CANON** | Canon organizacional de Crimson Nexus. |

*Parte de Crimson Nexus Corp. — Node MX-03*
