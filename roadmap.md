# M2Sim Roadmap

## Overview
Strategic plan for achieving H5: <20% average CPI error across 15+ benchmarks.
Last updated: February 19, 2026.

## Active Milestone

**M17b: Fix bicg load-use latency — IN PROGRESS**

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

**Latest CI-verified accuracy (from h5_accuracy_results.json, CI run 22204159767, commit 28f7ec1):**
- **15 benchmarks with error data** (11 micro + 4 PolyBench with HW CPI)
- **Overall average error: 23.70%** — does NOT yet meet <20% target
- **Key update from M17:** jacobi-1d reduced from 131.13% → 67.55% (target met). Bitfield+DataProc3Src forwarding gate merged. bicg still at 71.24% (load-use stall bottleneck — needs separate fix).

**Error breakdown (sorted by error, all CI-verified):**

| Benchmark | Category | Sim CPI | HW CPI | Error |
|-----------|----------|---------|--------|-------|
| bicg | polybench | 0.393 | 0.230 | 71.24% |
| jacobi-1d | polybench | 0.253 | 0.151 | 67.55% |
| branchheavy | micro | 0.970 | 0.714 | 35.85% |
| arithmetic | micro | 0.220 | 0.296 | 34.55% |
| atax | polybench | 0.183 | 0.219 | 19.40% |
| loadheavy | micro | 0.357 | 0.429 | 20.17% |
| reductiontree | micro | 0.419 | 0.480 | 14.56% |
| memorystrided | micro | 2.267 | 2.648 | 16.81% |
| storeheavy | micro | 0.522 | 0.612 | 17.24% |
| vectorsum | micro | 0.354 | 0.402 | 13.56% |
| strideindirect | micro | 0.600 | 0.528 | 13.64% |
| vectoradd | micro | 0.296 | 0.329 | 11.15% |
| mvt | polybench | 0.241 | 0.216 | 11.78% |
| dependency | micro | 1.020 | 1.088 | 6.67% |
| branch | micro | 1.320 | 1.303 | 1.30% |

**Infeasible:** gemm, 2mm, 3mm (polybench); crc32, edn, statemate, primecount, huffbench, matmult-int (embench)

## Path to H5: <20% Average Error Across 15+ Benchmarks

**Math:** Current sum of errors = ~355.5%. For 15 benchmarks at <20% avg, need sum < 300%. Must reduce by ~55.5 percentage points.

**Top priority:** bicg (71.24%) is the only benchmark keeping us from H5. If bicg reaches <20%, and arithmetic/branchheavy improve even slightly:
- bicg 71.24% → 20% saves 51 pts → sum ~304.5, avg ~20.3% — borderline
- bicg 71.24% → 20% + arithmetic 34.55% → 20% saves 51+14=65 pts → avg ~19.4% ✅ **H5 achieved**

**Root cause analysis (updated after M17):**
- **bicg** (sim too SLOW: 0.393 vs 0.230): Bottleneck is **LDR→MADD load-use latency** in the non-dcache code path. PolyBench accuracy CI runs without dcache (dcache_hits=0, dcache_misses=0). ALU forwarding cannot help — need to reduce the modeled load-use stall cycles to match M2's actual ~4-cycle L1 load-to-use latency.
- **jacobi-1d** ✅ FIXED (67.55%, below 70% target) — Bitfield+DataProc3Src forwarding gate.
- **arithmetic** (sim too FAST: 0.220 vs 0.296): In-order WAW limitation. Secondary target after bicg.
- **branchheavy** (sim too SLOW: 0.970 vs 0.714): Secondary target after bicg.

## Milestone Plan (M17b–M18)

### M17 OUTCOME (12 cycles, deadline missed)
- jacobi-1d ✅ FIXED: 131.13% → 67.55% (<70% target met). Bitfield+DataProc3Src forwarding gate implemented (commits e9a0185, 28f7ec1, branch leo/fix-fp-coissue).
- bicg ❌ NOT FIXED: 71.24% (target <50%). Root cause is LDR→MADD load-use latency, NOT ALU forwarding. The team exhausted forwarding approaches — need a different strategy.
- Overall avg improved: 29.46% → 23.70%.

### M17b: Fix bicg load-use latency (NEXT)
**Budget:** 6 cycles
**Goal:** Reduce bicg from 71.24% → <50% by tuning load-use stall cycles in the non-dcache pipeline path.
- **Root cause**: PolyBench tests run without dcache. Loads use a fixed-latency simple memory model. The modeled load-to-use latency (how many cycles until load result is available for dependent instructions) may exceed M2's actual ~4-cycle L1 latency.
- **Approach**: (1) Identify where load-use stall cycles are set in timing/pipeline/ for the non-dcache path; (2) Profile actual stall count for bicg; (3) Reduce stall cycles to match M2 hardware; (4) Open PR, run CI, verify no regressions.
- **Constraints**: Do NOT enable dcache for PolyBench. Do NOT change ALU forwarding logic. Keep jacobi-1d <70%, memorystrided ≤30%.
- **Success**: bicg < 50%, all other benchmarks at or better than current values.

### M18: Final calibration — achieve H5 target
**Budget:** 8 cycles
**Goal:** Achieve <20% average error across all 15 benchmarks. After bicg is fixed, address arithmetic (34.55%) and branchheavy (35.85%) to push overall avg below 20%.
**Success:** Average error < 20% across 15 benchmarks, all CI-verified.

**Total estimated remaining budget:** ~14 cycles

### H4: Multi-Core Support (deferred until H5 complete)

## Lessons Learned (from milestones 10–17b)

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
11. **Verify regressions with code analysis, not assumptions.** PR#106 was wrongly assumed to regress memorystrided — code analysis confirmed it didn't (D-cache gating only affects non-D-cache benchmarks).
12. **The top 2 errors are the main roadblock.** Fix jacobi-1d + bicg → H5 likely achieved.
13. **ALU forwarding has limits.** jacobi-1d yielded to forwarding fixes, but bicg's bottleneck is load-use latency — a different mechanism entirely. Always confirm which instruction type is stalling before choosing the fix.
14. **PolyBench accuracy CI runs WITHOUT dcache.** Cache-stage forwarding and D-cache path fixes have zero effect on PolyBench accuracy. Always check whether dcache is enabled when diagnosing PolyBench stalls.
15. **12 cycles is too many for one milestone.** M17 used all 12 cycles and only half-succeeded. Keep milestones to 6 cycles max for targeted fixes.
16. **One root cause per milestone.** M17 conflated two different bottlenecks (jacobi-1d = ALU forwarding; bicg = load-use latency). Each should have been its own milestone.
