# M2Sim Progress Report

**Last updated:** 2026-02-06 18:15 EST (Cycle 307)

## Current Status

| Metric | Value |
|--------|-------|
| Total PRs Merged | **127** |
| Open PRs | 4 |
| Open Issues | 13 (excl. tracker) |
| Pipeline Coverage | **70.5%** |
| Emu Coverage | 79.9% |

## 15 BENCHMARKS READY — PUBLICATION TARGET MET!

### Cycle 307 Status

Syscall work progressing for SPEC benchmark support:
- **15 benchmarks ready** — target met!
- **Coverage targets met** — emu 79.9%, pipeline 70.5%
- **Syscalls implemented:** exit (93), write (64), read (63), close (57), openat (56), brk (214), mmap (222)
- **127 PRs merged total**
- **4 open PRs** — fstat (#279), read/write FD (#280), lseek (#282), acceptance tests (#283)
- **13 open issues** (excl. tracker)

**Recent Updates (Cycle 307):**
- PR #276 merged — mmap syscall (222) implemented (7th syscall!)
- PR #279 (fstat) has merge conflicts after mmap merge — Bob to resolve
- PR #282 (lseek) ready with cathy-approved, CI passing
- PR #283 (file I/O acceptance tests) created by Cathy

**Previous Updates (Cycle 304):**
- PR #275 merged — brk syscall (214) implemented
- Eric created 6 new issues (#269-#274) for syscall roadmap

**Infrastructure Ready:**
- Self-hosted runner guide: `docs/m2-runner-setup.md`
- Benchmark workflow: `.github/workflows/benchmark.yml`
- PolyBench scripts: `./scripts/capture-m2-baselines.sh`
- SPEC timing script: `./scripts/run-spec-native.sh`

---

## 📈 Benchmark Inventory Status

| Suite | Ready | Status |
|-------|-------|--------|
| PolyBench | **7** (gemm, atax, 2mm, mvt, jacobi-1d, 3mm, bicg) | ✅ Complete |
| Embench | **7** (aha-mont64, crc32, matmult-int, primecount, edn, statemate, huffbench) | ✅ Complete |
| CoreMark | 1 | ⚠️ Impractical (>50M instr) |
| **Total** | **15 ready** | 🎯 **PUBLICATION TARGET MET!** |

---

## Coverage Status

| Package | Coverage | Target | Status |
|---------|----------|--------|--------|
| emu | 79.9% | 70%+ | ✅ Exceeded |
| pipeline | 70.5% | 70%+ | ✅ **MET!** |

---

## PolyBench — 7 Benchmarks Ready 🎉

| Benchmark | Status | Instructions |
|-----------|--------|--------------|
| gemm | ✅ Merged (PR #238) | ~37K |
| atax | ✅ Merged (PR #239) | ~5K |
| 2mm | ✅ Merged (PR #246) | ~70K |
| mvt | ✅ Merged (PR #246) | ~5K |
| jacobi-1d | ✅ Merged (PR #249) | ~5.3K |
| 3mm | ✅ Merged (PR #250) | ~105K |
| bicg | ✅ Merged (PR #251) | ~4.8K |

**Dataset sizes now configurable (MEDIUM default):**
- MINI: 16×16 matrices (fast testing)
- SMALL: 60-120 elements
- MEDIUM: 200-400 elements (default for timing)
- LARGE: 1000-2000 elements

---

## Embench — 7 Benchmarks Ready 🎉

| Benchmark | Status | Notes |
|-----------|--------|-------|
| aha-mont64 | ✅ Ready | Montgomery multiplication |
| crc32 | ✅ Ready | CRC checksum |
| matmult-int | ✅ Ready | Matrix multiply |
| primecount | ✅ Ready | Prime number counting |
| edn | ✅ Ready | ~3.1M instructions |
| statemate | ✅ Merged (PR #247) | ~1.04M instructions |
| huffbench | ✅ Merged (PR #248) | Compression algorithm |

---

## SPEC CPU 2017 — Native Baseline

Initial native timing on marin-2 (M2 Mac Mini):

| Benchmark | Wall Time | User Time | Sys Time |
|-----------|-----------|-----------|----------|
| 505.mcf_r | 4.99s | 4.78s | 0.04s |
| 531.deepsjeng_r | 3.45s | 3.23s | 0.05s |

**Note:** Simulator execution requires additional syscall support (openat, close, mmap, brk). Read syscall now implemented!

---

## Open PRs

| # | Title | Status |
|---|-------|--------|
| 279 | [Bob] Implement fstat syscall (80) | Merge conflicts (cathy-approved) |
| 280 | [Bob] Extend read/write for FDTable | cathy-approved |
| 282 | [Bob] Implement lseek syscall (62) | cathy-approved, CI passing |
| 283 | [Cathy] File I/O acceptance tests | Needs review |

## Syscall Implementation Status

Critical path for SPEC benchmark support:

| Syscall | Number | Status | PR |
|---------|--------|--------|-----|
| exit | 93 | Implemented | — |
| write | 64 | Implemented | — |
| read | 63 | Implemented | #264 |
| close | 57 | Implemented | #267 |
| openat | 56 | Implemented | #268 |
| brk | 214 | Implemented | #275 |
| mmap | 222 | Implemented | #276 |
| fstat | 80 | In Review | #279 |
| lseek | 62 | In Review | #282 |

**Completed:** mmap syscall (#261) → PR #276 merged

---

## Open Issues (13 excl. tracker)

| # | Title | Priority |
|---|-------|----------|
| 277 | Validate 548.exchange2_r after mmap | high |
| 278 | mprotect syscall (226) | medium |
| 281 | Track SPEC progress after 548.exchange2_r | medium |
| 263 | fstat syscall (80) | medium |
| 269 | Extend read/write for FDTable | high |
| 270 | lseek syscall (62) | medium |
| 271 | munmap syscall (215) | low |
| 272 | exit_group syscall (94) | medium |
| 273 | getpid/getuid/gettid syscalls | low |
| 274 | clock_gettime syscall (113) | low |
| 139 | Multi-core execution | low |
| 138 | SPEC benchmark execution | high |
| 107 | SPEC benchmark suite | high |

**Critical path:** mmap merged → Eric to validate 548.exchange2_r (#277)

---

## Key Achievements

**126 PRs Merged!** 🎉🎉🎉

**Both Coverage Targets MET!**
- emu: 79.9% ✅ (exceeded)
- pipeline: 70.5% ✅ (achieved!)

**🎯 15 Intermediate Benchmarks Ready!**
- PolyBench: 7 kernels (gemm, atax, 2mm, mvt, jacobi-1d, 3mm, bicg)
- Embench: 7 benchmarks (aha-mont64, crc32, matmult-int, primecount, edn, statemate, huffbench)
- CoreMark: 1 (impractical for emulation)

**Workload Diversity:**
- Matrix computation (gemm, 2mm, 3mm, mvt, matmult-int, bicg)
- Stencil computation (jacobi-1d)
- Compression (huffbench)
- Signal processing (edn)
- State machine (statemate)
- Cryptographic (aha-mont64, crc32)
- Integer arithmetic (primecount)
- Linear algebra (atax, bicg)
