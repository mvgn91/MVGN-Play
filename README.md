<div align="center">

# 🧠 MVGN — Multi-Agent Governance Framework v5.3

**Framework de gobierno multi-agente de Crimson Nexus Corp.** Incluye el patrón **Blackboard Pattern** como capa de abstracción para la creación, gestión y documentación de proyectos desde la idea hasta su publicación. Incluye **Audio Ingestion** con Faster-Whisper para transcribir notas de voz y reuniones directamente a PRDs. Cero costo. Cero dependencias de pago.

<p>
  <img src="https://img.shields.io/badge/version-5.3-teal?style=flat-square&labelColor=1A0B2E" alt="Version 5.3">
  <img src="https://img.shields.io/badge/blackboard-pattern-blue?style=flat-square&labelColor=1A0B2E" alt="Blackboard Pattern">
  <img src="https://img.shields.io/badge/observer--console-live-brightgreen?style=flat-square&labelColor=1A0B2E" alt="Observer Console">
  <img src="https://img.shields.io/badge/atenea--governance-scoring-orange?style=flat-square&labelColor=1A0B2E" alt="ATENEA Governance">
  <img src="https://img.shields.io/badge/audio--ingestion-faster--whisper-brightgreen?style=flat-square&labelColor=1A0B2E" alt="Audio Ingestion">
  <img src="https://img.shields.io/badge/governance-framework-violet?style=flat-square&labelColor=1A0B2E" alt="Governance Framework">
  <img src="https://img.shields.io/badge/collab--ready-yes-brightgreen?style=flat-square&labelColor=1A0B2E" alt="Collab Ready">
  <img src="https://img.shields.io/badge/license-MIT-brightgreen?style=flat-square&labelColor=1A0B2E" alt="License MIT">
</p>

**[¿Qué es MVGN?](#qué-es-mvgn)** •
**[Blackboard Pattern](#blackboard-pattern)** •
**[Audio Ingestion](#-audio-ingestion)** •
**[Empezar](#empezar-en-3-pasos)** •
**[Docs](docs/)** •
**[Ecosistema](#-ecosistema-mvgn)** •
**[Contribuir](#contribuir)**

</div>

---

## ¿Qué es MVGN?

**MVGN** es el framework de gobierno multi-agente de **Crimson Nexus Corp.** Proporciona estándares, trazabilidad, session management y múltiples capacidades para coordinar proyectos con IA.

Una de esas capacidades es el **Blackboard Pattern**: una capa de abstracción que coordina múltiples asistentes de IA dentro de un mismo proyecto usando estado compartido, handoff explícito y ejecución secuencial.

> Este repositorio contiene el starter kit del **Blackboard Pattern** de MVGN. Para el framework MVGN completo (KRONOS, ATENEA, estándares), consulta el vault oficial de Crimson Nexus.

```
  ┌──────────┐     ┌──────────┐     ┌──────────┐
  │ OpenCode │ ──→ │ FreeBuff │ ──→ │ MiMoCode │
  │ Fase 1   │     │ Fase 2   │     │ Fase 3   │
  └────┬─────┘     └────┬─────┘     └────┬─────┘
       │                │                │
       └────────────────┴────────────────┘
                        │
                 ┌──────▼──────┐
                 │  state.json │ ← Fuente de verdad
                 └─────────────┘
```

### Filosofía

> MVGN no intenta escribir mejor código. MVGN intenta que ningún cambio importante ocurra sin contexto, sin trazabilidad y sin responsabilidad definida.

### Características

| Característica | Detalle |
|:---------------|:--------|
| 🧠 **Multi-agente** | Coordina OpenCode, FreeBuff, MiMoCode y más |
| 📋 **Blackboard** | `state.json` como única fuente de verdad |
| 👁️ **Observation Console** | Terminal ANSI + Web dashboard en tiempo real con SSE |
| 📊 **ATENEA Scoring** | 9 condiciones de governance — health score combinado |
| 🤝 **Handoff explícito** | Contexto entre agentes vía `JOIN.md` + `handoff-latest.json` |
| 🎙️ **Audio Ingestion** | Transcribe audios (mp3, wav, m4a, flac, ogg) → contexto → PRD con Faster-Whisper |
| 💰 **Costo $0** | Sin APIs de pago, sin dependencias externas |
| 🔄 **Dual entry** | `NYX.md` (Markdown universal) + `NYX.ps1` (PowerShell) |
| 🏛️ **Legacy preservado** | Migraciones seguras con `archive/` |

---

## 🎙️ Audio Ingestion

MVGN puede **transcribir automáticamente archivos de audio** durante la fase **INIT** y convertirlos en documentos estructurados para generar PRDs.

### Pipeline de Audio

```
INIT → AUDIO_DISCOVERY → TRANSCRIBING → CONTEXT_BUILDING → CONTEXT_READY → PRD
```

### Cómo usarlo

1. Coloca tu archivo de audio (`.mp3`, `.wav`, `.m4a`, `.flac`, `.ogg`) en el directorio del proyecto
2. Ejecuta el pipeline de audio:

```powershell
.\scripts\audio\transcribe.ps1
```

3. O especifica un archivo directamente:

```powershell
.\scripts\audio\transcribe.ps1 -AudioFile "grabacion.mp3" -Model "turbo" -InstallDeps
```

### Dependencias

```bash
pip install faster-whisper
```

O usa el flag automático:
```powershell
.\scripts\audio\transcribe.ps1 -InstallDeps
```

### Artefactos generados

```
.mvgn/audio/{archivo}/
├── transcript.md         # Transcripción completa con timestamps
├── metadata.json         # Metadatos del modelo y audio
├── segments.json         # Segmentos temporizados
├── context.md            # Contexto estructurado para PRD
└── context-summary.json  # Resumen para state.json
```

---

## Empezar en 3 pasos

### 1. Instala MVGN en tu proyecto (una línea)

Desde la raíz de tu proyecto (nuevo o existente):

```powershell
irm https://raw.githubusercontent.com/mvgn91/mvgnlabs-starter-kit/main/tools/bootstrap.ps1 | iex
```

El bootstrap instala todo el motor bajo `.mvgn/` e inicializa el blackboard. Tu proyecto solo conserva lo suyo: `docs/`, `src/` y `.gitignore`.

> Alternativa para contribuir al kit: `git clone https://github.com/mvgn91/mvgnlabs-starter-kit.git && cd mvgnlabs-starter-kit`

### 2. Carga NYX.md en cualquier IA

Copia el contenido de [`.mvgn/NYX.md`](NYX.md) y pégalo en tu sesión de IA favorita:

**✅ ChatGPT · Claude · Gemini · DeepSeek · Cualquier modelo**

La IA leerá `NYX.md` y automáticamente:

1. **Detectará** el estado del proyecto (INIT / READY / UPGRADE)
2. **Creará** la estructura `.mvgn/blackboard/` si es nueva
3. **Escaneará** archivos de audio para transcribir (opcional)
4. **Te preguntará** qué quieres hacer

### 3. ¡Listo!

Empieza a dar instrucciones. La IA usará MVGN para coordinar todo.

### Mantente actualizado

```powershell
.\.mvgn\tools\bootstrap.ps1          # verifica y pregunta antes de actualizar
.\.mvgn\tools\bootstrap.ps1 -Force   # actualiza sin preguntar
```

Muestra el changelog delta entre tu versión y el último release, y preserva siempre tu blackboard, tus docs y tu `.gitignore`.

---

## Cómo funciona

### Pipeline de orquestación

Cuando ejecutas una tarea con `orchestrate.ps1`:

```powershell
# PowerShell
.\scripts\orchestrate.ps1 -Task "Implementar autenticación"

# Bash
bash scripts/orchestrate.sh --task "Implementar autenticación"
```

El pipeline ejecuta 3 fases secuenciales:

| Fase | Agente | Qué hace | Salida |
|:-----|:-------|:---------|:-------|
| 1 | **OpenCode** | Genera código | `src/` |
| 2 | **FreeBuff** | Revisa y optimiza | `src/review-report.md` |
| 3 | **MiMoCode** | Tests y documentación | `tests/` + `docs/` |

Cada fase actualiza `state.json` y deja contexto para la siguiente mediante handoff.

### Entry points

| Archivo | Propósito | Compatibilidad |
|:--------|:----------|:---------------|
| [`NYX.md`](NYX.md) | Entry point universal | **Cualquier** modelo de IA |
| [`NYX.ps1`](NYX.ps1) | Entry point PowerShell | PowerShell 5.1+ |

### Handoff entre agentes

```
1. Agente completa su fase
2. Actualiza state.json
3. Genera handoff-latest.json
4. Actualiza JOIN.md
5. Siguiente agente continúa
```

Para que otra IA se una al proyecto:
```
Lee .mvgn/blackboard/JOIN.md y ejecuta lo que dice.
```

---

## Estructura del proyecto

```
mi-proyecto/
├── .mvgn/
│   ├── blackboard/          ← Estado compartido (Blackboard)
│   │   ├── state.json       ← Fuente de verdad
│   │   ├── context.md       ← Instrucciones para agentes
│   │   ├── JOIN.md          ← Onboarding de otras IAs
│   │   └── logs/
│   └── audio/               ← Audio artifacts (transcripciones)
│       └── {archivo}/
│           ├── transcript.md
│           ├── context.md
│           └── metadata.json
├── docs/                    ← Documentación del proyecto
│   ├── mvgn-kernel.md       ← Reglas de comportamiento IA
│   ├── mvgn-recovery.md     ← Protocolo de recuperación
│   └── mvgn-finalization.md ← Protocolo de finalización
├── scripts/
│   ├── orchestrate.ps1      ← Motor de orquestación
│   ├── orchestrate.sh       ← Motor Bash
│   ├── agents/              ← Adaptadores por agente
│   └── audio/               ← Pipeline de audio
│       ├── transcribe.ps1   ← Orquestador
│       ├── transcriber.py   ← Transcripción Faster-Whisper
│       └── context-builder.py ← Generador de contexto
├── tools/                   ← Herramientas de sistema
│   ├── mvgn-init.ps1        ← Init de proyecto
│   └── validate-entry-points.ps1
├── src/                     ← Código fuente (vacío, para empezar)
├── tests/                   ← Tests (vacío, para empezar)
├── archive/                 ← Material legacy (si hay migraciones)
├── NYX.md                   ← Entry point universal
├── NYX.ps1                  ← Entry point PowerShell
├── mvgn-runtime.json        ← Manifiesto del runtime (en .mvgn/ desde LAYOUT/2.0)
└── README.md                ← Este archivo
```

---

## Documentación

| Documento | Propósito |
|:----------|:----------|
| [`NYX.md`](NYX.md) | Entry point universal |
| [`MVGN.md`](MVGN.md) | Documentación completa del sistema |
| [`docs/mvgn-kernel.md`](docs/mvgn-kernel.md) | Comportamiento IA y reglas |
| [`docs/mvgn-recovery.md`](docs/mvgn-recovery.md) | Protocolo de recuperación |
| [`docs/mvgn-finalization.md`](docs/mvgn-finalization.md) | Protocolo de finalización |
| [`docs/01_prd.md`](docs/01_prd.md) | Cómo escribir un PRD |

---

## Comandos rápidos

| Comando | Acción |
|:--------|:-------|
| `.\NYX.ps1 -Mode detect` | Detectar estado del proyecto |
| `.\NYX.ps1 -Mode ready` | Ver estado actual |
| `.\NYX.ps1 -Mode observer` | Observer en terminal |
| `.\NYX.ps1 -Mode observer -Web` | Observer en web dashboard (http://localhost:8080) |
| `.\tools\observe.bat` | Atajo doble-click (terminal + web) |
| `.\tools\install-observer.ps1` | Instalar observer en PATH del sistema |
| `.\tools\install-observer.ps1 -RegisterTask` | + inicio automático con Windows |
| `.\tools\mvgn-init.ps1` | Inicializar proyecto (modo INIT) |
| `.\tools\mvgn-init.ps1 -Collab` | Activar colaboración multi-agente |
| `.\scripts\orchestrate.ps1 -Task "desc"` | Ejecutar pipeline multi-agente |
| `.\scripts\orchestrate.ps1 -Status` | Ver estado del pipeline |
| `.\scripts\audio\transcribe.ps1` | Transcribir audio → contexto → PRD |
| `.\scripts\audio\transcribe.ps1 -InstallDeps` | Instalar Faster-Whisper + transcribir |
| `.\tools\validate-entry-points.ps1` | Validar sincronización NYX.md ↔ NYX.ps1 |

---

## Contribuir

1. Fork este repo
2. Crea tu branch: `git checkout -b feature/mi-feature`
3. Sigue el flujo MVGN para tu feature
4. Abre un Pull Request

¿Ideas? ¿Problemas? Abre un [issue](https://github.com/mvgn91/mvgnlabs-starter-kit/issues).

---

## 🌐 Ecosistema MVGN

MVGN es el framework de gobierno multi-agente de Crimson Nexus Corp. Este starter kit implementa una de sus capacidades: el **Blackboard Pattern**.

| Sistema | Propósito |
|:--------|:----------|
| **MVGN Framework** | Framework de gobierno multi-agente. Estándares, trazabilidad, sesiones. |
| **Blackboard Pattern** | ⬅️ **Este kit.** Capa de abstracción para proyectos de idea a publicación. |
| **Observation Console** | ⬅️ **Nuevo.** Terminal ANSI + Web dashboard con SSE, ATENEA health scoring. |
| **Audio Ingestion** | ⬅️ Transcripción de voz a contexto estructurado con Faster-Whisper. |
| **KRONOS Runtime** | Motor de estado operacional en Go para vaults Obsidian. |
| **ATENEA** | Activo Cognitivo de Governance & QA. Elabora OER. |
| **CARONTE** | Mapa de tránsito del dominio MVGN. |
| **CNX-CANON** | Canon de Crimson Nexus. |
| **AEON Protocol** | Preservación referencial y sellado de integridad. |

---

## Licencia

MIT — MVGN Labs / Crimson Nexus Corp.

---

<div align="center">

**MVGN v5.3** — Multi-Agent Governance Framework

*Parte del ecosistema Crimson Nexus Corp. — Node MX-03*

</div>
