# M2Sim Validation Baseline

This document captures the validation state of M2Sim before M3 (Timing Model) changes.

## Validation Date
2026-02-02 (Updated by Ethan validation - Issue #35)

## Milestone Status
- **M1: Foundation (MVP)** ✅ Complete
- **M2: Memory & Control Flow** ✅ Complete
- **M3: Timing Model** 🚧 Pending this validation

## Test Suite Results

### Unit Tests Summary

| Package | Tests | Status |
|---------|-------|--------|
| emu | 200 | ✅ PASS |
| driver | 17 | ✅ PASS |
| insts | 46 | ✅ PASS |
| loader | 17 | ✅ PASS |
| benchmarks | 14 | ✅ PASS |
| timing/cache | - | ✅ PASS |
| timing/mem | - | ✅ PASS |
| timing/core | - | ⚠️ Build failed (WIP - M3) |
| timing/pipeline | 94/99 | ⚠️ WIP (M3 timing work) |

**Total Functional Emulator Tests: 294 passing**

### Ethan Validation Suite Results

The following validation programs have been verified:

| Program | Description | Exit Code | Output | Instructions | Status |
|---------|-------------|-----------|--------|--------------|--------|
| simple_exit | Basic program termination | 42 | - | 3 | ✅ PASS |
| arithmetic | ADD operations (10+5) | 15 | - | 5 | ✅ PASS |
| subtraction | SUB operations (100-58) | 42 | - | 4 | ✅ PASS |
| loop | Conditional branch loop (3→0) | 0 | - | 9 | ✅ PASS |
| loop_sum | Sum 1+2+3+4+5 | 15 | - | 20 | ✅ PASS |
| hello | Write syscall | 0 | "Hello\n" | 8 | ✅ PASS |
| function_call | BL/RET subroutine | 15 | - | 6 | ✅ PASS |
| nested_calls | Nested BL/RET | 25 | - | 12 | ✅ PASS |
| logical_ops | AND/ORR/EOR | 255 | - | 7 | ✅ PASS |
| memory_ops | LDR/STR | 77 | - | 7 | ✅ PASS |
| cond_branch_eq | B.EQ condition | 5 | - | 5 | ✅ PASS |
| cond_branch_gt | B.GT condition | 10 | - | 5 | ✅ PASS |

**All 12 Ethan validation programs PASSED**

## Functional Emulator Capabilities

### Supported Instructions
- **ALU (Immediate)**: ADD, ADDS, SUB, SUBS
- **ALU (Register)**: ADD, ADDS, SUB, SUBS, AND, ANDS, ORR, EOR
- **Branch**: B, BL, B.cond (all conditions), BR, BLR, RET
- **Load/Store**: LDR (32/64-bit), STR (32/64-bit)
- **System**: SVC (syscall)

### Supported Syscalls
- **exit (93)**: Program termination with exit code
- **write (64)**: Write to file descriptor (stdout/stderr)

### Known Limitations
1. N-bit not handled in logical register instructions (BIC, ORN, EON)
2. No shifted register operands for ALU instructions
3. No SIMD/floating-point support
4. Pre-index/post-index addressing modes limited

## Regression Baseline

This baseline establishes the expected behavior before M3 timing changes:

```
=====================================
M2Sim Regression Baseline v1.0
Validated: 2026-02-02
=====================================

Packages:
  emu:        200 tests PASS
  driver:      17 tests PASS
  insts:       46 tests PASS
  loader:      17 tests PASS
  benchmarks:  14 tests PASS
  ---------------------------------
  Total:      294 tests PASS

Validation Programs:
  simple_exit:    exit(42)           → 42
  arithmetic:     10 + 5             → 15
  subtraction:    100 - 58           → 42
  loop:           count down 3→0     → 0
  loop_sum:       1+2+3+4+5          → 15
  hello:          write 'Hello\n'    → 0
  function_call:  BL/RET             → 15
  nested_calls:   nested BL/RET      → 25
  logical_ops:    AND/ORR/EOR        → 255
  memory_ops:     LDR/STR            → 77
  cond_branch_eq: B.EQ               → 5
  cond_branch_gt: B.GT               → 10

=====================================
```

## Notes for Timing Model Integration

When M3 timing changes are integrated:
1. **All existing unit tests must continue to pass**
2. **All validation programs must produce identical results**
3. **New timing-specific tests should be added**
4. Performance metrics can differ (that's expected)

### Commands to Verify Baseline

```bash
# Run all functional emulator tests
go test ./emu/... ./driver/... ./insts/... ./loader/... ./benchmarks/...

# Run only Ethan validation suite
go test ./emu/... -v 2>&1 | grep -A20 "Ethan Validation"
```

## Sign-off

- [x] **Bob**: Validation infrastructure created (Issue #35)
- [x] **Ethan**: Test programs validated
- [ ] **Alice**: Approved for M3 timing work

## Validation Test File

The Ethan validation suite is implemented in:
- `emu/ethan_validation_test.go` - Comprehensive validation tests with result tracking

To run the validation suite:
```bash
cd ~/dev/src/github.com/sarchlab/m2sim
go test ./emu/... -v 2>&1 | tail -50
```
