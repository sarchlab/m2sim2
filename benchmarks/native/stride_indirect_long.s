// stride_indirect_long.s - Long-running pointer chase benchmark
// 10M iterations of 8-hop pointer chase through array
// Purpose: Calibrate dependent-load (pointer chase) CPI on real M2 hardware
//
// Each iteration: 8 dependent loads (each load provides address for next)
// Tests: load-to-use latency chain (serialized memory access)

.global _main
.align 4

_main:
    // Allocate array on stack (8 * 8 = 64 bytes for index array)
    sub sp, sp, #128

    // Build pointer chase chain: 0→3→1→5→2→7→4→6→0
    mov x0, #3
    str x0, [sp, #0]       // A[0] = 3
    mov x0, #5
    str x0, [sp, #8]       // A[1] = 5
    mov x0, #7
    str x0, [sp, #16]      // A[2] = 7
    mov x0, #1
    str x0, [sp, #24]      // A[3] = 1
    mov x0, #6
    str x0, [sp, #32]      // A[4] = 6
    mov x0, #2
    str x0, [sp, #40]      // A[5] = 2
    mov x0, #0
    str x0, [sp, #48]      // A[6] = 0
    mov x0, #4
    str x0, [sp, #56]      // A[7] = 4

    // Initialize outer loop counter
    mov x10, #0              // iteration counter
    movz x11, #0x9680
    movk x11, #0x0098, lsl #16  // 10000000

.outer_loop:
    // --- Timing region: 8-hop pointer chase ---
    mov x2, #0              // start at index 0
    mov x3, #0              // hop counter

.chase_loop:
    // Compute address: base + index * 8
    lsl x5, x2, #3          // x5 = index * 8
    add x5, sp, x5          // x5 = base + offset
    ldr x2, [x5]            // x2 = A[index] (next index)
    add x3, x3, #1          // hop++
    cmp x3, #8              // 8 hops?
    b.lt .chase_loop
    // --- End timing region ---

    // Loop control
    add x10, x10, #1
    cmp x10, x11
    b.lt .outer_loop

    // Restore stack
    add sp, sp, #128

    // Exit with hop count = 8
    mov x0, x3
    and x0, x0, #0xFF

    // Exit syscall
    mov x16, #1
    svc #0x80
