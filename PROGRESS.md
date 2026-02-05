# M2Sim Progress Report

*Last updated: 2026-02-04 19:35 EST*

## Current Milestone: C1 - Execution Completeness

### Status Summary
- **M1-M5:** ✅ Complete
- **M6 (Validation):** 🚧 In Progress → Calibration milestones C1-C4
- **C1:** 🚧 In Progress (CoreMark at 2406 instructions)

### Recent Activity (2026-02-04)

**This cycle (19:35):**
- Grace: Skipped (cycle 190, not a 10th)
- Alice: Updated task board, action count 189→190
- Eric: Evaluated status, 15 open issues healthy
- Bob: **Fixed PR #175 tests** — added encodeMOVZ helper
- Cathy: Reviewed Bob's test fix — approved ✅
- Dana: Fixed gofmt formatting in PR #175, pushed CI fix

**Progress:**
- **PR #175** — ADD/SUB SP handling + NOP support (cathy-approved, CI re-running)
- Test fix: encodeMOVZ added for proper ARM64 immediate value setting
- Tests corrected to use MOVZ instead of incorrectly assuming ADD from SP gives zero
- **CoreMark: 2406 instructions** (now properly reports BRK trap)
- **40 PRs merged** total — excellent velocity!

### Blockers Status

**Previous blockers RESOLVED ✅**
- Cross-compiler: `aarch64-elf-gcc 15.2.0` ✅
- SPEC: `benchspec/CPU` exists ✅
- Logical immediate instructions ✅
- LSLV, UBFM, STR register offset, CCMP ✅
- BRK instruction ✅

**Current blockers:**
- PR #175 waiting on CI (lint fixed, awaiting re-run)
- CoreMark hits BRK #0x3e8 at PC=0x80BA8 (2406 instructions)
- Bob investigating why x21 becomes 0 in core_list_mergesort

### Open PRs

| PR | Title | Status |
|----|-------|--------|
| #175 | [Bob] Fix ADD/SUB SP handling + NOP | cathy-approved, CI re-running |

### Calibration Milestones

Per #167 discussion, Eric's proposal approved by Human:
- **C1: Execution Completeness** — Run CoreMark/Embench to completion 🚧
- **C2: Microbenchmark Accuracy (<30%)** — Tune timing parameters
- **C3: Intermediate Benchmark Accuracy (<20%)** — Validate overall timing
- **C4: SPEC Accuracy (stretch)** — Target <25%

### Next Steps

1. **Merge PR #175** — once CI passes
2. **Debug BRK trap** — trace execution to find why x21 becomes 0
3. **Complete CoreMark** — achieve C1 milestone
4. **Begin #122 refactor** — after C1 completes
5. Continue **Embench-IoT phase 2** after CoreMark validates

### Current Accuracy (microbenchmarks)

| Benchmark | Sim CPI | M2 CPI | Error | Root Cause |
|-----------|---------|--------|-------|------------|
| arithmetic_sequential | 0.400 | 0.268 | 49.3% | M2 has 8+ ALUs |
| branch_taken | 1.800 | 1.190 | 51.3% | Branch elim overhead |
| dependency_chain | 1.200 | 1.009 | 18.9% | Forwarding latency |
| **Average** | | | **39.8%** | |

**Note:** 20% target applies to INTERMEDIATE benchmarks, not microbenchmarks.

### Test Coverage

| Package | Coverage | Notes |
|---------|----------|-------|
| **insts** | **97%+** ✅ | instruction tests comprehensive |
| timing/cache | 89.1% | |
| benchmarks | 80.8% | |
| emu | 72.5% | |
| timing/latency | 71.8% | |
| timing/core | 60.0% | |
| timing/pipeline | 25.6% | #122 refactor pending |

### Recent Merges

| PR | Title | Status |
|----|-------|--------|
| #174 | [Bob] Add BRK instruction support | ✅ Merged |
| #173 | [Bob] Implement LSLV, UBFM, STR reg offset, CCMP | ✅ Merged |
