# M2Sim Roadmap

## Overview
This roadmap tracks milestone completion and strategic decisions for the M2Sim project. Last updated: February 17, 2026.

## Completed Milestones

### H1: Core Simulator ✅ COMPLETE
Foundation simulator with ARM64 decode, pipeline timing, cache hierarchy, branch prediction, 8-wide superscalar, and macro-op fusion.

### H2: SPEC Benchmark Enablement ✅ COMPLETE
Complete syscall infrastructure and ARM64 cross-compilation setup.

### H3: Initial Accuracy Calibration ✅ COMPLETE
Achieved <20% average CPI error on microbenchmarks (14.1%).

### Milestone 10: Stability Recovery ✅ COMPLETE (February 17, 2026)
Restored simulator stability: memorystrided regression fixed, all PolyBench timeouts resolved, 18 benchmarks with error data.

### Milestone 11 (cache verification portion): ✅ COMPLETE
Cache verification tests written and passed (PR #88, issue #183 closed). Akita caches behave as configured — no misconfigurations found. PR #87 merged.

## Failed Milestones

### Milestone 11: Reduce PolyBench CPI to <80% ❌ FAILED (25/25 cycles)
**Goal:** Reduce PolyBench average CPI error from 98% to <80%.
**Result:** Failed after 25 cycles. PolyBench average remained ~98%.
**Changes attempted:** OoO issue within fetch group (PR #85 - memory ports), instruction window 48→192, load-use stall bypass (PR #87).
**Key insight:** The in-order pipeline fundamentally overestimates CPI for loop-heavy PolyBench kernels. The M2's 330+ ROB enables massive loop-level parallelism that our pipeline doesn't model.

## Current State (February 17, 2026)

**Accuracy (CI-verified, post PR #87 merge):**
- **Microbenchmarks:** 17.00% average (11 benchmarks, meets <20% target)
- **PolyBench:** 98.15% average (7 benchmarks completing)
- **Overall:** 48.56% average (18 benchmarks with error data)

**PolyBench breakdown (sorted by error):**
| Benchmark | Sim CPI | HW CPI | Error |
|-----------|---------|--------|-------|
| gemm | 0.301 | 0.233 | 29.1% |
| bicg | 0.343 | 0.230 | 49.5% |
| mvt | 0.364 | 0.216 | 68.8% |
| atax | 0.396 | 0.219 | 81.2% |
| 3mm | 0.334 | 0.145 | 129.9% |
| 2mm | 0.328 | 0.144 | 128.6% |
| jacobi-1d | 0.453 | 0.151 | 200.0% |

**Pattern:** Sim CPI ranges 0.30–0.45 regardless of kernel. HW CPI ranges 0.14–0.23. The simulator is overestimating stalls uniformly. The worst offenders (jacobi-1d, 2mm, 3mm) are kernels with very low HW CPI (~0.14-0.15) suggesting high ILP that the in-order pipeline cannot exploit.

**Open issues requiring attention:**
- Issue #126 (human): pipeline.go is 6314 lines, needs refactoring into smaller files
- Issue #196: PR #89 (accuracy data update) needs merge
- Open PRs: #89 (accuracy data), #86 (stale, superseded by #89)

### Lessons Learned (cumulative)
1. **Break big problems into small ones.** Milestone 11 failed by targeting all 7 PolyBench kernels. Target 1-2 at a time.
2. **CI turnaround is the bottleneck.** PolyBench CI takes hours. Each cycle can only test one CI iteration. Budget cycles accordingly.
3. **Caches are correctly configured** (issue #183 resolved). The problem is purely in the pipeline timing model.
4. **Research before implementation.** Profile WHY sim CPI is high on specific kernels before changing pipeline parameters.
5. **pipeline.go needs refactoring.** At 6314 lines, it's a human-flagged maintainability issue that makes pipeline changes risky and slow.

## Milestone 12: Refactor pipeline.go + Profile PolyBench Bottlenecks

**Strategy:** Two parallel workstreams:
1. **Refactor pipeline.go** (issue #126) — split 6314-line file into logical modules. This makes future pipeline changes safer and faster. Human explicitly requested this.
2. **Profile PolyBench stalls** — add lightweight instrumentation to understand exactly where sim CPI is being overestimated (RAW stalls, structural hazards, cache misses, etc.) per kernel.

**Rationale:** Before attempting another accuracy milestone, we need (a) maintainable code to work with and (b) data on what to fix. Both are prerequisites for effective pipeline improvements.

**Estimated cycles:** 15

## Future Milestones (tentative)

### Milestone 13: Reduce gemm + bicg error to <30%
Target the two lowest-error PolyBench kernels. Use profiling data from milestone 12 to implement targeted fixes. Small, achievable goal.

### Milestone 14+: Incrementally reduce PolyBench error
Continue targeting individual kernels or small groups. Each milestone should target a specific error reduction.

### H4: Multi-Core Support (deferred)
Not started. Prerequisites: H5 accuracy target must be CI-verified first.
