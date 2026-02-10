# ARM64 Instruction Support Status

This document tracks ARM64 instructions and syscalls supported by M2Sim.

## Supported Instructions

### Data Processing (Immediate)

| Instruction | Description | Decoder | Emulator |
|-------------|-------------|---------|----------|
| ADD (imm)   | Add with immediate | ✅ | ✅ |
| ADDS (imm)  | Add with immediate, set flags | ✅ | ✅ |
| SUB (imm)   | Subtract with immediate | ✅ | ✅ |
| SUBS (imm)  | Subtract with immediate, set flags | ✅ | ✅ |

### Logical (Immediate)

| Instruction | Description | Decoder | Emulator |
|-------------|-------------|---------|----------|
| AND (imm)   | Bitwise AND with bitmask immediate | ✅ | ✅ |
| ANDS (imm)  | Bitwise AND with bitmask immediate, set flags | ✅ | ✅ |
| ORR (imm)   | Bitwise OR with bitmask immediate | ✅ | ✅ |
| EOR (imm)   | Bitwise XOR with bitmask immediate | ✅ | ✅ |

### Data Processing (Register)

| Instruction | Description | Decoder | Emulator |
|-------------|-------------|---------|----------|
| ADD (reg)   | Add registers | ✅ | ✅ |
| ADDS (reg)  | Add registers, set flags | ✅ | ✅ |
| SUB (reg)   | Subtract registers | ✅ | ✅ |
| SUBS (reg)  | Subtract registers, set flags | ✅ | ✅ |
| AND (reg)   | Bitwise AND | ✅ | ✅ |
| ANDS (reg)  | Bitwise AND, set flags | ✅ | ✅ |
| BIC (reg)   | Bitwise bit clear (AND NOT) | ✅ | ✅ |
| BICS (reg)  | Bitwise bit clear, set flags | ✅ | ✅ |
| ORR (reg)   | Bitwise OR | ✅ | ✅ |
| ORN (reg)   | Bitwise OR NOT | ✅ | ✅ |
| EOR (reg)   | Bitwise XOR | ✅ | ✅ |
| EON (reg)   | Bitwise exclusive OR NOT | ✅ | ✅ |

### Data Processing (2 Source)

| Instruction | Description | Decoder | Emulator |
|-------------|-------------|---------|----------|
| UDIV        | Unsigned divide | ✅ | ✅ |
| SDIV        | Signed divide | ✅ | ✅ |
| LSLV        | Logical shift left (variable) | ✅ | ✅ |
| LSRV        | Logical shift right (variable) | ✅ | ✅ |
| ASRV        | Arithmetic shift right (variable) | ✅ | ✅ |
| RORV        | Rotate right (variable) | ✅ | ✅ |

### Data Processing (3 Source)

| Instruction | Description | Decoder | Emulator |
|-------------|-------------|---------|----------|
| MADD        | Multiply-add (Rd = Ra + Rn * Rm) | ✅ | ✅ |
| MSUB        | Multiply-subtract (Rd = Ra - Rn * Rm) | ✅ | ✅ |

### Bitfield Instructions

| Instruction | Description | Decoder | Emulator |
|-------------|-------------|---------|----------|
| SBFM        | Signed bitfield move (ASR imm, SXTB, SXTH, SXTW, etc.) | ✅ | ✅ |
| BFM         | Bitfield move (BFI, BFXIL) | ✅ | ✅ |
| UBFM        | Unsigned bitfield move (LSL imm, LSR imm, UXTB, UXTH, etc.) | ✅ | ✅ |
| EXTR        | Extract register (bitfield from register pair) | ✅ | ✅ |

### Move Wide Instructions

| Instruction | Description | Decoder | Emulator |
|-------------|-------------|---------|----------|
| MOVZ        | Move wide with zero | ✅ | ✅ |
| MOVN        | Move wide with NOT | ✅ | ✅ |
| MOVK        | Move wide with keep | ✅ | ✅ |

### PC-Relative Addressing

| Instruction | Description | Decoder | Emulator |
|-------------|-------------|---------|----------|
| ADR         | PC-relative address | ✅ | ✅ |
| ADRP        | PC-relative page address | ✅ | ✅ |

### Conditional Select Instructions

| Instruction | Description | Decoder | Emulator |
|-------------|-------------|---------|----------|
| CSEL        | Conditional select | ✅ | ✅ |
| CSINC       | Conditional select increment | ✅ | ✅ |
| CSINV       | Conditional select invert | ✅ | ✅ |
| CSNEG       | Conditional select negate | ✅ | ✅ |

### Conditional Compare Instructions

| Instruction | Description | Decoder | Emulator |
|-------------|-------------|---------|----------|
| CCMN        | Conditional compare negative | ✅ | ✅ |
| CCMP        | Conditional compare | ✅ | ✅ |

### Branch Instructions

| Instruction | Description | Decoder | Emulator |
|-------------|-------------|---------|----------|
| B           | Unconditional branch | ✅ | ✅ |
| BL          | Branch with link | ✅ | ✅ |
| B.cond      | Conditional branch | ✅ | ✅ |
| BR          | Branch to register | ✅ | ✅ |
| BLR         | Branch with link to register | ✅ | ✅ |
| RET         | Return from subroutine | ✅ | ✅ |
| TBZ         | Test bit and branch if zero | ✅ | ✅ |
| TBNZ        | Test bit and branch if not zero | ✅ | ✅ |
| CBZ         | Compare and branch if zero | ✅ | ✅ |
| CBNZ        | Compare and branch if not zero | ✅ | ✅ |

### Exception Generation

| Instruction | Description | Decoder | Emulator |
|-------------|-------------|---------|----------|
| SVC         | Supervisor call (syscall trigger) | ✅ | ✅ |
| BRK         | Breakpoint (software debug trap) | ✅ | ✅ |

### Load/Store Instructions

| Instruction | Description | Decoder | Emulator |
|-------------|-------------|---------|----------|
| LDR (imm, 64-bit) | Load 64-bit register | ✅ | ✅ |
| LDR (imm, 32-bit) | Load 32-bit register (zero-extend) | ✅ | ✅ |
| LDR (literal)      | PC-relative load | ✅ | ✅ |
| STR (imm, 64-bit) | Store 64-bit register | ✅ | ✅ |
| STR (imm, 32-bit) | Store 32-bit register | ✅ | ✅ |
| LDP         | Load pair of registers | ✅ | ✅ |
| STP         | Store pair of registers | ✅ | ✅ |
| LDRB        | Load register byte (zero-extend) | ✅ | ✅ |
| STRB        | Store register byte | ✅ | ✅ |
| LDRSB       | Load register signed byte | ✅ | ✅ |
| LDRH        | Load register halfword (zero-extend) | ✅ | ✅ |
| STRH        | Store register halfword | ✅ | ✅ |
| LDRSH       | Load register signed halfword | ✅ | ✅ |
| LDRSW       | Load register signed word (32-bit to 64-bit) | ✅ | ✅ |

### SIMD Integer Instructions

| Instruction | Description | Decoder | Emulator |
|-------------|-------------|---------|----------|
| VADD        | Vector add | ✅ | ✅ |
| VSUB        | Vector subtract | ✅ | ✅ |
| VMUL        | Vector multiply | ✅ | ✅ |

### SIMD Floating-Point Instructions

| Instruction | Description | Decoder | Emulator |
|-------------|-------------|---------|----------|
| VFADD       | Vector FP add | ✅ | ✅ |
| VFSUB       | Vector FP subtract | ✅ | ✅ |
| VFMUL       | Vector FP multiply | ✅ | ✅ |

### SIMD Load/Store Instructions

| Instruction | Description | Decoder | Emulator |
|-------------|-------------|---------|----------|
| LDR Q       | Load 128-bit vector register | ✅ | ✅ |
| STR Q       | Store 128-bit vector register | ✅ | ✅ |

### SIMD Copy Instructions

| Instruction | Description | Decoder | Emulator |
|-------------|-------------|---------|----------|
| DUP         | Duplicate scalar to vector | ✅ | ✅ |
| VMOV        | Vector move | ✅ | ✅ |

### System Instructions

| Instruction | Description | Decoder | Emulator |
|-------------|-------------|---------|----------|
| MRS         | Move from system register | ✅ | ✅ |
| NOP         | No operation | ✅ | ✅ |

**Supported System Registers:**
- **DCZID_EL0**: Data Cache Zero ID register - Returns cache line size information (64-byte cache lines)

## Supported Syscalls

The emu package implements the full set of syscalls. The driver package provides a minimal subset (exit, write only).

### Emulator Syscalls (emu package)

| Syscall | Number | Description |
|---------|--------|-------------|
| openat  | 56     | Open file relative to directory fd |
| close   | 57     | Close file descriptor |
| lseek   | 62     | Reposition file offset |
| read    | 63     | Read from file descriptor |
| write   | 64     | Write to file descriptor |
| fstat   | 80     | Get file status |
| exit    | 93     | Terminate program with exit code |
| brk     | 214    | Change data segment size |
| mmap    | 222    | Map memory pages |
| mprotect| 226    | Set memory protection |

### Driver Syscalls (driver package, minimal)

| Syscall | Number | Description |
|---------|--------|-------------|
| exit    | 93     | Terminate program with exit code |
| write   | 64     | Write buffer to file descriptor |

### Syscall Convention (ARM64 Linux)
- Syscall number in X8
- Arguments in X0-X5
- Return value in X0

## Condition Codes Supported

| Code | Meaning | Condition |
|------|---------|-----------|
| EQ | Equal | Z == 1 |
| NE | Not equal | Z == 0 |
| CS/HS | Carry set / Unsigned higher or same | C == 1 |
| CC/LO | Carry clear / Unsigned lower | C == 0 |
| MI | Minus / Negative | N == 1 |
| PL | Plus / Positive or zero | N == 0 |
| VS | Overflow | V == 1 |
| VC | No overflow | V == 0 |
| HI | Unsigned higher | C == 1 && Z == 0 |
| LS | Unsigned lower or same | C == 0 || Z == 1 |
| GE | Signed greater than or equal | N == V |
| LT | Signed less than | N != V |
| GT | Signed greater than | Z == 0 && N == V |
| LE | Signed less than or equal | Z == 1 || N != V |
| AL | Always | (unconditional) |

## Instruction Formats Supported

- **FormatDPImm**: Data Processing with Immediate
- **FormatDPReg**: Data Processing with Register
- **FormatLogicalImm**: Logical with Bitmask Immediate
- **FormatBranch**: Unconditional Branch (Immediate)
- **FormatBranchCond**: Conditional Branch
- **FormatBranchReg**: Branch to Register
- **FormatTestBranch**: Test and Branch (TBZ, TBNZ)
- **FormatCompareBranch**: Compare and Branch (CBZ, CBNZ)
- **FormatLoadStore**: Load/Store with Immediate Offset
- **FormatLoadStorePair**: Load/Store Pair
- **FormatLoadStoreLit**: Load (PC-relative literal)
- **FormatPCRel**: PC-Relative Addressing (ADR, ADRP)
- **FormatMoveWide**: Move Wide (MOVZ, MOVN, MOVK)
- **FormatCondSelect**: Conditional Select (CSEL, CSINC, CSINV, CSNEG)
- **FormatCondCmp**: Conditional Compare (CCMN, CCMP)
- **FormatDataProc2Src**: Data Processing 2 Source (UDIV, SDIV, shifts)
- **FormatDataProc3Src**: Data Processing 3 Source (MADD, MSUB)
- **FormatBitfield**: Bitfield Operations (SBFM, BFM, UBFM)
- **FormatExtract**: Extract (EXTR)
- **FormatSIMDReg**: SIMD Data Processing (Register)
- **FormatSIMDLoadStore**: SIMD Load/Store
- **FormatSIMDCopy**: SIMD Copy (DUP, MOV, etc.)
- **FormatSystemReg**: System Register Operations (MRS, MSR)

## Known Limitations

### Missing Test Coverage

The following areas lack test coverage:

1. **Shifted register operands** - No tests verify correct decoding of shift
   type (LSL, LSR, ASR, ROR) and shift amount for register operands in
   ADD/SUB/logical instructions.

---

*Last updated: 2026-02-10*
*Consolidated from root SUPPORTED.md and insts/SUPPORTED.md*
