#!/usr/bin/env python3
"""
Convert PolyBench hardware baselines from CSV format to calibration_results.json format
for integration with the accuracy report framework.

This script processes M2 hardware timing measurements and creates calibration data
compatible with the existing accuracy analysis pipeline.
"""

import json
import csv
from pathlib import Path


def load_baselines_csv(csv_path: Path) -> dict:
    """Load PolyBench hardware baselines from CSV."""
    baselines = {}

    with open(csv_path, 'r') as f:
        reader = csv.DictReader(f)
        for row in reader:
            if row['benchmark'].strip():  # Skip empty rows
                baselines[row['benchmark']] = {
                    'avg_time_sec': float(row['avg_time_sec']),
                    'est_cycles': float(row['est_cycles']),
                    'instructions': int(row['instructions']),
                    'cpi': float(row['cpi'])
                }

    return baselines


def convert_to_calibration_format(baselines: dict, frequency_ghz: float = 3.5) -> dict:
    """Convert baseline CPI data to calibration_results.json format.

    Args:
        baselines: Dict from load_baselines_csv()
        frequency_ghz: M2 P-core frequency for converting CPI to latency

    Returns:
        Dict in calibration_results.json format
    """
    # Benchmark descriptions for PolyBench suite
    descriptions = {
        'atax': 'Matrix transpose and vector multiply (atax kernel)',
        'bicg': 'BiCG kernel - Biconjugate gradient solver',
        'gemm': 'General matrix multiply (GEMM) - compute C = AB + C',
        'mvt': 'Matrix-vector multiply and transpose operations',
        'jacobi-1d': '1D Jacobi iteration stencil computation',
        '2mm': 'Two matrix multiplies (2MM) - D = A * B; E = C * D',
        '3mm': 'Three matrix multiplies (3MM) - G = A * B; H = C * D; I = E * H'
    }

    calibration_data = {
        "methodology": "m2_hardware_baseline",
        "formula": "instruction_latency_ns = cpi / frequency_ghz",
        "frequency_ghz": frequency_ghz,
        "note": "Hardware baselines from M2 native ARM64 execution",
        "results": []
    }

    for benchmark, data in baselines.items():
        # Convert CPI to instruction latency using frequency
        # instruction_latency_ns = CPI / frequency_GHz
        instruction_latency_ns = data['cpi'] / frequency_ghz

        benchmark_result = {
            "benchmark": benchmark,
            "description": descriptions.get(benchmark, f"{benchmark} PolyBench kernel"),
            "calibrated": True,  # These are real hardware measurements
            "instruction_latency_ns": instruction_latency_ns,
            "m2_hardware_cpi": data['cpi'],
            "instructions_per_run": data['instructions'],
            "avg_time_sec": data['avg_time_sec'],
            "r_squared": 1.0,  # Perfect correlation since this is direct measurement
            "data_points": [
                {
                    "instructions": data['instructions'],
                    "time_ms": data['avg_time_sec'] * 1000  # Convert to ms
                }
            ]
        }

        calibration_data["results"].append(benchmark_result)

    return calibration_data


def main():
    """Convert PolyBench baselines to calibration format."""
    repo_root = Path(__file__).parent

    # Input: baselines.csv with M2 hardware measurements
    baselines_csv = repo_root / "reports" / "m2-baselines" / "baselines.csv"

    # Output: PolyBench calibration results JSON
    output_json = repo_root / "benchmarks" / "polybench_calibration_results.json"

    print(f"Loading PolyBench baselines from: {baselines_csv}")
    baselines = load_baselines_csv(baselines_csv)

    print(f"Found {len(baselines)} PolyBench benchmarks:")
    for name, data in baselines.items():
        print(f"  {name}: {data['cpi']:.3f} CPI, {data['instructions']} instructions")

    print("\nConverting to calibration format...")
    calibration_data = convert_to_calibration_format(baselines)

    print(f"Writing calibration data to: {output_json}")
    with open(output_json, 'w') as f:
        json.dump(calibration_data, f, indent=2)

    print(f"✅ PolyBench calibration data created: {len(calibration_data['results'])} benchmarks")

    # Verify the output format
    print("\nGenerated calibration data summary:")
    for result in calibration_data["results"]:
        latency_ns = result["instruction_latency_ns"]
        cpi = result["m2_hardware_cpi"]
        print(f"  {result['benchmark']}: {latency_ns:.4f} ns/inst ({cpi:.3f} CPI)")


if __name__ == "__main__":
    main()