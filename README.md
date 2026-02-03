# M2Sim

A cycle-accurate Apple M2 CPU simulator built on the [Akita](https://github.com/sarchlab/akita) simulation framework.

## Current Status

🟢 **Milestone 3 (Timing Model) Complete** — Now working on M4 (Cache Hierarchy)

| Milestone | Status |
|-----------|--------|
| M1: Foundation | ✅ Complete |
| M2: Memory & Control Flow | ✅ Complete |
| M3: Timing Model | ✅ Complete |
| M4: Cache Hierarchy | 🔄 In Progress |
| M5: Advanced Features | ⏳ Planned |
| M6: Validation | ⏳ Planned |

See [SPEC.md](SPEC.md) for detailed milestone definitions.

## Overview

M2Sim provides both functional emulation and timing simulation for ARM64 user-space programs. It can:

- Execute ARM64 binaries correctly (functional emulation)
- Predict execution time with cycle-level accuracy (timing simulation)

The simulator is designed for computer architecture research, enabling detailed analysis of CPU behavior without requiring access to physical hardware.

## Features

### Functional Emulation
- ARM64 instruction set support (see [SUPPORTED.md](SUPPORTED.md) for details)
- Linux syscall emulation (`exit`, `write`)
- ELF binary loading

### Timing Simulation
- 5-stage pipeline (Fetch, Decode, Execute, Memory, Writeback)
- Instruction latency modeling
- L1/L2 cache hierarchy
- Data forwarding and hazard detection

## Installation

### Prerequisites
- Go 1.21 or later
- Git

### Building from Source

```bash
# Clone the repository
git clone https://github.com/sarchlab/m2sim.git
cd m2sim

# Build all packages
go build ./...

# Run tests to verify installation
go test ./...
```

## Usage

### Running a Simulation

```bash
# Build the simulator
go build -o m2sim ./cmd/m2sim

# Run a program (functional emulation)
./m2sim path/to/program.elf

# Run with timing simulation
./m2sim --timing path/to/program.elf
```

### Compiling Test Programs

To compile ARM64 programs for the simulator:

```bash
# On macOS with Apple Silicon
clang -o program program.c

# Cross-compilation from other platforms
clang -target arm64-apple-macos -o program program.c
```

## Project Structure

```
m2sim/
├── emu/           # Functional ARM64 emulator
├── timing/        # Cycle-accurate timing model
│   ├── core/      # CPU core timing
│   ├── cache/     # L1/L2 cache hierarchy
│   ├── latency/   # Instruction latency modeling
│   ├── mem/       # Memory timing model
│   └── pipeline/  # 5-stage pipeline implementation
├── insts/         # ARM64 instruction definitions and decoder
├── driver/        # OS service emulation (syscalls)
├── loader/        # ELF binary loader
├── benchmarks/    # Test programs and validation
├── samples/       # Example programs
└── cmd/m2sim/     # Command-line interface
```

## Documentation

- [SPEC.md](SPEC.md) - Project specification and milestones
- [CLAUDE.md](CLAUDE.md) - Development guidelines
- [DESIGN.md](DESIGN.md) - Design philosophy
- [SUPPORTED.md](SUPPORTED.md) - Supported ARM64 instructions

## Testing

```bash
# Run all tests
go test ./...

# Run tests with Ginkgo (more detailed output)
ginkgo -r

# Run specific package tests
go test ./emu/... -v
```

## Related Projects

- [Akita](https://github.com/sarchlab/akita) - Simulation framework
- [MGPUSim](https://github.com/sarchlab/mgpusim) - GPU simulator using Akita

## License

This project is developed by the [SARCH Lab](https://github.com/sarchlab).
