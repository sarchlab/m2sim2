// memory_random_long.s - Long-running random memory access benchmark
// 10M iterations of pseudo-random store/load operations
// Purpose: Calibrate random memory access CPI on real M2 hardware
//
// Each iteration: 5 store/load pairs at scattered offsets (10 memory instructions)
// Uses a large buffer (4096 bytes) with non-sequential access pattern
// to stress cache line utilization.

.global _main
.align 4

_main:
    // Reserve a 4096-byte buffer on stack for random access
    sub sp, sp, #4096

    // Initialize loop counter
    mov x10, #0              // iteration counter
    // Load 10000000 (0x989680) into x11
    movz x11, #0x9680
    movk x11, #0x0098, lsl #16

    // Initialize value
    mov x0, #0

.loop:
    // --- Timing region: Random (scattered) memory operations ---
    // Offsets chosen to scatter across cache lines:
    // 0, 2048, 512, 3072, 1536 (each at least 64 bytes apart = different cache lines)
    str x0, [sp]             // offset 0
    ldr x1, [sp]
    add x0, x1, #1

    str x0, [sp, #2048]      // offset 2048
    ldr x1, [sp, #2048]
    add x0, x1, #1

    str x0, [sp, #512]       // offset 512
    ldr x1, [sp, #512]
    add x0, x1, #1

    str x0, [sp, #3072]      // offset 3072
    ldr x1, [sp, #3072]
    add x0, x1, #1

    str x0, [sp, #1536]      // offset 1536
    ldr x1, [sp, #1536]
    add x0, x1, #1           // 5 increments per iteration
    // --- End timing region ---

    // Loop control
    add x10, x10, #1
    cmp x10, x11
    b.lt .loop

    // Restore stack
    add sp, sp, #4096

    // x0 = 50M, exit code is 8-bit: 50000000 mod 256 = 128
    and x0, x0, #0xFF

    // Exit syscall
    mov x16, #1
    svc #0x80
