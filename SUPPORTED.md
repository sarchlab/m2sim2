# ARM64 Decoder - Supported Instructions

This document lists the ARM64 instructions currently supported by the decoder,
along with known limitations.

## Supported Instructions

### Data Processing (Immediate)

| Instruction | Description |
|-------------|-------------|
| ADD/ADDS    | Add (immediate) |
| SUB/SUBS    | Subtract (immediate) |

### Data Processing (Register)

| Instruction | Description |
|-------------|-------------|
| ADD/ADDS    | Add (shifted register) |
| SUB/SUBS    | Subtract (shifted register) |
| AND/ANDS    | Bitwise AND |
| ORR         | Bitwise inclusive OR |
| EOR         | Bitwise exclusive OR |

### Branches

| Instruction | Description |
|-------------|-------------|
| B           | Branch (unconditional) |
| BL          | Branch with link |
| B.cond      | Branch (conditional) |
| BR          | Branch to register |
| BLR         | Branch with link to register |
| RET         | Return from subroutine |

## Known Limitations

### Logical Register Operations - N-bit Not Handled

**Issue:** The N-bit (bit 21) in logical register instructions is not currently
decoded. This bit controls whether the second operand (Rm) is inverted before
the operation.

**Affected Instructions:**
- **BIC** (Bitwise Bit Clear) - decoded incorrectly as AND
- **ORN** (Bitwise OR NOT) - decoded incorrectly as ORR  
- **EON** (Bitwise Exclusive OR NOT) - decoded incorrectly as EOR

**Encoding Reference:**
```
Logical (shifted register): sf | opc | 01010 | shift | N | Rm | imm6 | Rn | Rd
                                                       ^
                                                   bit 21 (N)
```

When N=1, the Rm value should be bitwise inverted (~Rm) before the logical
operation is applied. The current decoder ignores this bit.

**Status:** Not implemented. Tracked for future work.

### Missing Test Coverage

The following areas lack test coverage:

1. **Shifted register operands** - No tests verify correct decoding of shift
   type (LSL, LSR, ASR, ROR) and shift amount for register operands in
   ADD/SUB/logical instructions.

## Version History

- **PR #7** - Initial ARM64 decoder implementation
