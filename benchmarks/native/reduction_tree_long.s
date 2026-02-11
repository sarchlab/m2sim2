// reduction_tree_long.s - Long-running parallel reduction benchmark
// 10M iterations of 16-element tree reduction
// Purpose: Calibrate ILP-heavy reduction CPI on real M2 hardware
//
// Each iteration: 16 loads + 15 adds (4 levels) = 31 instructions
// Tests: instruction-level parallelism in reduction pattern

.global _main
.align 4

_main:
    // Allocate array on stack (16 * 8 = 128 bytes)
    sub sp, sp, #256

    // Initialize array: A[i] = i + 1
    mov x0, #1
    str x0, [sp, #0]
    mov x0, #2
    str x0, [sp, #8]
    mov x0, #3
    str x0, [sp, #16]
    mov x0, #4
    str x0, [sp, #24]
    mov x0, #5
    str x0, [sp, #32]
    mov x0, #6
    str x0, [sp, #40]
    mov x0, #7
    str x0, [sp, #48]
    mov x0, #8
    str x0, [sp, #56]
    mov x0, #9
    str x0, [sp, #64]
    mov x0, #10
    str x0, [sp, #72]
    mov x0, #11
    str x0, [sp, #80]
    mov x0, #12
    str x0, [sp, #88]
    mov x0, #13
    str x0, [sp, #96]
    mov x0, #14
    str x0, [sp, #104]
    mov x0, #15
    str x0, [sp, #112]
    mov x0, #16
    str x0, [sp, #120]

    // Initialize outer loop counter
    mov x20, #0              // iteration counter
    movz x21, #0x9680
    movk x21, #0x0098, lsl #16  // 10000000

.loop:
    // --- Timing region: 16-element tree reduction ---
    // Load all 16 elements
    ldr x0, [sp, #0]
    ldr x2, [sp, #8]
    ldr x3, [sp, #16]
    ldr x4, [sp, #24]
    ldr x5, [sp, #32]
    ldr x6, [sp, #40]
    ldr x7, [sp, #48]
    ldr x9, [sp, #56]
    ldr x10, [sp, #64]
    ldr x11, [sp, #72]
    ldr x12, [sp, #80]
    ldr x13, [sp, #88]
    ldr x14, [sp, #96]
    ldr x15, [sp, #104]
    ldr x16, [sp, #112]
    ldr x17, [sp, #120]

    // Level 1: 8 independent pairwise sums
    add x0, x0, x2          // 1+2 = 3
    add x3, x3, x4          // 3+4 = 7
    add x5, x5, x6          // 5+6 = 11
    add x7, x7, x9          // 7+8 = 15
    add x10, x10, x11       // 9+10 = 19
    add x12, x12, x13       // 11+12 = 23
    add x14, x14, x15       // 13+14 = 27
    add x16, x16, x17       // 15+16 = 31

    // Level 2: 4 independent sums
    add x0, x0, x3          // 3+7 = 10
    add x5, x5, x7          // 11+15 = 26
    add x10, x10, x12       // 19+23 = 42
    add x14, x14, x16       // 27+31 = 58

    // Level 3: 2 independent sums
    add x0, x0, x5          // 10+26 = 36
    add x10, x10, x14       // 42+58 = 100

    // Level 4: final sum
    add x0, x0, x10         // 36+100 = 136
    // --- End timing region ---

    // Loop control
    add x20, x20, #1
    cmp x20, x21
    b.lt .loop

    // Restore stack
    add sp, sp, #256

    // x0 = 136 each iteration, exit code = 136 & 0xFF = 136
    and x0, x0, #0xFF

    // Exit syscall
    mov x16, #1
    svc #0x80
