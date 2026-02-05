# M2Sim Progress Report

**Last updated:** 2026-02-05 16:44 EST (Cycle 257)

## Current Status

| Metric | Value |
|--------|-------|
| Total PRs Merged | 74 |
| Open PRs | 1 |
| Open Issues | 14 |
| Pipeline Coverage | 72.8% ✅ |
| Emu Coverage | 79.9% ✅ |

## Cycle 257 Updates

- **PR #233** (Bob: Hot branch benchmark) — CI still failing despite PSTATE fixes
  - Bob rebased on main with Cathy's same-cycle fix (48851e7)
  - Build ✅, Lint ✅, Unit Tests ✅, **Acceptance Tests ❌** (timeout)
  - Unit tests pass locally but acceptance tests (8-wide mode) still hang
  - Deeper investigation needed: may be another issue beyond PSTATE forwarding
- **Root cause analysis ongoing:**
  - Eric documented in `docs/timing-sim-debugging.md`
  - Unit tests use single-issue (default), acceptance tests use 8-wide (`WithOctupleIssue`)
  - Cathy's fix IS in the branch, but something else may be blocking

**Open PRs:**
- PR #233: cathy-approved ✅, CI failing (Acceptance Tests timeout even with PSTATE fix)

**Critical blocker:** PR #233 hangs in acceptance tests despite all PSTATE fixes being included. Needs deeper timing sim investigation.

## Cycle 255 Updates

- **Cathy: Fixed PSTATE flag forwarding for ALL superscalar slots (2-8)** — root cause found!
  - Previous fix only covered slot 1; CMP in slot 2 + B.NE in slot 3 still read stale flags
  - Added SetsFlags/FlagN/Z/C/V to all EXMEM registers (3-8)
  - Changed Execute→ExecuteWithFlags for slots 2-8
  - TestCountdownLoop and TestBackwardBranch now pass ✅
- **PR #233** (Bob: Hot branch benchmark) — Rebased with PSTATE fix, CI running
  - Build ✅, Lint ✅, Unit Tests ✅, Acceptance Tests in progress
- **Issue #236 CLOSED** — PSTATE forwarding fix is complete

## Key Achievements

**Emu Coverage Target Exceeded!**
| Package | Coverage | Status |
|---------|----------|--------|
| emu | 79.9% | ✅ Above 70% target! |

**8-Wide Infrastructure Validated!**
| Benchmark | CPI | IPC | Error vs M2 |
|-----------|-----|-----|-------------|
| arithmetic_8wide | 0.250 | 4.0 | **6.7%** ✅ |

## Accuracy Status (Microbenchmarks)

| Benchmark | Simulator CPI | M2 Real CPI | Error | Priority |
|-----------|---------------|-------------|-------|----------|
| arithmetic_8wide | 0.250 | 0.268 | **6.7%** | ✅ Target met! |
| dependency_chain | 1.200 | 1.009 | **18.9%** | ✅ Near target |
| branch_taken_conditional | 1.600 | 1.190 | **34.5%** | ⚠️ Waiting for PR #233 |

**Target:** <20% average error

## Optimization Progress

| Priority | Optimization | Status |
|----------|--------------|--------|
| 1 | ✅ CMP + B.cond fusion (PR #212) | Merged |
| 2 | ✅ 8-wide decode infrastructure (PR #215) | Merged |
| 3 | ✅ BTB size increase 512→2048 (PR #227) | Merged |
| 4 | ✅ Zero-cycle predicted-taken branches (PR #230) | Merged |
| 5 | ✅ PSTATE forwarding for all slots (9d7c2e6, 48851e7) | Merged to main |
| 6 | 🔄 Hot branch benchmark (PR #233) | Needs rebase on main |

## Coverage Analysis

| Package | Coverage | Status |
|---------|----------|--------|
| timing/cache | 89.1% | ✅ |
| timing/pipeline | 72.8% | ✅ |
| timing/latency | 73.3% | ✅ |
| timing/core | 100% | ✅ |
| emu | 79.9% | ✅ Target exceeded! |

## Completed Optimizations

1. ✅ CMP + B.cond fusion (PR #212) — 62.5% → 34.5% branch error
2. ✅ 8-wide decode infrastructure (PR #215)
3. ✅ 8-wide benchmark enable (PR #220)
4. ✅ arithmetic_8wide benchmark (PR #223) — validates 8-wide, 6.7% error
5. ✅ BTB size increase 512→2048 (PR #227)
6. ✅ Emu coverage 79.9% (PRs #214, #217, #218, #222, #225, #226, #228, #229)
7. ✅ Zero-cycle predicted-taken branches (PR #230)
8. ✅ PSTATE forwarding fixes (9d7c2e6, 48851e7) — all 8 superscalar slots

## Stats

- 74 PRs merged total
- 1 open PR (#233 hot branch benchmark — needs rebase)
- 258+ tests passing
- All coverage targets exceeded ✓
- 8-wide arithmetic accuracy: **6.7%** ✓
- Emu coverage: **79.9%** ✓
- Pipeline coverage: **72.8%** ✓
