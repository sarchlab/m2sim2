# Embench-IoT Build Guide for M2Sim

*Research by Eric | 2026-02-04*

## Overview

Embench-IoT benchmarks provide intermediate-scale workloads for validating M2Sim timing accuracy. This guide documents how to build them for bare-metal ARM64 execution.

## Repository

```bash
git clone https://github.com/embench/embench-iot.git
```

## Key Benchmarks (Phase 2)

| Benchmark | Type | Source |
|-----------|------|--------|
| aha-mont64 | Integer ALU | Montgomery multiplication |
| crc32 | Bit manipulation | Cyclic redundancy check |
| matmult-int | Memory + ALU | Integer matrix multiplication |

## Build Requirements

Same toolchain as CoreMark:
- `aarch64-elf-gcc` (bare-metal cross-compiler)
- No libc dependency (standalone execution)

## Adaptation Strategy

Each Embench benchmark needs these adaptations:

### 1. Create `support.h` replacement

Embench benchmarks use `support.h` which defines:
- `CPU_MHZ` - clock frequency (unused for M2Sim, set to 1)
- `benchmark()` - entry point
- `initialise_benchmark()` - setup
- `verify_benchmark()` - validation

For M2Sim, create minimal support header:

```c
/* m2sim_support.h */
#define CPU_MHZ 1

void initialise_benchmark(void);
int benchmark(void);
int verify_benchmark(int r);
```

### 2. Create `startup.S`

Same pattern as CoreMark:
```asm
.section .text.startup
.global _start
.type _start, %function

_start:
    adrp x0, __stack_top
    add x0, x0, :lo12:__stack_top
    mov sp, x0
    
    /* Clear BSS */
    adrp x0, __bss_start
    add x0, x0, :lo12:__bss_start
    adrp x1, __bss_end
    add x1, x1, :lo12:__bss_end
    mov x2, #0
1:
    cmp x0, x1
    b.ge 2f
    str x2, [x0], #8
    b 1b
2:
    
    bl initialise_benchmark
    bl benchmark
    
    /* Store result and halt */
3:
    wfi
    b 3b
```

### 3. Create `linker.ld`

Same as CoreMark:
```ld
ENTRY(_start)

SECTIONS {
    . = 0x80000;
    .text : { *(.text.startup) *(.text*) }
    .rodata : { *(.rodata*) }
    .data : { *(.data*) }
    .bss : { __bss_start = .; *(.bss*) *(COMMON) __bss_end = .; }
    . = ALIGN(16);
    __stack_top = . + 0x10000;
}
```

### 4. Build command

```bash
aarch64-elf-gcc -O2 -nostdlib -nostartfiles -ffreestanding \
    -march=armv8-a -mabi=lp64 \
    -I. \
    -T linker.ld \
    startup.S benchmark.c -o benchmark.elf
```

## aha-mont64 Specifics

**Source files:** `mont64.c`

**Dependencies:** 
- Uses `__int128` if available (for 128-bit multiply)
- No stdlib needed (pure integer math)

**Modifications needed:**
1. Remove `#include <stdlib.h>` (unused after modification)
2. Remove `#include <stdio.h>` (unused)
3. Replace `#include "support.h"` with `#include "m2sim_support.h"`

**Workload:** Montgomery modular multiplication - tests integer ALU throughput.

## crc32 Specifics

**Source files:** `crc_32.c`

**Dependencies:** Minimal, pure integer operations

**Workload:** CRC calculation - tests bit manipulation instructions.

## matmult-int Specifics

**Source files:** `matmult-int.c`

**Dependencies:** Minimal arrays

**Workload:** Integer matrix multiplication - tests memory access patterns + ALU.

## Integration with M2Sim

Once built, run with:
```go
builder := cpu.MakeBuilder().
    WithELFBinary("benchmarks/embench/aha-mont64/aha-mont64.elf")
```

## Priority

Build order (based on complexity and value):
1. **aha-mont64** - Pure ALU, smallest, validates integer timing
2. **crc32** - Bit manipulation, good for testing shifts/xor
3. **matmult-int** - Memory patterns, validates cache model

## Notes

- LOCAL_SCALE_FACTOR controls iteration count (affects runtime)
- Reduce scale factor for faster testing during development
- M2Sim WFI instruction terminates execution (same as CoreMark)
