# PolyBench for M2Sim

Bare-metal adaptations of PolyBench/C benchmarks for M2Sim timing validation.

## Overview

This directory contains bare-metal versions of PolyBench kernels adapted for
the M2Sim simulator. The benchmarks use:

- Integer arithmetic (no FPU dependencies)
- Static array allocation (no malloc/libc)
- SMALL dataset sizes (60-120 element dimensions)
- Exit via syscall (bare-metal compatible)

## Benchmarks

### Phase 1 (Current)

| Benchmark | Description | Operations |
|-----------|-------------|------------|
| gemm | Matrix multiply C=αAB+βC | 16×16×16 = 4,096 MACs |

### Phase 2 (Planned)

| Benchmark | Description |
|-----------|-------------|
| atax | Matrix transpose + vector multiply |

## Building

Requires `aarch64-elf-gcc` cross-compiler.

```bash
# Build all benchmarks
./build.sh

# Build specific benchmark
./build.sh gemm

# Clean build artifacts
./build.sh clean
```

## Output Files

- `gemm_m2sim.elf` - Bare-metal ELF binary
- `gemm_m2sim.dis` - Disassembly listing

## Running with M2Sim

```bash
# From repository root
go test -run TestPolybenchGemm -v ./benchmarks/
```

## References

- [PolyBench/C 4.2.1](https://github.com/MatthiasJReisinger/PolyBenchC-4.2.1)
- Issue #237: PolyBench Phase 1
- `docs/polybench-implementation-approach.md`
