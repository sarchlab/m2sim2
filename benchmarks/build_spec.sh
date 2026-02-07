#!/bin/bash
# Build SPEC CPU 2017 benchmarks as static ARM64 Linux ELF binaries for M2Sim.
#
# Prerequisites:
#   - brew install FiloSottile/musl-cross/musl-cross  (provides aarch64-linux-musl-gcc)
#   - Docker (for Fortran benchmarks)
#   - SPEC source at benchmarks/spec (symlink to SPEC installation)
#
# Usage: ./build_spec.sh [benchmark]
#   benchmark: 505.mcf_r | 531.deepsjeng_r | 548.exchange2_r | all (default: all)

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SPEC_DIR="${SCRIPT_DIR}/spec"

if [ ! -d "${SPEC_DIR}/benchspec/CPU" ]; then
    echo "Error: SPEC not found at ${SPEC_DIR}"
    echo "Create a symlink: ln -s /path/to/spec ${SPEC_DIR}"
    exit 1
fi

CC=aarch64-linux-musl-gcc
CXX=aarch64-linux-musl-g++

build_505_mcf_r() {
    echo "Building 505.mcf_r..."
    local SRC="${SPEC_DIR}/benchspec/CPU/505.mcf_r/src"
    local EXE_DIR="${SPEC_DIR}/benchspec/CPU/505.mcf_r/exe"
    mkdir -p "${EXE_DIR}"

    ${CC} -static -O2 -DSPEC -DSPEC_LP64 \
        -I"${SRC}" -I"${SRC}/spec_qsort" \
        "${SRC}"/*.c "${SRC}"/spec_qsort/spec_qsort.c \
        -o "${EXE_DIR}/mcf_r_base.arm64" -lm

    echo "  -> $(file "${EXE_DIR}/mcf_r_base.arm64")"

    # Set up test run directory
    local RUN_DIR="${SPEC_DIR}/benchspec/CPU/505.mcf_r/run/run_base_test_arm64.0000"
    mkdir -p "${RUN_DIR}"
    cp "${SPEC_DIR}/benchspec/CPU/505.mcf_r/data/test/input/inp.in" "${RUN_DIR}/"
    echo "  Test input staged at ${RUN_DIR}"
}

build_531_deepsjeng_r() {
    echo "Building 531.deepsjeng_r..."
    local SRC="${SPEC_DIR}/benchspec/CPU/531.deepsjeng_r/src"
    local EXE_DIR="${SPEC_DIR}/benchspec/CPU/531.deepsjeng_r/exe"
    mkdir -p "${EXE_DIR}"

    ${CXX} -static -O2 -DSPEC -DSMALL_MEMORY \
        -I"${SRC}" \
        "${SRC}"/*.cpp \
        -o "${EXE_DIR}/deepsjeng_r_base.arm64"

    echo "  -> $(file "${EXE_DIR}/deepsjeng_r_base.arm64")"

    # Set up test run directory
    local RUN_DIR="${SPEC_DIR}/benchspec/CPU/531.deepsjeng_r/run/run_base_test_arm64.0000"
    mkdir -p "${RUN_DIR}"
    cp "${SPEC_DIR}/benchspec/CPU/531.deepsjeng_r/data/test/input/test.txt" "${RUN_DIR}/"
    echo "  Test input staged at ${RUN_DIR}"
}

build_548_exchange2_r() {
    echo "Building 548.exchange2_r (via Docker, Fortran)..."
    local SRC="${SPEC_DIR}/benchspec/CPU/548.exchange2_r/src"
    local EXE_DIR="${SPEC_DIR}/benchspec/CPU/548.exchange2_r/exe"
    mkdir -p "${EXE_DIR}"

    local TMPDIR
    TMPDIR=$(mktemp -d)

    docker run --rm --platform linux/arm64 \
        -v "${SRC}:/src:ro" \
        -v "${TMPDIR}:/out" \
        alpine:latest sh -c \
        "apk add --no-cache gfortran musl-dev && \
         gfortran -static -no-pie -O2 -DSPEC -cpp \
         -o /out/exchange2_r_base.arm64 /src/exchange2.F90"

    cp "${TMPDIR}/exchange2_r_base.arm64" "${EXE_DIR}/"
    rm -rf "${TMPDIR}"

    echo "  -> $(file "${EXE_DIR}/exchange2_r_base.arm64")"

    # Set up test run directory
    local RUN_DIR="${SPEC_DIR}/benchspec/CPU/548.exchange2_r/run/run_base_test_arm64.0000"
    mkdir -p "${RUN_DIR}"
    cp "${SPEC_DIR}/benchspec/CPU/548.exchange2_r/data/test/input/control" "${RUN_DIR}/"
    cp "${SPEC_DIR}/benchspec/CPU/548.exchange2_r/data/all/input/puzzles.txt" "${RUN_DIR}/"
    echo "  Test input staged at ${RUN_DIR}"
}

TARGET="${1:-all}"

case "${TARGET}" in
    505.mcf_r)
        build_505_mcf_r
        ;;
    531.deepsjeng_r)
        build_531_deepsjeng_r
        ;;
    548.exchange2_r)
        build_548_exchange2_r
        ;;
    all)
        build_505_mcf_r
        build_531_deepsjeng_r
        build_548_exchange2_r
        ;;
    *)
        echo "Unknown benchmark: ${TARGET}"
        echo "Usage: $0 [505.mcf_r|531.deepsjeng_r|548.exchange2_r|all]"
        exit 1
        ;;
esac

echo "Done!"
