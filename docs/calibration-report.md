# M2Sim Calibration Report

**Date:** 2026-02-03  
**Author:** Cathy (QA Agent)  
**Milestone:** M6 Validation

## Executive Summary

This report documents the baseline M2 hardware measurements and compares them with M2Sim simulator predictions. The <2% accuracy target is the focus of M6.

**Current Status:** ❌ Accuracy needs improvement (>100% error on some benchmarks)

## Methodology

We use **linear regression calibration** to measure real M2 instruction latencies. This technique separates process startup overhead (~2-3ms) from actual instruction execution time by running benchmarks with varying instruction counts.

**Key insight:** `time = latency × instructions + overhead`

By varying instruction count and fitting a line, we extract:
- Slope = instruction latency (ns/instruction)
- Intercept = process overhead (ms)
- R² = fit quality (>0.999 = excellent)

See `benchmarks/native/METHODOLOGY.md` for full details.

## M2 Hardware Baseline

| Benchmark | Description | Latency (ns/inst) | CPI | IPC | R² |
|-----------|-------------|-------------------|-----|-----|-----|
| arithmetic | 20 independent ADDs | 0.0766 | 0.27 | 3.73 | 0.9996 |
| dependency | 20 dependent ADDs (RAW chain) | 0.2883 | 1.01 | 0.99 | 0.9999 |
| branch | 5 taken branches | 0.3399 | 1.19 | 0.84 | 0.9997 |

**Key Observations:**
- **arithmetic**: IPC of 3.73 shows M2's superscalar execution (multiple independent ADDs per cycle)
- **dependency**: IPC of ~1.0 shows serialization from RAW hazards
- **branch**: Well-predicted branches have low overhead

## Simulator Results

| Benchmark | Sim Cycles | Sim Instructions | Sim CPI | Sim Latency (ns) |
|-----------|------------|------------------|---------|------------------|
| arithmetic | 24 | 20 | 1.20 | 0.343 |
| dependency | 44 | 20 | 2.20 | 0.629 |
| branch | 29 | 10 | 2.90 | 0.829 |

## Accuracy Comparison

| Benchmark | Real Latency | Sim Latency | Error |
|-----------|--------------|-------------|-------|
| arithmetic | 0.077 ns | 0.343 ns | **346.6%** |
| dependency | 0.288 ns | 0.629 ns | **118.1%** |
| branch | 0.340 ns | 0.829 ns | **143.3%** |

**Average Error:** 202.7%  
**Target:** <2%

## Root Cause Analysis

### Why Arithmetic Error Is Highest (346.6%)

**Real M2:** IPC = 3.73 (executes nearly 4 independent ADDs per cycle)
**Simulator:** CPI = 1.2 (treats as mostly sequential)

The simulator doesn't model M2's wide superscalar execution. Real M2 P-cores can issue 6+ micro-ops per cycle.

### Why Dependency Is Better (118.1%)

Dependent instructions serialize naturally, limiting ILP. Both simulator and hardware show ~1 CPI when operations depend on each other.

### Why Branch Error Is Moderate (143.3%)

The simulator models branch prediction but may have:
- Higher assumed misprediction rate
- Higher branch resolution latency

## Calibration Priorities

| Issue | Impact | Fix |
|-------|--------|-----|
| No superscalar execution | Critical | Model instruction-level parallelism |
| Branch prediction accuracy | Medium | Tune branch predictor parameters |
| Pipeline depth assumptions | Medium | Validate pipeline stage latencies |

## Next Steps

1. **Issue #97**: Accuracy analysis and tuning
   - Analyze instruction scheduling in simulator
   - Consider modeling issue width (6+ for M2)
   - Tune branch predictor parameters

2. **Expand Benchmark Coverage**
   - Add memory benchmarks (L1/L2/L3 cache)
   - Add SIMD benchmarks
   - Add mixed workloads

3. **Performance Counter Validation**
   - Use xctrace for direct cycle counts
   - Validate linear regression results

## Files

| File | Purpose |
|------|---------|
| `benchmarks/native/m2_baseline.json` | Authoritative baseline data |
| `benchmarks/native/METHODOLOGY.md` | Measurement methodology |
| `benchmarks/native/calibration_results.json` | Raw calibration output |
| `benchmarks/native/accuracy_results.json` | Latest accuracy comparison |

## Conclusion

The simulator's CPI predictions are 2-4x higher than real M2 hardware. The primary cause is that M2's aggressive out-of-order, superscalar execution is not modeled. The simulator appears to model closer to an in-order pipeline.

Achieving <2% accuracy will require significant changes to the execution model.
