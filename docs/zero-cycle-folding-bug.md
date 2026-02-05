# Zero-Cycle Branch Folding Bug Analysis

**Author:** Eric (Researcher)  
**Date:** 2026-02-05  
**Issue:** PR #233 acceptance tests hang even after branch handling fix (d159a73)

## Summary

The zero-cycle branch folding feature in `tickOctupleIssue()` eliminates conditional branches at fetch time without any verification mechanism. This causes infinite loops when the branch prediction becomes wrong.

## Root Cause

In `pipeline.go` around line 5555, conditional branches can be "folded" (eliminated at fetch time):

```go
// Zero-cycle branch folding: high-confidence predicted-taken conditional branches
// are eliminated similar to unconditional B.
if isCond, _ := isFoldableConditionalBranch(word, fetchPC); isCond {
    pred := p.branchPredictor.Predict(fetchPC)
    if pred.TargetKnown && pred.Taken && pred.Confidence >= 3 {
        fetchPC = pred.Target
        p.stats.FoldedBranches++
        continue  // ← Branch never enters pipeline!
    }
}
```

**The Problem:**
1. When a branch is folded, it **never enters the pipeline**
2. It is **never decoded or executed**
3. The branch condition is **NEVER VERIFIED**
4. If prediction was wrong, there is **NO RECOVERY MECHANISM**

## Bug Reproduction

For `branch_hot_loop` benchmark:
```
SUB X0, X0, #1   // 0x1000
CMP X0, #0       // 0x1004  
B.NE loop        // 0x1008 (target = 0x1000)
SVC 0            // 0x100c
```

**Trace:**
```
Cycle  8: PC=0x1000, Flushes=2, Folded=0  (2nd misprediction → BTB trained)
Cycle  9: PC=0x1008, Flushes=2, Folded=3  (← B.NE now folded!)
Cycle 10: PC=0x1004, Flushes=2, Folded=4  (PC stuck, X0 keeps decrementing)
...
```

After 2 branch updates (both taken), the bimodal counter reaches 2. But the tournament predictor + gshare can reach confidence=3 quickly due to hash aliasing.

When X0 becomes 0:
- B.NE should NOT be taken (Z=1, NE=false)
- But the branch is FOLDED (never checked)
- fetchPC keeps redirecting to 0x1000
- Loop runs forever

## Why Single-Issue Works

Single-issue mode does NOT have zero-cycle branch folding. Every branch instruction:
1. Enters the pipeline (IFID)
2. Gets decoded (IDEX)
3. Gets executed with flag evaluation
4. Branch outcome is verified
5. Mispredictions are recovered via flush + PC redirect

## Proposed Fix

**Option 1 (Conservative):** Disable folding for conditional branches:
```go
// Only fold unconditional branches, not conditional
if isEliminableBranch(word) {
    // ... fold unconditional B ...
}
// Remove the conditional branch folding code entirely
```

**Option 2 (Correct but Complex):** Add verification mechanism:
1. Create "folded branch" record with PC and predicted outcome
2. At some point, verify the prediction was correct
3. If wrong, flush pipeline and redirect PC to correct target
4. This is how real processors implement zero-cycle branches

## Impact

- PR #233 (hot branch benchmark) cannot pass until this is fixed
- All conditional branch loops in 8-wide mode are affected
- Single-issue, dual-issue, etc. modes work correctly

## Test Case

```go
// Minimal reproduction:
pipe := pipeline.NewPipeline(regFile, memory, pipeline.WithOctupleIssue())
// Run branch_hot_loop benchmark
// Hangs at PC=0x1004 with FoldedBranches incrementing each cycle
```

## Related

- Bob's branch handling fix (d159a73) is CORRECT and NECESSARY
- But it cannot fix branches that never reach the execute stage
- The root cause is fetch-stage elimination without verification
