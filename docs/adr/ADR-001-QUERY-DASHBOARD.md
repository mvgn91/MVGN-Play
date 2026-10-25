# ADR-001: Activación de QUERY Dashboard

> **Estado:** Aprobado
> **Versión:** 1.0
> **Clasificación:** Architecture Decision Record
> **Fecha:** 2026-07-30
> **Contexto:** ATHN1-AUD-20260730-OBS-CONSOLE-001 (v2)
> **Evaluación ATENEA:** VERDE CONDICIONADO

---

## Decisión

Activar **QUERY Dashboard** como capability de observación en vivo para proyectos MVGN.

La observación en vivo se implementa como binario independiente `mvgn-query.exe` que:
- Lee `.mvgn/blackboard/state.json`
- Renderiza semáforo de estado (⚪🔵🟡🟢🔴)
- Soporta modo `--watch` con polling configurable
- Aplica Scientific Glass UI (CNX-CANON-002)

---

## Alternativas consideradas

| Alternativa | Voto | Motivo del rechazo |
|:------------|:----:|:-------------------|
| Extender `mvgn-scan` con `--watch` | ❌ | Mezcla responsabilidades: scan ≠ observe (CNX-CANON-008) |
| Script PowerShell standalone | ❌ | No portable, no compilable a .exe, sin type safety |
| Dashboard web con servidor HTTP | ❌ | Fuera del alcance MVP. Mayor complejidad. |
| **✅ Binario `mvgn-query` independiente** | ✅ | Clara separación, reutiliza display de mvgn-scan, compila a .exe |

---

## Consecuencias

### Positivas
- Separación clara entre escaneo y observación
- Binario autocontenido (un .exe, cero dependencias externas)
- Reutiliza el sistema display ANSI de `mvgn-scan`
- Puede ejecutarse en cualquier proyecto MVGN clonado
- Fácil integración en pipelines CI/CD

### Negativas / Riesgos
- Dependencia del esquema de `state.json` (debe versionarse)
- El polling agrega I/O periódico (mitigado con intervalo configurable)
- Si `state.json` cambia de schema, hay que actualizar el lector

---

## Especificación técnica

### Comando
```bash
mvgn-query --watch [--interval 5s] [--project ./] [--no-color]
```

### Flags
| Flag | Default | Descripción |
|:-----|:--------|:------------|
| `--watch` | `false` | Modo observación continua |
| `--interval` | `5s` | Intervalo de polling |
| `--project` | `.` | Ruta del proyecto MVGN |
| `--no-color` | `false` | Deshabilitar colores ANSI |

### Pipeline de datos
```
state.json → FileStateSource → validación → ObservedState → display
```

### Criterios P0 cumplidos
| ID | Criterio | Estado |
|:---|:---------|:-------|
| P0-01 | Lector de state.json | ✅ |
| P0-02 | Validación mínima de esquema | ✅ |
| P0-03 | Modelo interno ObservedState | ✅ |
| P0-04 | Flag --watch | ✅ |
| P0-05 | Loop con contexto y cancelación | ✅ |
| P0-06 | Render del semáforo ⚪🔵🟡🟢🔴 | ✅ |
| P0-07 | Manejo de ausencia de blackboard | ✅ |
| P0-08 | Manejo de JSON corrupto | ✅ |
| P0-09 | Timestamp de última actualización | ✅ |
| P0-10 | Detección de estado stale | ✅ |

### Exclusiones explícitas del MVP
- ❌ Timeline avanzado (Fase 2)
- ❌ Health score (Fase 2)
- ❌ Dashboard HTML (Fase 4)
- ❌ Persistencia de timeline (Fase 4)
- ❌ Integración con telemetry (Fase 2)
- ❌ Modo servidor (Fase 4)

## Tareas post-MVP (obligatorias antes de declarar capability activa)

| # | Tarea | Responsable | Prioridad |
|:--|:------|:------------|:----------|
| PM-01 | Actualizar `_kronos/capabilities-registry.json`: `query.status → "active"`, `query.enabled → true` | Operador | P0.5 |
| PM-02 | Actualizar `_kronos/runtime-contract.json`: `query.status → "active"` | Operador | P0.5 |
| PM-03 | Verificar que `mvgn-query.exe` no interfiere con `mvgn-scan.exe` | Buffy | P1 |
| PM-04 | Agregar entrada en `docs/04_changelog.md` para la activación de QUERY | Buffy | P1 |

---

## Referencias

- `CNX-CANON-000` — Filosofía Crimson Nexus
- `CNX-CANON-001` — Identidad de Activos Cognitivos
- `CNX-CANON-002` — Scientific Glass UI
- `CNX-CANON-008` — Integración Activos Cognitivos + MVGN
- `ATHN1-AUD-20260730-OBS-CONSOLE-001` — Auditoría de viabilidad (v2)
- `_kronos/capabilities-registry.json` — QUERY: inactive
- `_kronos/runtime-contract.json` — QUERY: inactive

---

*Registrado por Buffy (agent-buffy) — MVGN Labs v5.2*
