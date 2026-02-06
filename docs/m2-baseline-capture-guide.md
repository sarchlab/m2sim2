# M2 Baseline Capture Guide

**Author:** Eric (AI Researcher)  
**Date:** 2026-02-06 (Cycle 277)  
**Purpose:** Document the process for capturing M2 hardware baselines for benchmark validation

## Overview

Per issue #141, M6 validation requires comparing M2Sim timing predictions against real M2 hardware measurements using **intermediate-size benchmarks** (NOT microbenchmarks).

**Current Status:** 15 benchmarks ready, 0 M2 baselines captured

## Step-by-Step Process

### Phase 1: Build Native Benchmarks

The current benchmarks are **bare-metal ELFs** for simulation. For M2 baselines, we need **native macOS executables**.

#### PolyBench Native Builds

```bash
cd benchmarks/polybench

# Build all PolyBench kernels for native execution
for kernel in gemm atax 2mm mvt jacobi-1d 3mm bicg; do
  clang -O2 -DPOLYBENCH_TIME \
    -I. $kernel/$kernel.c \
    -o ${kernel}_native
done
```

**Note:** May need to modify polybench.h for native timing:
```c
// Add for native builds
#ifdef POLYBENCH_TIME
#include <time.h>
static struct timespec _polybench_start, _polybench_end;
#define polybench_start_instruments clock_gettime(CLOCK_MONOTONIC, &_polybench_start)
#define polybench_stop_instruments clock_gettime(CLOCK_MONOTONIC, &_polybench_end)
#endif
```

#### Embench Native Builds

Embench benchmarks need header adjustments for native macOS:

```bash
cd benchmarks/

# For each Embench benchmark, create native build
# (May require adapting support code for stdio/stdlib)
```

### Phase 2: Capture Cycle Counts

#### Option A: Apple Instruments (xctrace)

```bash
# List available templates
xctrace list templates

# Record CPU counters for a benchmark run
xctrace record --template 'CPU Counters' \
  --output gemm.trace \
  -- ./gemm_native

# Extract cycle data
xctrace export --input gemm.trace --xpath '//row[@name="CYCLES"]' > gemm_cycles.xml
```

#### Option B: Performance Counter API

```c
// Use mach performance counters
#include <mach/mach.h>
#include <mach/thread_act.h>

// Get cycles using thread_info (user+system)
thread_basic_info_data_t info;
mach_msg_type_number_t count = THREAD_BASIC_INFO_COUNT;
thread_info(mach_thread_self(), THREAD_BASIC_INFO, (thread_info_t)&info, &count);
// info.user_time + info.system_time gives wall time
```

#### Option C: Wall Clock with High Iteration Count

For benchmarks without direct cycle access:

```bash
# Run with many iterations, divide by count
time -p ./gemm_native  # Repeat 1000 times, average
```

### Phase 3: Record Baseline Data

Create a baseline data file:

```yaml
# m2_baselines.yaml
hardware:
  chip: Apple M2
  cores: 4P + 4E
  frequency: 3.5 GHz (P-cores)
  
benchmarks:
  gemm:
    instructions: 37000  # From simulation
    cycles: ????  # Fill from hardware measurement
    cpi: ????  # cycles / instructions
    
  atax:
    instructions: 5000
    cycles: ????
    cpi: ????
    
  # ... repeat for all 15 benchmarks
```

### Phase 4: Calculate Accuracy

```python
# accuracy.py
def calculate_error(sim_cpi, m2_cpi):
    return abs(sim_cpi - m2_cpi) / m2_cpi * 100

# Example
benchmarks = {
    'gemm': {'sim_cpi': 0.25, 'm2_cpi': 0.27},  # 7.4% error
    'atax': {'sim_cpi': 0.30, 'm2_cpi': 0.28},  # 7.1% error
    # ...
}

avg_error = sum(calculate_error(b['sim_cpi'], b['m2_cpi']) 
                for b in benchmarks.values()) / len(benchmarks)
print(f"Average error: {avg_error:.1f}%")
```

## Data Needed from Human

| Benchmark | Metric Needed |
|-----------|---------------|
| gemm | Cycles for ~37K instructions |
| atax | Cycles for ~5K instructions |
| 2mm | Cycles for ~70K instructions |
| mvt | Cycles for ~5K instructions |
| jacobi-1d | Cycles for ~5.3K instructions |
| 3mm | Cycles for ~105K instructions |
| bicg | Cycles for ~4.8K instructions |
| aha-mont64 | Cycles (instruction count varies) |
| crc32 | Cycles |
| matmult-int | Cycles |
| primecount | Cycles |
| edn | Cycles for ~3.1M instructions |
| statemate | Cycles for ~1.04M instructions |
| huffbench | Cycles |

## Current Blocker

The agent team **cannot capture M2 baselines autonomously** because:

1. Requires physical access to M2 hardware
2. Requires native macOS builds (not bare-metal ELFs)
3. Requires Apple's performance counter APIs or Instruments
4. May require sudo/admin for some performance counters

**Estimated human effort:** 2-4 hours for all 15 benchmarks

## Fallback: Microbenchmark Extrapolation

If M2 baseline capture is delayed, we can estimate intermediate benchmark accuracy from microbenchmark results:

| Microbenchmark | Error | Notes |
|----------------|-------|-------|
| arithmetic_8wide | 7.2% | Best case (simple ops) |
| dependency_chain | 18.9% | With dependencies |
| branch_conditional | 34.5% | Worst case (folding disabled) |

**Weighted average:** ~20.2%

However, per #141, this does NOT satisfy M6 requirements.

## Success Criteria

M6 is complete when:
- [ ] ≥10 intermediate benchmarks have M2 baselines
- [ ] Average accuracy < 20% across those benchmarks
- [ ] Results documented in PROGRESS.md

---
*This guide supports M6 Validation milestone and issue #141.*
