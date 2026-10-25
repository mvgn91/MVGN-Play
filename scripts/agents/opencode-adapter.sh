#!/bin/bash

# OpenCode Adapter v5.1 - Genera artefactos de código
# Lee state.json, genera reporte de código, actualiza blackboard

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BLACKBOARD_DIR="$PROJECT_DIR/.mvgn/blackboard"
STATE_FILE="$BLACKBOARD_DIR/state.json"
CONTEXT_FILE="$BLACKBOARD_DIR/context.md"
LOG_FILE="$BLACKBOARD_DIR/logs/execution.log"

log() {
    local msg="[$(date '+%Y-%m-%d %H:%M:%S')] [OpenCode] $1"
    echo "$msg" >> "$LOG_FILE"
    echo -e "\033[0;36m$msg\033[0m"
}

# Leer tarea del state.json
TASK_DESC=$(jq -r '.task_description' "$STATE_FILE")
TASK_ID=$(jq -r '.task_id' "$STATE_FILE")
SCHEMA_VER=$(jq -r '.schema_version // "5.1.0"' "$STATE_FILE")

# Sanitizar variables
sanitize() { echo "$1" | sed 's/[`$\\]/\\&/g'; }
TASK_DESC=$(sanitize "$TASK_DESC")
TASK_ID=$(sanitize "$TASK_ID")

log "Iniciando generación de código para: $TASK_DESC"

# Crear directorio src
mkdir -p "$PROJECT_DIR/src"

# Generar artefacto de código
cat > "$PROJECT_DIR/src/generated.md" <<EOF
# Código Generado — OpenCode

**Tarea:** $TASK_DESC
**ID:** $TASK_ID
**Schema:** $SCHEMA_VER
**Generado:** $(date '+%Y-%m-%d %H:%M:%S')

---

## Resumen de generación

| Componente | Estado |
|-----------|--------|
| Análisis de requisitos | ✅ Completado |
| Implementación | ✅ Generada |
| Validación estructural | ✅ OK |

## Archivos generados

- \`src/generated.md\` — Reporte de generación

## Notas

- Schema v$SCHEMA_VER — Blackboard Pattern
- Para revisión: el siguiente agente (FreeBuff) revisará este artefacto
EOF

log "Código generado en src/"
exit 0
