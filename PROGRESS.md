# M2Sim Progress Report

*Last updated: 2026-02-04 17:29 EST*

## Current Milestone: M6 - Validation

### Status Summary
- **M1-M5:** ✅ Complete
- **M6:** 🚧 In Progress

### Recent Activity (2026-02-04)

**This cycle (17:18):**
- Grace: Skipped (cycle 184, not a 10th)
- Alice: Updated task board, assigned CoreMark testing, action count 183→184
- Eric: Responded to #167 with calibration milestone proposal (C1-C4)
- Bob: Found emulator bug! PR #166 had decoder but no emulator support
  - Created #168, fixed with PR #170
  - Created #169 for next blocker (logical immediate)
- Cathy: Reviewed and approved PR #170
- Dana: Merged PR #170 ✅

**Progress:**
- ✅ **PR #170 merged** — emulator support for CSEL, UDIV, MADD, TBZ, CBZ families
- ✅ **38 PRs merged** total — excellent velocity!
- ✅ **0 open PRs** — clean slate again
- 🔄 CoreMark: 2103 → 2106 instructions (emulator fix helped!)

### Blockers Status

**Previous blockers RESOLVED ✅**
- Cross-compiler: `aarch64-elf-gcc 15.2.0` ✅
- SPEC: `benchspec/CPU` exists ✅
- PRs #153-159, #166, #170 all merged ✅
- Emulator instruction support for decoded formats ✅

**Current blocker:**
- #169: Decoder missing logical immediate (AND, ORR, EOR, ANDS with bitmask)
- Failing instruction: `ands x1, x1, #0xffffffffffff` at PC=0x800B8

### Next Steps

1. **Implement #169** — logical immediate decoder support
2. **Continue CoreMark execution** — aim for full completion
3. **Consider calibration milestones** — per #167 discussion (C1-C4 proposal)
4. Begin **Embench-IoT phase 2** (#163, #164, #165) after CoreMark validates

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
| **insts** | **96%+** ✅ | 18 new tests in #166 |
| timing/cache | 89.1% | |
| benchmarks | 80.8% | |
| emu | 72.5% | |
| timing/latency | 71.8% | |
| timing/core | 60.0% | |
| timing/pipeline | 25.6% | #122 refactor pending |

### Open PRs

None — clean slate! 🎉

### Open Issues

| Issue | Priority | Status |
|-------|----------|--------|
| #169 | High | NEW: Logical immediate decoder (blocker) |
| #167 | High | Human: Consider calibration milestones |
| #165 | Medium | Embench: matmult-int (phase 2) |
| #164 | Medium | Embench: crc32 (phase 2) |
| #163 | Medium | Embench: aha-mont64 (phase 2) |
| #152 | — | Human directive (blockers resolved) |
| #146 | High | SPEC CPU 2017 setup |
| #145 | Low | Reduce Claude.md |
| #141 | High | Approve 20% error target |
| #138 | High | Spec benchmark execution |
| #132 | High | Intermediate benchmarks research |
| #122 | Medium | Pipeline refactor |
| #115 | High | Accuracy gaps |
| #107 | High | Spec benchmark available |

### Calibration Milestone Proposal (from #167)

Eric proposed restructuring calibration into clearer milestones:
- **C1: Execution Completeness** — Run CoreMark/Embench to completion
- **C2: Microbenchmark Accuracy (<30%)** — Tune timing parameters
- **C3: Intermediate Benchmark Accuracy (<20%)** — Validate overall timing
- **C4: SPEC Accuracy (stretch)** — Target <25% for complex benchmarks

Awaiting Alice/Human decision on adoption.
