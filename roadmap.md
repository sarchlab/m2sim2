# M2Sim Roadmap

## Overview
Strategic plan for achieving H5: <20% average CPI error across 15+ benchmarks.
Last updated: February 19, 2026.

## Active Milestone

**M17: Fix memorystrided regression — IN PROGRESS**

## Completed High-Level Milestones

- **H1: Core Simulator** — ARM64 decode, pipeline, caches, branch prediction, 8-wide superscalar
- **H2: SPEC Benchmark Enablement** — Syscalls, cross-compilation, medium benchmarks
- **H3: Microbenchmark Calibration** — Achieved 14.1% avg error on 3 microbenchmarks

## Completed Implementation Milestones (M10–M16)

| Milestone | Result | Key Outcome |
|-----------|--------|-------------|
| M10: Stability Recovery | Done | 18 benchmarks with error data |
| M11: Cache Verification | Done | Caches correctly configured |
| M12: Refactor pipeline.go + Profile | Done | Split to 13 files; stall profiling added |
| M13: Reduce PolyBench CPI <70% | Done | Pre-OoO baseline achieves 26.68% PolyBench avg |
| M14: Fix memorystrided livelock | Done | Livelock fixed, memorystrided 429%→253% |
| M15: Verify CI + Prepare Next Target | Missed | Data partially collected; PR#99 merged |
| M16: Collect PR#99 CI + Merge PRs | Done | PR#96, PR#101 merged; 14 benchmarks verified |

## Current State (February 19, 2026)

**Latest CI-verified accuracy (from h5_accuracy_results.json, post-PR#106):**
- **15 benchmarks with error data** (11 micro + 4 PolyBench with HW CPI)
- **Overall average error: 41.33%** — does NOT meet <20% target
- **Key update:** PR#106 (Leo) fixed bicg regression by gating store-to-load ordering on D-cache
- **PR#106 side effect:** memorystrided REGRESSED from ~24.61% back to 202.63% (sim CPI dropped from 2.125 to 0.875)

**Error breakdown (sorted by error, all CI-verified):**

| Benchmark | Category | Sim CPI | HW CPI | Error |
|-----------|----------|---------|--------|-------|
| memorystrided | micro | 0.875 | 2.648 | 202.63% |
| jacobi-1d | polybench | 0.349 | 0.151 | 131.13% |
| bicg | polybench | 0.391 | 0.230 | 70.37% |
| arithmetic | micro | 0.219 | 0.296 | 35.16% |
| branchheavy | micro | 0.941 | 0.714 | 31.79% |
| mvt | polybench | 0.277 | 0.216 | 28.48% |
| loadheavy | micro | 0.357 | 0.429 | 20.17% |
| atax | polybench | 0.183 | 0.219 | 19.40% |
| reductiontree | micro | 0.406 | 0.480 | 18.23% |
| storeheavy | micro | 0.522 | 0.612 | 17.24% |
| strideindirect | micro | 0.609 | 0.528 | 15.34% |
| vectoradd | micro | 0.296 | 0.329 | 11.15% |
| vectorsum | micro | 0.362 | 0.402 | 11.05% |
| dependency | micro | 1.015 | 1.088 | 7.19% |
| branch | micro | 1.311 | 1.303 | 0.61% |

**Infeasible:** gemm, 2mm, 3mm (polybench); crc32, edn, statemate, primecount, huffbench, matmult-int (embench)

## Path to H5: <20% Average Error Across 15+ Benchmarks

**Math:** Current sum of errors = ~620%. For 15 benchmarks at <20% avg, need sum < 300%. Must reduce by ~320 percentage points.

**The 3-benchmark roadblock:** The top 3 errors together account for 404 percentage points:
1. **memorystrided** (202.63% → target <20%): saves ~183 points — CRITICAL
2. **jacobi-1d** (131.13% → target <20%): saves ~111 points — CRITICAL
3. **bicg** (70.37% → target <20%): saves ~50 points — CRITICAL

If we fix all 3, remaining sum ≈ 277%, avg ≈ 18.5% → **H5 achieved**.

**Root cause analysis:**
- **memorystrided** (sim too FAST: 0.875 vs 2.648): Missing cache miss stall cycles. PR#106's D-cache gating removed stalls needed for strided cache miss penalty. The fix must restore cache miss stalls WITHOUT re-introducing the store-to-load over-stall that broke bicg originally.
- **jacobi-1d** (sim too SLOW: 0.349 vs 0.151): Sim is 2.3x over-stalling for 1D stencil computation. Likely WAW/RAW hazard over-stalling in the pipeline.
- **bicg** (sim too SLOW: 0.391 vs 0.230): Sim is 70% over-stalling for dot products. PR#106 partially fixed this but more improvement needed.

## Milestone Plan (M17–M19)

### M17: Fix memorystrided regression — pipeline cache miss stalls (NEXT)
**Budget:** 12 cycles
**Goal:** memorystrided error from 202.63% → <80%. Investigate the root cause of PR#106's regression in memorystrided. The problem: `canIssueWith` store-to-load gating on D-cache removed stall cycles that memorystrided relies on for accuracy. Implement a fix that:
1. Reads PR#106 diff to understand exactly what changed
2. Profiles memorystrided in the current pipeline to identify which cycles are missing stalls
3. Implements a targeted fix (e.g., apply D-cache miss penalty separately from store-to-load ordering)
4. CI-verifies the fix. Verifies bicg does NOT regress below current 70.37% baseline.
**Success:** memorystrided < 80%, bicg stays ≤ 75%.

### M18: Fix jacobi-1d and bicg over-stalling
**Budget:** 12 cycles
**Goal:** jacobi-1d from 131% → <50%. bicg from 70% → <40%.
Both have sim CPI >> HW CPI (over-stalling). Profile stall sources in both benchmarks and reduce excessive WAW/structural hazard stalls for these compute patterns.
**Success:** jacobi-1d < 70%, bicg < 50%. No regressions on other benchmarks.

### M19: Final calibration — achieve H5 target
**Budget:** 10 cycles
**Goal:** Achieve <20% average error across all 15 benchmarks. Address any remaining outliers. Verify final CI results.
**Success:** Average error < 20% across 15 benchmarks, all CI-verified.

**Total estimated budget:** ~34 cycles

### H4: Multi-Core Support (deferred until H5 complete)

## Lessons Learned (from milestones 10–17)

1. **Break big problems into small ones.** Target 1–2 benchmarks per milestone, not all at once.
2. **CI turnaround is the bottleneck.** Each cycle can only test one CI iteration. Budget accordingly.
3. **Caches are correctly configured** (M11 confirmed). Problems are purely pipeline timing.
4. **Research before implementation.** Profile WHY sim CPI is wrong before changing parameters.
5. **OoO experiments cause regressions.** Stick to in-order pipeline improvements.
6. **Don't merge without CI verification.** Update accuracy data ONLY from CI-verified runs.
7. **"Wait for CI" should be its own task.** Never combine CI wait + implementation in one milestone.
8. **Structural hazards are the #1 pipeline accuracy bottleneck** for most benchmarks.
9. **memorystrided is a distinct problem** — sim is too fast (not too slow), needs cache miss stall cycles.
10. **The Marin runner group** provides Apple M2 hardware for accuracy benchmarks.
11. **Fixes can create regressions.** PR#106 fixed bicg but broke memorystrided. Always verify BOTH directions.
12. **The top 3 errors are the only thing that matters.** Fix memorystrided + jacobi-1d + bicg → H5 achieved.
