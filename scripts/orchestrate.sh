#!/bin/bash

# MVGN Orchestrate - Blackboard Pattern Multi-Agent Orchestrator
# Ejecuta secuencialmente: OpenCode → FreeBuff → MiMoCode
# Costo: $0 | Dependencias: bash, jq, git

set -e

# ─── CONFIG ──────────────────────────────────────────────────
BLACKBOARD_DIR=".mvgn/blackboard"
STATE_FILE="$BLACKBOARD_DIR/state.json"
CONTEXT_FILE="$BLACKBOARD_DIR/context.md"
LOG_FILE="$BLACKBOARD_DIR/logs/execution.log"
SLEEP_BETWEEN_AGENTS=2

# ─── COLORES ─────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

# ─── FUNCIONES ───────────────────────────────────────────────

log() {
    local msg="[$(date '+%Y-%m-%d %H:%M:%S')] $1"
    echo -e "${CYAN}$msg${NC}"
    echo "$msg" >> "$LOG_FILE"
}

error() {
    local msg="[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: $1"
    echo -e "${RED}$msg${NC}"
    echo "$msg" >> "$LOG_FILE"
}

success() {
    local msg="[$(date '+%Y-%m-%d %H:%M:%S')] SUCCESS: $1"
    echo -e "${GREEN}$msg${NC}"
    echo "$msg" >> "$LOG_FILE"
}

warn() {
    local msg="[$(date '+%Y-%m-%d %H:%M:%S')] WARN: $1"
    echo -e "${YELLOW}$msg${NC}"
    echo "$msg" >> "$LOG_FILE"
}

# Actualizar state.json de forma atómica
update_state() {
    local tmp_file="$STATE_FILE.tmp"
    jq "$1" "$STATE_FILE" > "$tmp_file" && mv "$tmp_file" "$STATE_FILE"
}

# Obtener valor del state.json
get_state() {
    jq -r "$1" "$STATE_FILE"
}

# ─── VALIDACIONES ────────────────────────────────────────────

validate_deps() {
    log "Validando dependencias..."
    
    if ! command -v jq &> /dev/null; then
        error "jq no encontrado. Instala con: apt install jq"
        exit 1
    fi
    
    if ! command -v git &> /dev/null; then
        error "git no encontrado."
        exit 1
    fi
    
    success "Dependencias OK"
}

validate_state() {
    if [ ! -f "$STATE_FILE" ]; then
        error "state.json no encontrado en $BLACKBOARD_DIR"
        exit 1
    fi
    
    local status=$(get_state '.status')
    if [ "$status" = "completed" ]; then
        warn "Tarea ya completada"
        exit 0
    fi
    
    if [ "$status" = "failed" ]; then
        error "Tarea en estado failed. Último error: $(get_state '.last_error')"
        exit 1
    fi
}

# ─── EJECUCIÓN DE AGENTES ───────────────────────────────────

ADAPTERS_DIR="$SCRIPT_DIR/agents"

run_opencode() {
    log "Ejecutando OpenCode (Fase 1: Generación)..."
    
    update_state '.status = "in_progress" | .current_phase = 1'
    
    # Ejecutar adaptador de OpenCode
    if bash "$ADAPTERS_DIR/opencode-adapter.sh"; then
        update_state '.completed_agents += ["opencode"] | .status = "completed_phase_1" | .artifacts_generated += ["src/"]'
        success "OpenCode completado"
    else
        update_state '.status = "failed" | .last_error = "OpenCode falló"'
        error "OpenCode falló"
        exit 1
    fi
}

run_freebuff() {
    log "Ejecutando FreeBuff (Fase 2: Revisión)..."
    
    update_state '.current_phase = 2'
    
    # Ejecutar adaptador de FreeBuff
    if bash "$ADAPTERS_DIR/freebuff-adapter.sh"; then
        update_state '.completed_agents += ["freebuff"] | .status = "completed_phase_2"'
        success "FreeBuff completado"
    else
        update_state '.status = "failed" | .last_error = "FreeBuff falló"'
        error "FreeBuff falló"
        exit 1
    fi
}

run_mimocode() {
    log "Ejecutando MiMoCode (Fase 3: Tests y Docs)..."
    
    update_state '.current_phase = 3'
    
    # Ejecutar adaptador de MiMoCode
    if bash "$ADAPTERS_DIR/mimocode-adapter.sh"; then
        update_state '.completed_agents += ["mimocode"] | .status = "completed"'
        success "MiMoCode completado"
    else
        update_state '.status = "failed" | .last_error = "MiMoCode falló"'
        error "MiMoCode falló"
        exit 1
    fi
}

# ─── FLUJO PRINCIPAL ────────────────────────────────────────

main() {
    echo -e "${CYAN}╔══════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║     MVGN Orchestrator - Blackboard      ║${NC}"
    echo -e "${CYAN}╚══════════════════════════════════════════╝${NC}"
    echo ""
    
    # Validar
    validate_deps
    validate_state
    
    local task_desc=$(get_state '.task_description')
    log "Tarea: $task_desc"
    
    # Ejecutar secuencia
    run_opencode
    sleep "$SLEEP_BETWEEN_AGENTS"
    
    run_freebuff
    sleep "$SLEEP_BETWEEN_AGENTS"
    
    run_mimocode
    
    # Commit automático
    log "Realizando commit automático..."
    git add -A
    git commit -m "[MVGN-ORCH] Feature completada: $task_desc"
    
    echo ""
    success "═══════════════════════════════════════════"
    success "Pipeline completado exitosamente"
    success "═══════════════════════════════════════════"
}

# ─── PUNTO DE ENTRADA ───────────────────────────────────────

if [ "$1" = "--task" ] && [ -n "$2" ]; then
    # Inicializar nueva tarea
    mkdir -p "$BLACKBOARD_DIR/logs"
    
    cat > "$STATE_FILE" <<EOF
{
  "task_id": "$(uuidgen 2>/dev/null || echo "task-$(date +%s)")",
  "task_description": "$2",
  "status": "pending",
  "current_phase": 0,
  "agent_sequence": ["opencode", "freebuff", "mimocode"],
  "completed_agents": [],
  "artifacts_generated": [],
  "last_error": null,
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
    
    log "Tarea inicializada: $2"
    main
elif [ "$1" = "--resume" ]; then
    # Reanudar tarea existente
    main
else
    echo "Uso:"
    echo "  $0 --task \"Descripción de la tarea\""
    echo "  $0 --resume"
    exit 1
fi