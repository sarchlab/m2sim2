#!/bin/bash
# M2 Baseline Capture Script
# Author: Eric (AI Researcher) - Cycle 281
# Purpose: Automate native builds and timing capture for M2 baseline validation
#
# Usage: ./scripts/capture-m2-baselines.sh [benchmark|all]
# Example: ./scripts/capture-m2-baselines.sh gemm
#          ./scripts/capture-m2-baselines.sh all

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
OUTPUT_DIR="$REPO_ROOT/reports/m2-baselines"
ITERATIONS=10  # Run each benchmark 10 times for stable timing

mkdir -p "$OUTPUT_DIR"

# PolyBench benchmarks and their expected instruction counts (from simulation)
declare -A POLYBENCH_INSTR=(
    ["gemm"]=37000
    ["atax"]=5000
    ["2mm"]=70000
    ["mvt"]=5000
    ["jacobi-1d"]=5349
    ["3mm"]=105410
    ["bicg"]=4830
)

build_native() {
    local bench=$1
    local src_dir="$REPO_ROOT/benchmarks/polybench/$bench"
    local src_file="$src_dir/$bench.c"
    local out_file="$OUTPUT_DIR/${bench}_native"
    
    if [[ ! -f "$src_file" ]]; then
        echo "ERROR: Source not found: $src_file"
        return 1
    fi
    
    echo "Building native $bench..."
    clang -O2 -DPOLYBENCH_TIME \
        -I"$REPO_ROOT/benchmarks/polybench" \
        "$src_file" \
        -o "$out_file" 2>/dev/null || {
        # Fallback: build without timing macros
        clang -O2 \
            -I"$REPO_ROOT/benchmarks/polybench" \
            "$src_file" \
            -o "$out_file"
    }
    echo "Built: $out_file"
}

measure_timing() {
    local bench=$1
    local exe="$OUTPUT_DIR/${bench}_native"
    
    if [[ ! -x "$exe" ]]; then
        echo "ERROR: Executable not found: $exe"
        return 1
    fi
    
    echo "Measuring $bench ($ITERATIONS iterations)..."
    
    local total_time=0
    for i in $(seq 1 $ITERATIONS); do
        # Use time command, capture real time
        local start=$(python3 -c "import time; print(time.perf_counter())")
        "$exe" > /dev/null 2>&1
        local end=$(python3 -c "import time; print(time.perf_counter())")
        local elapsed=$(python3 -c "print($end - $start)")
        total_time=$(python3 -c "print($total_time + $elapsed)")
    done
    
    local avg_time=$(python3 -c "print($total_time / $ITERATIONS)")
    echo "  Average time: ${avg_time}s"
    
    # Estimate cycles (assuming 3.5 GHz for M2 P-core)
    local cycles=$(python3 -c "print(int($avg_time * 3.5e9))")
    local instr=${POLYBENCH_INSTR[$bench]:-0}
    local cpi=$(python3 -c "print(round($cycles / $instr, 3) if $instr > 0 else 0)")
    
    echo "  Estimated cycles: $cycles"
    echo "  Instructions: $instr"
    echo "  CPI: $cpi"
    
    # Append to results file
    echo "$bench,$avg_time,$cycles,$instr,$cpi" >> "$OUTPUT_DIR/baselines.csv"
}

capture_benchmark() {
    local bench=$1
    echo "=== Capturing M2 baseline for: $bench ==="
    build_native "$bench"
    measure_timing "$bench"
    echo ""
}

main() {
    local target="${1:-all}"
    
    # Initialize CSV
    echo "benchmark,avg_time_sec,est_cycles,instructions,cpi" > "$OUTPUT_DIR/baselines.csv"
    
    if [[ "$target" == "all" ]]; then
        echo "Capturing M2 baselines for all PolyBench benchmarks..."
        echo "Output: $OUTPUT_DIR"
        echo ""
        
        for bench in gemm atax 2mm mvt jacobi-1d 3mm bicg; do
            capture_benchmark "$bench"
        done
    else
        capture_benchmark "$target"
    fi
    
    echo "=== Results ==="
    cat "$OUTPUT_DIR/baselines.csv"
    echo ""
    echo "Results saved to: $OUTPUT_DIR/baselines.csv"
    echo ""
    echo "NOTE: Cycle counts are estimated from wall time @ 3.5 GHz."
    echo "For precise cycles, use 'xctrace' or Apple performance counters."
}

main "$@"
