# M2Sim Progress Report

*Last updated: 2026-02-04 10:36 EST*

## Current Milestone: M6 - Validation

### Status Summary
- **M1-M5:** ✅ Complete
- **M6:** 🚧 In Progress

### Recent Activity (2026-02-04)

**This cycle (10:36):**
- **PR #140 MERGED** ✅ Tournament branch predictor
  - Upgraded from simple bimodal to tournament predictor (bimodal + gshare + choice)
  - Increased BHT size from 1024 to 4096 entries
  - Increased BTB size from 256 to 512 entries
  - Added 12-bit global history for gshare correlation
  - Updated misprediction penalty from 12 to 14 cycles (M2 spec)
  - Issue #135 closed

**Research update:**
- Eric commented on #134 with accuracy target analysis
- Recommended adjusted target: <20% average error
- <2% is unrealistic for in-order simulation

**Previous cycle (10:08):**
- PR #137 MERGED ✅ SPEC benchmark CI workflow
- Issues #132-136 created for accuracy work

**Current Accuracy:**
| Benchmark | Sim CPI | M2 CPI | Error |
|-----------|---------|--------|-------|
| arithmetic_sequential | 0.400 | 0.268 | 49.3% |
| dependency_chain | 1.200 | 1.009 | 18.9% |
| branch_taken | 1.800 | 1.190 | 51.3% |
| **Average** | | | **39.8%** |

*Note: Accuracy will be re-measured after tournament predictor merge.*

### Open Issues

| Issue | Priority | Status |
|-------|----------|--------|
| #138 | High | SPEC benchmark execution |
| #136 | High | Memory latency tuning |
| #134 | High | Accuracy target discussion |
| #132 | High | Intermediate benchmarks research |
| #139 | Low | Multi-core execution (long-term) |
| #129 | Low | README update |
| #122 | Low | Pipeline.go refactoring |
| #115 | Medium | M6 - Investigate accuracy gaps |
| #107 | High | SPEC benchmarks available |

### Open PRs
None - PR #140 merged this cycle!

### Accuracy Work Progress
- Phase 1: ✅ Branch predictor tuning (PR #140)
- Phase 2: 🔜 Memory latency tuning (#136)
- Phase 3: 🔜 Re-measure accuracy after tuning

### Blockers
- Fundamental accuracy limitation: M2Sim is in-order, M2 is out-of-order
- Recommendation: Adjust target to <20% for in-order simulation
- Decision needed on #134 (accuracy target) for M6 completion criteria

### Next Steps
1. Run benchmarks to measure accuracy after tournament predictor
2. Implement memory latency tuning (#136)
3. Finalize accuracy target decision (#134)
4. README update (#129)

## Milestones Overview

| Milestone | Description | Status |
|-----------|-------------|--------|
| M1 | Foundation (MVP) | ✅ Complete |
| M2 | Memory & Control Flow | ✅ Complete |
| M3 | Timing Model | ✅ Complete |
| M4 | Cache Hierarchy | ✅ Complete |
| M5 | Advanced Features | ✅ Complete |
| M6 | Validation | 🚧 In Progress |
