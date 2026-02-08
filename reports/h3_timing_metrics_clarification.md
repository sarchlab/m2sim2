# H3 Timing Metrics Clarification: Virtual Time vs Simulation Time

**Author**: Alex
**Date**: February 8, 2026
**Purpose**: Resolve confusion between simulation time and virtual time metrics for accurate calibration analysis

## Executive Summary

This document clarifies the critical distinction between simulation time and virtual time metrics in M2Sim timing analysis, addressing confusion that could undermine H3 calibration validation.

## Key Definitions

### Virtual Time (Primary Calibration Target)
- **Definition**: Simulated hardware execution time - the time it would take real hardware to execute the program
- **Units**: CPU cycles, nanoseconds of simulated hardware time
- **Examples**:
  - CPI (Cycles Per Instruction): 1.2 cycles per instruction
  - Virtual execution time: 0.077ns per instruction (at 3.2GHz simulated frequency)
- **Purpose**: Accuracy target for timing simulation calibration against real hardware

### Simulation Time (Performance Diagnostic)
- **Definition**: Real-time execution speed of the simulator itself
- **Units**: Wall-clock time, instructions per second throughput
- **Examples**:
  - Simulation speed: 139-144ns real time per simulated instruction
  - Throughput: 7M instructions per second
- **Purpose**: Performance optimization and benchmarking of simulator implementation

## Current Analysis Results Context

### CPI Comparison Results (Virtual Time)
The `cpi_comparison_results.json` contains **virtual time metrics**:
- `full_pipeline_cpi: 1.8` = simulated hardware executes this benchmark at 1.8 cycles per instruction
- `fast_timing_cpi: 1.2` = fast timing model predicts 1.2 cycles per instruction
- `divergence_pct: -34.5%` = fast timing underestimates virtual time by 34.5%

### Previous Confusion
Earlier analysis incorrectly mixed:
- **139-144ns per instruction** → simulation speed (real-time performance)
- **0.077ns per instruction** → virtual time (simulated hardware timing)

These represent completely different measurements and cannot be directly compared.

## Calibration Implications

### Virtual Time Accuracy (Primary Goal)
- **Target**: Virtual time should match real M2 hardware baseline within ~15% error
- **Current Status**: Average error 22.8% (arithmetic 35.2%, dependency 10.3%, branch 22.7%)
- **Measurement**: Compare simulated CPI against real hardware CPI measurements

### Simulation Speed (Secondary Goal)
- **Target**: Optimize simulator performance for faster calibration iterations
- **Current Status**: Fast timing provides ~10x speedup over full pipeline
- **Measurement**: Instructions per second throughput, wall-clock execution time

## Framework Requirements

### Analysis Pipeline Updates
1. **Clear metric labeling**: All outputs must specify virtual vs simulation time
2. **Separate analysis tracks**: Virtual time accuracy vs simulation performance
3. **Calibration focus**: Virtual time is the primary calibration validation metric

### Output Format Standards
```json
{
  "virtual_time": {
    "cpi": 1.2,
    "frequency_ghz": 3.2,
    "execution_time_ns": 0.375
  },
  "simulation_performance": {
    "wall_time_per_instruction_ns": 144,
    "throughput_mips": 6.94
  }
}
```

## Action Items

### Immediate (1-2 cycles)
- [x] Document virtual vs simulation time distinction
- [ ] Update analysis scripts to correctly parse virtual time
- [ ] Standardize output formats with clear metric labeling

### Short-term (2-3 cycles)
- [ ] Validate virtual time against real M2 hardware measurements
- [ ] Update calibration framework to focus on virtual time accuracy
- [ ] Create simulation performance tracking as separate diagnostic

### Medium-term (3-4 cycles)
- [ ] Integrate clarified metrics into fast timing analysis framework
- [ ] Update all documentation and reports with corrected terminology
- [ ] Train team on proper metric interpretation

## Success Criteria

- [ ] No confusion between virtual time and simulation time in any analysis
- [ ] Virtual time accuracy correctly measured against hardware baseline
- [ ] Calibration targets virtual time precision within 15% error range
- [ ] Simulation performance tracked separately for optimization

## Strategic Impact

Correct virtual time analysis is essential for:
- **H3 calibration validation**: Ensuring simulator accuracy matches real hardware
- **Fast timing validation**: Confirming fast timing maintains virtual time precision
- **Parameter tuning**: Optimizing simulation parameters for hardware accuracy

This clarification enables confident deployment of calibrated timing simulation for H3.3 advanced capabilities.