# CoreMark M2Sim Port

Cross-compiled CoreMark for bare-metal ARM64 execution in M2Sim.

## Prerequisites

1. Clone CoreMark source:
   ```bash
   git clone https://github.com/eembc/coremark.git ../coremark
   ```

2. Install cross-compiler:
   ```bash
   brew install aarch64-elf-gcc
   ```

## Building

Prerequisites:
- `aarch64-elf-gcc` (install via `brew install aarch64-elf-gcc`)

Build:
```bash
./build.sh
```

Output:
- `coremark_m2sim.elf` - ARM64 ELF binary for M2Sim
- `coremark_m2sim.dis` - Disassembly for debugging

## Configuration

- **Iterations:** 10 (configurable via ITERATIONS)
- **Memory:** Static allocation (6000 bytes)
- **Timing:** Simple counter (M2Sim measures actual cycles)

## Running in M2Sim

**Status:** The ELF binary is correctly built, but M2Sim currently lacks support for
some instructions used by compiled code (ADRP, MOV as ORR alias, LDR literal).
See issue #156 for tracking instruction expansion.

Once M2Sim supports the required instructions:
```bash
# From M2Sim root directory
go run . --timing benchmarks/coremark/m2sim/coremark_m2sim.elf
```

## Files

| File | Description |
|------|-------------|
| `core_portme.h` | Port-specific header (types, config) |
| `core_portme.c` | Port-specific implementation (timing, memory) |
| `startup.S` | ARM64 bare-metal startup code |
| `linker.ld` | Linker script for bare-metal |
| `build.sh` | Build script |

## Notes

- No printf output (stubbed) - M2Sim measures cycles internally
- Uses WFI instruction to halt after completion
- Entry point at 0x80000
