# M2 Timing Baseline Measurement Methodology

**Document Version:** 1.0  
**Last Updated:** 2026-02-03  
**Author:** Cathy (QA Agent)

## Purpose

This document describes the methodology used to collect real Apple M2 hardware timing data for validating M2Sim's accuracy against the <2% error target.

## The Problem

Direct measurement of benchmark execution time on macOS is challenging:

1. **Process Startup Overhead**: Creating a process adds ~18ms overhead
2. **Short Benchmarks**: Our benchmarks execute ~20 instructions (nanoseconds)
3. **Overhead >> Signal**: Process overhead is millions of times larger than benchmark time

Simple timing like `time ./benchmark` is useless for micro-benchmarks.

## Solution: Linear Regression Calibration

We use linear regression to separate process overhead from instruction latency.

### Mathematical Model

```
execution_time = instruction_latency × instruction_count + process_overhead
```

Where:
- `execution_time` (ms): Measured wall-clock time
- `instruction_latency` (ns/instruction): What we want to measure
- `instruction_count`: Number of instructions executed
- `process_overhead` (ms): Constant startup/teardown cost

### Procedure

1. **Generate benchmarks with varying instruction counts**
   - Same benchmark structure, different iteration counts
   - E.g., 1M, 2M, 4M, 8M, 16M, 32M iterations

2. **Run each benchmark multiple times**
   - 15 runs per configuration
   - 3 warmup runs (discarded)
   - Trimmed mean (remove 20% outliers)

3. **Fit linear regression**
   - X-axis: instruction count
   - Y-axis: execution time
   - Slope = instruction latency
   - Intercept = process overhead

4. **Validate fit quality**
   - R² > 0.999 indicates excellent fit
   - All our benchmarks achieve R² > 0.999

### Why This Works

By varying instruction count while keeping process overhead constant, we can isolate the instruction-dependent component. The linear regression naturally separates the two.

Example from calibration:
```
arithmetic: time = 0.0766 ns × instructions + 2.47 ms
R² = 0.9996 (excellent fit)
```

## Benchmark Descriptions

| Benchmark | Description | Instructions/Iter | Target Measurement |
|-----------|-------------|-------------------|-------------------|
| arithmetic | 20 independent ADDs | 20 | ALU throughput (superscalar) |
| dependency | 20 dependent ADDs (RAW chain) | 20 | Pipeline forwarding latency |
| branch | 5 taken branches | 5 | Branch predictor performance |

## Hardware Configuration

- **Processor**: Apple M2 (Performance Cores @ 3.5 GHz)
- **OS**: macOS (Apple Silicon native)
- **Conditions**: System idle, no thermal throttling

## Derived Metrics

From instruction latency, we calculate:

**Cycles Per Instruction (CPI)**:
```
CPI = latency_ns × frequency_GHz
CPI = latency_ns × 3.5
```

**Instructions Per Cycle (IPC)**:
```
IPC = 1 / CPI
```

## Calibration Results Summary

| Benchmark | Latency (ns/inst) | CPI @ 3.5 GHz | R² |
|-----------|------------------|---------------|-------|
| arithmetic | 0.077 | 0.27 | 0.9996 |
| dependency | 0.288 | 1.01 | 0.9999 |
| branch | 0.340 | 1.19 | 0.9997 |

## Accuracy Target

M2Sim targets <2% error against these baselines.

**Error Formula**:
```
error = abs(sim_latency - real_latency) / min(sim_latency, real_latency)
```

## Running Calibration

To regenerate baseline data:

```bash
cd benchmarks/native
python3 linear_calibration.py
```

Output: `calibration_results.json`

To compare simulator vs baseline:

```bash
python3 accuracy_report.py
```

Output: `accuracy_results.json`, `accuracy_report.md`

## Alternative Methods

### Apple Instruments (xctrace)

For direct performance counter access:
```bash
xctrace record --template 'CPU Counters' --output trace.trace --launch -- ./benchmark
```

**Pros**: Direct cycle counts, no process overhead issues  
**Cons**: More complex setup, requires Xcode

### perf (Linux)

If running on Linux with M-series via VM:
```bash
perf stat ./benchmark
```

## Data Files

| File | Description | Format |
|------|-------------|--------|
| `calibration_results.json` | Raw calibration data | JSON |
| `m2_baseline.json` | Authoritative baseline for comparison | JSON |
| `accuracy_results.json` | Latest accuracy comparison | JSON |

## Statistical Considerations

1. **Warmup**: 3 runs discarded to warm caches
2. **Outlier Removal**: 20% trimmed mean
3. **Multiple Runs**: 15 timed runs per configuration
4. **Multiple Scales**: 6 different instruction counts (1M-32M iters)

This produces statistically robust latency estimates with R² > 0.999.

## Limitations

1. **P-cores Only**: Assumes execution on performance cores (3.5 GHz)
2. **Idle System**: Results vary under load
3. **No E-cores**: Efficiency cores not characterized
4. **Warm Cache**: All measurements assume cache hits

## Revision History

| Date | Version | Changes |
|------|---------|---------|
| 2026-02-03 | 1.0 | Initial methodology document |
