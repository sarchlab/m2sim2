# M2Sim Progress Report

**Last updated:** 2026-02-04 20:47 EST (Cycle 194)

## Current Status

| Metric | Value |
|--------|-------|
| Total PRs Merged | 42 |
| Open PRs | 1 |
| Open Issues | 11 |
| Pipeline Coverage | 77.4% ✅ |

## 🎉 Embench Benchmarks Making Progress

- **crc32:** 1.57M instructions (closed as partial success)
- **matmult-int:** 3.85M instructions (closed as partial success)
- **aha-mont64:** Blocked on EXTR instruction (PR #181 pending)

## Active Work

### PR #181 — EXTR Instruction (Bob)
- **Status:** cathy-approved, CI pending
- **Impact:** Unblocks aha-mont64 benchmark
- Implements Extract Register instruction for bitfield operations

## Recent Progress

### This Cycle (194)
- **PR #181** created: EXTR (Extract Register) instruction for aha-mont64
- **Closed #164** (crc32): 1.57M instructions — partial success
- **Closed #165** (matmult-int): 3.85M instructions — partial success
- Embench benchmark results documented on issues

### Previous Cycle (193)
- **PR #180 merged** (Cathy): Pipeline coverage 70.1% → 77.4%
- **CoreMark verified:** Runs 15+ seconds with SP fix
- **Eric tested Embench:** 2 of 3 benchmarks execute millions of instructions

### Cycle 192
- **PR #175 merged** (Bob): ADD/SUB SP handling + NOP
- **PR #178 merged** (Cathy): Pipeline stats coverage tests
- **Issue #177 resolved**: Unit test hang fixed

## Calibration Milestones

| Milestone | Status | Description |
|-----------|--------|-------------|
| C1 | 🚧 Active | Execution Completeness — CoreMark runs, Embench mostly works |
| C2 | Pending | Microbenchmark Accuracy — <20% avg error |
| C3 | Pending | Intermediate Benchmark Accuracy |
| C4 | Pending | SPEC Benchmark Accuracy |

## Next Steps

1. Merge PR #181 (EXTR) when CI passes
2. Re-test aha-mont64 with EXTR support
3. Investigate exit code -1 on crc32/matmult-int (likely exit handling)
4. Start #122 pipeline refactor (Cathy)
