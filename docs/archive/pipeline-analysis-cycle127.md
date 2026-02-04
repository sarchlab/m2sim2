# Pipeline Performance Analysis - Cycle #127

## Current State

| Benchmark | Cycles | Instructions | CPI | IPC |
|-----------|--------|--------------|-----|-----|
| arithmetic_sequential (5 regs) | 8 | 20 | 0.400 | 2.50 |
| arithmetic_6wide (6 regs) | 8 | 24 | 0.333 | 3.00 |

Real M2 achieves **0.268 CPI (3.7 IPC)** on similar arithmetic workloads.

## Analysis

### Why the Pipeline Fill/Drain Overhead Matters

Both benchmarks execute in 8 cycles due to **pipeline fill/drain overhead**:
- 5-stage pipeline: IF → ID → EX → MEM → WB
- First instruction retires at cycle 5
- Steady-state throughput only achieved after pipeline fills

**arithmetic_sequential (5 registers):**
- Uses X0-X4, each register repeats every 5 instructions
- RAW hazard between inst 0 (writes X0) and inst 5 (reads X0) in same fetch window
- Effective issue limited to 5/cycle within each fetch window
- 20 instructions / (4 issue cycles + 4 drain cycles) = 2.5 apparent IPC

**arithmetic_6wide (6 registers):**
- Uses X0-X5, 6 independent instructions per group
- All 6 can issue together within each cycle
- 24 instructions / (4 issue cycles + 4 drain cycles) = 3.0 apparent IPC

### Steady-State vs Apparent IPC

The 3.0 apparent IPC includes pipeline overhead. **Steady-state throughput is 6 IPC** (6 instructions issued per cycle). For short benchmarks, the 4-cycle pipeline drain dominates.

To measure true throughput, need longer benchmarks:
- 120 instructions: 20 issue + 4 drain = 24 cycles → 5.0 apparent IPC
- 600 instructions: 100 issue + 4 drain = 104 cycles → 5.77 apparent IPC

### Why M2 Achieves Higher IPC (3.7)

1. **Out-of-Order Execution**: M2's ~630-entry ROB can overlap pipeline fill/drain with execution
2. **7 Integer ALUs**: M2 has 7 ALUs vs our 6 (but limited by RAW hazards anyway)
3. **8-wide Decode**: M2 decodes 8 instructions/cycle vs our 6
4. **Deeper Pipeline**: More stages means more instructions in-flight during steady state
5. **Register Renaming**: Eliminates false WAW/WAR dependencies

### Recommendations

1. ✅ **Added arithmetic_6wide benchmark** - properly tests 6-wide issue capability
2. **Create longer benchmarks** to measure steady-state throughput accurately
3. **Consider OoO model** for M6 accuracy target (significant undertaking)
4. **Accept in-order limits** - 6 IPC steady-state is theoretical maximum

## Changes Made

1. Added `arithmetic_6wide` benchmark to microbenchmarks.go
2. Uses 6 distinct registers (X0-X5) for proper 6-wide testing
3. 24 instructions in 4 groups of 6

---
*Analysis by Bob during M2Sim cycle #127*
