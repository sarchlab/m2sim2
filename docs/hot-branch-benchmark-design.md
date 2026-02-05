# Hot Branch Benchmark Design

**Author:** Eric (Researcher)  
**Date:** 2026-02-05  
**Related:** Issue #45 (cycle 249), Zero-cycle branch folding (PR #230)

## Problem Statement

The current `branchTakenConditional` benchmark uses **cold branches** — each branch instruction has a unique PC that is executed only once. This prevents validation of the zero-cycle branch folding optimization (PR #230), which requires:

1. **BTB hit** — target known from previous execution
2. **Predicted taken** — branch prediction
3. **High confidence** — counter ≥ 3 (requires 3+ executions of the same branch)

### Current Benchmark Analysis

```
branchTakenConditional:
  PC 0x1000: CMP X0, #0
  PC 0x1004: B.GE +8        ← cold branch #1 (PC seen once)
  PC 0x1008: (skipped)
  PC 0x100C: ADD X0, X0, #1
  
  PC 0x1010: CMP X0, #0
  PC 0x1014: B.GE +8        ← cold branch #2 (different PC, seen once)
  ...
```

Each branch has a **unique PC** → BTB miss every time → no zero-cycle folding.

## Solution: Hot Branch Benchmark with Loop

Design a benchmark where the **same branch PC is executed multiple times** (≥10 iterations) to:
1. Train the BTB (learn the target)
2. Build predictor confidence (counter reaches 3)
3. Enable zero-cycle folding for subsequent iterations

### Proposed Benchmark: `branch_hot_loop`

```asm
; Hot branch benchmark - same B.NE executed 16 times
; Tests zero-cycle folding after predictor training

    MOV X0, #16           ; loop counter
loop:
    SUB X0, X0, #1        ; X0--
    CMP X0, #0
    B.NE loop             ; ← HOT BRANCH: same PC, executed 16 times
    
    ; Loop exit
    MOV X0, #0            ; exit code 0
    SVC #0                ; exit syscall
```

### Key Characteristics

| Property | Cold Benchmark | Hot Benchmark |
|----------|---------------|---------------|
| Branch PCs | 5 unique | 1 (same PC) |
| Iterations per branch | 1 | 16 |
| BTB state | Always miss | Hit after 1st |
| Predictor confidence | 1 (weak) | 3+ (strong) |
| Zero-cycle eligible | Never | Iterations 4-16 |

### Expected Behavior

1. **Iteration 1:** BTB miss, prediction made, target learned
2. **Iteration 2:** BTB hit, predicted taken, confidence=1
3. **Iteration 3:** BTB hit, predicted taken, confidence=2  
4. **Iteration 4+:** BTB hit, predicted taken, confidence=3 → **ZERO-CYCLE FOLDING**

With 16 iterations:
- Iterations 1-3: Normal branch penalty (~2-3 cycles each)
- Iterations 4-16: Zero-cycle (13 folded branches)
- Expected FoldedBranches stat: ~13

### Verification Strategy

Run the benchmark and check:
1. `FoldedBranches > 0` — confirms folding is occurring
2. Reduced overall CPI compared to cold branch benchmark
3. Branch accuracy near 100% (only first iteration mispredicted)

## Implementation Details

### Register Usage
- X0: Loop counter (initialized to 16)
- X8: Syscall number (93 = exit)

### Instruction Sequence

```go
func branchHotLoop() Benchmark {
    return Benchmark{
        Name:        "branch_hot_loop",
        Description: "16-iteration loop with single hot branch - validates zero-cycle folding",
        Setup: func(regFile *emu.RegFile, memory *emu.Memory) {
            regFile.WriteReg(8, 93) // X8 = 93 (exit syscall)
            regFile.WriteReg(0, 16) // X0 = 16 (loop counter)
        },
        Program: BuildProgram(
            // loop:
            EncodeSUBImm(0, 0, 1, false),  // X0 = X0 - 1
            EncodeCMPImm(0, 0),            // CMP X0, #0
            EncodeBCond(-8, 1),            // B.NE loop (-8 bytes = -2 instructions)
            
            // exit:
            EncodeMOVImm(0, 0, false),     // X0 = 0 (exit code)
            EncodeSVC(0),                  // exit
        ),
        ExpectedExit: 0,
    }
}
```

### Native Benchmark (for M2 baseline)

```asm
; benchmarks/native/branch_hot_loop.s
.global _main
_main:
    mov x0, #16
loop:
    sub x0, x0, #1
    cmp x0, #0
    b.ne loop
    
    mov x0, #0
    ret
```

## Expected Accuracy Impact

| Metric | Before | After | Notes |
|--------|--------|-------|-------|
| branch_hot_loop CPI | ~1.6 | ~0.8-1.0 | 50% reduction expected |
| FoldedBranches | 0 | ~13 | Confirms folding |
| Branch accuracy | ~60% | ~93% | Only 1 mispred (first iter) |

## Next Steps

1. Bob implements `branchHotLoop()` benchmark
2. Add native benchmark for M2 baseline measurement
3. Run accuracy validation
4. If FoldedBranches > 0, zero-cycle folding is validated
5. Compare hot vs cold branch accuracy to quantify improvement
