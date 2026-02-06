# Literature Survey: CPU Simulator Validation Methodology

**Author:** Eric (AI Researcher)  
**Date:** 2026-02-05  
**Issue:** #240 - Publication-Ready M2Sim: Research-Driven Milestone Planning

## Executive Summary

This survey examines recent (2019-2024) top-tier computer architecture publications to understand benchmark standards, validation methodologies, and accuracy expectations for cycle-accurate CPU simulators targeting venues like ISCA, MICRO, HPCA, and ASPLOS.

## 1. Benchmarks Used

### Standard Benchmark Suites

| Suite | Typical Usage | Benchmark Count | Notes |
|-------|---------------|-----------------|-------|
| **SPEC CPU 2017** | Primary industry standard | 43 (23 INT, 20 FP) | Required for publication credibility |
| **SPEC CPU 2006** | Legacy, still cited | 29 total | Being phased out |
| **CoreMark** | Embedded/mobile validation | 1 score | Industry standard for ARM |
| **Embench-IoT** | Embedded workloads | 19 benchmarks | Modern, diverse workloads |
| **PolyBench/C** | Kernel-level validation | 30 kernels | Good for loop/memory patterns |
| **MiBench** | Embedded domains | 35 benchmarks | Automotive, telecom, security |

### Key Finding
**SPEC CPU 2017 is the gold standard** for publication. Papers without SPEC validation face reviewer skepticism. However, for specialized simulators (embedded, mobile), CoreMark + Embench provide acceptable alternatives.

## 2. Benchmark Quantity

### Typical Counts in Publications

| Venue | Paper Type | Benchmarks Used | Notes |
|-------|-----------|-----------------|-------|
| ISCA/MICRO | Full simulator | 15-43 (SPEC subset or full) | Often SimPoint sampling |
| HPCA | Microarch study | 10-20 representative | Selected by diversity |
| IEEE CAL | Short paper | 5-10 benchmarks | Minimum viable set |
| Workshop | Preliminary | 3-5 benchmarks | Proof of concept |

### Recommendations
- **For M2Sim MVP publication:** 10-15 diverse benchmarks minimum
- **For top-tier venue:** SPEC CPU 2017 subset (~20 benchmarks) + CoreMark
- **Shortcut:** Use representative subset methodology (see SPECcast paper)

## 3. Success Metrics / Accuracy Expectations

### gem5 Validation Studies (Reference Point)

From "Validation of the gem5 Simulator for x86 Architectures" (SC 2019):

| Benchmark Category | Average IPC Error | Notes |
|-------------------|-------------------|-------|
| Control benchmarks | 39% | Branch prediction sensitivity |
| Dependency benchmarks | 8.5% | Well-modeled |
| Execution benchmarks | 458% | Outliers, specific features |
| Memory benchmarks | 38.7% | Cache/memory system dependent |
| **Overall average** | ~11-18% | After calibration |

From "Performance Error Evaluation of gem5 for ARM Server" (2023):
- **Average absolute error: within 25%** for multi-threaded workloads
- More than twice as accurate as one-IPC simulation

### Industry-Accepted Accuracy Targets

| Accuracy Level | Error Range | Publication Suitability |
|----------------|-------------|------------------------|
| Excellent | <10% | Novel methodology claims |
| Good | 10-20% | Standard for new simulators |
| Acceptable | 20-30% | With justification |
| Marginal | 30-50% | Requires explanation |
| Poor | >50% | Unlikely to pass review |

### M2Sim Current Status
- **Microbenchmarks:** 20.2% average error (does NOT count per #141)
- **Intermediate benchmarks:** Not yet measured (PolyBench awaiting M2 baselines)
- **Target:** <20% average error on intermediate benchmarks

## 4. Validation Methodology

### Standard Approaches

1. **Hardware Baseline Capture**
   - Run benchmarks on real hardware
   - Use hardware performance counters
   - Collect: cycles, instructions, IPC, cache miss rates

2. **Simulator Execution**
   - Run same benchmarks in simulator
   - Match configuration to hardware (cache sizes, latencies, etc.)

3. **Error Calculation**
   ```
   IPC_error = |IPC_sim - IPC_hw| / IPC_hw × 100%
   CPI_error = |CPI_sim - CPI_hw| / CPI_hw × 100%
   ```

4. **Reporting Standards**
   - Report per-benchmark error
   - Report average (geometric mean preferred)
   - Report max error
   - Explain outliers

### What Reviewers Look For

1. **Methodology transparency**: How was hardware data collected?
2. **Configuration matching**: Does simulator match hardware specs?
3. **Benchmark diversity**: Multiple categories (int, fp, memory, control)
4. **Error analysis**: Why do errors exist? Which features cause them?
5. **Reproducibility**: Can others verify the results?

## 5. Recommendations for M2Sim

### Immediate Actions (Phase 1)
1. **Capture M2 baselines for PolyBench** (gemm, atax) — requires human
2. **Implement CoreMark validation** — CoreMark ELF exists, needs instruction support
3. **Document validation methodology** — create `docs/validation-methodology.md`

### Short-Term (Phase 2)
1. **Add 5-10 more PolyBench benchmarks**
2. **Run Embench-IoT subset** (aha-mont64, crc32, matmult-int)
3. **Establish ~15 benchmark baseline**

### Medium-Term (Phase 3)
1. **SPEC CPU 2017 test inputs** (reduced size for simulation)
2. **Cross-validate with gem5 ARM results**
3. **Publication draft preparation**

### Target Metrics for Publication

| Metric | Target | Current | Gap |
|--------|--------|---------|-----|
| Benchmark count | 15+ intermediate | 2 (gemm, atax) | +13 |
| Average IPC error | <20% | Unknown (awaiting baselines) | TBD |
| Max IPC error | <40% | Unknown | TBD |
| CoreMark score error | <15% | Not measured | TBD |

## 6. Key References

1. Akram & Sawalha. "Validation of the gem5 Simulator for x86 Architectures." SC Workshops 2019.
2. "Performance Error Evaluation of gem5 Simulator for ARM Server." 2023.
3. "SPECcast: Methodology for Fast Performance Evaluation with SPEC CPU 2017." ICPP 2020.
4. "Wait of a Decade: Did SPEC CPU 2017 Broaden the Performance Horizon?" HPCA.
5. "A Workload Characterization of the SPEC CPU2017 Benchmark Suite." ISPASS 2018.

## 7. Conclusion

M2Sim's <20% accuracy target is well-aligned with published simulator validation standards. The critical gap is:

1. **Benchmark coverage**: Need 15+ intermediate benchmarks (currently 2)
2. **M2 baselines**: Human must capture hardware data for gemm/atax
3. **Validation methodology document**: Formal documentation for publication

**Priority recommendation**: Focus on PolyBench expansion + CoreMark, as these provide publication-credible validation without requiring SPEC CPU licensing.

---
*This survey addresses Issue #240 requirements for understanding publication standards in computer architecture simulation.*
