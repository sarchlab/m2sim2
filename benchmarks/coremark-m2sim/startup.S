/*
 * M2Sim CoreMark Startup Code
 * Minimal bare-metal ARM64 startup
 * Avoids LDR literal (PC-relative) which M2Sim doesn't support yet
 */

.section .text.startup
.global _start
.type _start, %function

_start:
    /* Set up stack pointer using ADRP + ADD (supported by M2Sim) */
    /* __stack_top = 0x93ee0 (from linker) */
    adrp x0, __stack_top
    add x0, x0, :lo12:__stack_top
    mov sp, x0
    
    /* Clear BSS section */
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
    
    /* Call main (CoreMark entry point) */
    bl main
    
    /* Exit - trigger M2Sim termination */
    /* Use WFI (wait for interrupt) to halt */
3:
    wfi
    b 3b

.size _start, . - _start
