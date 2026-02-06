# M6 Completion Criteria Evaluation

**Author:** Eric (AI Researcher)  
**Date:** 2026-02-05 (Cycle 271)  
**Purpose:** Evaluate what's needed to complete M6 Validation milestone

## M6 Definition (from SPEC.md)

> **M6: Validation** — Final accuracy validation
> - [x] Run microbenchmark suite
> - [ ] Compare with real M2 timing data
> - [ ] Achieve <20% average error

## Current Status vs Requirements

| Requirement | Status | Notes |
|-------------|--------|-------|
| Microbenchmark suite | ✅ Complete | 3 microbenchmarks run |
| M2 timing comparison | ❌ Blocked | Requires human to run on real M2 |
| <20% average error | ⏸️ Pending | Need intermediate benchmarks per #141 |

## Critical Finding: #141 Caveats

Per issue #141, Human explicitly stated:

> "Must use at least **intermediate-size benchmarks**"  
> "**Microbenchmarks should NOT be included** in the accuracy measurement"

**This means:** The 20.2% microbenchmark accuracy does NOT count for M6 completion!

## What Actually Counts for M6

### Required: Intermediate Benchmarks

| Suite | Count | Status |
|-------|-------|--------|
| PolyBench | 4 | ✅ gemm, atax, 2mm, mvt ready |
| Embench | 5 | ✅ aha-mont64, crc32, matmult-int, primecount, edn ready |
| **Total** | **10** | ✅ Sufficient for validation |

### Required: M2 Baselines

**Current:** 0 baselines captured  
**Needed:** At least 10 benchmarks with M2 hardware measurements

**How to capture baselines:**
1. Build native versions (not bare-metal ELFs)
2. Run on real M2 hardware
3. Use performance counters (`perf` or Apple's instruments)
4. Record cycle counts for each benchmark

### Required: <20% Average Error

**Current:** Unknown (no intermediate benchmark accuracy measured)  
**Microbenchmark accuracy:** 20.2% (does NOT count per #141)

## Blocking Issues

### 1. M2 Baseline Capture (Human Required)

The team cannot capture M2 baselines autonomously:
- Requires physical access to M2 hardware
- Requires native macOS builds (not bare-metal ELFs)
- Requires performance counter integration

**Estimated effort:** 2-4 hours for 10 benchmarks

### 2. Native Benchmark Builds

Current benchmarks are **bare-metal ELFs** for simulation.
For M2 baselines, we need **native macOS executables**.

PolyBench native builds would be straightforward:
```bash
clang -O2 -o gemm_native benchmarks/polybench/gemm/gemm.c
```

Embench benchmarks may need modification for native execution.

## M6 Completion Checklist

- [x] Implement intermediate benchmarks (10 ready)
- [ ] Build native versions for M2 testing
- [ ] Capture M2 baselines (human required)
- [ ] Run timing simulation for each benchmark
- [ ] Calculate accuracy against M2 baselines
- [ ] Verify <20% average error

## Timeline Estimate

| Task | Owner | Duration |
|------|-------|----------|
| Native benchmark builds | Bob | 2-4h |
| M2 baseline capture | Human | 2-4h |
| Timing simulation runs | Bob | 1-2h |
| Accuracy analysis | Eric | 1h |
| **Total** | — | **6-11h** |

## Recommendations

### Immediate (Agents Can Do)

1. ✅ Benchmark count is sufficient (10 ready)
2. Bob: Implement statemate (#245) to reach 11
3. Eric: Document simulation results for ready benchmarks
4. Cathy: Maintain coverage toward 70% target

### Requires Human Involvement

1. **Decision:** Accept PolyBench + Embench for M6 validation?
2. **Action:** Build native versions of gemm/atax for M2
3. **Action:** Run on M2 with performance counters
4. **Action:** Provide baseline cycle counts

### If Human is Unavailable

We can:
1. Continue expanding benchmark suite toward 15+
2. Continue pipeline coverage work (69.6% → 70%)
3. Research alternative validation approaches
4. Document expected accuracy based on microbenchmark results

## Publication Readiness

Per docs/literature-survey-simulator-validation.md:
- gem5 validation papers show 11-25% IPC error is typical
- 15+ benchmarks recommended for publication credibility
- Our 10 benchmarks are acceptable, 15+ is better

## Conclusion

**M6 completion is blocked on M2 baseline capture**, which requires human involvement. The agent team has:

1. ✅ Implemented 10 intermediate benchmarks
2. ✅ Fixed all timing simulator bugs
3. ✅ Achieved 79.9% emu coverage, 69.6% pipeline coverage
4. ⏸️ Awaiting M2 baselines for accuracy validation

**Next step:** Human to capture M2 baselines for gemm, atax, 2mm, mvt (PolyBench) and/or Embench benchmarks.

---
*This evaluation supports issue #141 (accuracy target approval) and #240 (publication readiness).*
