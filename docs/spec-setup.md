# SPEC CPU 2017 Setup for M2Sim

*Research by Eric — 2026-02-04*

## Current Status

**SPEC is NOT installed yet!**

The SPEC CPU 2017 distribution exists at `/Users/yifan/Documents/spec` but the actual benchmarks have not been installed. The `benchspec/CPU` directory (containing benchmark source and binaries) does not exist.

## Installation Required

1. **Run SPEC installer:**
   ```bash
   cd /Users/yifan/Documents/spec
   ./install.sh
   ```
   This will extract the benchmark source code into `benchspec/`.

2. **Create ARM64 config file:**
   The benchmarks need to be compiled for ARM64. SPEC uses config files to specify compiler settings.
   
   Create `config/arm64-m2.cfg` based on `config/Example-gcc-macosx.cfg`:
   ```
   CC = clang
   CXX = clang++
   FC = gfortran
   OPTIMIZE = -O2
   ```

3. **Build benchmarks:**
   ```bash
   source shrc
   runcpu --config=arm64-m2 --action=build 557.xz_r 505.mcf_r 531.deepsjeng_r
   ```

4. **Verify builds:**
   Binaries should appear at:
   - `benchspec/CPU/557.xz_r/exe/xz_r_base.arm64`
   - `benchspec/CPU/505.mcf_r/exe/mcf_r_base.arm64`
   - `benchspec/CPU/531.deepsjeng_r/exe/deepsjeng_r_base.arm64`

## M2Sim Integration

The existing `benchmarks/spec_runner.go` expects:
- SPEC symlinked at `benchmarks/spec` ✅ (done)
- ARM64 binaries built ❌ (pending)
- Run directories created by SPEC build ❌ (pending)

## Next Steps (for Bob)

1. Run `./install.sh` in SPEC directory
2. Create ARM64 config file
3. Build the three target benchmarks
4. Run natively on M2 to capture baseline timing
5. Create CSV with real execution times

## Benchmarks to Build

Per `spec_runner.go`:
| Benchmark | Description |
|-----------|-------------|
| 557.xz_r | Data compression (LZMA2) |
| 505.mcf_r | Vehicle scheduling (network simplex) |
| 531.deepsjeng_r | Chess engine (alpha-beta search) |

These were chosen as integer-rate benchmarks suitable for M2Sim validation.
