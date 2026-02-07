#!/bin/bash
# Bootstrap script - reset agent system to clean state
# Usage: ./bootstrap.sh

set -e
cd "$(dirname "$0")"

echo "=== M2Sim Agent Bootstrap ==="

# 1. Stop any running processes
echo "Stopping running processes..."
pkill -f "node orchestrator.js" 2>/dev/null && echo "  Killed orchestrator" || echo "  No orchestrator running"
pkill -f "monitor/node_modules" 2>/dev/null && echo "  Killed monitor" || echo "  No monitor running"

# 2. Remove temporary files
echo "Removing temporary files..."
rm -f state.json && echo "  Removed state.json" || true
rm -f orchestrator.log && echo "  Removed orchestrator.log" || true
rm -f nohup.out && echo "  Removed nohup.out" || true
rm -f monitor/monitor.log && echo "  Removed monitor/monitor.log" || true
rm -f monitor/server.log && echo "  Removed monitor/server.log" || true
rm -rf .DS_Store && echo "  Removed .DS_Store" || true

# 3. Remove all workers
echo "Removing workers..."
if [ -d "workers" ]; then
  rm -rf workers/*
  echo "  Cleared workers/"
else
  mkdir -p workers
  echo "  Created empty workers/"
fi

# 4. Remove workspace
echo "Removing workspace..."
if [ -d "../workspace" ]; then
  rm -rf ../workspace
  echo "  Removed workspace/"
else
  echo "  No workspace/ to remove"
fi

# 5. Remove messages
echo "Removing messages..."
rm -rf messages/* 2>/dev/null && echo "  Cleared messages/" || mkdir -p messages && echo "  Created empty messages/"

echo ""
echo "=== Bootstrap complete ==="
echo "To start fresh: ./run.sh"
