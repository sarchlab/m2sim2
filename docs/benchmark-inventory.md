# M2Sim Benchmark Inventory

**Author:** Eric (AI Researcher)  
**Date:** 2026-02-05  
**Purpose:** Track all available intermediate benchmarks for M6 validation

## Summary

Per Issue #141, microbenchmarks do NOT count for accuracy validation. We need intermediate-size benchmarks.

| Suite | Ready | Pending | Notes |
|-------|-------|---------|-------|
| PolyBench | 2 | Many more available | gemm, atax ready |
| Embench-IoT | 4 | 1 (edn needs build) | Good diversity |
| CoreMark | 1 | - | Impractical (50M+ instr) |
| **Total** | **7** | **1** | Target: 15+ for publication |

## Ready Benchmarks (with ELFs)

### PolyBench/C

| Benchmark | Location | Instructions | Status |
|-----------|----------|--------------|--------|
| gemm | benchmarks/polybench/gemm_m2sim.elf | ~37K | ✅ Ready |
| atax | benchmarks/polybench/atax_m2sim.elf | ~5K | ✅ Ready |

### Embench-IoT

| Benchmark | Location | Workload Type | Status |
|-----------|----------|---------------|--------|
| aha-mont64 | benchmarks/aha-mont64-m2sim/ | Modular arithmetic | ✅ Ready |
| crc32 | benchmarks/crc32-m2sim/ | Checksum/bit ops | ✅ Ready |
| matmult-int | benchmarks/matmult-int-m2sim/ | Matrix multiply | ✅ Ready |
| primecount | benchmarks/primecount-m2sim/ | Integer math | ✅ Ready |

### CoreMark

| Benchmark | Location | Instructions | Status |
|-----------|----------|--------------|--------|
| coremark | benchmarks/coremark-m2sim/ | >50M/iteration | ⚠️ Impractical |

## Pending Benchmarks

### Embench-IoT (Need Build)

| Benchmark | Location | Issue | Notes |
|-----------|----------|-------|-------|
| edn | benchmarks/edn-m2sim/ | Needed | Has sources, no ELF |

### Embench-IoT Phase 2 (Alice Approved)

Per Issue #183, these are approved for implementation:

| Benchmark | Workload Type | Priority |
|-----------|---------------|----------|
| huffbench | Huffman coding | High |
| statemate | State machine | High |

### PolyBench Expansion

Additional PolyBench kernels could be added:

| Benchmark | Type | Complexity |
|-----------|------|------------|
| 2mm | Matrix multiply chain | Medium |
| 3mm | Matrix multiply chain | Medium |
| mvt | Matrix-vector ops | Low |
| bicg | Bi-conjugate gradient | Medium |
| doitgen | Multi-resolution | Medium |

## Publication Gap Analysis

| Metric | Current | Target | Gap |
|--------|---------|--------|-----|
| Ready benchmarks | 6 | 15+ | +9 needed |
| M2 baselines | 0 | 15+ | Requires human |
| Accuracy measured | 0 | 15+ | Awaiting baselines |

## Recommended Priorities

### Immediate (Bob)
1. Build edn ELF (low effort, sources exist)
2. Add 2-3 more PolyBench kernels (2mm, mvt, bicg)

### Requires Human
1. Capture M2 baselines for gemm, atax
2. Decision on accepting intermediate benchmarks for M6

### Medium-term
1. Add huffbench, statemate from Embench
2. Reach 15+ benchmarks for publication credibility

## Verification Commands

```bash
# List all ready ELFs
for elf in benchmarks/*/*m2sim.elf benchmarks/polybench/*m2sim.elf; do
  echo "$(basename $elf): $(du -h $elf | cut -f1)"
done

# Test a benchmark
go run cmd/m2sim/main.go benchmarks/polybench/gemm_m2sim.elf
```

---
*This inventory supports Issue #141 (intermediate benchmark requirement) and Issue #240 (publication readiness).*
