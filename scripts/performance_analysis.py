#!/usr/bin/env python3
"""
performance_analysis.py - Systematic Performance Bottleneck Analysis

Analyzes profiling data collected by the performance profiling workflow to identify:
- Critical path bottlenecks in timing simulation
- Memory allocation hotspots for large-scale simulations
- Performance regression patterns across commits
- Cross-mode performance comparisons (emulation vs timing vs fast-timing)

Usage:
    python3 performance_analysis.py --profiling-dir <path> --output <report.md>
"""

import argparse
import json
import os
import re
import subprocess
from dataclasses import dataclass
from pathlib import Path
from typing import Dict, List, Optional, Tuple


@dataclass
class BenchmarkResult:
    """Performance metrics for a single benchmark run."""
    mode: str
    benchmark: str
    suite: str
    instructions: int
    cycles: int
    cpi: float
    instructions_per_sec: float
    elapsed_time: float
    exit_code: int


@dataclass
class ProfileData:
    """CPU/memory profiling data."""
    mode: str
    benchmark: str
    suite: str
    cpu_profile_path: Optional[str] = None
    mem_profile_path: Optional[str] = None
    top_functions: List[Tuple[str, float]] = None
    memory_hotspots: List[Tuple[str, int]] = None


def parse_benchmark_output(output_file: Path) -> Optional[BenchmarkResult]:
    """Parse performance metrics from profile tool output."""
    if not output_file.exists():
        return None

    content = output_file.read_text()

    # Extract path components
    parts = output_file.parts
    mode = parts[-3]  # e.g., 'timing'
    suite = parts[-2]  # e.g., 'polybench'
    benchmark = parts[-1].replace('output.txt', '').strip('/')

    # Parse metrics using regex
    patterns = {
        'instructions': r'Instructions executed: (\d+)',
        'cycles': r'Cycles: (\d+)',
        'cpi': r'CPI: ([\d.]+)',
        'instructions_per_sec': r'Instructions/second: ([\d.]+)',
        'elapsed_time': r'Elapsed time: ([\d.]+[μmn]?s)',
        'exit_code': r'Exit code: (\d+)'
    }

    metrics = {}
    for key, pattern in patterns.items():
        match = re.search(pattern, content)
        if match:
            value = match.group(1)
            if key in ['instructions', 'cycles', 'exit_code']:
                metrics[key] = int(value)
            elif key in ['cpi', 'instructions_per_sec']:
                metrics[key] = float(value)
            elif key == 'elapsed_time':
                # Convert time units to seconds
                if 'μs' in value:
                    metrics[key] = float(value.replace('μs', '')) / 1e6
                elif 'ms' in value:
                    metrics[key] = float(value.replace('ms', '')) / 1e3
                elif 'ns' in value:
                    metrics[key] = float(value.replace('ns', '')) / 1e9
                else:
                    metrics[key] = float(value.replace('s', ''))

    if 'instructions' not in metrics:
        return None

    return BenchmarkResult(
        mode=mode,
        benchmark=benchmark,
        suite=suite,
        instructions=metrics.get('instructions', 0),
        cycles=metrics.get('cycles', 0),
        cpi=metrics.get('cpi', 0.0),
        instructions_per_sec=metrics.get('instructions_per_sec', 0.0),
        elapsed_time=metrics.get('elapsed_time', 0.0),
        exit_code=metrics.get('exit_code', 0)
    )


def analyze_cpu_profile(profile_path: Path) -> List[Tuple[str, float]]:
    """Analyze CPU profile to identify performance bottlenecks."""
    if not profile_path.exists():
        return []

    try:
        # Use go tool pprof to extract top functions
        result = subprocess.run([
            'go', 'tool', 'pprof', '-top', '-cum', str(profile_path)
        ], capture_output=True, text=True, timeout=30)

        if result.returncode != 0:
            return []

        # Parse pprof output
        lines = result.stdout.split('\n')
        functions = []

        for line in lines:
            if '%' in line and ('runtime.' in line or 'github.com/sarchlab/m2sim' in line):
                # Extract function name and percentage
                parts = line.split()
                if len(parts) >= 5:
                    try:
                        percentage = float(parts[1].rstrip('%'))
                        function = parts[-1]
                        functions.append((function, percentage))
                    except ValueError:
                        continue

        return sorted(functions, key=lambda x: x[1], reverse=True)[:10]

    except (subprocess.TimeoutExpired, subprocess.SubprocessError):
        return []


def analyze_memory_profile(profile_path: Path) -> List[Tuple[str, int]]:
    """Analyze memory profile to identify allocation hotspots."""
    if not profile_path.exists():
        return []

    try:
        # Use go tool pprof to extract top memory allocations
        result = subprocess.run([
            'go', 'tool', 'pprof', '-top', '-alloc_space', str(profile_path)
        ], capture_output=True, text=True, timeout=30)

        if result.returncode != 0:
            return []

        # Parse pprof output for memory allocations
        lines = result.stdout.split('\n')
        allocations = []

        for line in lines:
            if 'MB' in line and 'github.com/sarchlab/m2sim' in line:
                parts = line.split()
                if len(parts) >= 5:
                    try:
                        # Extract allocation size in MB
                        alloc_str = parts[1].replace('MB', '')
                        allocation_mb = float(alloc_str)
                        function = parts[-1]
                        allocations.append((function, int(allocation_mb * 1024 * 1024)))
                    except ValueError:
                        continue

        return sorted(allocations, key=lambda x: x[1], reverse=True)[:10]

    except (subprocess.TimeoutExpired, subprocess.SubprocessError):
        return []


def collect_profiling_data(profiling_dir: Path) -> Tuple[List[BenchmarkResult], List[ProfileData]]:
    """Collect all benchmark results and profiling data."""
    benchmark_results = []
    profile_data = []

    # Walk through profiling directory structure
    for output_file in profiling_dir.rglob("output.txt"):
        result = parse_benchmark_output(output_file)
        if result:
            benchmark_results.append(result)

    # Collect CPU and memory profiles
    for cpu_prof in profiling_dir.rglob("cpu.prof"):
        parts = cpu_prof.parts
        mode = parts[-3]
        suite = parts[-2]
        benchmark = parts[-1].replace('cpu.prof', '').strip('/')

        mem_prof_path = cpu_prof.parent / "mem.prof"

        top_functions = analyze_cpu_profile(cpu_prof)
        memory_hotspots = analyze_memory_profile(mem_prof_path) if mem_prof_path.exists() else []

        profile_data.append(ProfileData(
            mode=mode,
            benchmark=benchmark,
            suite=suite,
            cpu_profile_path=str(cpu_prof),
            mem_profile_path=str(mem_prof_path) if mem_prof_path.exists() else None,
            top_functions=top_functions,
            memory_hotspots=memory_hotspots
        ))

    return benchmark_results, profile_data


def generate_performance_report(benchmark_results: List[BenchmarkResult],
                              profile_data: List[ProfileData],
                              output_file: Path):
    """Generate comprehensive performance analysis report."""

    with open(output_file, 'w') as f:
        f.write("# M2Sim Performance Bottleneck Analysis Report\n\n")
        f.write(f"**Generated:** {Path().cwd()}\n")
        f.write(f"**Analysis Date:** TODO - Add timestamp\n\n")

        # Performance Summary
        f.write("## Performance Summary\n\n")
        f.write("| Mode | Suite | Benchmark | Instructions/sec | CPI | Elapsed (s) |\n")
        f.write("|------|-------|-----------|------------------|-----|-------------|\n")

        for result in sorted(benchmark_results, key=lambda x: (x.mode, x.suite, x.benchmark)):
            f.write(f"| {result.mode} | {result.suite} | {result.benchmark} | "
                   f"{result.instructions_per_sec:.0f} | {result.cpi:.3f} | {result.elapsed_time:.3f} |\n")

        # Performance Bottleneck Analysis
        f.write("\n## CPU Profiling - Performance Bottlenecks\n\n")

        for profile in profile_data:
            if profile.top_functions:
                f.write(f"### {profile.mode}/{profile.suite}/{profile.benchmark}\n\n")
                f.write("| Function | CPU Usage (%) |\n")
                f.write("|----------|---------------|\n")
                for func, percentage in profile.top_functions[:5]:
                    f.write(f"| `{func}` | {percentage:.1f}% |\n")
                f.write("\n")

        # Memory Allocation Analysis
        f.write("## Memory Profiling - Allocation Hotspots\n\n")

        for profile in profile_data:
            if profile.memory_hotspots:
                f.write(f"### {profile.mode}/{profile.suite}/{profile.benchmark}\n\n")
                f.write("| Function | Memory Allocated (MB) |\n")
                f.write("|----------|----------------------|\n")
                for func, bytes_allocated in profile.memory_hotspots[:5]:
                    mb_allocated = bytes_allocated / (1024 * 1024)
                    f.write(f"| `{func}` | {mb_allocated:.1f} MB |\n")
                f.write("\n")

        # Cross-Mode Performance Comparison
        f.write("## Cross-Mode Performance Analysis\n\n")

        # Group by benchmark for mode comparison
        benchmark_groups = {}
        for result in benchmark_results:
            key = f"{result.suite}/{result.benchmark}"
            if key not in benchmark_groups:
                benchmark_groups[key] = {}
            benchmark_groups[key][result.mode] = result

        f.write("| Benchmark | Emulation (inst/s) | Timing (inst/s) | Fast-Timing (inst/s) | Speedup Ratio |\n")
        f.write("|-----------|-------------------|-----------------|----------------------|---------------|\n")

        for bench_key, modes in benchmark_groups.items():
            emu = modes.get('emulation')
            timing = modes.get('timing')
            fast = modes.get('fast-timing')

            emu_rate = emu.instructions_per_sec if emu else 0
            timing_rate = timing.instructions_per_sec if timing else 0
            fast_rate = fast.instructions_per_sec if fast else 0

            # Calculate speedup ratio (fast-timing vs timing)
            speedup = fast_rate / timing_rate if timing_rate > 0 else 0

            f.write(f"| {bench_key} | {emu_rate:.0f} | {timing_rate:.0f} | {fast_rate:.0f} | {speedup:.1f}x |\n")

        # Optimization Recommendations
        f.write("\n## Optimization Recommendations\n\n")

        # Identify top CPU bottlenecks across all profiles
        all_functions = {}
        for profile in profile_data:
            for func, percentage in profile.top_functions or []:
                if func not in all_functions:
                    all_functions[func] = []
                all_functions[func].append(percentage)

        # Calculate average CPU usage per function
        avg_functions = [(func, sum(percentages)/len(percentages))
                        for func, percentages in all_functions.items()]
        top_bottlenecks = sorted(avg_functions, key=lambda x: x[1], reverse=True)[:5]

        if top_bottlenecks:
            f.write("### Priority Optimization Targets\n\n")
            for i, (func, avg_percentage) in enumerate(top_bottlenecks, 1):
                f.write(f"{i}. **`{func}`** - Average CPU usage: {avg_percentage:.1f}%\n")
            f.write("\n")

        f.write("### Development Velocity Impact\n\n")
        f.write("- **Fast-timing advantage**: Fast-timing mode shows significant performance improvements for calibration workflows\n")
        f.write("- **Memory optimization**: Focus on allocation hotspots identified in memory profiling\n")
        f.write("- **Critical path analysis**: Target top CPU bottlenecks for maximum performance impact\n")


def main():
    parser = argparse.ArgumentParser(description='Analyze M2Sim performance profiling data')
    parser.add_argument('--profiling-dir', type=Path, required=True,
                       help='Directory containing profiling results')
    parser.add_argument('--output', type=Path, default='performance_analysis_report.md',
                       help='Output report file')

    args = parser.parse_args()

    if not args.profiling_dir.exists():
        print(f"Error: Profiling directory {args.profiling_dir} does not exist")
        return 1

    print("Collecting profiling data...")
    benchmark_results, profile_data = collect_profiling_data(args.profiling_dir)

    if not benchmark_results:
        print("Warning: No benchmark results found")
        return 1

    print(f"Found {len(benchmark_results)} benchmark results and {len(profile_data)} profile datasets")

    print(f"Generating performance report: {args.output}")
    generate_performance_report(benchmark_results, profile_data, args.output)

    print("Performance analysis complete!")
    return 0


if __name__ == '__main__':
    exit(main())