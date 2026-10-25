# mvgn-kernel — Comportamiento IA y Orquestación

> **Refactorizado desde:** `.mvgn/session-contract.md` + `.mvgn/kernel-spec.md`
> **Versión:** v5.1 — Blackboard Pattern
> **Fecha:** 2026-07-30

---

## 1. Activación del sistema MVGN

El sistema se considera activo **únicamente** cuando:

- [ ] `.mvgn/blackboard/state.json` existe y tiene estado válido
- [ ] `.mvgn/blackboard/context.md` está cargado como referencia
- [ ] El agente conoce el estado actual del proyecto
- [ ] Este documento (mvgn-kernel.md) está cargado en la sesión actual

Si alguna condición falla → el agente opera en **modo no-MVGN** (estado inválido).

---

## 2. Reglas de comportamiento obligatorio

### 2.1 Prioridad de fuentes

| Prioridad | Fuente | Qué define |
|-----------|--------|-----------|
| 1 (máxima) | **state.json** | Estado actual del proyecto y tarea activa |
| 2 | **JOIN.md** | Instrucciones para el agente actual |
| 3 | **context.md** | Contexto general del proyecto |
| 4 | **mvgn-kernel.md** | Comportamiento de la IA en esta sesión |

### 2.2 La IA debe

- Operar exclusivamente dentro de la tarea asignada
- Leer `state.json` antes de cada acción
- Actualizar `state.json` al completar su fase
- Si detecta una inconsistencia, reportarla y detenerse

### 2.3 La IA NO puede

- Modificar archivos fuera de su tarea asignada
- Ejecutar múltiples tareas simultáneas
- Modificar documentos aprobados sin nueva autorización
- Ignorar `state.json` como fuente de verdad

---

## 3. Flujo obligatorio de respuesta

```
PASO 1 — LEER STATE
  Leer .mvgn/blackboard/state.json → estado, tarea, agentes completados

PASO 2 — DETERMINAR FASE
  Según completed_agents y agent_sequence, determinar qué sigue

PASO 3 — EJECUTAR ACCIÓN
  Ejecutar dentro de la fase asignada

PASO 4 — ACTUALIZAR ESTADO
  Actualizar state.json con resultados + artifacts

PASO 5 — REPORTAR
  Informar resultado y estado actual
```

---

## 4. Progressive Loading (carga mínima)

El sistema debe operar con el **mínimo de documentos activos** para cada estado:

| Estado | Docs requeridos |
|--------|----------------|
| `pending` | `state.json`, `JOIN.md`, `context.md` |
| `in_progress` | anteriores + `mvgn-kernel.md` |
| `completed` | `state.json`, `04_changelog.md` |
| `failed` | anteriores + `mvgn-recovery.md` |

---

## 5. Control de integridad

### 5.1 Checks obligatorios

| # | Check | Qué valida | Falla si |
|---|-------|-----------|----------|
| K-01 | State accesible | `state.json` existe y es JSON válido | No existe / formato inválido |
| K-02 | Estado único | Un solo estado declarado | Múltiples estados o ninguno |
| K-03 | Secuencia válida | `agent_sequence` contiene al menos un agente | Vacío o nulo |
| K-04 | Sin bloqueos fantasma | `last_error` coincide con realidad | Error declarado pero no real |
| K-05 | Post-ejecución estable | Después de operación, sistema sigue coherente | Estado cambió sin transición válida |

### 5.2 Niveles de integridad

| Nivel | Significado | Acción |
|-------|-------------|--------|
| **OK** | Todos los checks K-01 a K-05 pasan | Operación normal |
| **DEGRADED** | Checks menores fallan (K-04/K-05) | Permitir con warnings |
| **CORRUPTED** | K-01 o K-02 fallan | Detener todo, ejecutar recovery |

---

## 6. Árbol de decisión

```
state.json leído
├── ¿status == "pending"?
│   └── Sí → ejecutar fase actual
├── ¿status == "in_progress"?
│   └── Sí → continuar desde donde se quedó
├── ¿status == "completed"?
│   └── Sí → reportar resumen, esperar nueva tarea
├── ¿status == "failed"?
│   └── Sí → reportar error, ofrecer resume o recovery
└── ¿status desconocido?
    └── CORRUPTED → ejecutar recovery
```

---

## 7. Formato de respuesta

Al final de cada interacción, responder con:

```
MVGN SESSION STATE
────────────────────
STATE:         <pending | in_progress | completed | failed>
MODE:          <NORMAL | COLLAB>
INTEGRITY:     <OK | DEGRADED | CORRUPTED>
────────────────────
LAST ACTION:   <qué se acaba de hacer>
NEXT ACTION:   <qué toca hacer ahora>
ERRORS:        <ninguno | descripción>
────────────────────
```

---

## 8. Guards de esquema (Schema v3.0 → v5.1)

### 8.1 Escritura atómica obligatoria

Todo cambio a `state.json` DEBE seguir:

```
1. Escribir cambios en state.json.tmp
2. Validar que el JSON es sintácticamente correcto
3. Renombrar atómico: .tmp → state.json
```

### 8.2 Validación de campos obligatorios

`state.json` debe contener siempre:
- `task_id`, `task_description`, `status`, `current_phase`
- `agent_sequence`, `completed_agents`, `artifacts_generated`
- `last_error`, `timestamp`

---

*Documento refactorizado para Blackboard v5.1 — 2026-07-30*
*Basado en session-contract.md v2.1 + kernel-spec.md v2.1*
