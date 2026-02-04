# 4-Wide Superscalar Validation Plan

## Overview
This document outlines the validation strategy for the 4-wide superscalar implementation (Issue #111).

## Expected Impact on Benchmarks

### Benchmarks That Should Improve

| Benchmark | Current CPI | Expected 4-Wide CPI | Rationale |
|-----------|-------------|---------------------|-----------|
| arithmetic_sequential | 0.7 (dual) | ~0.35 | 20 independent ADDs can issue 4/cycle |
| loop_simulation | ~1.0 | ~0.6 | Mix of independent ops per iteration |
| matrix_operations | ~0.9 | ~0.5 | Parallel loads and independent ALU ops |
| mixed_operations | ~1.2 | ~0.8 | More ILP available with wider issue |

### Benchmarks With Minimal Change Expected

| Benchmark | Current CPI | Expected 4-Wide CPI | Rationale |
|-----------|-------------|---------------------|-----------|
| dependency_chain | 1.2 | ~1.2 | Serial dependency limits ILP |
| memory_sequential | ~2.0 | ~2.0 | Single memory port bottleneck |
| branch_taken | 1.4 | ~1.4 | Branch overhead already minimized |
| function_calls | ~1.5 | ~1.5 | BL/RET serialize execution |

## Validation Tests

### 1. Functional Correctness
All benchmarks must produce correct results:
```go
// Tests to verify:
func TestFourWideCorrectness(t *testing.T) {
    config := DefaultConfig()
    config.EnableQuadIssue = true
    harness := NewHarness(config)
    harness.AddBenchmarks(GetMicrobenchmarks())
    results := harness.RunAll()
    
    for _, r := range results {
        if r.ExitCode != r.ExpectedExit {
            t.Errorf("%s: got exit %d, want %d", r.Name, r.ExitCode, r.ExpectedExit)
        }
    }
}
```

### 2. Performance Regression
No benchmark should get *slower* with 4-wide:
```go
func TestFourWideNoRegression(t *testing.T) {
    // Run with dual-issue
    config2 := DefaultConfig()
    config2.EnableDualIssue = true
    results2 := runBenchmarks(config2)
    
    // Run with quad-issue
    config4 := DefaultConfig()
    config4.EnableQuadIssue = true
    results4 := runBenchmarks(config4)
    
    for name, r4 := range results4 {
        r2 := results2[name]
        if r4.CPI > r2.CPI * 1.05 { // Allow 5% tolerance
            t.Errorf("%s regressed: dual=%.3f, quad=%.3f", name, r2.CPI, r4.CPI)
        }
    }
}
```

### 3. ILP Extraction
Verify 4-wide actually issues 4 instructions when possible:
```go
func TestFourWideILP(t *testing.T) {
    // Create benchmark with 4 independent instructions
    program := BuildProgram(
        EncodeADDImm(0, 0, 1, false), // X0 += 1
        EncodeADDImm(1, 1, 1, false), // X1 += 1
        EncodeADDImm(2, 2, 1, false), // X2 += 1
        EncodeADDImm(3, 3, 1, false), // X3 += 1
        // Repeat pattern...
        EncodeSVC(0),
    )
    
    // With 40 independent ADDs + syscall
    // Ideal 4-wide: 10 cycles for ADDs + overhead
    // Should see CPI approaching 0.25 for the ADD block
}
```

### 4. Hazard Detection
Verify hazards are correctly detected across all 4 slots:
```go
func TestFourWideHazards(t *testing.T) {
    // RAW hazard across slots
    program := BuildProgram(
        EncodeADDImm(0, 0, 1, false), // X0 += 1 (slot 0)
        EncodeADDReg(1, 0, 0, false), // X1 = X0 + X0 (slot 1 depends on slot 0)
        EncodeADDReg(2, 1, 1, false), // X2 = X1 + X1 (slot 2 depends on slot 1)
        EncodeADDReg(3, 2, 2, false), // X3 = X2 + X2 (slot 3 depends on slot 2)
        EncodeSVC(0),
    )
    // These should NOT issue together due to dependencies
    // Expect serial execution, CPI near 1.0 for this block
}
```

### 5. Memory Port Contention
Only one memory operation per cycle:
```go
func TestFourWideMemoryPort(t *testing.T) {
    // Four loads should serialize
    program := BuildProgram(
        EncodeLDR64(0, 1, 0),
        EncodeLDR64(2, 1, 1),
        EncodeLDR64(4, 1, 2),
        EncodeLDR64(6, 1, 3),
        EncodeSVC(0),
    )
    // Only one load/cycle, should NOT improve from dual-issue
}
```

## Accuracy Comparison

### Against M2 Baseline

| Benchmark | M2 Real CPI | Current (dual) | Target (4-wide) | Notes |
|-----------|-------------|----------------|-----------------|-------|
| arithmetic | 0.268 | 0.700 (161% err) | ~0.350 (30% err) | Major improvement |
| dependency | 1.009 | 1.200 (19% err) | ~1.200 (19% err) | No change expected |
| branch | 1.190 | 1.400 (18% err) | ~1.400 (18% err) | Already optimized |

### CPI Targets by Benchmark

For 4-wide to be considered successful:

1. **arithmetic_sequential**: CPI ≤ 0.40 (vs 0.70 dual)
2. **loop_simulation**: CPI ≤ 0.70 (vs ~1.0 dual)
3. **matrix_operations**: CPI ≤ 0.60 (vs ~0.9 dual)
4. **dependency_chain**: CPI unchanged (~1.2)
5. **All tests pass** with correct exit codes

## Implementation Checklist

Before merging 4-wide PR:

- [ ] All existing tests pass
- [ ] No correctness regressions (exit codes match expected)
- [ ] No performance regressions (CPI not worse than dual)
- [ ] arithmetic_sequential shows significant improvement
- [ ] Hazard detection works across 4 slots
- [ ] Memory port limitation enforced
- [ ] Branch handling works correctly
- [ ] Forwarding works across all pipeline stages

## Architecture Notes for Bob

### Key Complexity Points

1. **Hazard Detection**: Must check dependencies between all 4 slots in decode:
   - Slot 1 vs Slot 0
   - Slot 2 vs Slots 0,1
   - Slot 3 vs Slots 0,1,2

2. **Forwarding Network**: Grows significantly:
   - Each execute slot needs forwarding from EX/MEM and MEM/WB for *all* slots
   - Same-cycle forwarding between slots executing together

3. **Branch Handling**: Branch in any slot should:
   - Flush all later slots in same bundle
   - Flush IF/ID pipeline registers
   - Work with prediction (already implemented for dual)

4. **Memory Port**: Only one memory op per cycle
   - Detect multiple loads/stores in decode
   - Issue only the first, stall others

### Suggested Incremental Approach

1. **Phase 1**: Extend fetch to 4 instructions
   - Add ifid3, ifid4 registers
   - Fetch 4 words per cycle (when no branches)

2. **Phase 2**: Extend decode to 4 slots
   - Add idex3, idex4 registers  
   - Implement canQuadIssue() with full dependency checking
   - Keep memory port constraint (only 1 mem op/cycle)

3. **Phase 3**: Extend execute
   - Add exmem3, exmem4 registers
   - Extend forwarding network
   - Handle branches in any slot

4. **Phase 4**: Extend memory/writeback
   - Add memwb3, memwb4 registers
   - Multiple writebacks per cycle

## Notes

- The partially-implemented 4-wide code in the codebase includes:
  - `TertiaryIFIDRegister`, `TertiaryIDEXRegister` (only slot 3)
  - `tickQuadIssue()` skeleton
  - `ifid3`, `idex3` pipeline registers
  
- Slot 4 structures still need to be added
- The validation tests above should be implemented in `benchmarks/timing_validation_test.go`
