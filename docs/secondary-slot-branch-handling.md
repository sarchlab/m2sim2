# Secondary Slot Branch Handling — Fix Pattern

**Author:** Eric (Researcher)
**Date:** 2026-02-05
**Status:** CRITICAL — Required for PR #233 to pass

## Problem Summary

PR #233 (hot branch benchmark) times out because **branch handling only exists for slot 0 (primary slot)**.

In 8-wide mode, the hot branch loop executes:
```
Slot 0: SUB X0, X0, #1   (p.idex)  - not a branch
Slot 1: CMP X0, #0       (p.idex2) - sets flags ✅ (Cathy fix 48851e7)
Slot 2: B.NE loop        (p.idex3) - IS a branch but NO HANDLING ❌
```

### What Works
1. ✅ PSTATE flags forwarded correctly (Cathy fix 9d7c2e6)
2. ✅ Same-cycle flag forwarding for 8-wide (Cathy fix 48851e7)
3. ✅ `ExecuteWithFlags()` computes `BranchTaken = true` correctly

### What's Missing
4. ❌ **No code checks `p.idex3.IsBranch`** (or idex2, idex4, etc.)
5. ❌ **PC is never redirected** for branches in secondary slots
6. ❌ **Pipeline is never flushed** when branch in slots 2-8 mispredicts

**Result:** Loop counter decrements, B.NE evaluates `BranchTaken=true`, but nothing happens → infinite loop.

## Root Cause Analysis

### Current Code Structure (slot 0 only)

```go
// Line 3981 in tickOctupleIssue()
if p.idex.IsBranch {
    actualTaken := execResult.BranchTaken
    actualTarget := execResult.BranchTarget
    // ... prediction verification ...
    if wasMispredicted {
        p.pc = branchTarget
        p.flushAllIFID()
        p.flushAllIDEX()
        // Clear secondary EXMEM registers
        p.exmem2.Clear()
        p.exmem3.Clear()
        // ...
        return
    }
}
```

### Why Unit Tests Pass

Unit tests use **single-issue mode** (default). In single-issue mode:
- B.NE gets placed in slot 0 (`p.idex`)
- The existing branch handling code at line 3981 executes
- Branches work correctly

### Why Acceptance Tests Fail

Acceptance harness uses **8-wide mode** (`WithOctupleIssue`). In 8-wide mode:
- The decoder packs multiple instructions per cycle
- B.NE ends up in slot 2 or 3 (`p.idex3`)
- **No branch handling exists for slots 2-8**
- Branch result computed but never acted upon

## Fix Pattern

For each secondary slot (idex2 through idex8), add branch handling AFTER the execution code but BEFORE moving to the next slot.

### Template for Each Slot

```go
// Execute slot N
if p.idexN.Valid && !memStall && !execStall {
    // ... existing flag forwarding code ...
    // ... execution code ...
    
    nextEXMEMN = /* build EXMEM register */
    
    // >>> ADD THIS: Branch handling for slot N <<<
    if p.idexN.IsBranch {
        actualTaken := execResult.BranchTaken
        actualTarget := execResult.BranchTarget
        
        p.stats.BranchPredictions++
        
        predictedTaken := p.idexN.PredictedTaken
        predictedTarget := p.idexN.PredictedTarget
        earlyResolved := p.idexN.EarlyResolved
        
        wasMispredicted := false
        if actualTaken {
            if !predictedTaken {
                wasMispredicted = true
            } else if predictedTarget != actualTarget {
                wasMispredicted = true
            }
        } else {
            if predictedTaken {
                wasMispredicted = true
            }
        }
        
        if earlyResolved && actualTaken {
            wasMispredicted = false
        }
        
        p.branchPredictor.Update(p.idexN.PC, actualTaken, actualTarget)
        
        if wasMispredicted {
            p.stats.BranchMispredictions++
            branchTarget := actualTarget
            if !actualTaken {
                branchTarget = p.idexN.PC + 4
            }
            p.pc = branchTarget
            p.flushAllIFID()
            p.flushAllIDEX()
            p.stats.Flushes++
            
            // Latch MEMWB results, clear later EXMEM registers
            if !memStall {
                p.memwb = nextMEMWB
                p.memwb2 = nextMEMWB2
                // ... all memwb ...
                p.exmem = nextEXMEM
                // Latch EXMEM up to current slot
                p.exmem2 = nextEXMEM2  // for slot 3+
                // Clear EXMEM from current slot onward
                p.exmemN.Clear()  // current slot
                p.exmemN+1.Clear()  // later slots
                // ...
            }
            return
        }
        p.stats.BranchCorrect++
    }
}
```

## Specific Line Locations

| Slot | IDEX Register | Insert After Line | EXMEM to Latch | EXMEM to Clear |
|------|---------------|-------------------|----------------|----------------|
| 2 | `p.idex2` | ~4098 (after nextEXMEM2 built) | exmem | exmem2-8 |
| 3 | `p.idex3` | ~4172 (after nextEXMEM3 built) | exmem, exmem2 | exmem3-8 |
| 4 | `p.idex4` | ~4246 (after nextEXMEM4 built) | exmem, exmem2, exmem3 | exmem4-8 |
| 5 | `p.idex5` | ~4320 (after nextEXMEM5 built) | exmem-4 | exmem5-8 |
| 6 | `p.idex6` | ~4394 (after nextEXMEM6 built) | exmem-5 | exmem6-8 |
| 7 | `p.idex7` | ~4468 (after nextEXMEM7 built) | exmem-6 | exmem7-8 |
| 8 | `p.idex8` | ~4542 (after nextEXMEM8 built) | exmem-7 | exmem8 |

## Important Considerations

### 1. EXMEM Latching on Misprediction

When a branch in slot N mispredicts:
- **Latch** EXMEM registers for slots 0 through N-1 (instructions BEFORE the branch)
- **Clear** EXMEM registers for slots N through 8 (branch and instructions AFTER)

This is because instructions fetched after the branch (in same packet) are from wrong path.

### 2. Required Fields in SecondaryIDEXRegister

Verify `SecondaryIDEXRegister` (and Tertiary, etc.) have these fields:
- `IsBranch bool`
- `PredictedTaken bool`
- `PredictedTarget uint64`
- `EarlyResolved bool`

Check `timing/pipeline/superscalar.go` for the register definitions.

### 3. Test Strategy

After implementing, verify:
1. `TestCountdownLoop` still passes
2. `TestBackwardBranch` still passes
3. Run acceptance tests locally: `go test -v ./tests/acceptance/...`
4. Check PR #233 CI passes

## Estimated Effort

~50-80 lines of code per slot × 7 slots = 350-560 lines total.

The code is repetitive but critical for correctness. Consider refactoring into a helper function after it works.

## References

- Primary slot branch handling: `timing/pipeline/pipeline.go` lines 3981-4050
- PSTATE forwarding fix: commit 48851e7
- Same-cycle forwarding: commit 9d7c2e6
- Hot branch benchmark: PR #233
