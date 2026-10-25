# 03 — Tasks

> Desglose de tareas del proyecto — Ciclo v5.3 "Separation" (SEP-001).

---

## Ciclo v5.3 — Separation (SEP-001)

> Objetivo: separar el DESARROLLO de MVGN (este lab) de su IMPLEMENTACIÓN en proyectos
> del usuario. Un usuario externo obtiene MVGN con un solo comando (`irm ... | iex`)
> que trae siempre la última versión desde GitHub Releases, sin historial git ni
> artefactos de desarrollo. Instancia por proyecto; todo lo runtime queda excluido
> al publicar.

### Fase 1 — Cleanup

| ID | Descripción | Prioridad | Estado |
|:---|:------------|:----------|:-------|
| T-101 | Eliminar directorios vacíos heredados (analytics, core, adapters, kernel, event-history, session, snapshots, policies, profiles, skills, output*, traceability*) y auditar qué queda | Alta | ✅ Completada |
| T-102 | Consolidar `.gitignore` canónico: cubrir TODA ruta runtime de MVGN (events/, reports/, .mvgn/audio/, telemetría, event-history, snapshots, session, output) | Alta | ✅ Completada* |
| T-103 | Renumerar docs duplicados (dos `07_*`) y fijar convención de numeración `NN_` | Media | ✅ Completada |

### Fase 2 — Payload y templates

| ID | Descripción | Prioridad | Estado |
|:---|:------------|:----------|:-------|
| T-104 | Definir manifiesto de payload (`templates/payload-manifest.json`): lista exacta de archivos que constituyen la herramienta distribuida | Alta | ✅ Completada |
| T-105 | Poblar `templates/` con docs base como archivos fuente únicos (00_idea, 01_prd, 02_architecture, 03_tasks, 04_changelog) — fin de los here-strings en scripts | Alta | ✅ Completada |
| T-106 | Escribir contrato de layout versionado (`templates/LAYOUT.md`): estructura canónica de un proyecto MVGN, única fuente de verdad de carpetas | Alta | ✅ Completada |
| T-107 | Crear `.gitignore` de publicación para proyectos que adoptan MVGN (excluye todo lo runtime) dentro del payload | Alta | ✅ Completada* |
| T-115 | Estandarizar `templates/JOIN.md` como protocolo canónico de comunicación inter-modelo (JOIN/2.0) con cosmogonía CNX: panteón operativo, designadores AAAA-##, reglas de tránsito y contrato de estado | Alta | 🔍 En revisión |

### Fase 3 — Scaffolder

| ID | Descripción | Prioridad | Estado |
|:---|:------------|:----------|:-------|
| T-108 | Crear `tools/bootstrap.ps1`: entry point de una línea (`irm raw.../bootstrap.ps1 \| iex`) — descarga último release, extrae payload en cwd sin `.git`, invoca init | Alta | ✅ Completada |
| T-109 | Refactorizar `tools/mvgn-init.ps1`: eliminar here-strings, copiar docs desde `templates/` según LAYOUT.md — init determinista | Alta | ✅ Completada |

### Fase 4 — Release pipeline

| ID | Descripción | Prioridad | Estado |
|:---|:------------|:----------|:-------|
| T-110 | Crear `tools/package-release.ps1`: arma `mvgn-vX.Y.Z.zip` SOLO con el manifiesto de payload + bump de versión | Alta | ✅ Completada |
| T-111 | Publicar GitHub Release `v5.3.0` con zip de payload adjunto | Alta | ⏸️ Esperando autorización del usuario |
| T-112 | Extender `NYX.ps1 -Mode update`: actualizar vía descarga de release (reemplazo de archivos) en vez de `git pull` | Alta | ✅ Completada |

### Fase 5 — Documentación y validación

| ID | Descripción | Prioridad | Estado |
|:---|:------------|:----------|:-------|
| T-113 | Actualizar NYX.md, MVGN.md, README.md: flujo de adopción externa (bootstrap → init → work → update) | Media | 🔄 Corrección (H-15: README) |
| T-114 | Validación E2E: carpeta limpia → bootstrap → init → verificar estructura == LAYOUT.md → update vía release | Alta | 🔄 Corrección (H-14: versión en init) |

### Fase 6 — Integración CLI (ampliación SEP-001)

| ID | Descripción | Prioridad | Estado |
|:---|:------------|:----------|:-------|
| T-116 | tools/install-profile.ps1: bloque idempotente en perfil PowerShell con función global mvgn (resolución ascendente de NYX.ps1) + tab-completion de modos | Alta | ✅ Completada |
| T-117 | cnx.ps1 (KRONOS RUNTIME): Invoke-GitClone reemplazado por flujo bootstrap vía GitHub Releases; textos de menú actualizados | Alta | ✅ Completada* |
| T-118 | Importar módulo de ingesta de audio probado (GPU/CUDA: transcriber.py + transcribe.ps1 v1.2 con fix DLLs CUDA, batch, processed.json); simplificar a punto único tools/mvgn-audio.ps1; disponible en INIT y en retomos (READY) para ciclos de features/correcciones; NYX.md AUDIO actualizado | Alta | ✅ Completada |
| T-119 | Consolidación LAYOUT/2.0: TODO MVGN bajo .mvgn/ — manifiesto 2.0, gitignore.template de 1 línea, init/bootstrap/audio con ProjectRoot derivado de script, perfil bloque v2 (.mvgn/NYX.ps1), cnx detecta .mvgn/, NYX.md documentado; incidente bootstrap revertido reportado y resuelto; tests fresh+update E2E verdes | Alta | 🔍 En revisión |

> Supervisión: FreeBuff revisa los artefactos al cierre de CADA fase antes de pasar a la siguiente.
> *output\*: decidir en T-101 si se conserva o se migra a ruta ignorada.
>
> **Auditoría pre-sellado Fase 5 (FreeBuff, 2026-08-25): ❌ RECHAZADA.** H-14 (ALTA):
> `mvgn-init.ps1:135` hardcodea runtime_version 5.2.0 — instalación fresca v5.3.0 se
> autorreporta desactualizada. H-15 (MEDIA): README sin flujo de adopción. Re-sellar tras correcciones.
>
> **Re-auditoría Fase 4 (FreeBuff, 2026-08-25): ✅ APROBADA.** H-6/H-3 corregidas y verificadas empíricamente
> (zip 24 archivos diff=0; fresh install == LAYOUT/2.0; update E2E en sandbox preservando blackboard/docs).
> Abiertos: H-9 (media), H-13 (baja). T-111 espera autorización del usuario. **Fase 5 habilitada.**
>
> **Veredicto Fase 3 (FreeBuff, 2026-08-25): APROBADA.** Init determinista verificado empíricamente (2 sandboxes idénticas + corrida doble idempotente); bootstrap instala desde release sin .git (payload por manifiesto).
>
> **Veredicto Fase 4 (FreeBuff, 2026-08-25): RECHAZADA — requiere corrección.**
> Zip local == manifiesto ✅ (21 archivos, diff=0; H-4 resuelto). Pero: H-6 (ALTA) `bootstrap.ps1:175` — en flujo de actualización `$tmp\extracted` nunca se crea (la descarga va a `$tmpPre`, líneas 128–146, y se borra en `finally`), por lo que `-Mode update` y la auto-actualización de bootstrap fallan siempre. Además H-3 sigue abierto. T-111 queda a espera de confirmación del usuario y de Fase 4 aprobada.
>
> **Veredicto Fase 1 (FreeBuff, 2026-08-25): APROBADA CON OBSERVACIONES.**
> Hallazgos: H-1 (baja) `.kimchi/ferments` vacío sin justificar · H-2 (baja) `mvgn-scan/internal/bridge` vacío sin justificar ·
> H-3 (media) `.gitignore:20` — regla `events/` engloba `events/event.schema.json` (fuente trackeado).
> **H-3 es prerrequisito para aprobar la Fase 4.**

---

## Tareas Completadas (historial)

| ID | Descripción | Prioridad | Estado |
|:---|:------------|:----------|:-------|
| T-001 | Crear scripts/audio/transcriber.py — Wrapper Python para Faster-Whisper + OpenAI Whisper | Alta | ✅ Completada |
| T-002 | Crear scripts/audio/context-builder.py — Generador de context.md estructurado con 12 secciones | Alta | ✅ Completada |
| T-003 | Crear scripts/audio/transcribe.ps1 — Orquestador PowerShell (detecta, selecciona, transcribe, genera contexto) | Alta | ✅ Completada |
| T-004 | Actualizar tools/mvgn-init.ps1 — Agregar AUDIO_DISCOVERY durante INIT | Alta | ✅ Completada |
| T-005 | Actualizar NYX.md — Nuevo modo AUDIO, flujo INIT con audio, estructura .mvgn/audio/ | Alta | ✅ Completada |
| T-006 | Actualizar MVGN.md — Documentar Audio Ingestion, estados extendidos, dependencias | Media | ✅ Completada |
| T-007 | Actualizar mvgn-runtime.json — v5.2, capability audio_ingestion, config de audio | Alta | ✅ Completada |
| T-008 | Actualizar README.md — Badge audio, comandos, estructura, sección de audio | Media | ✅ Completada |
| T-009 | Fix post-review: Agregar parámetro -Engine (whisper/faster-whisper) | Alta | ✅ Completada |
| T-010 | Fix post-review: Reforzar RF-006 (context.md como única fuente de PRD) | Alta | ✅ Completada |
| T-011 | Fix post-review: Sincronizar versiones NYX.md + mvgn-runtime.json a v5.2 | Media | ✅ Completada |
| T-012 | Documentación inversa: PRD, Arquitectura, Tasks, Changelog, Lessons, State Report | Media | ✅ Completada |

---

## Leyenda

- **Pendiente** — No iniciada
- **En Progreso** — En ejecución
- **Completada** — Finalizada y verificada
- **Bloqueada** — Requiere acción externa

---

*Documento generado por MVGN — v5.2*
