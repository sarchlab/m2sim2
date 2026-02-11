// vector_sum_long.s - Long-running vector sum benchmark
// 10M iterations of array element accumulation loop
// Purpose: Calibrate load+accumulate loop CPI on real M2 hardware
//
// Each iteration: 16 loads + 16 adds + loop overhead = ~35 instructions
// Tests: load throughput in accumulation pattern

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
    mov x10, #0              // iteration counter
    movz x11, #0x9680
    movk x11, #0x0098, lsl #16  // 10000000

.outer_loop:
    // --- Timing region: sum 16 array elements ---
    mov x0, #0              // sum = 0
    mov x1, sp              // array pointer
    mov x2, #0              // i = 0
    mov x3, #16             // N = 16

.inner_loop:
    ldr x4, [x1]            // load A[i]
    add x0, x0, x4          // sum += A[i]
    add x1, x1, #8          // ptr += 8
    add x2, x2, #1          // i++
    cmp x2, x3              // i < N?
    b.lt .inner_loop
    // --- End timing region ---

    // Loop control
    add x10, x10, #1
    cmp x10, x11
    b.lt .outer_loop

    // Restore stack
    add sp, sp, #256

    // x0 = 136 (sum of 1..16), exit code = 136 & 0xFF = 136
    and x0, x0, #0xFF

    // Exit syscall
    mov x16, #1
    svc #0x80
