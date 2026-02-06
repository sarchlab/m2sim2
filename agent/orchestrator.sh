#!/usr/bin/env bash
# M2Sim Multi-Agent Orchestrator
# Runs agents in sequence with 3-minute intervals between each call.
# Bypasses OpenClaw cron (which has a bug in 2026.2.3-1).

set -euo pipefail

REPO_DIR="$(cd "$(dirname "$0")/.." && pwd)"
SKILL_PATH="/Users/yifan/.openclaw/workspace/skills/multi-agent-dev"
TRACKER_ISSUE=45
INTERVAL_SECONDS=180  # 3 minutes

# Agent sequence: Alice → Eric → Bob → Cathy → Dana → Grace
AGENTS=("alice" "eric" "bob" "cathy" "dana" "grace")

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"
}

get_next_agent() {
    cd "$REPO_DIR"
    local next_label
    next_label=$(gh issue view "$TRACKER_ISSUE" --json labels -q '.labels[].name' | grep -E '^next:' | sed 's/next://' | head -1)
    echo "${next_label:-alice}"
}

is_agent_active() {
    cd "$REPO_DIR"
    gh issue view "$TRACKER_ISSUE" --json labels -q '.labels[].name' | grep -qE '^active:'
}

check_session_running() {
    local agent="$1"
    openclaw sessions list --active-minutes 15 2>/dev/null | grep -q "${agent}-m2sim" && return 0 || return 1
}

spawn_agent() {
    local agent="$1"
    log "Spawning agent: $agent"
    
    openclaw cron run 99a4d4c6-5cb5-4c47-b73b-973587c051cf --force
}

main() {
    log "M2Sim Orchestrator started"
    log "Interval: ${INTERVAL_SECONDS}s, Repo: $REPO_DIR"
    
    while true; do
        log "--- Checking cycle ---"
        
        # Pull latest
        cd "$REPO_DIR"
        git pull --rebase --quiet 2>/dev/null || true
        
        # Check if an agent is active
        if is_agent_active; then
            local active_agent
            active_agent=$(gh issue view "$TRACKER_ISSUE" --json labels -q '.labels[].name' | grep -E '^active:' | sed 's/active://' | head -1)
            
            # Verify session is actually running
            if check_session_running "$active_agent"; then
                log "Agent '$active_agent' is active, waiting..."
            else
                log "Stale active label for '$active_agent', clearing..."
                gh issue edit "$TRACKER_ISSUE" --remove-label "active:$active_agent" 2>/dev/null || true
                
                # Spawn next agent
                local next_agent
                next_agent=$(get_next_agent)
                spawn_agent "$next_agent"
            fi
        else
            # No active agent, spawn next one
            local next_agent
            next_agent=$(get_next_agent)
            spawn_agent "$next_agent"
        fi
        
        log "Sleeping ${INTERVAL_SECONDS}s..."
        sleep "$INTERVAL_SECONDS"
    done
}

# Handle signals for graceful shutdown
trap 'log "Shutting down..."; exit 0' SIGINT SIGTERM

main "$@"
