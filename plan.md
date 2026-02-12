# Leo Plan: Next Accuracy Target Selection (Issue #438)

## Decision: PolyBench Simulation + Baseline Fix > vectorsum/strideindirect > SPEC

### Current State Summary

**Microbenchmarks (11 benchmarks):** ~10.8% average error — STRONG
- strideindirect: 13.6% (down from 74.3% after PR #459)
- vectorsum: 19.4%, vectoradd: 19.8% (within 20% target)
- All other microbenchmarks: <17%

**PolyBench (7 benchmarks):** 235,453% average error — BROKEN BASELINES
- Sim CPI values are reasonable (1-5 CPI → 0.3-1.4 ns/inst)
- Hardware baselines are wrong: 956-9236 ns/inst (should be ~0.3 ns/inst)
- Root cause: 16x16 ELF total runtime divided by instruction count, without accounting for startup overhead or using multi-scale linear regression

**SPEC:** Blocked on infrastructure (no self-hosted runner)

### Plan

#### Step 1: Create PolyBench CI workflow (issue #464)
- Add GitHub Actions workflow that runs `TestPolybench*` without `-short` flag
- Capture actual sim CPI values for all 7 PolyBench benchmarks
- 5-minute timeout per benchmark should be sufficient for 16x16 MINI size

#### Step 2: Diagnose PolyBench baseline methodology failure
- Document that current PolyBench baselines in calibration_results.json are invalid
- The PolyBench entries have `hardware_baseline: true` but the instruction_latency_ns values (956-9236 ns) are 3000-30000x higher than real M2 per-instruction latency (~0.3 ns)
- This is because the measurement divided total ELF runtime by instruction count without multi-scale regression to remove startup overhead

#### Step 3: Fix PolyBench baseline methodology
- Build PolyBench at multiple sizes (MINI through MEDIUM)
- Measure on real M2 hardware using linear regression (same methodology as microbenchmarks)
- Replace broken baselines with proper per-instruction latency values
- This requires native ARM64 compilation at multiple sizes — may need to modify build.sh

#### Step 4 (lower priority): vectorsum/vectoradd improvement
- Both at ~19.5% error, within 20% target
- Root cause is in-order pipeline limitation (can't overlap independent work with load-use stalls)
- Would require OOO-like bypass — significant complexity for <5% improvement
- Defer until PolyBench baselines are fixed

### Next Execute Cycle Deliverable
- Lock issue #464
- Create GH Actions workflow for PolyBench timing execution
- Comment on issue #463 with root cause analysis of PolyBench baseline failure
