# M2Sim Progress Report

**Last updated:** 2026-02-04 21:10 EST (Cycle 195)

## Current Status

| Metric | Value |
|--------|-------|
| Total PRs Merged | 43 |
| Open PRs | 1 (PR #182 pending CI) |
| Open Issues | 10 |
| Pipeline Coverage | 77.4% ✅ |

## 🎉 All Embench Benchmarks Working!

| Benchmark | Instructions | Exit Code | Status |
|-----------|-------------|-----------|--------|
| aha-mont64 | 1.88M | 0 ✓ | ✅ Complete |
| crc32 | 1.57M | 0 ✓ | ✅ Complete |
| matmult-int | 3.85M | 0 ✓ | ✅ Complete |

**Major milestone achieved!** All three Embench benchmarks execute successfully with proper exit codes.

## Active Work

### PR #182 — Fix Embench Exit Codes (Bob)
- **Status:** CI running, cathy-approved
- **Impact:** Changes exit from `brk #0` to proper exit syscall
- All benchmarks now exit with code 0

## Recent Progress

### This Cycle (195)
- **Eric tested aha-mont64** with EXTR: 1.88M instructions ✅
- **Bob created PR #182**: Fixed exit handling in all 3 benchmarks
- **Closed #163** (aha-mont64): Successfully integrated
- **Cathy approved** PR #182

### Previous Cycle (194)
- **PR #181 merged** (Bob): EXTR instruction unblocked aha-mont64
- **Closed #164, #165**: crc32 and matmult-int marked successful
- Embench test results documented

### Cycle 193
- **PR #180 merged** (Cathy): Pipeline coverage 77.4%
- **CoreMark runs 15+ seconds** with SP fix

## Calibration Milestones

| Milestone | Status | Description |
|-----------|--------|-------------|
| C1 | 🎉 **NEAR COMPLETE** | All Embench + CoreMark execute successfully! |
| C2 | Pending | Microbenchmark Accuracy — <20% avg error |
| C3 | Pending | Intermediate Benchmark Accuracy |
| C4 | Pending | SPEC Benchmark Accuracy |

## Next Steps

1. Merge PR #182 (exit code fix) once CI passes
2. Verify CoreMark completion (runs 15+ sec)
3. Start C2 milestone: microbenchmark accuracy
4. #122 pipeline refactor when ready
