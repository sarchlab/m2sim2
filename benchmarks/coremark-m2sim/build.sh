#!/bin/bash
# Build CoreMark for M2Sim (aarch64-elf bare-metal)

set -e

# CoreMark source directory (clone from https://github.com/eembc/coremark)
COREMARK_DIR="$(dirname "$0")/../coremark"
PORT_DIR="$(dirname "$0")"

if [ ! -d "${COREMARK_DIR}" ]; then
    echo "Error: CoreMark source not found at ${COREMARK_DIR}"
    echo "Clone it first: git clone https://github.com/eembc/coremark.git ${COREMARK_DIR}"
    exit 1
fi
OUTPUT="${PORT_DIR}/coremark_m2sim.elf"

CC=aarch64-elf-gcc
AS=aarch64-elf-as
LD=aarch64-elf-ld
OBJDUMP=aarch64-elf-objdump

# Compiler flags
CFLAGS="-O2 -static -nostdlib -nostartfiles -ffreestanding"
CFLAGS+=" -march=armv8-a -mgeneral-regs-only"
CFLAGS+=" -DPERFORMANCE_RUN=1 -DITERATIONS=10"
CFLAGS+=" -fno-builtin -fno-stack-protector"
CFLAGS+=" -I${COREMARK_DIR} -I${PORT_DIR}"

# Source files
SRCS="${COREMARK_DIR}/core_list_join.c \
      ${COREMARK_DIR}/core_main.c \
      ${COREMARK_DIR}/core_matrix.c \
      ${COREMARK_DIR}/core_state.c \
      ${COREMARK_DIR}/core_util.c \
      ${PORT_DIR}/core_portme.c"

echo "Building CoreMark for M2Sim..."

# Compile startup assembly
echo "  Compiling startup.S..."
${AS} -o "${PORT_DIR}/startup.o" "${PORT_DIR}/startup.S"

# Compile C sources
echo "  Compiling C sources..."
for src in ${SRCS}; do
    obj="${PORT_DIR}/$(basename "${src}" .c).o"
    echo "    $(basename ${src})..."
    ${CC} ${CFLAGS} -c -o "${obj}" "${src}"
done

# Link
echo "  Linking..."
OBJS="${PORT_DIR}/startup.o \
      ${PORT_DIR}/core_list_join.o \
      ${PORT_DIR}/core_main.o \
      ${PORT_DIR}/core_matrix.o \
      ${PORT_DIR}/core_state.o \
      ${PORT_DIR}/core_util.o \
      ${PORT_DIR}/core_portme.o"

${CC} ${CFLAGS} -T "${PORT_DIR}/linker.ld" -o "${OUTPUT}" ${OBJS} -lgcc

# Generate disassembly for debugging
echo "  Generating disassembly..."
${OBJDUMP} -d "${OUTPUT}" > "${PORT_DIR}/coremark_m2sim.dis"

echo "Done! Output: ${OUTPUT}"
echo "Size: $(wc -c < "${OUTPUT}") bytes"

# Show entry point
echo "Entry point:"
${OBJDUMP} -f "${OUTPUT}" | grep "start address"
