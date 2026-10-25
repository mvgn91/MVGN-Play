# mvgn-finalization — Protocolo de Finalización

> **Refactorizado desde:** `.mvgn/finalization-protocol.md`
> **Versión:** v5.1 — Blackboard Pattern
> **Fecha:** 2026-07-30

---

## 0. Glosario

| Término | Definición |
|---------|-----------|
| **Finalización** | Proceso de cierre de un proyecto o versión |
| **Release** | Versión publicada |
| **Sellado** | Estado inmutable documentado con respaldo |
| **Nomenclatura** | Sistema de nombrado estandarizado |
| **SLUG** | Identificador URL-safe del proyecto |

---

## 1. Activación

Se activa cuando se cumplen **todas**:

| # | Condición | Verificación |
|---|-----------|-------------|
| F-01 | Todas las tareas completadas | `state.json` → `status == "completed"` |
| F-02 | Sin errores activos | `state.json` → `last_error == null` |
| F-03 | Humano autoriza | "Finalizar" o "Release" |

---

## 2. Fase 1 — VERIFY (Verificación pre-cierre)

### Checklist

- [ ] F-01: Todas las tareas completadas
- [ ] F-02: Sin errores activos
- [ ] F-03: Autorización humana recibida
- [ ] V-01: `04_changelog.md` actualizado
- [ ] V-02: `05_lessons_learned.md` sin incidentes abiertos
- [ ] V-03: `state.json` refleja estado real

### Si la verificación falla

| Falla | Acción |
|-------|--------|
| V-01 a V-02 | Registrar en lessons. Pausar finalización. |
| V-03 | Corregir state.json antes de continuar. |

---

## 3. Fase 2 — DOCUMENT (Documentación de cierre)

### Documentos a generar

| Documento | Contenido | Recomendado para |
|-----------|-----------|------------------|
| `README.md` | README del proyecto | **OBLIGATORIO** si se sube a GitHub |
| `08_release_notes.md` | Notas de release | Proyectos con release público |
| `08_technical_summary.md` | Resumen técnico | Equipos > 2 devs |

### Formato de documentos

```markdown
# NN — Título

> **Proyecto:** {NOMBRE_PROYECTO}
> **Versión:** {SEMVER}
> **Generado:** {YYYY-MM-DD} por MVGN Finalization Protocol

---
```

---

## 4. Fase 3 — NOMENCLATE (Nomenclatura)

### Sistema NOM

| Código | Elemento | Formato | Ejemplo |
|--------|----------|---------|---------|
| NOM-001 | Directorio raíz | `{SLUG}_v{SEMVER}_{TIER}` | `mi-proyecto_v1.0.0_PRODUCTION` |
| NOM-002 | Release tag | `v{SEMVER}-{YYYYMMDD}` | `v1.0.0-20260622` |
| NOM-003 | Docs de cierre | `{SLUG}_docs_v{SEMVER}.md` | `mi-proyecto_docs_v1.0.0.md` |
| NOM-004 | Rama principal | `main` | — |

### Reglas

| Regla | Descripción |
|-------|-------------|
| R-NOM-01 | SLUG del nombre oficial: minúsculas, espacios → guiones |
| R-NOM-02 | SEMVER: `MAJOR.MINOR.PATCH` |
| R-NOM-03 | TIER: `PRODUCTION`, `ARCHIVE`, `SANDBOX`, `LEGACY` |
| R-NOM-04 | SLUG no cambia una vez asignado |

---

## 5. Fase 4 — SEAL (Sellado)

| # | Acción | Descripción |
|---|--------|-------------|
| S-01 | Backup de docs | Copia de `docs/` a `{SLUG}_docs_v{SEMVER}.zip` |
| S-02 | Snapshot final | Registrar estado final en `06_state_report.md` |
| S-03 | Sugar commit | `git add -A && git commit -m "release: v{SEMVER}"` + tag |

### Snapshot de sellado

```markdown
## Snapshot de sellado — {YYYY-MM-DD}

| Componente | Estado |
|------------|--------|
| `docs/` | ✅ Documentación completa |
| `src/` | ✅ Código fuente |
| `scripts/` | ✅ Orquestación activa |
| `archive/` | ✅ Legacy preservado |
| **Proyecto sellado como:** | `{SLUG}_v{SEMVER}_{TIER}` |
```

---

## 6. Fase 5 — RELEASE (Transición a COMPLETED)

### Acciones finales

1. Actualizar `state.json`: `status = "completed"`
2. Registrar en `04_changelog.md` la finalización
3. Si hay repo git: commit + tag

### Mensaje de finalización

```
MVGN FINALIZATION COMPLETE

  Proyecto: {NOMBRE_PROYECTO}
  Versión:  {SEMVER}
  Estado:   ✅ COMPLETED

  Documentación generada:
    ✅ 08_release_notes.md
    ✅ README.md

  Sellado:   ✅ Backup completado
```

---

## 7. Reapertura

Si un proyecto completado necesita modificaciones:

1. Humano autoriza reapertura
2. Se crea nueva tarea en `state.json`
3. Se actualiza `04_changelog.md`: "Proyecto reabierto — {motivo}"
4. La versión SEMVER avanza según alcance

---

## 8. Casos borde

### Proyecto sin deploy

TIER = `ARCHIVE`. No se genera reporte de deploy.

### Múltiples releases

Ejecución ligera: solo verificar nuevas tareas, actualizar release notes, avanzar SEMVER.

### Finalización cancelada

1. Detener protocolo
2. Restaurar estado anterior en `state.json`
3. Registrar en `05_lessons_learned.md`

---

*Documento refactorizado para Blackboard v5.1 — 2026-07-30*
*Basado en finalization-protocol.md v1.0*
