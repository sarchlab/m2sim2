#!/usr/bin/env python3
"""
linear_calibration.py - Linear Regression Calibration Tool

Uses the methodology suggested by Human in Issue #88:
Run benchmarks with varying instruction counts and use linear regression
to separate process startup overhead from actual instruction latency.

Formula: time = instruction_latency * instruction_count + constant_overhead

This script:
1. Generates assembly benchmarks with varying iteration counts
2. Builds and runs them, collecting timing data
3. Fits a linear regression model
4. Reports the instruction latency and overhead
"""

import os
import subprocess
import sys
import tempfile
import time
from dataclasses import dataclass
from pathlib import Path
from typing import List, Tuple

# Check for numpy/scipy availability
try:
    import numpy as np
    from scipy import stats
    HAS_SCIPY = True
except ImportError:
    HAS_SCIPY = False
    print("Note: scipy not available, using simple linear regression fallback")

# Benchmark templates - each generates N iterations of target instructions
BENCHMARK_TEMPLATES = {
    "arithmetic": {
        "description": "20 independent ADDs per iteration (ALU throughput)",
        "instructions_per_iter": 20,
        "template": """
// arithmetic_calibration.s - Generated for {iterations} iterations
.global _main
.align 4

_main:
    mov x10, #0              // iteration counter
{load_iterations}

    mov x0, #0
    mov x1, #0
    mov x2, #0
    mov x3, #0
    mov x4, #0

.loop:
    add x0, x0, #1
    add x1, x1, #1
    add x2, x2, #1
    add x3, x3, #1
    add x4, x4, #1

    add x0, x0, #1
    add x1, x1, #1
    add x2, x2, #1
    add x3, x3, #1
    add x4, x4, #1

    add x0, x0, #1
    add x1, x1, #1
    add x2, x2, #1
    add x3, x3, #1
    add x4, x4, #1

    add x0, x0, #1
    add x1, x1, #1
    add x2, x2, #1
    add x3, x3, #1
    add x4, x4, #1

    add x10, x10, #1
    cmp x10, x11
    b.lt .loop

    mov x0, #0
    mov x16, #1
    svc #0x80
"""
    },
    "dependency": {
        "description": "20 dependent ADDs per iteration (RAW hazards)",
        "instructions_per_iter": 20,
        "template": """
// dependency_calibration.s - Generated for {iterations} iterations
.global _main
.align 4

_main:
    mov x10, #0              // iteration counter
{load_iterations}

    mov x0, #0

.loop:
    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1

    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1

    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1

    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1
    add x0, x0, #1

    add x10, x10, #1
    cmp x10, x11
    b.lt .loop

    mov x0, #0
    mov x16, #1
    svc #0x80
"""
    },
    "branch": {
        "description": "5 taken branches per iteration (branch predictor)",
        "instructions_per_iter": 5,
        "template": """
// branch_calibration.s - Generated for {iterations} iterations
.global _main
.align 4

_main:
    mov x10, #0              // iteration counter
{load_iterations}

.loop:
    b .b1
.b1:
    b .b2
.b2:
    b .b3
.b3:
    b .b4
.b4:
    b .b5
.b5:

    add x10, x10, #1
    cmp x10, x11
    b.lt .loop

    mov x0, #0
    mov x16, #1
    svc #0x80
"""
    },
    "memorystrided": {
        "description": "10 store/load pairs with stride-4 access per iteration (strided memory pattern)",
        "instructions_per_iter": 20,
        "template": """
// memorystrided_calibration.s - Generated for {iterations} iterations
.global _main
.align 4

_main:
    mov x10, #0              // iteration counter
{load_iterations}

    sub sp, sp, #320         // allocate buffer (enough for stride offsets)
    mov x0, #7               // value to store/load
    mov x1, sp               // base address

.loop:
    // 10 store/load pairs at stride-4 offsets (32-byte stride)
    str x0, [x1, #0]
    ldr x0, [x1, #0]
    str x0, [x1, #32]
    ldr x0, [x1, #32]
    str x0, [x1, #64]
    ldr x0, [x1, #64]
    str x0, [x1, #96]
    ldr x0, [x1, #96]
    str x0, [x1, #128]
    ldr x0, [x1, #128]
    str x0, [x1, #160]
    ldr x0, [x1, #160]
    str x0, [x1, #192]
    ldr x0, [x1, #192]
    str x0, [x1, #224]
    ldr x0, [x1, #224]
    str x0, [x1, #256]
    ldr x0, [x1, #256]
    str x0, [x1, #288]
    ldr x0, [x1, #288]

    add x10, x10, #1
    cmp x10, x11
    b.lt .loop

    add sp, sp, #320
    mov x0, #0
    mov x16, #1
    svc #0x80
"""
    },
    "loadheavy": {
        "description": "20 independent loads per iteration (load throughput)",
        "instructions_per_iter": 20,
        "template": """
// loadheavy_calibration.s - Generated for {iterations} iterations
.global _main
.align 4

_main:
    mov x10, #0              // iteration counter
{load_iterations}

    sub sp, sp, #160         // allocate buffer (20 * 8 bytes)
    mov x1, sp               // base address

    // Pre-fill buffer with known values
    mov x2, #1
    str x2, [x1, #0]
    mov x2, #2
    str x2, [x1, #8]
    mov x2, #3
    str x2, [x1, #16]
    mov x2, #4
    str x2, [x1, #24]
    mov x2, #5
    str x2, [x1, #32]
    mov x2, #6
    str x2, [x1, #40]
    mov x2, #7
    str x2, [x1, #48]
    mov x2, #8
    str x2, [x1, #56]
    mov x2, #9
    str x2, [x1, #64]
    mov x2, #10
    str x2, [x1, #72]
    mov x2, #11
    str x2, [x1, #80]
    mov x2, #12
    str x2, [x1, #88]
    mov x2, #13
    str x2, [x1, #96]
    mov x2, #14
    str x2, [x1, #104]
    mov x2, #15
    str x2, [x1, #112]
    mov x2, #16
    str x2, [x1, #120]
    mov x2, #17
    str x2, [x1, #128]
    mov x2, #18
    str x2, [x1, #136]
    mov x2, #19
    str x2, [x1, #144]
    mov x2, #20
    str x2, [x1, #152]

.loop:
    // 20 loads to independent registers (no RAW hazards)
    ldr x0, [x1, #0]
    ldr x2, [x1, #8]
    ldr x3, [x1, #16]
    ldr x4, [x1, #24]
    ldr x5, [x1, #32]
    ldr x6, [x1, #40]
    ldr x7, [x1, #48]
    ldr x9, [x1, #56]
    ldr x12, [x1, #64]
    ldr x13, [x1, #72]
    ldr x14, [x1, #80]
    ldr x15, [x1, #88]
    ldr x16, [x1, #96]
    ldr x17, [x1, #104]
    ldr x18, [x1, #112]
    ldr x19, [x1, #120]
    ldr x20, [x1, #128]
    ldr x21, [x1, #136]
    ldr x22, [x1, #144]
    ldr x23, [x1, #152]

    add x10, x10, #1
    cmp x10, x11
    b.lt .loop

    add sp, sp, #160
    mov x0, #0
    mov x16, #1
    svc #0x80
"""
    },
    "storeheavy": {
        "description": "20 independent stores per iteration (store throughput)",
        "instructions_per_iter": 20,
        "template": """
// storeheavy_calibration.s - Generated for {iterations} iterations
.global _main
.align 4

_main:
    mov x10, #0              // iteration counter
{load_iterations}

    sub sp, sp, #160         // allocate buffer (20 * 8 bytes)
    mov x1, sp               // base address
    mov x2, #99              // value to store

.loop:
    // 20 stores to sequential addresses (no data dependencies)
    str x2, [x1, #0]
    str x2, [x1, #8]
    str x2, [x1, #16]
    str x2, [x1, #24]
    str x2, [x1, #32]
    str x2, [x1, #40]
    str x2, [x1, #48]
    str x2, [x1, #56]
    str x2, [x1, #64]
    str x2, [x1, #72]
    str x2, [x1, #80]
    str x2, [x1, #88]
    str x2, [x1, #96]
    str x2, [x1, #104]
    str x2, [x1, #112]
    str x2, [x1, #120]
    str x2, [x1, #128]
    str x2, [x1, #136]
    str x2, [x1, #144]
    str x2, [x1, #152]

    add x10, x10, #1
    cmp x10, x11
    b.lt .loop

    add sp, sp, #160
    mov x0, #0
    mov x16, #1
    svc #0x80
"""
    },
    "branchheavy": {
        "description": "10 conditional branches per iteration (alternating taken/not-taken)",
        "instructions_per_iter": 10,
        "template": """
// branchheavy_calibration.s - Generated for {iterations} iterations
// Uses alternating taken/not-taken pattern to stress branch predictor
.global _main
.align 4

_main:
    mov x10, #0              // iteration counter
{load_iterations}

.loop:
    // 10 conditional branches (CMP + B.LT pattern)
    // Even branches: compare X10 >= 0, always taken (forward skip)
    // Odd branches: compare 0 >= X10, never taken (falls through)
    // This creates alternating taken/not-taken to stress predictor

    // Branch 1: taken (X10 >= 0 is always true since X10 is unsigned counter)
    cmp x10, #0
    b.ge .t1
    nop
.t1:
    // Branch 2: not taken (0 < X10 when X10 > 0, first iter 0 == 0 so taken)
    cmp xzr, x10
    b.ge .t2
    nop
.t2:
    // Branch 3: taken
    cmp x10, #0
    b.ge .t3
    nop
.t3:
    // Branch 4: not taken
    cmp xzr, x10
    b.ge .t4
    nop
.t4:
    // Branch 5: taken
    cmp x10, #0
    b.ge .t5
    nop
.t5:
    // Branch 6: not taken
    cmp xzr, x10
    b.ge .t6
    nop
.t6:
    // Branch 7: taken
    cmp x10, #0
    b.ge .t7
    nop
.t7:
    // Branch 8: not taken
    cmp xzr, x10
    b.ge .t8
    nop
.t8:
    // Branch 9: taken
    cmp x10, #0
    b.ge .t9
    nop
.t9:
    // Branch 10: not taken
    cmp xzr, x10
    b.ge .t10
    nop
.t10:

    add x10, x10, #1
    cmp x10, x11
    b.lt .loop

    mov x0, #0
    mov x16, #1
    svc #0x80
"""
    },
    "vectorsum": {
        "description": "16-element array sum loop per iteration (load+accumulate)",
        "instructions_per_iter": 100,  # 4 setup + 16 inner iters * 6 insts/inner iter = 100
        "template": """
// vectorsum_calibration.s - Generated for {iterations} outer iterations
// Inner loop: 16 loads + accumulates (6 insts * 16 = 96 inner + 4 setup = 100)
.global _main
.align 4

_main:
    // Allocate array on stack (16 * 8 = 128 bytes, 256 with alignment padding)
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

    mov x10, #0              // outer iteration counter
{load_iterations}

.outer_loop:
    // Setup (4 insts)
    mov x0, #0              // sum = 0
    mov x1, sp              // array pointer
    mov x2, #0              // i = 0
    mov x3, #16             // N = 16

.inner_loop:
    // Inner loop body (6 insts * 16 iters = 96)
    ldr x4, [x1]            // load A[i]
    add x0, x0, x4          // sum += A[i]
    add x1, x1, #8          // ptr += 8
    add x2, x2, #1          // i++
    cmp x2, x3              // i < N?
    b.lt .inner_loop

    add x10, x10, #1
    cmp x10, x11
    b.lt .outer_loop

    add sp, sp, #256
    mov x0, #0
    mov x16, #1
    svc #0x80
"""
    },
    "vectoradd": {
        "description": "16-element vector add loop per iteration (2 loads+add+store)",
        "instructions_per_iter": 165,  # 5 setup + 16 inner iters * 10 insts/inner iter = 165
        "template": """
// vectoradd_calibration.s - Generated for {iterations} outer iterations
// Inner loop: C[i]=A[i]+B[i] (10 insts * 16 = 160 inner + 5 setup = 165)
.global _main
.align 4

_main:
    // Allocate 3 arrays: A, B, C (each 128 bytes = 16*8 = 384, 512 with padding)
    sub sp, sp, #512

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

    // Initialize B[i] = 2*(i+1)
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

    mov x10, #0              // outer iteration counter
{load_iterations}

.outer_loop:
    // Setup (5 insts)
    add x1, sp, #0          // A ptr
    add x2, sp, #128        // B ptr
    add x3, sp, #256        // C ptr
    mov x4, #0              // i = 0
    mov x5, #16             // N = 16

.inner_loop:
    // Inner loop body (10 insts * 16 iters = 160)
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

    add x10, x10, #1
    cmp x10, x11
    b.lt .outer_loop

    add sp, sp, #512
    mov x0, #0
    mov x16, #1
    svc #0x80
"""
    },
    "reductiontree": {
        "description": "16-element parallel reduction tree per iteration (16 loads + 15 adds)",
        "instructions_per_iter": 31,  # 16 loads + 15 adds = 31 flat instructions
        "template": """
// reductiontree_calibration.s - Generated for {iterations} iterations
// Flat body: 16 loads + 15 tree-reduction adds = 31 insts per iteration
.global _main
.align 4

_main:
    // Allocate array on stack (16 * 8 = 128 bytes, 256 with alignment padding)
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

    mov x20, #0              // iteration counter
{load_iterations_x21}

.loop:
    // Load all 16 elements (16 insts)
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

    // Level 1: 8 pairwise sums
    add x0, x0, x2
    add x3, x3, x4
    add x5, x5, x6
    add x7, x7, x9
    add x10, x10, x11
    add x12, x12, x13
    add x14, x14, x15
    add x16, x16, x17

    // Level 2: 4 sums
    add x0, x0, x3
    add x5, x5, x7
    add x10, x10, x12
    add x14, x14, x16

    // Level 3: 2 sums
    add x0, x0, x5
    add x10, x10, x14

    // Level 4: final sum
    add x0, x0, x10

    add x20, x20, #1
    cmp x20, x21
    b.lt .loop

    add sp, sp, #256
    mov x0, #0
    mov x16, #1
    svc #0x80
"""
    },
    "strideindirect": {
        "description": "8-hop pointer chase per iteration (dependent load chain)",
        "instructions_per_iter": 50,  # 2 setup + 8 inner iters * 6 insts = 50
        "template": """
// strideindirect_calibration.s - Generated for {iterations} outer iterations
// Inner loop: 8 dependent pointer chases (6 insts * 8 = 48 inner + 2 setup = 50)
.global _main
.align 4

_main:
    // Allocate index array on stack (8 * 8 = 64 bytes, 128 with alignment padding)
    sub sp, sp, #128

    // Build pointer chase chain: 0->3->1->5->2->7->4->6->0
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

    mov x10, #0              // outer iteration counter
{load_iterations}

.outer_loop:
    // Setup (2 insts)
    mov x2, #0              // start at index 0
    mov x3, #0              // hop counter

.chase_loop:
    // Inner loop body (6 insts * 8 hops = 48)
    lsl x5, x2, #3          // x5 = index * 8
    add x5, sp, x5          // x5 = base + offset
    ldr x2, [x5]            // x2 = A[index] (next index)
    add x3, x3, #1          // hop++
    cmp x3, #8              // 8 hops?
    b.lt .chase_loop

    add x10, x10, #1
    cmp x10, x11
    b.lt .outer_loop

    add sp, sp, #128
    mov x0, #0
    mov x16, #1
    svc #0x80
"""
    },
}


def load_iterations_asm(n: int, reg: str = "x11") -> str:
    """Generate ARM64 assembly to load iteration count into given register."""
    if n <= 0xFFFF:
        return f"    movz {reg}, #{n}"
    elif n <= 0xFFFFFFFF:
        low = n & 0xFFFF
        high = (n >> 16) & 0xFFFF
        lines = [f"    movz {reg}, #{low}"]
        if high > 0:
            lines.append(f"    movk {reg}, #{high}, lsl #16")
        return "\n".join(lines)
    else:
        raise ValueError(f"Iteration count {n} too large (max 2^32-1)")


def generate_benchmark(template_name: str, iterations: int) -> str:
    """Generate assembly source for a benchmark with given iteration count."""
    tmpl = BENCHMARK_TEMPLATES[template_name]
    return tmpl["template"].format(
        iterations=iterations,
        load_iterations=load_iterations_asm(iterations),
        load_iterations_x21=load_iterations_asm(iterations, "x21"),
    )


def build_and_run(asm_source: str, runs: int = 5, warmup: int = 2) -> List[float]:
    """Build assembly source and run multiple times, returning execution times in seconds.
    
    Includes warmup runs (discarded) to warm up caches and reduce variance.
    """
    times = []
    
    with tempfile.TemporaryDirectory() as tmpdir:
        asm_path = Path(tmpdir) / "benchmark.s"
        obj_path = Path(tmpdir) / "benchmark.o"
        exe_path = Path(tmpdir) / "benchmark"
        
        # Write assembly
        asm_path.write_text(asm_source)
        
        # Assemble
        result = subprocess.run(
            ["as", "-o", str(obj_path), str(asm_path)],
            capture_output=True, text=True
        )
        if result.returncode != 0:
            raise RuntimeError(f"Assembly failed: {result.stderr}")
        
        # Find SDK path dynamically (works on both Xcode and CommandLineTools)
        sdk_result = subprocess.run(
            ["xcrun", "--show-sdk-path"],
            capture_output=True, text=True
        )
        sdk_path = sdk_result.stdout.strip() if sdk_result.returncode == 0 else "/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk"

        # Link
        result = subprocess.run(
            ["ld", "-o", str(exe_path), str(obj_path),
             "-lSystem", "-L", sdk_path + "/usr/lib",
             "-syslibroot", sdk_path,
             "-e", "_main", "-arch", "arm64"],
            capture_output=True, text=True
        )
        if result.returncode != 0:
            raise RuntimeError(f"Link failed: {result.stderr}")
        
        # Warmup runs (discarded)
        for _ in range(warmup):
            subprocess.run([str(exe_path)], capture_output=True)
        
        # Timed runs
        for _ in range(runs):
            start = time.perf_counter()
            subprocess.run([str(exe_path)], capture_output=True)
            end = time.perf_counter()
            times.append(end - start)
    
    return times


@dataclass
class CalibrationResult:
    """Result of linear regression calibration."""
    benchmark: str
    description: str
    instruction_latency_ns: float  # nanoseconds per instruction
    overhead_ms: float             # process startup overhead in milliseconds
    r_squared: float               # goodness of fit
    data_points: List[Tuple[int, float]]  # (instruction_count, avg_time_ms)


def simple_linear_regression(x: List[float], y: List[float]) -> Tuple[float, float, float]:
    """Simple linear regression without scipy. Returns (slope, intercept, r_squared)."""
    n = len(x)
    sum_x = sum(x)
    sum_y = sum(y)
    sum_xy = sum(xi * yi for xi, yi in zip(x, y))
    sum_x2 = sum(xi * xi for xi in x)
    sum_y2 = sum(yi * yi for yi in y)
    
    slope = (n * sum_xy - sum_x * sum_y) / (n * sum_x2 - sum_x * sum_x)
    intercept = (sum_y - slope * sum_x) / n
    
    # R-squared
    ss_tot = sum((yi - sum_y/n)**2 for yi in y)
    ss_res = sum((yi - (slope * xi + intercept))**2 for xi, yi in zip(x, y))
    r_squared = 1 - (ss_res / ss_tot) if ss_tot > 0 else 0
    
    return slope, intercept, r_squared


def trimmed_mean(values: List[float], trim_percent: float = 0.2) -> float:
    """Calculate trimmed mean, removing top and bottom trim_percent of values."""
    if len(values) < 3:
        return sum(values) / len(values)
    
    sorted_vals = sorted(values)
    n = len(sorted_vals)
    trim_count = int(n * trim_percent)
    
    if trim_count == 0:
        return sum(sorted_vals) / n
    
    trimmed = sorted_vals[trim_count:-trim_count]
    return sum(trimmed) / len(trimmed) if trimmed else sum(sorted_vals) / n


def calibrate_benchmark(template_name: str, iteration_counts: List[int], 
                        runs_per_count: int = 15, verbose: bool = True) -> CalibrationResult:
    """Run calibration for a single benchmark type.
    
    Uses warmup runs and trimmed mean to reduce variance.
    """
    tmpl = BENCHMARK_TEMPLATES[template_name]
    instr_per_iter = tmpl["instructions_per_iter"]
    
    if verbose:
        print(f"\n{'='*60}")
        print(f"Calibrating: {template_name}")
        print(f"Description: {tmpl['description']}")
        print(f"{'='*60}")
    
    data_points = []
    instruction_counts = []
    times_ms = []
    
    for iterations in iteration_counts:
        total_instructions = iterations * instr_per_iter
        if verbose:
            print(f"  {iterations:>10,} iterations ({total_instructions:>12,} instructions)... ", end="", flush=True)
        
        asm_source = generate_benchmark(template_name, iterations)
        run_times = build_and_run(asm_source, runs=runs_per_count, warmup=3)
        
        # Use trimmed mean to reduce impact of outliers
        run_times_ms = [t * 1000 for t in run_times]
        avg_time_ms = trimmed_mean(run_times_ms, 0.2)
        
        # Calculate std of trimmed values for reporting
        sorted_times = sorted(run_times_ms)
        trim_count = int(len(sorted_times) * 0.2)
        trimmed = sorted_times[trim_count:-trim_count] if trim_count > 0 else sorted_times
        std_time_ms = (sum((t - avg_time_ms)**2 for t in trimmed) / len(trimmed)) ** 0.5 if trimmed else 0
        
        if verbose:
            print(f"{avg_time_ms:7.2f} ms (±{std_time_ms:.2f})")
        
        data_points.append((total_instructions, avg_time_ms))
        instruction_counts.append(total_instructions)
        times_ms.append(avg_time_ms)
    
    # Linear regression: time_ms = slope * instructions + intercept
    if HAS_SCIPY:
        slope, intercept, r_value, p_value, std_err = stats.linregress(instruction_counts, times_ms)
        r_squared = r_value ** 2
    else:
        slope, intercept, r_squared = simple_linear_regression(instruction_counts, times_ms)
    
    # Convert slope from ms/instruction to ns/instruction
    latency_ns = slope * 1e6
    
    return CalibrationResult(
        benchmark=template_name,
        description=tmpl['description'],
        instruction_latency_ns=latency_ns,
        overhead_ms=intercept,
        r_squared=r_squared,
        data_points=data_points
    )


def print_results(results: List[CalibrationResult]):
    """Print calibration results in a readable format."""
    print("\n" + "="*70)
    print("CALIBRATION RESULTS")
    print("="*70)
    print(f"{'Benchmark':<15} {'Latency (ns)':<14} {'Overhead (ms)':<14} {'R²':<10}")
    print("-"*70)
    
    for r in results:
        print(f"{r.benchmark:<15} {r.instruction_latency_ns:>11.4f}   {r.overhead_ms:>11.2f}   {r.r_squared:>8.6f}")
    
    print("\n" + "="*70)
    print("INTERPRETATION")
    print("="*70)
    
    for r in results:
        freq_ghz = 1 / (r.instruction_latency_ns) if r.instruction_latency_ns > 0 else 0
        cpi = r.instruction_latency_ns * 3.5  # Assume 3.5 GHz M2 P-core
        
        print(f"\n{r.benchmark}:")
        print(f"  Formula: time = {r.instruction_latency_ns:.4f} ns × instructions + {r.overhead_ms:.2f} ms")
        print(f"  Implied throughput: {freq_ghz:.2f} G instructions/sec")
        print(f"  At 3.5 GHz: {cpi:.2f} cycles per instruction (CPI)")
        print(f"  Process startup overhead: {r.overhead_ms:.2f} ms")
        print(f"  R² = {r.r_squared:.6f} {'(excellent fit)' if r.r_squared > 0.999 else '(good fit)' if r.r_squared > 0.99 else ''}")


def main():
    """Run the full calibration suite."""
    import argparse
    import json

    parser = argparse.ArgumentParser(description="M2Sim Linear Regression Calibration Tool")
    parser.add_argument("--benchmarks", nargs="*", default=None,
                        help="Specific benchmarks to calibrate (default: all)")
    parser.add_argument("--runs", type=int, default=15,
                        help="Runs per data point (default: 15)")
    parser.add_argument("--output", type=str, default=None,
                        help="Output JSON path (default: calibration_results.json)")
    args = parser.parse_args()

    print("="*70)
    print("M2Sim Linear Regression Calibration Tool")
    print("Methodology: Issue #88 (Human's suggestion)")
    print("="*70)

    # Iteration counts following Human's suggestion: 100, 200, 400, 800, 1600...
    # Scaled up significantly so instruction time >> overhead (~20ms)
    # At ~3.5B instr/sec, need >70M instructions to get >20ms instruction time
    # Using: 1M, 2M, 4M, 8M, 16M, 32M iterations
    # With 20 instructions/iter, this gives 20M to 640M total instructions
    iteration_counts = [
        1_000_000,
        2_000_000,
        4_000_000,
        8_000_000,
        16_000_000,
        32_000_000,
    ]

    # Select which benchmarks to run
    benchmark_names = args.benchmarks if args.benchmarks else list(BENCHMARK_TEMPLATES.keys())
    for name in benchmark_names:
        if name not in BENCHMARK_TEMPLATES:
            print(f"Error: unknown benchmark '{name}'")
            print(f"Available: {', '.join(BENCHMARK_TEMPLATES.keys())}")
            sys.exit(1)

    results = []

    for template_name in benchmark_names:
        result = calibrate_benchmark(
            template_name,
            iteration_counts,
            runs_per_count=args.runs,
            verbose=True
        )
        results.append(result)

    print_results(results)

    # Save results to JSON
    output = {
        "methodology": "linear_regression",
        "formula": "time_ms = latency_ns * instruction_count / 1e6 + overhead_ms",
        "results": [
            {
                "benchmark": r.benchmark,
                "description": r.description,
                "calibrated": True,
                "instruction_latency_ns": r.instruction_latency_ns,
                "overhead_ms": r.overhead_ms,
                "r_squared": r.r_squared,
                "data_points": [{"instructions": d[0], "time_ms": d[1]} for d in r.data_points]
            }
            for r in results
        ]
    }

    output_path = Path(args.output) if args.output else Path(__file__).parent / "calibration_results.json"
    output_path.write_text(json.dumps(output, indent=2))
    print(f"\nResults saved to: {output_path}")


if __name__ == "__main__":
    main()
