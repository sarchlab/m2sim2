# Arithmetic Throughput Gap Analysis

**Author:** Cathy (QA)  
**Date:** 2026-02-04  
**Issue:** M2Sim arithmetic_sequential shows 0.450 CPI vs M2's 0.268 CPI (67.9% error)

## Executive Summary

The arithmetic throughput gap stems from M2Sim's 4-wide issue configuration vs the real M2's **8-wide decode** and **6 integer ALUs**. Additionally, the M2 eliminates unconditional branches entirely (0-cycle), which we don't model.

## Research Findings: Apple M2 Microarchitecture

Based on [Dougall Johnson's Firestorm documentation](https://dougallj.github.io/applecpu/firestorm.html) and [AnandTech's M1 deep dive](https://www.anandtech.com/show/16226/apple-silicon-m1-a14-deep-dive/2):

### Integer Execution Units (6 total)

| Unit | Capabilities |
|------|--------------|
| 1 | ALU + bitfield + flags + branch + adr + msr/mrs nzcv + mrs |
| 2 | ALU + bitfield + flags + branch + adr + msr/mrs nzcv + indirect branch + ptrauth |
| 3 | ALU + bitfield + flags + mov-from-simd/fp |
| 4 | ALU + bitfield + mov-from-simd/fp |
| 5 | ALU + bitfield + mul + div |
| 6 | ALU + bitfield + mul + madd + crc + bfm/extr |

**All 6 units can execute basic ALU operations (ADD, SUB, AND, OR, XOR).**

### Pipeline Characteristics

- **Decode width:** 8 uops per cycle
- **Taken branches:** 1 per cycle
- **Integer ALUs:** 6 (all can do ADD)
- **Multiply units:** 2 (units 5 and 6)
- **Load units:** 3 (units 9, 10, and 8 can load)
- **Store units:** 2 (units 7 and 8)

### Instruction Elimination (Zero-Cycle Operations)

The M2 can **eliminate** certain instructions completely:

- `mov x0, #0` - handled by renaming
- `mov x0, x1` - handled by renaming (64-bit only)
- `mov x0, #immediate` - up to 2 per 8 instructions via renaming
- `nop` - never issues
- **`b` (unconditional branch) - NEVER ISSUES!**

This is significant: **unconditional branches cost 0 cycles** on M2!

### Instruction Fusion

Certain instruction pairs fuse into single uops:

- `adds/subs/ands/cmp/tst + b.cc` → fuses when reading ≤4 registers per 6 instructions
- `add/sub/and/orr/eor + cbnz/cbz` → fuses if destination matches cbz operand
- AES instruction pairs for cryptography

## Analysis: arithmetic_sequential Benchmark

### The Benchmark Code

```go
// 20 independent ADDs to different registers (X0-X4, repeated)
EncodeADDImm(0, 0, 1, false),  // X0 = X0 + 1
EncodeADDImm(1, 1, 1, false),  // X1 = X1 + 1  
EncodeADDImm(2, 2, 1, false),  // X2 = X2 + 1
EncodeADDImm(3, 3, 1, false),  // X3 = X3 + 1
EncodeADDImm(4, 4, 1, false),  // X4 = X4 + 1
// ... repeated 4 times ...
EncodeSVC(0)  // syscall
```

**Total:** 20 ADDs + 1 SVC = 21 instructions

### Expected Behavior on Real M2

With 6 ALUs and 8-wide decode:
- 20 independent ADDs can execute at 6/cycle max = 3.33 cycles for ALU work
- Pipeline fill/drain adds ~4-5 cycles (5-stage pipeline)
- Total: ~7-8 cycles for 21 instructions
- Expected CPI: 7-8 / 21 ≈ **0.33-0.38**

The measured 0.268 CPI suggests even better ILP, likely due to:
- Overlapped pipeline stages
- Eliminated instructions
- Potential for >6 issues with contrived patterns

### Current M2Sim Behavior

With 4-wide issue:
- 20 ADDs at 4/cycle = 5 cycles for ALU work  
- Pipeline overhead: ~4 cycles
- Total: ~9 cycles
- CPI: 9 / 20 ≈ **0.45** ✓ matches our measurement!

## Proposed Configuration Changes

### Option 1: Increase Issue Width to 6

This is the most accurate change. M2 has 6 integer ALUs, so a 6-wide superscalar config would be appropriate.

```go
// Add to superscalar.go
func SextupleIssueConfig() SuperscalarConfig {
    return SuperscalarConfig{
        IssueWidth: 6,
    }
}
```

**Expected improvement:**
- 20 ADDs / 6 = 3.33 cycles + 4 overhead = ~7.33 cycles
- CPI: 7.33 / 20 ≈ 0.37 (still ~38% error, but improved from 68%)

### Option 2: Implement Branch Elimination

Since unconditional branches (B, BL) never issue on M2, we should add elimination:

```go
// In execute stage, check for eliminated instructions
func (p *Pipeline) isEliminated(inst *insts.Instruction) bool {
    switch inst.Op {
    case insts.OpB:  // unconditional branch
        return true
    // Future: add MOV elimination
    }
    return false
}
```

### Option 3: Full 8-Wide (Maximum Accuracy)

For maximum accuracy, match M2's 8-wide decode. This would require:
- Octuple issue registers (8 pipeline slots)
- More complex dependency checking
- Higher implementation complexity

## Recommendation

**Phase 1 (Immediate):** Increase issue width to 6
- Moderate implementation effort
- Significant accuracy improvement
- Matches M2's actual integer ALU count

**Phase 2 (Future):** Add branch/move elimination
- More complex but adds realism
- Helps with branch_taken benchmark too

**Phase 3 (Long-term):** Consider 8-wide decode
- Only if Phase 1-2 don't achieve target accuracy

## Impact on Other Benchmarks

| Benchmark | Current Error | Expected After 6-Wide |
|-----------|---------------|----------------------|
| arithmetic_sequential | 67.9% | ~30-40% |
| dependency_chain | 18.9% | Similar (bottleneck is dependencies, not width) |
| branch_taken | 101.7% | Slightly improved (branches still limit to 1/cycle) |

## References

1. [Dougall Johnson - Firestorm Overview](https://dougallj.github.io/applecpu/firestorm.html)
2. [AnandTech - Apple M1 Deep Dive](https://www.anandtech.com/show/16226/apple-silicon-m1-a14-deep-dive/2)
3. [ocxtal - M1 Optimization Notes](https://github.com/ocxtal/insn_bench_aarch64/blob/master/optimization_notes_apple_m1.md)
