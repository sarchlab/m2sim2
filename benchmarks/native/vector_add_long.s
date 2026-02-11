// vector_add_long.s - Long-running vector add benchmark
// 10M iterations of C[i] = A[i] + B[i] loop
// Purpose: Calibrate dual-load+store loop CPI on real M2 hardware
//
// Each iteration: inner loop of 16 elements with 2 loads + 1 add + 1 store
// Tests: load+store throughput in compute loop

.global _main
.align 4

_main:
    // Allocate arrays on stack: A, B, C each 128 bytes (16 * 8)
    sub sp, sp, #512

    // A starts at sp+0, B at sp+128, C at sp+256

    // Initialize A[i] = i + 1
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

    // Initialize B[i] = 2 * (i + 1)
    mov x0, #2
    str x0, [sp, #128]
    mov x0, #4
    str x0, [sp, #136]
    mov x0, #6
    str x0, [sp, #144]
    mov x0, #8
    str x0, [sp, #152]
    mov x0, #10
    str x0, [sp, #160]
    mov x0, #12
    str x0, [sp, #168]
    mov x0, #14
    str x0, [sp, #176]
    mov x0, #16
    str x0, [sp, #184]
    mov x0, #18
    str x0, [sp, #192]
    mov x0, #20
    str x0, [sp, #200]
    mov x0, #22
    str x0, [sp, #208]
    mov x0, #24
    str x0, [sp, #216]
    mov x0, #26
    str x0, [sp, #224]
    mov x0, #28
    str x0, [sp, #232]
    mov x0, #30
    str x0, [sp, #240]
    mov x0, #32
    str x0, [sp, #248]

    // Initialize outer loop counter
    mov x10, #0              // iteration counter
    movz x11, #0x9680
    movk x11, #0x0098, lsl #16  // 10000000

.outer_loop:
    // --- Timing region: C[i] = A[i] + B[i] ---
    add x1, sp, #0          // A ptr
    add x2, sp, #128        // B ptr
    add x3, sp, #256        // C ptr
    mov x4, #0              // i = 0
    mov x5, #16             // N = 16

.inner_loop:
    ldr x6, [x1]            // load A[i]
    ldr x7, [x2]            // load B[i]
    add x9, x6, x7          // A[i] + B[i]
    str x9, [x3]            // store C[i]
    add x1, x1, #8          // A ptr++
    add x2, x2, #8          // B ptr++
    add x3, x3, #8          // C ptr++
    add x4, x4, #1          // i++
    cmp x4, x5              // i < N?
    b.lt .inner_loop
    // --- End timing region ---

    // Loop control
    add x10, x10, #1
    cmp x10, x11
    b.lt .outer_loop

    // Load C[0] as result
    ldr x0, [sp, #256]      // C[0] = A[0]+B[0] = 1+2 = 3

    // Restore stack
    add sp, sp, #512

    // Exit with C[0] = 3
    and x0, x0, #0xFF

    // Exit syscall
    mov x16, #1
    svc #0x80
