#!/bin/bash
# run-spec-native.sh - Run SPEC CPU 2017 benchmarks natively and record timing
# [Bob] Created for #138 - baseline timing collection
#
# Usage: ./scripts/run-spec-native.sh [output.csv]
#
# Runs SPEC benchmarks on the host machine and records:
# - Wall-clock execution time
# - User/system CPU time
#
# Currently supported: mcf_r, deepsjeng_r
# TODO: xz_r requires complex input setup

set -e

SPEC_ROOT="${SPEC_ROOT:-/Users/yifan/Documents/spec}"
M2SIM_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUTPUT_CSV="${1:-$M2SIM_ROOT/results/spec-native-timing.csv}"

# Create results directory
mkdir -p "$(dirname "$OUTPUT_CSV")"

# Benchmark list (excluding xz_r which needs complex input setup)
BENCHMARK_LIST="505.mcf_r 531.deepsjeng_r"

# Check SPEC installation
if [ ! -d "$SPEC_ROOT/benchspec/CPU" ]; then
    echo "ERROR: SPEC not found at $SPEC_ROOT"
    echo "Set SPEC_ROOT environment variable"
    exit 1
fi

echo "SPEC CPU 2017 Native Timing Collection"
echo "======================================="
echo "Host: $(hostname)"
echo "Date: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
echo "Output: $OUTPUT_CSV"
echo ""

# Write CSV header
echo "benchmark,exe,dataset,wall_time_sec,user_time_sec,sys_time_sec,exit_code,timestamp" > "$OUTPUT_CSV"

# Get executable and input for a benchmark
get_exe_name() {
    local bench="$1"
    case "$bench" in
        505.mcf_r) echo "mcf_r_base.arm64" ;;
        531.deepsjeng_r) echo "deepsjeng_r_base.arm64" ;;
        557.xz_r) echo "xz_r_base.arm64" ;;
        *) echo "" ;;
    esac
}

get_input_args() {
    local bench="$1"
    case "$bench" in
        505.mcf_r) echo "inp.in" ;;
        531.deepsjeng_r) echo "test.txt" ;;
        557.xz_r) echo "" ;;
        *) echo "" ;;
    esac
}

# Function to run a benchmark
run_benchmark() {
    local bench="$1"
    local exe_name=$(get_exe_name "$bench")
    local input=$(get_input_args "$bench")
    
    local bench_dir="$SPEC_ROOT/benchspec/CPU/$bench"
    local exe_path="$bench_dir/exe/$exe_name"
    local data_dir="$bench_dir/data/test/input"
    
    if [ ! -f "$exe_path" ]; then
        echo "  SKIP: $exe_path not found"
        return 1
    fi
    
    echo "Running $bench..."
    
    # Create temp run directory
    local run_dir=$(mktemp -d)
    cp "$exe_path" "$run_dir/"
    
    # Copy input files
    if [ -d "$data_dir" ]; then
        cp "$data_dir"/* "$run_dir/" 2>/dev/null || true
    fi
    
    cd "$run_dir"
    
    local timestamp=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    local timing_file=$(mktemp)
    
    # Run benchmark with time
    local exit_code=0
    if [ -n "$input" ]; then
        /usr/bin/time "./$exe_name" "$input" > /dev/null 2> "$timing_file" || exit_code=$?
    else
        /usr/bin/time "./$exe_name" > /dev/null 2> "$timing_file" || exit_code=$?
    fi
    
    # Parse macOS time output: "X.XX real Y.YY user Z.ZZ sys"
    local real_time=$(awk '{print $1}' "$timing_file" 2>/dev/null || echo "0")
    local user_time=$(awk '{print $3}' "$timing_file" 2>/dev/null || echo "0")
    local sys_time=$(awk '{print $5}' "$timing_file" 2>/dev/null || echo "0")
    
    echo "  Wall: ${real_time}s, User: ${user_time}s, Sys: ${sys_time}s, Exit: $exit_code"
    
    # Write to CSV
    echo "$bench,$exe_name,test,$real_time,$user_time,$sys_time,$exit_code,$timestamp" >> "$OUTPUT_CSV"
    
    # Cleanup
    rm -rf "$run_dir" "$timing_file"
    cd - > /dev/null
    
    return 0
}

# Run all benchmarks
echo ""
for bench in $BENCHMARK_LIST; do
    run_benchmark "$bench" || true
    echo ""
done

echo "======================================="
echo "Results written to: $OUTPUT_CSV"
echo ""
cat "$OUTPUT_CSV"
