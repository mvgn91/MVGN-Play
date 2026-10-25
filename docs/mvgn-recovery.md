# mvgn-recovery — Protocolo de Recuperación

> **Refactorizado desde:** `.mvgn/recovery-protocol.md`
> **Versión:** v5.1 — Blackboard Pattern
> **Fecha:** 2026-07-30

---

## 0. Cuándo se activa

El recovery protocol se activa **obligatoriamente** cuando:

| Disparador | Origen | Descripción |
|------------|--------|-------------|
| `state.json` corrupto | Kernel | K-01 o K-02 fallan (archivo no existe, JSON inválido, estado indeterminado) |
| Inconsistencia documental | Agente | Estado declarado no coincide con evidencia |
| Humano lo solicita | Humano | "Ejecuta recovery protocol" |
| Error en cascada | Agente | Múltiples fallos simultáneos que impiden operación normal |

---

## 1. Fase 1 — FREEZE (Congelación)

### Acciones

1. Declarar estado `failed` en `state.json` con `last_error = "Recovery en curso"`
2. Congelar toda modificación de archivos (read-only)
3. Generar snapshot del estado actual

### Snapshot

```json
{
  "snapshot_pre_recovery": {
    "timestamp": "YYYY-MM-DDTHH:mm:ssZ",
    "state_declarado": "...",
    "evidence": {
      "state.json": "✅ / ⚠️ / ❌",
      "context.md": "✅ / ⚠️ / ❌",
      "scripts/": "✅ / ⚠️ / ❌"
    },
    "inconsistencias": ["lista de contradicciones"]
  }
}
```

---

## 2. Fase 2 — DIAGNOSE (Diagnóstico)

### Evaluación de documentos

| Criterio | Evaluación |
|----------|------------|
| ¿El archivo existe? | Sí / No |
| ¿Tiene contenido válido? | Sí / No |
| ¿Es consistente con state.json? | Sí / No |

### Determinación del último estado consistente

1. Si `state.json` es válido → estado mínimo `pending`
2. Si hay `completed_agents` → estado mínimo `in_progress`
3. Si `artifacts_generated` tiene entradas → estado `completed`
4. Si `last_error` no es null → estado `failed`

### Reporte de diagnóstico

```
MVGN RECOVERY DIAGNOSIS

Archivos válidos:
  state.json:      ✅ / ⚠️ / ❌
  context.md:      ✅ / ⚠️ / ❌
  JOIN.md:         ✅ / ⚠️ / ❌

Último estado consistente: <estado>
Propuesta: reset a <estado>
Conservando: state.json, context.md, JOIN.md
Reconstruir: <lista>
```

---

## 3. Fase 3 — REBUILD (Reconstrucción)

### 3.1 Autorización

Esperar instrucciones del humano antes de ejecutar cambios.

### 3.2 Reglas de conservación

| Archivo | ¿Se conserva? | Condición |
|---------|--------------|-----------|
| `state.json` | Siempre ✅ | Se regenera si está corrupto |
| `context.md` | Siempre ✅ | Se conserva o regenera |
| `JOIN.md` | Siempre ✅ | Se conserva o regenera |
| `scripts/` | Siempre ✅ | No se modifica |
| `archive/` | Siempre ✅ | No se modifica |

### 3.3 Procedimiento

**state.json** (si corrupto):
1. Si existe backup → restaurar
2. Si no → regenerar con estado `pending`
3. Marcar como `"recovered": true`

**context.md** (si corrupto):
1. Regenerar desde state.json + plantilla estándar
2. Marcar como recuperado

---

## 4. Fase 4 — VALIDATE (Revalidación)

### Checklist post-recovery

- [ ] `state.json` tiene todos los campos obligatorios
- [ ] `state.json.tmp` no existe (limpiar si quedó)
- [ ] `context.md` es legible
- [ ] Todos los scripts activos existen
- [ ] No hay referencias rotas

---

## 5. Fase 5 — RESUME (Reanudación)

### 5.1 Actualización final

1. `state.json` refleja estado post-recovery
2. El snapshot pre-recovery se conserva en `05_lessons_learned.md`
3. Se elimina bloqueo de recovery

### 5.2 Registro en lecciones aprendidas

```markdown
## Recovery: <fecha>

**Disparador:** state.json corrupto / inconsistencia / humano
**Causa raíz:** <qué causó el problema>
**Archivos afectados:** <lista>
**Estado post-recovery:** <estado>
**Prevención:** <cómo evitar que ocurra de nuevo>
```

### 5.3 Reanudación

- Estado = último estado consistente
- Modo = NORMAL
- Agente reporta resultados al humano

---

## 6. Casos borde

### Recovery parcial

Si solo `state.json` está corrupto → reconstruir solo ese archivo. No es necesario recovery completo.

### Corrupción irrecuperable

Si `state.json`, `context.md` y `scripts/` están todos irrecuperables:
1. Sistema no puede determinar estado
2. Declarar estado `INIT`
3. Conservar solo `archive/` como historial
4. Humano redefine desde cero

### Falso positivo

Si se activa recovery pero no hay inconsistencia real:
1. Cancelar recovery
2. Restaurar estado anterior
3. Registrar en lecciones como falso positivo

---

*Documento refactorizado para Blackboard v5.1 — 2026-07-30*
*Basado en recovery-protocol.md v2.1*
