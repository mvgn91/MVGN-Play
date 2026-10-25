# 07 — Release Notes

> **Proyecto:** MVGN Labs — Starter Kit
> **Versión:** v5.2.0
> **Generado:** 2026-07-30 por MVGN Finalization Protocol
> **Estado:** ✅ COMPLETED

---

## Resumen

MVGN Starter Kit v5.2.0 incluye el **Observation Console** (terminal ANSI + web dashboard en tiempo real), **ATENEA Governance Scoring** (9 condiciones de health), **Update Checker** integrado en NYX.ps1, y **Audio Ingestion Pipeline** con Faster-Whisper.

---

## Novedades v5.2.0

### Observation Console
- Terminal ANSI con semáforo en vivo: `.\mvgn-scan\mvgn-scan.exe -watch`
- Dashboard web con SSE: `.\mvgn-scan\mvgn-scan.exe -observe -port 8080`
- Modo unificado (terminal + web): `.\mvgn-scan\mvgn-scan.exe -watch -observe`
- Health score combinado: baseline status + artifacts + agents + ATENEA governance (20%)
- ATENEA score: 9 condiciones de governance (0-80) en tiempo real
- Telemetry timeline: últimos eventos desde `events/*.jsonl`

### ATENEA Governance
- Modelo `AteneaConditions` con 9 campos booleanos
- Evaluación automática en cada ciclo de observación
- Persistencia en `state.json` y `dashboard-data.json`
- Blend 20% en health score final

### MVGN Update Checker
- `.\NYX.ps1 -Mode update` — compara local vs remoto
- Check: runtime_version, último git tag, commit HEAD
- Ofrece git pull / git pull --rebase / cancel
- Auto-check al final del modo READY

### Instalación
- `.\tools\install-observer.ps1` — one-command deploy
- `.\tools\install-observer.ps1 -RegisterTask` — inicio automático con Windows

---

## Capacidades

| Capacidad | Estado |
|:----------|:-------|
| Blackboard Pattern | ✅ v5.2 |
| Multi-Agent Orchestration | ✅ v5.2 |
| Handoff entre Agentes | ✅ v5.2 |
| Dual Entry (NYX.md + NYX.ps1) | ✅ v5.2 |
| Audio Ingestion (Faster-Whisper) | ✅ v5.2 |
| Observation Console (Terminal) | ✅ v5.2 |
| Observation Console (Web) | ✅ v5.2 |
| ATENEA Governance Scoring | ✅ v5.2 |
| Update Checker | ✅ v5.2 |
| One-command Install | ✅ v5.2 |

---

## Comandos Rápidos

```powershell
.\NYX.ps1 -Mode detect                    # Detectar estado
.\NYX.ps1 -Mode ready                     # Ver estado actual
.\NYX.ps1 -Mode observer                  # Observer terminal
.\NYX.ps1 -Mode observer -Web             # Observer dashboard web
.\NYX.ps1 -Mode update                    # Verificar actualizaciones
.\NYX.ps1 -Mode update -Force             # Actualizar sin preguntar
.\tools\install-observer.ps1              # Instalar observer
.\tools\observe.bat                       # Atajo doble-click
```

---

## Dependencias

- PowerShell 5.1+ (scripts)
- Go 1.21+ (para build de mvgn-scan)
- Git (para update checker)
- Python + Faster-Whisper (opcional, solo audio)
- **Costo: $0**

---

## Notas de Migración

Si venís de v4.x: Ejecutar `.\NYX.ps1 -Mode upgrade` primero.

Si venís de v5.0: Solo `git pull` — el blackboard es retrocompatible.

---

## Sellado

| Componente | Estado |
|------------|--------|
| `docs/` | ✅ Documentación completa (11 docs + ADR) |
| `mvgn-scan/` | ✅ Código Go compila sin errores |
| `scripts/` | ✅ Orquestación activa |
| `tools/` | ✅ Herramientas de instalación y validación |
| `NYX.md` / `NYX.ps1` | ✅ Entry points sincronizados |
| **Proyecto sellado como:** | `mvgnlabs-starter-kit_v5.2.0_STARTER` |

---

*MVGN v5.2.0 — Crimson Nexus Corp. — Node MX-03*
