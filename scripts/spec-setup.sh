#!/bin/bash
# SPEC CPU 2017 Setup Script for M2Sim
# [Bob] Created for ARM64 benchmark compilation
#
# Prerequisites:
# - SPEC CPU 2017 ISO/installer at /Users/yifan/Documents/spec
# - Clang/LLVM for ARM64 cross-compilation (or native ARM64 machine)
#
# Usage: ./scripts/spec-setup.sh [install|build|check]

set -e

SPEC_ROOT="${SPEC_ROOT:-/Users/yifan/Documents/spec}"
M2SIM_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SPEC_LINK="$M2SIM_ROOT/benchmarks/spec"

# Target benchmarks for M2Sim validation
BENCHMARKS=(
    "505.mcf_r"
    "531.deepsjeng_r"
    "557.xz_r"
)

usage() {
    echo "Usage: $0 [install|build|check|link]"
    echo ""
    echo "Commands:"
    echo "  install  - Run SPEC installer (interactive)"
    echo "  build    - Build ARM64 binaries for target benchmarks"
    echo "  check    - Check which binaries are available"
    echo "  link     - Create symlink from benchmarks/spec to SPEC_ROOT"
    echo ""
    echo "Environment:"
    echo "  SPEC_ROOT - Path to SPEC installation (default: /Users/yifan/Documents/spec)"
}

check_spec() {
    if [ ! -f "$SPEC_ROOT/shrc" ]; then
        echo "ERROR: SPEC not found at $SPEC_ROOT"
        echo "Set SPEC_ROOT environment variable or install SPEC first"
        exit 1
    fi
}

cmd_link() {
    if [ -L "$SPEC_LINK" ]; then
        echo "Symlink already exists: $SPEC_LINK -> $(readlink "$SPEC_LINK")"
    elif [ -e "$SPEC_LINK" ]; then
        echo "ERROR: $SPEC_LINK exists but is not a symlink"
        exit 1
    else
        ln -s "$SPEC_ROOT" "$SPEC_LINK"
        echo "Created symlink: $SPEC_LINK -> $SPEC_ROOT"
    fi
}

cmd_install() {
    if [ ! -f "$SPEC_ROOT/install.sh" ]; then
        echo "ERROR: SPEC installer not found at $SPEC_ROOT/install.sh"
        exit 1
    fi
    
    echo "Running SPEC installer..."
    echo "Note: This is interactive. Accept defaults for most options."
    cd "$SPEC_ROOT"
    ./install.sh
}

cmd_build() {
    check_spec
    
    # Check if config exists
    CONFIG="$SPEC_ROOT/config/arm64-m2sim.cfg"
    if [ ! -f "$CONFIG" ]; then
        echo "Creating ARM64 config file..."
        cp "$M2SIM_ROOT/scripts/arm64-m2sim.cfg" "$CONFIG"
    fi
    
    cd "$SPEC_ROOT"
    source shrc
    
    echo "Building ARM64 binaries for benchmarks:"
    for bench in "${BENCHMARKS[@]}"; do
        echo "  Building $bench..."
        runcpu --config=arm64-m2sim --action=build --tune=base "$bench" || {
            echo "WARNING: Failed to build $bench"
        }
    done
    
    echo ""
    echo "Build complete. Run '$0 check' to see available binaries."
}

cmd_check() {
    check_spec
    
    echo "Checking ARM64 binaries:"
    echo ""
    
    for bench in "${BENCHMARKS[@]}"; do
        # Extract benchmark number and name
        num=$(echo "$bench" | cut -d. -f1)
        name=$(echo "$bench" | cut -d. -f2 | sed 's/_r$//')
        
        # Look for ARM64 binary
        binary_path="$SPEC_ROOT/benchspec/CPU/$bench/exe/${name}_r_base.arm64"
        
        if [ -f "$binary_path" ]; then
            echo "  ✅ $bench - $(ls -lh "$binary_path" | awk '{print $5}')"
        else
            echo "  ❌ $bench - not built"
        fi
    done
}

# Main
case "${1:-}" in
    install) cmd_install ;;
    build) cmd_build ;;
    check) cmd_check ;;
    link) cmd_link ;;
    *) usage ;;
esac
