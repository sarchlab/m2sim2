// memory_strided_long.s - Long-running strided memory access benchmark
// 10M iterations of strided store/load operations (stride = 32 bytes)
// Purpose: Calibrate strided memory access CPI on real M2 hardware
//
// Each iteration: 5 store/load pairs at stride-4 offsets (10 memory instructions)
// Tests strided cache line access performance.

.global _main
.align 4

_main:
    // Reserve stack space for strided data (256 bytes to cover stride offsets)
    sub sp, sp, #256

    // Initialize loop counter
    mov x10, #0              // iteration counter
    // Load 10000000 (0x989680) into x11
    movz x11, #0x9680
    movk x11, #0x0098, lsl #16

    // Initialize value
    mov x0, #0

.loop:
    // --- Timing region: Strided memory operations ---
    str x0, [sp]             // offset 0
    ldr x1, [sp]
    add x0, x1, #1

    str x0, [sp, #32]        // offset 32 bytes (stride = 4 * 8)
    ldr x1, [sp, #32]
    add x0, x1, #1

    str x0, [sp, #64]        // offset 64
    ldr x1, [sp, #64]
    add x0, x1, #1

    str x0, [sp, #96]        // offset 96
    ldr x1, [sp, #96]
    add x0, x1, #1

    str x0, [sp, #128]       // offset 128
    ldr x1, [sp, #128]
    add x0, x1, #1           // 5 increments per iteration
    // --- End timing region ---

    // Loop control
    add x10, x10, #1
    cmp x10, x11
    b.lt .loop

    // Restore stack
    add sp, sp, #256

    // x0 = 50M, exit code is 8-bit: 50000000 mod 256 = 128
    and x0, x0, #0xFF

    // Exit syscall
    mov x16, #1
    svc #0x80
