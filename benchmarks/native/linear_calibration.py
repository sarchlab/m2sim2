#!/usr/bin/env python3
"""
linear_calibration.py - Linear Regression Calibration Tool

Uses the methodology suggested by Human in Issue #88:
Run benchmarks with varying instruction counts and use linear regression
to separate process startup overhead from actual instruction latency.

Formula: time = instruction_latency * instruction_count + constant_overhead

This script:
1. Generates assembly benchmarks with varying iteration counts
2. Builds and runs them, collecting timing data
3. Fits a linear regression model
4. Reports the instruction latency and overhead
"""

import os
import subprocess
import sys
import tempfile
import time
from dataclasses import dataclass
from pathlib import Path
from typing import List, Tuple

# Check for numpy/scipy availability
try:
    import numpy as np
    from scipy import stats
    HAS_SCIPY = True
except ImportError:
    HAS_SCIPY = False
    print("Note: scipy not available, using simple linear regression fallback")

# Benchmark templates - each generates N iterations of target instructions
BENCHMARK_TEMPLATES = {
    "arithmetic": {
        "description": "20 independent ADDs per iteration (ALU throughput)",
        "instructions_per_iter": 20,
        "template": """
// arithmetic_calibration.s - Generated for {iterations} iterations
.global _main
.align 4

_main:
    mov x10, #0              // iteration counter
{load_iterations}

    mov x0, #0
    mov x1, #0
    mov x2, #0
    mov x3, #0
    mov x4, #0

.loop:
    add x0, x0, #1
    add x1, x1, #1
    add x2, x2, #1
    add x3, x3, #1
    add x4, x4, #1

    add x0, x0, #1
    add x1, x1, #1
    add x2, x2, #1
    add x3, x3, #1
    add x4, x4, #1

    add x0, x0, #1
    add x1, x1, #1
    add x2, x2, #1
    add x3, x3, #1
    add x4, x4, #1

    add x0, x0, #1
    add x1, x1, #1
    add x2, x2, #1
    add x3, x3, #1
    add x4, x4, #1

    add x10, x10, #1
    cmp x10, x11
    b.lt .loop

    mov x0, #0
    mov x16, #1
    svc #0x80
"""
    },
    "dependency": {
        "description": "20 dependent ADDs per iteration (RAW hazards)",
        "instructions_per_iter": 20,
        "template": """
// dependency_calibration.s - Generated for {iterations} iterations
.global _main
.align 4

_main:
    mov x10, #0              // iteration counter
{load_iterations}

    mov x0, #0

.loop:
    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1

    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1

    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1

    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1

    add x10, x10, #1
    cmp x10, x11
    b.lt .loop

    mov x0, #0
    mov x16, #1
    svc #0x80
"""
    },
    "branch": {
        "description": "5 taken branches per iteration (branch predictor)",
        "instructions_per_iter": 5,
        "template": """
// branch_calibration.s - Generated for {iterations} iterations
.global _main
.align 4

_main:
    mov x10, #0              // iteration counter
{load_iterations}

.loop:
    b .b1
.b1:
    b .b2
.b2:
    b .b3
.b3:
    b .b4
.b4:
    b .b5
.b5:

    add x10, x10, #1
    cmp x10, x11
    b.lt .loop

    mov x0, #0
    mov x16, #1
    svc #0x80
"""
    },
}


def load_iterations_asm(n: int) -> str:
    """Generate ARM64 assembly to load iteration count into x11."""
    if n <= 0xFFFF:
        return f"    movz x11, #{n}"
    elif n <= 0xFFFFFFFF:
        low = n & 0xFFFF
        high = (n >> 16) & 0xFFFF
        lines = [f"    movz x11, #{low}"]
        if high > 0:
            lines.append(f"    movk x11, #{high}, lsl #16")
        return "\n".join(lines)
    else:
        raise ValueError(f"Iteration count {n} too large (max 2^32-1)")


def generate_benchmark(template_name: str, iterations: int) -> str:
    """Generate assembly source for a benchmark with given iteration count."""
    tmpl = BENCHMARK_TEMPLATES[template_name]
    return tmpl["template"].format(
        iterations=iterations,
        load_iterations=load_iterations_asm(iterations)
    )


def build_and_run(asm_source: str, runs: int = 5, warmup: int = 2) -> List[float]:
    """Build assembly source and run multiple times, returning execution times in seconds.
    
    Includes warmup runs (discarded) to warm up caches and reduce variance.
    """
    times = []
    
    with tempfile.TemporaryDirectory() as tmpdir:
        asm_path = Path(tmpdir) / "benchmark.s"
        obj_path = Path(tmpdir) / "benchmark.o"
        exe_path = Path(tmpdir) / "benchmark"
        
        # Write assembly
        asm_path.write_text(asm_source)
        
        # Assemble
        result = subprocess.run(
            ["as", "-o", str(obj_path), str(asm_path)],
            capture_output=True, text=True
        )
        if result.returncode != 0:
            raise RuntimeError(f"Assembly failed: {result.stderr}")
        
        # Link
        result = subprocess.run(
            ["ld", "-o", str(exe_path), str(obj_path),
             "-lSystem", "-syslibroot", "/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk",
             "-e", "_main", "-arch", "arm64"],
            capture_output=True, text=True
        )
        if result.returncode != 0:
            raise RuntimeError(f"Link failed: {result.stderr}")
        
        # Warmup runs (discarded)
        for _ in range(warmup):
            subprocess.run([str(exe_path)], capture_output=True)
        
        # Timed runs
        for _ in range(runs):
            start = time.perf_counter()
            subprocess.run([str(exe_path)], capture_output=True)
            end = time.perf_counter()
            times.append(end - start)
    
    return times


@dataclass
class CalibrationResult:
    """Result of linear regression calibration."""
    benchmark: str
    description: str
    instruction_latency_ns: float  # nanoseconds per instruction
    overhead_ms: float             # process startup overhead in milliseconds
    r_squared: float               # goodness of fit
    data_points: List[Tuple[int, float]]  # (instruction_count, avg_time_ms)


def simple_linear_regression(x: List[float], y: List[float]) -> Tuple[float, float, float]:
    """Simple linear regression without scipy. Returns (slope, intercept, r_squared)."""
    n = len(x)
    sum_x = sum(x)
    sum_y = sum(y)
    sum_xy = sum(xi * yi for xi, yi in zip(x, y))
    sum_x2 = sum(xi * xi for xi in x)
    sum_y2 = sum(yi * yi for yi in y)
    
    slope = (n * sum_xy - sum_x * sum_y) / (n * sum_x2 - sum_x * sum_x)
    intercept = (sum_y - slope * sum_x) / n
    
    # R-squared
    ss_tot = sum((yi - sum_y/n)**2 for yi in y)
    ss_res = sum((yi - (slope * xi + intercept))**2 for xi, yi in zip(x, y))
    r_squared = 1 - (ss_res / ss_tot) if ss_tot > 0 else 0
    
    return slope, intercept, r_squared


def trimmed_mean(values: List[float], trim_percent: float = 0.2) -> float:
    """Calculate trimmed mean, removing top and bottom trim_percent of values."""
    if len(values) < 3:
        return sum(values) / len(values)
    
    sorted_vals = sorted(values)
    n = len(sorted_vals)
    trim_count = int(n * trim_percent)
    
    if trim_count == 0:
        return sum(sorted_vals) / n
    
    trimmed = sorted_vals[trim_count:-trim_count]
    return sum(trimmed) / len(trimmed) if trimmed else sum(sorted_vals) / n


def calibrate_benchmark(template_name: str, iteration_counts: List[int], 
                        runs_per_count: int = 15, verbose: bool = True) -> CalibrationResult:
    """Run calibration for a single benchmark type.
    
    Uses warmup runs and trimmed mean to reduce variance.
    """
    tmpl = BENCHMARK_TEMPLATES[template_name]
    instr_per_iter = tmpl["instructions_per_iter"]
    
    if verbose:
        print(f"\n{'='*60}")
        print(f"Calibrating: {template_name}")
        print(f"Description: {tmpl['description']}")
        print(f"{'='*60}")
    
    data_points = []
    instruction_counts = []
    times_ms = []
    
    for iterations in iteration_counts:
        total_instructions = iterations * instr_per_iter
        if verbose:
            print(f"  {iterations:>10,} iterations ({total_instructions:>12,} instructions)... ", end="", flush=True)
        
        asm_source = generate_benchmark(template_name, iterations)
        run_times = build_and_run(asm_source, runs=runs_per_count, warmup=3)
        
        # Use trimmed mean to reduce impact of outliers
        run_times_ms = [t * 1000 for t in run_times]
        avg_time_ms = trimmed_mean(run_times_ms, 0.2)
        
        # Calculate std of trimmed values for reporting
        sorted_times = sorted(run_times_ms)
        trim_count = int(len(sorted_times) * 0.2)
        trimmed = sorted_times[trim_count:-trim_count] if trim_count > 0 else sorted_times
        std_time_ms = (sum((t - avg_time_ms)**2 for t in trimmed) / len(trimmed)) ** 0.5 if trimmed else 0
        
        if verbose:
            print(f"{avg_time_ms:7.2f} ms (±{std_time_ms:.2f})")
        
        data_points.append((total_instructions, avg_time_ms))
        instruction_counts.append(total_instructions)
        times_ms.append(avg_time_ms)
    
    # Linear regression: time_ms = slope * instructions + intercept
    if HAS_SCIPY:
        slope, intercept, r_value, p_value, std_err = stats.linregress(instruction_counts, times_ms)
        r_squared = r_value ** 2
    else:
        slope, intercept, r_squared = simple_linear_regression(instruction_counts, times_ms)
    
    # Convert slope from ms/instruction to ns/instruction
    latency_ns = slope * 1e6
    
    return CalibrationResult(
        benchmark=template_name,
        description=tmpl['description'],
        instruction_latency_ns=latency_ns,
        overhead_ms=intercept,
        r_squared=r_squared,
        data_points=data_points
    )


def print_results(results: List[CalibrationResult]):
    """Print calibration results in a readable format."""
    print("\n" + "="*70)
    print("CALIBRATION RESULTS")
    print("="*70)
    print(f"{'Benchmark':<15} {'Latency (ns)':<14} {'Overhead (ms)':<14} {'R²':<10}")
    print("-"*70)
    
    for r in results:
        print(f"{r.benchmark:<15} {r.instruction_latency_ns:>11.4f}   {r.overhead_ms:>11.2f}   {r.r_squared:>8.6f}")
    
    print("\n" + "="*70)
    print("INTERPRETATION")
    print("="*70)
    
    for r in results:
        freq_ghz = 1 / (r.instruction_latency_ns) if r.instruction_latency_ns > 0 else 0
        cpi = r.instruction_latency_ns * 3.5  # Assume 3.5 GHz M2 P-core
        
        print(f"\n{r.benchmark}:")
        print(f"  Formula: time = {r.instruction_latency_ns:.4f} ns × instructions + {r.overhead_ms:.2f} ms")
        print(f"  Implied throughput: {freq_ghz:.2f} G instructions/sec")
        print(f"  At 3.5 GHz: {cpi:.2f} cycles per instruction (CPI)")
        print(f"  Process startup overhead: {r.overhead_ms:.2f} ms")
        print(f"  R² = {r.r_squared:.6f} {'(excellent fit)' if r.r_squared > 0.999 else '(good fit)' if r.r_squared > 0.99 else ''}")


def main():
    """Run the full calibration suite."""
    print("="*70)
    print("M2Sim Linear Regression Calibration Tool")
    print("Methodology: Issue #88 (Human's suggestion)")
    print("="*70)
    
    # Iteration counts following Human's suggestion: 100, 200, 400, 800, 1600...
    # Scaled up significantly so instruction time >> overhead (~20ms)
    # At ~3.5B instr/sec, need >70M instructions to get >20ms instruction time
    # Using: 1M, 2M, 4M, 8M, 16M, 32M iterations
    # With 20 instructions/iter, this gives 20M to 640M total instructions
    iteration_counts = [
        1_000_000,
        2_000_000,
        4_000_000,
        8_000_000,
        16_000_000,
        32_000_000,
    ]
    
    results = []
    
    for template_name in BENCHMARK_TEMPLATES:
        result = calibrate_benchmark(
            template_name,
            iteration_counts,
            runs_per_count=15,  # More runs for better statistics
            verbose=True
        )
        results.append(result)
    
    print_results(results)
    
    # Save results to JSON
    import json
    output = {
        "methodology": "linear_regression",
        "formula": "time_ms = latency_ns * instruction_count / 1e6 + overhead_ms",
        "results": [
            {
                "benchmark": r.benchmark,
                "description": r.description,
                "instruction_latency_ns": r.instruction_latency_ns,
                "overhead_ms": r.overhead_ms,
                "r_squared": r.r_squared,
                "data_points": [{"instructions": d[0], "time_ms": d[1]} for d in r.data_points]
            }
            for r in results
        ]
    }
    
    output_path = Path(__file__).parent / "calibration_results.json"
    output_path.write_text(json.dumps(output, indent=2))
    print(f"\nResults saved to: {output_path}")


if __name__ == "__main__":
    main()
