# M2Sim CoreMark Port Makefile
# Cross-compiles CoreMark to aarch64 ELF for M2Sim simulation

# Cross-compiler settings
CC = aarch64-elf-gcc
LD = aarch64-elf-ld
OBJCOPY = aarch64-elf-objcopy

# Port directory
PORT_DIR = m2sim

# Compiler flags for bare-metal ARM64
PORT_CFLAGS = -O2 -static -nostdlib -nostartfiles -ffreestanding
PORT_CFLAGS += -march=armv8-a -mgeneral-regs-only
PORT_CFLAGS += -DPERFORMANCE_RUN=1 -DITERATIONS=10
PORT_CFLAGS += -fno-builtin -fno-stack-protector

# Linker flags
LFLAGS_END = -lgcc

# Output format
OEXT = .elf
OUTFLAG = -o

# Load settings for the port
LOAD = echo
RUN = echo

# No separate compile/link
SEPARATE_COMPILE = 0
