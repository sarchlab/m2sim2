#!/bin/bash
# Wrapper script for orchestrator with hot reload support
# Exit code 75 = reload requested, restart immediately
# Any other non-zero = error, wait before retry

cd "$(dirname "$0")"

while true; do
  echo "[wrapper] Starting orchestrator..."
  node orchestrator.js
  code=$?
  
  if [ $code -eq 0 ]; then
    echo "[wrapper] Orchestrator exited cleanly"
    exit 0
  elif [ $code -eq 75 ]; then
    echo "[wrapper] Hot reload requested, restarting..."
    sleep 1
  else
    echo "[wrapper] Orchestrator crashed (code $code), restarting in 30s..."
    sleep 30
  fi
done
