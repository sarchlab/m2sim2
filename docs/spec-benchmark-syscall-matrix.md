# SPEC CPU 2017 Benchmark-to-Syscall Matrix

This document maps specific SPEC CPU 2017 benchmarks to their syscall requirements,
helping prioritize syscall implementation order.

## Current M2Sim Syscall Status

| Syscall | Number | Status | Issue/PR |
|---------|--------|--------|----------|
| exit | 93 | ✅ Implemented | - |
| write | 64 | ✅ Implemented | - |
| read | 63 | ✅ Implemented | PR #264 merged |
| close | 57 | ✅ Implemented | PR #267 merged |
| openat | 56 | ✅ Implemented | PR #268 merged |
| brk | 214 | 📋 Planned | #260 |
| mmap | 222 | 📋 Planned | #261 |
| fstat | 80 | 📋 Planned | #263 |
| lseek | 62 | 📋 Planned | #270 |
| munmap | 215 | 📋 Planned | #271 |
| exit_group | 94 | 📋 Planned | #272 |

**Dependencies:** ✅ File descriptor table (#262) → PR #266 merged.

**Note:** read/write syscalls currently only support stdin/stdout/stderr. Issue #269 tracks extending them to use FDTable for opened files.

## Benchmark Syscall Requirements Matrix

### Integer Rate Benchmarks (SPECrate 2017 Integer)

| Benchmark | read | openat | close | brk | mmap | fstat | Notes |
|-----------|------|--------|-------|-----|------|-------|-------|
| 500.perlbench_r | X | X | X | X | X | X | Heavy I/O, complex |
| 502.gcc_r | X | X | X | X | X | X | Compiler, heavy I/O |
| 505.mcf_r | X | X | X | X | - | X | Single input file, simpler |
| 520.omnetpp_r | X | X | X | X | X | X | Network sim |
| 523.xalancbmk_r | X | X | X | X | X | X | XML processing |
| 525.x264_r | X | X | X | X | X | X | Video encoding |
| 531.deepsjeng_r | X | X | X | X | X | - | Chess, moderate I/O |
| 541.leela_r | X | X | X | X | X | - | Go AI, minimal I/O |
| 548.exchange2_r | - | - | - | X | X | - | Sudoku, no file I/O |
| 557.xz_r | X | X | X | X | X | X | Compression |

### Simplest to Most Complex (Recommended Order)

1. **548.exchange2_r** - Sudoku solver
   - Syscalls: brk, mmap only (no file I/O!)
   - Best first target after memory syscalls

2. **541.leela_r** - Go AI
   - Syscalls: read, openat, close, brk, mmap
   - Minimal file I/O (reads network weights once)

3. **531.deepsjeng_r** - Chess engine
   - Syscalls: read, openat, close, brk, mmap, (fstat optional)
   - Reads position file once

4. **505.mcf_r** - Vehicle scheduling
   - Syscalls: read, openat, close, brk, fstat
   - Single input file, simpler than others

## Syscall Implementation Priority

Based on the matrix above, recommended implementation order:

### Phase 1: Complete File I/O (Enable 505.mcf_r, 531.deepsjeng_r)
1. **File descriptor table (#262)** - Required foundation
2. **close (#258)** - Simple, enables cleanup
3. **openat (#259)** - Enables file opening

### Phase 2: Memory Management (Enable 548.exchange2_r)
4. **brk (#260)** - Simple heap growth
5. **mmap (#261)** - Anonymous memory mapping

### Phase 3: File Metadata
6. **fstat (#263)** - File statistics

## Benchmark-Specific Details

### 548.exchange2_r (Sudoku Solver)
- **Why simplest:** Pure computation, no file I/O
- **Memory:** Uses stack + heap for puzzle state
- **Critical syscalls:** brk (heap), mmap (large allocations)
- **Testing:** Can run without any file I/O infrastructure

### 505.mcf_r (Vehicle Scheduling)
- **Input:** Single network specification file (~500KB)
- **Output:** inp.out, mcf.out
- **Memory:** Integer arithmetic, pointer structures
- **Critical syscalls:** openat, read, close, fstat, brk

### 531.deepsjeng_r (Chess Engine)
- **Input:** FEN position file
- **Output:** Analysis results
- **Memory:** ~700MB for transposition tables
- **Critical syscalls:** openat, read, close, brk, mmap

### 541.leela_r (Go AI)
- **Input:** Network weights file
- **Output:** Move analysis
- **Memory:** Monte Carlo tree search
- **Critical syscalls:** openat, read, close, brk, mmap

## Testing Strategy

### Stage 1: Memory-Only Benchmark
Run 548.exchange2_r once brk + mmap implemented:
- No file I/O needed
- Tests memory management
- Fast validation

### Stage 2: Simple File I/O
Run 505.mcf_r with FD table + file syscalls:
- Single file read pattern
- Validates complete file I/O path

### Stage 3: Full Validation
Run remaining benchmarks incrementally.

## Estimated Implementation Effort

| Syscall | Complexity | Dependencies |
|---------|------------|--------------|
| close | Low | FD table |
| openat | Medium | FD table, OS interface |
| brk | Low | Track heap boundary |
| mmap | High | Memory region tracking |
| fstat | Low | FD table |

---
*Research compiled by Eric (Cycle 302)*
*Updated by Eric (Cycle 304) — FD table, close, openat merged*
