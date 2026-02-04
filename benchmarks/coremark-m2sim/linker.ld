/* M2Sim CoreMark Linker Script */
/* Bare-metal ARM64 ELF layout */

ENTRY(_start)

SECTIONS
{
    . = 0x80000;  /* Load address for M2Sim */
    
    .text : {
        *(.text.startup)
        *(.text*)
    }
    
    .rodata : {
        *(.rodata*)
    }
    
    .data : {
        *(.data*)
    }
    
    .bss : {
        __bss_start = .;
        *(.bss*)
        *(COMMON)
        __bss_end = .;
    }
    
    . = ALIGN(16);
    __stack_top = . + 0x10000;  /* 64KB stack */
}
