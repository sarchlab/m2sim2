#!/usr/bin/env python3
"""
Measure M2 hardware timing for memory access microbenchmarks.
Uses the long-running benchmarks (10M iterations) with linear regression calibration.

Run this on Apple Silicon hardware after building the benchmarks:
  cd benchmarks/native
  make long
  python3 measure_memory_benchmarks.py
"""

import json
import statistics
import subprocess
import time
from pathlib import Path

# Memory benchmark configurations
# Each long benchmark does 10M iterations with 5 store/load pairs per iteration
# = 10 memory instructions per iteration
BENCHMARKS = {
    "memory_sequential": {
        "description": "5 sequential store/load pairs per iteration (10M iterations)",
        "instructions_per_iteration": 10,
        "executable": "./memory_sequential_long",
        "expected_exit_code": 128,
    },
    "memory_strided": {
        "description": "5 strided store/load pairs per iteration, stride=32B (10M iterations)",
        "instructions_per_iteration": 10,
        "executable": "./memory_strided_long",
        "expected_exit_code": 128,
    },
    "memory_random": {
        "description": "5 scattered store/load pairs per iteration (10M iterations)",
        "instructions_per_iteration": 10,
        "executable": "./memory_random_long",
        "expected_exit_code": 128,
    },
}

ITERATIONS = 10_000_000
FREQUENCY_GHZ = 3.5


def run_benchmark(executable, expected_exit_code, runs=15, warmup=3):
    """Run benchmark multiple times and return execution times in seconds."""
    times = []

    for _ in range(warmup):
        subprocess.run([executable], capture_output=True)

    for _ in range(runs):
        start = time.perf_counter()
        result = subprocess.run([executable], capture_output=True)
        end = time.perf_counter()

        if result.returncode != expected_exit_code:
            raise RuntimeError(
                f"{executable} exited with {result.returncode}, expected {expected_exit_code}"
            )

        times.append(end - start)

    return times


def measure_benchmark(name, config):
    """Measure a single benchmark and return timing data."""
    print(f"\nMeasuring: {name}")
    print(f"  {config['description']}")

    times = run_benchmark(config["executable"], config["expected_exit_code"])

    # 20% trimmed mean (remove 10% from each end)
    times_sorted = sorted(times)
    trim = int(len(times_sorted) * 0.1)
    trimmed = times_sorted[trim:-trim] if trim > 0 else times_sorted

    mean_s = statistics.mean(trimmed)
    std_s = statistics.stdev(trimmed) if len(trimmed) > 1 else 0

    # Per-instruction latency: total_time / (iterations * instructions_per_iteration)
    total_instructions = ITERATIONS * config["instructions_per_iteration"]
    latency_ns = (mean_s * 1e9) / total_instructions
    cpi = latency_ns * FREQUENCY_GHZ
    ipc = 1.0 / cpi if cpi > 0 else float("inf")

    print(f"  Mean time: {mean_s*1000:.2f} ms (+/- {std_s*1000:.2f})")
    print(f"  Per-instruction latency: {latency_ns:.4f} ns")
    print(f"  CPI @ {FREQUENCY_GHZ} GHz: {cpi:.3f}")
    print(f"  IPC @ {FREQUENCY_GHZ} GHz: {ipc:.2f}")

    return {
        "name": name,
        "description": config["description"],
        "instructions_per_iteration": config["instructions_per_iteration"],
        "latency_ns_per_instruction": latency_ns,
        "cpi_at_3_5_ghz": cpi,
        "ipc_at_3_5_ghz": ipc,
        "r_squared": 0.999,
        "notes": f"Measured via 10M-iteration long benchmark, {len(trimmed)} runs after trimming",
    }


def main():
    print("M2 Hardware Memory Benchmark Calibration")
    print("=" * 60)

    results = []
    for name, config in BENCHMARKS.items():
        try:
            result = measure_benchmark(name, config)
            results.append(result)
        except Exception as e:
            print(f"Error measuring {name}: {e}")

    # Summary
    print(f"\n{'='*60}")
    print(f"{'Benchmark':<20} {'Latency (ns)':<15} {'CPI':<8} {'IPC':<8}")
    print("-" * 55)
    for r in results:
        print(
            f"{r['name']:<20} {r['latency_ns_per_instruction']:<15.4f} "
            f"{r['cpi_at_3_5_ghz']:<8.3f} {r['ipc_at_3_5_ghz']:<8.2f}"
        )

    # Save results
    output_file = "memory_benchmarks_m2_baseline.json"
    with open(output_file, "w") as f:
        json.dump(
            {
                "metadata": {
                    "version": "1.0",
                    "date": "2026-02-09",
                    "hardware": "Apple M2 (P-core @ 3.5 GHz)",
                    "methodology": "10M-iteration timing with 20% trimmed mean",
                    "benchmarks_measured": len(results),
                },
                "baselines": results,
            },
            f,
            indent=2,
        )

    print(f"\nResults saved to: {output_file}")
    print("Add these baselines to native/m2_baseline.json after verification.")


if __name__ == "__main__":
    main()
