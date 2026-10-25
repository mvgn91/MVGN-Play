# 07 — Auditoría: Observation Console (Semáforo MVGN)

> **ID:** AUD-ATHN1-OBS-001
> **Propósito:** Evaluar la viabilidad técnica de una consola de observación en vivo para proyectos MVGN
> **Auditor:** Buffy (DeepSeek v4 Flash) vía Freebuff
> **Destino:** Mailbox ATHN1 vía KRONOS
> **Fecha:** 2026-07-30
> **Versión MVGN:** v5.2.0

---

## Resumen Ejecutivo

**Idea propuesta:** Una consola de observación que se ejecute dentro de cualquier proyecto que contenga MVGN y muestre:
1. Versión actual de MVGN
2. Última tarea realizada
3. Estado en tiempo real de lo que se está trabajando
4. Indicador tipo **semáforo** (🟢/🟡/🔴) del estado del proyecto

**Veredicto:** ✅ **Completamente viable** — El 70% de los building blocks ya existen. La implementación requiere 1 sprint (estimación: 2-3 días) y aprovecha el código Go existente de `mvgn-scan` + `mvgn-telemetry` + Blackboard.

---

## 1. Diagnóstico de Arquitectura Actual

### 1.1 Componentes existentes que ya sirven

| Componente | Propósito | Relevancia para Observation Console |
|:-----------|:----------|:-----------------------------------|
| **mvgn-scan** | CLI Go que escanea proyectos MVGN | ⭐ **Base** — Ya genera `dashboard-data.json`, tiene display UI ANSI, scanner y validator |
| **mvgn-telemetry** | Session Trace Engine en Go | ⭐ **Vital** — Captura eventos SESSION_STARTED/CLOSED, HEARTBEAT, STATE_CHANGED. TTL caching de 5s. |
| **Blackboard state.json** | Fuente de verdad del estado actual | ⭐ **Crítico** — Status: idle/pending/in_progress/completed/failed, fase actual, agentes completados |
| **mvgn-runtime.json** | Manifiesto de capacidades | ✅ Versión, actores, perfil, audio config |
| **display/cnx.go** | UI de terminal ANSI | ✅ Ya implementa: banners, boxes, metric cards, status tags con colores 🟢🟡🔴 |
| **docs/06_state_report.md** | Reporte de estado estático | Útil como snapshot inicial |

### 1.2 Mapa de fuentes de datos actuales y sus carencias

```
┌─────────────────────────────────────────────────────────┐
│                    SNAPSHOT (hoy)                        │
├─────────────────────────────────────────────────────────┤
│  mvgn-scan dashboard-data.json   ───  ✅ Snapshot frío  │
│  Blackboard state.json           ───  ✅ Estado actual  │
│  Telemetry events/*.jsonl        ───  ✅ Historial      │
│  mvgn-runtime.json               ───  ✅ Versión        │
│  .mvgn-context.json              ───  ❌ OBSOLETO       │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│                   OBSERVATION CONSOLE (target)           │
├─────────────────────────────────────────────────────────┤
│  Live polling state.json + events    ───  ⬜ Por hacer  │
│  Semáforo visual                     ───  ⬜ Por hacer  │
│  Timeline de tareas recientes        ───  ⬜ Por hacer  │
│  Health score + métricas             ───  ⬜ Por hacer  │
└─────────────────────────────────────────────────────────┘
```

---

## 2. Análisis de Viabilidad por Componente

### 2.1 Semáforo de estado (🟢🟡🔴⚪)

| Estado Blackboard | Color Semáforo | Acción recomendada |
|:-----------------|:---------------|:-------------------|
| `idle` | ⚪ **Gris** | Sin tarea activa — sistema en reposo |
| `pending` | 🔵 **Azul** | Tarea pendiente de ejecutar — no iniciada |
| `in_progress` | 🟡 **Amarillo** | Trabajando activamente — semáforo en precaución |
| `completed` | 🟢 **Verde** | Última tarea completada — semáforo en verde |
| `failed` | 🔴 **Rojo** | Error activo — semáforo en rojo, requiere recovery |

**Viabilidad:** ✅ **Inmediata.** El `display/cnx.go` ya tiene `StatusTag()` que renderiza colores verde/amarillo/rojo. Solo hay que mapear `state.json.status` a los colores, respetando la semántica de semáforo universal: ⚪=idle, 🔵=pending, 🟡=in_progress, 🟢=completed, 🔴=failed.

### 2.2 Versión actual

**Fuente:** `mvgn-runtime.json` → `runtime_version`
**Viabilidad:** ✅ **Ya existe.** `mvgn-scan` ya lee y muestra este campo.

### 2.3 Última tarea realizada

**Fuentes disponibles:**
- `state.json.task_description` + `state.json.completed_agents`
- Telemetry `events/*.jsonl` → SESSION_CLOSED events
- `docs/06_state_report.md` → Última acción
- `docs/04_changelog.md` → Historial de cambios

**Viabilidad:** ✅ **Inmediata.** `state.json` ya contiene `task_description` y `completed_agents`.

### 2.4 Monitoreo en tiempo real

**Opción A (Recomendada): Polling ligero**
- Watch `state.json` cada N segundos (configurable: 2s-10s)
- También leer `events/*.jsonl` para últimos eventos
- Caching TTL de 5s (ya implementado en telemetry)
- ✅ Mínimo impacto — `state.json` es <1KB

**Opción B (Futuro): File system watcher**
- Usar `fsnotify` en Go para detectar cambios en tiempo real
- Elimina la necesidad de polling
- ⚠️ Más complejo, no crítico para MVP

**Viabilidad:** ✅ **Alta.** La opción A es trivial de implementar.

### 2.5 Componentes que NO existen y habría que crear

| Componente | Esfuerzo | Prioridad | Notas |
|:-----------|:---------|:----------|:------|
| **--watch mode** en mvgn-scan | 🔷 Bajo (4h) | P0 | Agregar flag `-watch` con polling loop |
| **Traffic light display** | 🔷 Bajo (2h) | P0 | Mapear status_tag a símbolos de semáforo |
| **Timeline de tareas** | 🔷 Bajo (3h) | P1 | Leer últimos N eventos de telemetry |
| **Health score gauge** | 🔶 Medio (6h) | P2 | Score compuesto: completitud, errores, actividad |
| **Instalación one-command** | 🔶 Medio (4h) | P1 | Script `tools/install-observer.ps1` |
| **Dashboard HTML export** | 🔴 Alto (16h) | P3 | Versión web HTML+JS bonita |

---

## 3. Arquitectura Propuesta para la Observation Console

#### ⚠️ Gap arquitectónico importante: mvgn-scan NO lee state.json

El scanner actual (`scanner.go`) busca exclusivamente dos archivos:
- `mvgn-runtime.json` → metadatos estáticos del runtime
- `.mvgn-context.json` → contexto del proyecto (schema v3.0)

**Pero el blackboard de MVGN v5.2 usa `state.json`** (`.mvgn/blackboard/state.json`), que es un archivo diferente con un schema distinto. El `consolidator` actualmente **ignora por completo** `state.json`.

```
Lo que mvgn-scan lee hoy:          Lo que necesita leer:
┌──────────────────────┐           ┌──────────────────────┐
│ mvgn-runtime.json    │   ✅      │ mvgn-runtime.json    │
│ .mvgn-context.json   │   ✅      │ state.json       ← ❌ FALTA
└──────────────────────┘           └──────────────────────┘
```

**Impacto:** Para la Observation Console, habrá que agregar un **nuevo data source** que lea `state.json` y extraiga: `status`, `task_description`, `completed_agents`, `current_phase`, `last_error`. Esto es ~30 líneas de Go sobre el patrón ya existente en `consolidator.go`.

---

## 3.2 Extensión de mvgn-scan (enfoque recomendado)

```
mvgn-scan
├── cmd/mvgn-scan/main.go
│   ├── (existente) scan, validate, consolidate
│   └── (NUEVO)  -watch              ← modo observación
│                 -interval 5         ← intervalo de polling (segundos)
│                 -project "./"       ← proyecto a observar
│
├── internal/
│   ├── scanner/       (existente)    ← escanea proyectos
│   ├── validator/     (existente)    ← valida archivos
│   ├── consolidator/  (existente)    ← genera dashboard
│   ├── display/       (existente)    ← UI ANSI 🟢🟡🔴
│   ├── models/        (existente)    ← estructuras de datos
│   └── observer/      (NUEVO)       ← lógica de observación en vivo
│       ├── watch.go                  ← polling loop + semáforo
│       ├── timeline.go              ← historial de eventos
│       └── status.go                ← estado en vivo
│
└── output/
    └── dashboard-data.json  (existente)
```

### 3.3 Pipeline de datos en vivo

```
┌──────────────┐     poll cada Ns     ┌──────────────────────┐
│  state.json  │ ───────────────────→ │                      │
│  (blackboard)│                      │   OBSERVER ENGINE    │
└──────────────┘                      │   (watch.go)         │
                                      │                      │
┌──────────────┐     poll cada Ns     │  ├── Leer state.json │
│ events/*.json│ ───────────────────→ │  ├── Leer runtime    │
│  (telemetry) │                      │  ├── Leer events     │
└──────────────┘                      │  └── Calcular score  │
                                      │                      │
┌──────────────┐     poll cada Ns     │       │              │
│ runtime.json │ ───────────────────→ │       ▼              │
│  (manifest)  │                      │  ┌─────────────┐     │
└──────────────┘                      │  │   DISPLAY   │     │
                                      │  │  ┌───────┐  │     │
                                      │  │  │ 🟢  🔵 │  │     │
                                      │  │  │ 🟡  ⚪ │  │     │
                                      │  │  │ 🔴     │  │     │
                                      │  │  └───────┘  │     │
                                      │  └─────────────┘     │
                                      └──────────────────────┘
```

---

## 4. Evaluación de Esfuerzo

### Por enfoque

| Enfoque | Esfuerzo | Complejidad | Resultado |
|:--------|:---------|:------------|:----------|
| **🧩 A) Extender mvgn-scan** (recomendado) | 2-3 días | 🔷 Baja | CLI ANSI con semáforo, watch mode, timeline |
| **🌐 B) Dashboard web standalone** | 5-7 días | 🔴 Alta | Servidor web con HTML/JS/SSE, charts, histórico |
| **⚡ C) Script PowerShell** | 1-2 días | 🔷 Baja | Loop PS que lee state.json cada Ns, colorea output |
| **🔌 D) Plugin para editors** | 10-15 días | 🔴🔴 Muy alta | Extensión VS Code + Obsidian |

### Recomendación: Enfoque A (extender mvgn-scan) + C (script PS rápido como complemento)

> **Fundamento:** `mvgn-scan` ya tiene display, scanner, y lógica de lectura. Agregar `--watch` mode es **~50 líneas de Go** sobre infraestructura existente. La UI ANSI con semáforo ya existe en `display/cnx.go`.

---

## 5. Roadmap Sugerido

| Fase | Qué | Tiempo | Dependencias |
|:-----|:----|:-------|:-------------|
| **MVP** | `mvgn-scan --watch` con semáforo 🟢🟡🔴⚪ + versión + última tarea | 2 días | Ninguna |
| **Fase 2** | Timeline de últimos eventos + health score | +2 días | MVP |
| **Fase 3** | Script `tools/observe.ps1` para ejecución one-liner | +1 día | MVP |
| **Fase 4 (futuro)** | Dashboard HTML export + live web server | +5 días | Fase 2 |

---

## 6. Riesgos y Mitigaciones

| Riesgo | Impacto | Probabilidad | Mitigación |
|:-------|:--------|:-------------|:-----------|
| state.json no actualizado en tiempo real | 🟡 Semáforo desactualizado | Media | Mostrar timestamp de última actualización. El polling loop lo detecta. |
| Proyecto sin blackboard (UPGRADE necesario) | 🔴 No muestra nada | Baja | El scanner ya detecta esto. Mostrar mensaje "Ejecuta UPGRADE primero". |
| Múltiples proyectos escaneados | 🟡 ¿Cuál observar? | Baja | Flag `-project` explicito. En `--watch` solo se observa 1 proyecto. |
| Rendimiento con polling agresivo | 🟠 CPU | Baja | Intervalo default 5s. Configurable. state.json <1KB. |

---

## 7. Conclusión

La Observation Console de MVGN es **completamente viable** con la arquitectura actual. Los building blocks existen:

- ✅ **Datos**: state.json, mvgn-runtime.json, telemetry events
- ✅ **Display**: display/cnx.go con ANSI, colores, status tags
- ✅ **Scanner**: Ya encuentra proyectos MVGN
- ✅ **Telemetry**: Ya captura eventos de sesión con TTL caching

Lo único que falta es:
1. Un **modo `--watch`** en `mvgn-scan` que haga polling cíclico
2. **Mapeo status→semáforo** (10 líneas de Go)
3. **Timeline** de últimos eventos desde telemetry

**Estimación total: ~2-3 días de implementación por un desarrollador Go.**

> 🟢 **Luz verde para proceder con la implementación.**

---

## Apéndice A: Mockup de la Observation Console

```
MVGN-SCN-001 · OBSERVATION INTERFACE

┌─────────────────────────────────────────────────────────────┐
│  ╔═══════════════════════════════════════════════════════╗  │
│  ║   ███╗   ███╗██╗   ██╗ ██████╗ ███╗   ██╗           ║  │
│  ║   ████╗ ████║██║   ██║██╔════╝ ████╗  ██║           ║  │
│  ║   ██╔████╔██║██║   ██║██║  ███╗██╔██╗ ██║           ║  │
│  ║   ██║╚██╔╝██║╚██╗ ██╔╝██║   ██║██║╚██╗██║           ║  │
│  ║   ██║ ╚═╝ ██║ ╚████╔╝ ╚██████╔╝██║ ╚████║           ║  │
│  ║   ╚═╝     ╚═╝  ╚═══╝   ╚═════╝ ╚═╝  ╚═══╝           ║  │
│  ╚═══════════════════════════════════════════════════════╝  │
│                                                             │
│  MVGN-SCN-002 · LIVE OBSERVATION                            │
│                                                             │
│  ┌──────────────────────┐  ┌──────────────────────┐        │
│  │         5.2          │  │        🟢            │        │
│  │       VERSION        │  │       STATUS         │        │
│  └──────────────────────┘  └──────────────────────┘        │
│                                                             │
│  ╔══════════════════════════════════════════════════════════╗│
│  ║  LATEST TASK                                             ║│
│  ║  ─────────────────────────────────────────────────────  ║│
│  ║  PRD-MVGN-UPDATE-001 · Audio Ingestion (ATHN1)          ║│
│  ║  Completed agents: Buffy, OpenCode                       ║│
│  ║  Artifacts: 12 generated                                 ║│
│  ╚══════════════════════════════════════════════════════════╝│
│                                                             │
│  ╔══════════════════════════════════════════════════════════╗│
│  ║  EVENT TIMELINE                                          ║│
│  ║  ─────────────────────────────────────────────────────  ║│
│  ║  [14:32:01] 🟢 STATE_CHANGED: pending → in_progress     ║│
│  ║  [14:32:05] ♥ HEARTBEAT — system alive                  ║│
│  ║  [14:32:10] 📋 TASK_UPDATED: Fase 2/3                  ║│
│  ║  [14:32:15] 🟡 STATE_CHANGED: in_progress → completed   ║│
│  ╚══════════════════════════════════════════════════════════╝│
│                                                             │
│  ───────────────────────────────────────────────────────────│
│  Watching ./project · Polling: every 5s · Last update: ... │
│  Crimson Nexus Corporation · MVGN Labs · Node MX-03        │
│  ───────────────────────────────────────────────────────────│
```

---

## Apéndice B: Manejo de corrupción en --watch mode

Si `state.json` se corrompe durante un ciclo de polling (ej: archivo truncado durante escritura, JSON malformado), el modo `--watch` debe:

1. **No crashear** — mostrar `⚠️ state.json corrupto — reintentando...` con timestamp
2. **Conservar el último estado válido** en memoria para no perder el semáforo
3. **Reintentar** en el siguiente ciclo de polling
4. Si persiste > 3 ciclos consecutivos, sugerir: "Ejecuta recovery: Lee docs/mvgn-recovery.md"

---

## Apéndice C: Mapeo a sistemas del ecosistema MVGN

| Sistema | Rol en esta iniciativa |
|:--------|:----------------------|
| **MVGN Framework** | Framework base — la Observation Console es una capacidad de MVGN |
| **Blackboard Pattern** | Fuente de datos viva (`state.json`) |
| **KRONOS Runtime** | 🚀 **Destino del reporte** — Mailbox ATHN1 |
| **ATENEA** | Governance & QA — Revisión de la auditoría |
| **CARONTE** | Mapa de tránsito — La Observation Console sería un nuevo nodo |
| **NYX** | Entry point — El modo OBSERVE se integraría en NYX.md |
| **mvgn-scan** | 🛠️ **Base de implementación** |

---

*Documento generado por Buffy (Freebuff) para Mailbox ATHN1 — MVGN Labs v5.2*
*Crimson Nexus Corp. — Node MX-03*
