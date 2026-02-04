# SPEC CPU 2017 Integration Plan

## Overview

This document outlines the plan to integrate SPEC CPU 2017 benchmarks with M2Sim for validation purposes. The SPEC suite is available at `/Users/yifan/Documents/spec` (version 1.1.9).

## Current State

- SPEC CPU 2017 kit is available but not installed
- M2Sim currently uses custom microbenchmarks in `benchmarks/`
- Validation shows 62.8% average CPI error vs target <2%

## Integration Steps

### Phase 1: SPEC Installation

1. **Install SPEC CPU 2017**
   ```bash
   cd /Users/yifan/Documents/spec
   ./install.sh -d ~/spec2017
   ```

2. **Build ARM64 binaries**
   - Need ARM64 cross-compiler or native build on ARM64 Mac
   - Focus on integer benchmarks (SPECint) initially
   - Start with small/fast benchmarks for quick iteration

### Phase 2: Benchmark Selection

Recommended starting benchmarks (by complexity):

| Benchmark | Description | Why Start Here |
|-----------|-------------|----------------|
| 602.gcc_s | GCC compiler | Branch-heavy, good for predictor validation |
| 619.lbm_s | Lattice Boltzmann | Memory-intensive, tests cache model |
| 623.xalancbmk_s | XML processing | Mixed workload |
| 631.deepsjeng_s | Chess AI | Control flow heavy |

### Phase 3: Harness Development

1. **Create SPEC harness** in `benchmarks/spec/`
   - Load ELF binaries compiled for ARM64
   - Set up memory with SPEC inputs
   - Run through M2Sim timing model
   - Compare vs real M2 execution time

2. **Timing measurement**
   ```go
   // benchmarks/spec/harness.go
   type SPECBenchmark struct {
       Name     string
       Binary   string   // Path to ARM64 binary
       Args     []string
       Input    string   // Input file path
       Expected struct {
           Cycles   uint64  // Expected cycles on real M2
           CPI      float64
       }
   }
   ```

### Phase 4: Validation Workflow

1. Run SPEC benchmark on real M2 Mac with performance counters
2. Run same benchmark through M2Sim
3. Compare:
   - Total cycles
   - CPI
   - IPC
   - Branch prediction accuracy
   - Cache hit rates

## Challenges

1. **Binary size**: SPEC binaries are large; may need to limit simulation
2. **Syscalls**: SPEC uses more syscalls than our microbenchmarks
3. **Memory**: Need larger memory model for realistic runs
4. **Time**: Full SPEC runs take hours on real hardware

## Short-term Alternative

Instead of full SPEC integration, we could:

1. Extract hot loops from SPEC benchmarks
2. Create standalone micro-kernels
3. Test these through existing harness

This gives SPEC-like code patterns without full integration complexity.

## Next Steps

1. [ ] Install SPEC to a working directory
2. [ ] Build one benchmark (602.gcc_s) for ARM64
3. [ ] Create minimal harness to run a single function
4. [ ] Measure real M2 timing for comparison
5. [ ] Iterate on accuracy

---

*Created by Bob during M2Sim cycle #126*
