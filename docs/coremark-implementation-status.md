# CoreMark Implementation Status

**Author:** Eric (AI Researcher)  
**Date:** 2026-02-05  
**Related:** Task board research item

## Current State

### What Exists

A CoreMark port for M2Sim already exists at `benchmarks/coremark/m2sim/`:

| Component | Status | Notes |
|-----------|--------|-------|
| `core_portme.h` | ✅ Complete | ARM64 types, bare-metal config |
| `core_portme.c` | ✅ Complete | Timing, memory stubs |
| `startup.S` | ✅ Complete | ARM64 bare-metal startup |
| `linker.ld` | ✅ Complete | Entry at 0x80000 |
| `build.sh` | ✅ Complete | Cross-compile script |
| `coremark_m2sim.elf` | ✅ Built | 81KB ELF binary |

### Configuration

- **Iterations:** 10 (configurable via ITERATIONS)
- **Memory:** Static allocation (6000 bytes)
- **Compiler:** aarch64-elf-gcc 15.2.0
- **Flags:** -O2 -static -nostdlib -ffreestanding

## Blocker

Per the README in `benchmarks/coremark/m2sim/`:

> **Status:** The ELF binary is correctly built, but M2Sim currently lacks support for some instructions used by compiled code (ADRP, MOV as ORR alias, LDR literal).

### Missing Instructions

| Instruction | Usage | Priority |
|-------------|-------|----------|
| ADRP | Address computation | High |
| MOV (ORR alias) | Data movement | High |
| LDR (literal) | PC-relative loads | High |

These are common ARM64 instructions used by compiled C code. Without them, CoreMark cannot execute.

## Next Steps

### Option 1: Add Missing Instructions (Recommended)

Bob should implement the missing instructions:

1. **ADRP** - Add Page (PC-relative)
   - Forms PC-relative address to 4KB page
   - Common in position-independent code

2. **MOV as ORR** - Move register
   - `MOV Xd, Xn` encoded as `ORR Xd, XZR, Xn`
   - May already work if ORR is complete

3. **LDR (literal)** - Load from PC-relative offset
   - Loads 32/64-bit value from code section
   - Essential for constants

### Option 2: Compiler Flags (Workaround)

Try different compiler flags to avoid problematic instructions:
```bash
-fno-pic -fno-pie -mcmodel=large
```

This may eliminate ADRP usage but could introduce other issues.

## Verification

Once instructions are added:

```bash
cd ~/dev/src/github.com/sarchlab/m2sim
go run cmd/m2sim/main.go -timing benchmarks/coremark/m2sim/coremark_m2sim.elf
```

Expected output:
- Exit code 0
- ~10 iterations completed
- Cycle count for CPI calculation

## M2 Baseline Reference

For comparison, Apple M2 CoreMark score:
- **35,120 iterations/sec** (single-threaded, per previous research)

This gives us a target to validate against once simulator execution works.

## Recommendations

1. **Create concrete issue** for missing instruction implementation
2. **Prioritize ADRP** — most commonly used for address computation
3. **Test incrementally** — try running after each instruction added
4. **Document any additional missing instructions** as they're discovered

---
*This document addresses the task board item: "Research CoreMark implementation approach for industry-standard validation"*
