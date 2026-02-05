#!/bin/bash
# [Eric] Batch timing simulation script for Embench benchmarks
# Created: 2026-02-04

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
REPORT_DIR="$PROJECT_ROOT/reports"
RESULTS_FILE="$REPORT_DIR/embench-timing-$(date +%Y-%m-%d).json"

mkdir -p "$REPORT_DIR"

# List of Embench benchmarks to test
EMBENCH_BENCHMARKS=(
    "aha-mont64-m2sim"
    "crc32-m2sim"
    "matmult-int-m2sim"
    "primecount-m2sim"
    "edn-m2sim"
)

echo "=== M2Sim Embench Timing Simulation ===" 
echo "Date: $(date)"
echo "Output: $RESULTS_FILE"
echo ""

cd "$PROJECT_ROOT"

# Run timing tests for each benchmark
echo "{"
echo '  "benchmarks": ['

first=true
for bench in "${EMBENCH_BENCHMARKS[@]}"; do
    if [ -d "benchmarks/$bench" ]; then
        elf_file="benchmarks/$bench/${bench//-m2sim/}_m2sim.elf"
        if [ -f "$elf_file" ]; then
            echo "Running timing simulation: $bench"
            
            # Run timing test (captures output)
            result=$(go test -v -run "TestBenchmark.*${bench//-m2sim/}" ./benchmarks/ 2>&1 || true)
            
            if [ "$first" = true ]; then
                first=false
            else
                echo ","
            fi
            
            echo "    {"
            echo "      \"name\": \"$bench\","
            echo "      \"status\": \"attempted\","
            echo "      \"elf\": \"$elf_file\""
            echo -n "    }"
        else
            echo "Warning: $elf_file not found"
        fi
    else
        echo "Warning: benchmarks/$bench not found"
    fi
done

echo ""
echo "  ],"
echo "  \"generated\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\""
echo "}"

echo ""
echo "=== Timing simulation complete ==="
