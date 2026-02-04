# M2Sim Progress Report

*Last updated: 2026-02-04 07:10 EST*

## Current Milestone: M6 - Validation

### Status Summary
- **M1-M5:** ✅ Complete
- **M6:** 🚧 In Progress

### Recent Activity (2026-02-04)

**Committed to main:**
- [Bob] Added branch predictor stats to timing harness (d9124d0)
  - BranchPredictorStats() method on Pipeline
  - Displays predictions, correct, mispredictions, accuracy %
- [Bob] Created `docs/spec-integration-plan.md` - roadmap for SPEC CPU 2017
- [Cathy] Created `docs/m2-microarchitecture-research.md` - M2 Avalanche specs

**Key Research Findings:**
- M2 Avalanche core: 8-wide decode, 7 ALUs, 4 LSUs, ~630 entry ROB
- M2Sim: 6-wide issue, 6 ALUs, in-order execution
- Root cause of accuracy gap: narrower issue width + no OoO execution

**Current Accuracy:**
| Benchmark | Sim CPI | M2 CPI | Error |
|-----------|---------|--------|-------|
| arithmetic_sequential | 0.400 | 0.268 | 49.3% |
| dependency_chain | 1.200 | 1.009 | 18.9% |
| branch_taken | 1.800 | 1.190 | 51.3% |
| **Average** | | | **39.8%** |

### Open Issues

| Issue | Priority | Status |
|-------|----------|--------|
| #122 | Medium | Quality - pipeline.go refactoring (analysis complete) |
| #115 | High | M6 - Investigate accuracy gaps for <2% target |
| #107 | High | [Human] SPEC benchmark suite available (integration planned) |

### Open PRs
None

### Blockers
- Fundamental accuracy limitation: M2Sim is in-order, M2 is out-of-order
- For <2% accuracy, may need OoO simulation or accept higher target

### Next Steps
1. Increase issue width to 8-wide to better match M2
2. Consider adding more ALU slots (7 vs current 6)
3. Begin SPEC benchmark integration per plan in docs/
4. Evaluate if OoO execution is required for accuracy target

## Milestones Overview

| Milestone | Description | Status |
|-----------|-------------|--------|
| M1 | Foundation (MVP) | ✅ Complete |
| M2 | Memory & Control Flow | ✅ Complete |
| M3 | Timing Model | ✅ Complete |
| M4 | Cache Hierarchy | ✅ Complete |
| M5 | Advanced Features | ✅ Complete |
| M6 | Validation | 🚧 In Progress |
