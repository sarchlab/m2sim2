# EDN M2Sim Benchmark

Embench-IoT `edn` (embedded digital/signal processing) benchmark adapted for M2Sim bare-metal execution.

## Building

```bash
./build.sh
```

Requires `aarch64-elf-gcc` cross-compiler.

## Running

```bash
cd ../..
go run ./cmd/m2sim benchmarks/edn-m2sim/edn_m2sim.elf
```

## Benchmark Details

- **Type:** Embedded signal processing (vector multiply, FIR, IIR, DCT)
- **Source:** Embench-IoT
- **Complexity:** Medium
- **Focus:** Array operations, fixed-point arithmetic
