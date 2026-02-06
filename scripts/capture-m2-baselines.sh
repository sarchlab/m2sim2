#!/bin/bash
# M2 Baseline Capture Script
# Author: Eric (AI Researcher) - Cycle 281, Updated Cycle 290
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

# Get instruction count for benchmark (from simulation data)
get_instructions() {
    case "$1" in
        gemm) echo 37000 ;;
        atax) echo 5000 ;;
        twomm) echo 70000 ;;
        mvt) echo 5000 ;;
        jacobi1d) echo 5349 ;;
        threemm) echo 105410 ;;
        bicg) echo 4830 ;;
        *) echo 0 ;;
    esac
}

# Get directory name for benchmark (handles numeric prefixes)
get_dir_name() {
    case "$1" in
        twomm) echo "2mm" ;;
        jacobi1d) echo "jacobi-1d" ;;
        threemm) echo "3mm" ;;
        *) echo "$1" ;;
    esac
}

build_native() {
    local bench=$1
    local dir_name=$(get_dir_name "$bench")
    local src_dir="$REPO_ROOT/benchmarks/polybench/$dir_name"
    local src_file="$src_dir/$dir_name.c"
    local out_file="$OUTPUT_DIR/${bench}_native"
    
    if [[ ! -f "$src_file" ]]; then
        echo "ERROR: Source not found: $src_file"
        return 1
    fi
    
    echo "Building native $bench (from $dir_name)..."
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
        # Use Python for precise timing
        # Note: Benchmarks return checksum as exit code (non-zero is OK)
        local start=$(python3 -c "import time; print(time.perf_counter())")
        "$exe" > /dev/null 2>&1 || true  # Ignore exit code (benchmark returns checksum)
        local end=$(python3 -c "import time; print(time.perf_counter())")
        local elapsed=$(python3 -c "print($end - $start)")
        total_time=$(python3 -c "print($total_time + $elapsed)")
    done
    
    local avg_time=$(python3 -c "print($total_time / $ITERATIONS)")
    echo "  Average time: ${avg_time}s"
    
    # Estimate cycles (assuming 3.5 GHz for M2 P-core)
    local cycles=$(python3 -c "print(int($avg_time * 3.5e9))")
    local instr=$(get_instructions "$bench")
    local cpi=$(python3 -c "print(round($cycles / $instr, 3) if $instr > 0 else 0)")
    
    echo "  Estimated cycles: $cycles"
    echo "  Instructions: $instr"
    echo "  CPI: $cpi"
    
    # Append to results file (use directory name for reporting)
    local dir_name=$(get_dir_name "$bench")
    echo "$dir_name,$avg_time,$cycles,$instr,$cpi" >> "$OUTPUT_DIR/baselines.csv"
}

capture_benchmark() {
    local bench=$1
    local dir_name=$(get_dir_name "$bench")
    echo "=== Capturing M2 baseline for: $dir_name ==="
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
        
        # Use normalized names (no numeric prefixes/dashes for bash safety)
        for bench in gemm atax twomm mvt jacobi1d threemm bicg; do
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
