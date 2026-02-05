package emu_test

import (
	"encoding/binary"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sarchlab/m2sim/emu"
)

// Encoder helpers for conditional compare instructions.

// encodeCCMP encodes a CCMP (conditional compare) instruction.
// Format: sf | op | 1 | 11010010 | Rm/imm5 | cond | 1/0 | o2 | Rn | o3 | nzcv
// op[30]: 1 for CCMP (subtract)
// imm[11]: 0=register form, 1=immediate form
func encodeCCMP(rn, rm uint8, cond uint8, nzcv uint8, is64Bit bool, isImm bool) uint32 {
	var sf uint32
	if is64Bit {
		sf = 1
	}
	var immBit uint32
	if isImm {
		immBit = 1
	}
	// sf | op=1 | 1 | 11010010 | Rm/imm5 | cond | immBit | o2=0 | Rn | o3=0 | nzcv
	return (sf << 31) | (1 << 30) | (1 << 29) | (0b11010010 << 21) |
		(uint32(rm) << 16) | (uint32(cond) << 12) | (immBit << 11) |
		(uint32(rn) << 5) | uint32(nzcv)
}

// encodeCCMN encodes a CCMN (conditional compare negative) instruction.
// op[30]: 0 for CCMN (add)
func encodeCCMN(rn, rm uint8, cond uint8, nzcv uint8, is64Bit bool, isImm bool) uint32 {
	var sf uint32
	if is64Bit {
		sf = 1
	}
	var immBit uint32
	if isImm {
		immBit = 1
	}
	// sf | op=0 | 1 | 11010010 | Rm/imm5 | cond | immBit | o2=0 | Rn | o3=0 | nzcv
	return (sf << 31) | (0 << 30) | (1 << 29) | (0b11010010 << 21) |
		(uint32(rm) << 16) | (uint32(cond) << 12) | (immBit << 11) |
		(uint32(rn) << 5) | uint32(nzcv)
}

// Condition codes
const (
	CondEQ_CC = 0b0000 // Equal (Z==1)
	CondNE_CC = 0b0001 // Not equal (Z==0)
	CondGE_CC = 0b1010 // Signed greater or equal (N==V)
	CondLT_CC = 0b1011 // Signed less than (N!=V)
	CondAL_CC = 0b1110 // Always
)

func condCmpProgram(inst uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, inst)
	return buf
}

var _ = Describe("Conditional Compare Instructions", func() {
	var e *emu.Emulator

	BeforeEach(func() {
		e = emu.NewEmulator()
	})

	Describe("CCMP (Conditional Compare)", func() {
		Context("64-bit register form", func() {
			It("should compare when condition is true (Z flag set)", func() {
				// Set up: X1=100, X2=50, condition EQ with Z=1
				e.RegFile().WriteReg(1, 100)
				e.RegFile().WriteReg(2, 50)
				e.RegFile().PSTATE.Z = true // Make EQ condition true

				// CCMP X1, X2, #0, EQ
				// If EQ is true: compare X1 - X2 (100-50=50)
				// Result: N=0, Z=0, C=1 (no borrow), V=0
				inst := encodeCCMP(1, 2, CondEQ_CC, 0b0000, true, false)
				program := condCmpProgram(inst)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().PSTATE.N).To(BeFalse())
				Expect(e.RegFile().PSTATE.Z).To(BeFalse())
				Expect(e.RegFile().PSTATE.C).To(BeTrue()) // 100 >= 50
				Expect(e.RegFile().PSTATE.V).To(BeFalse())
			})

			It("should set nzcv when condition is false", func() {
				// Set up: condition EQ with Z=0 (false)
				e.RegFile().WriteReg(1, 100)
				e.RegFile().WriteReg(2, 50)
				e.RegFile().PSTATE.Z = false // Make EQ condition false

				// CCMP X1, X2, #0b1010, EQ
				// If EQ is false: set flags to nzcv=1010 (N=1, Z=0, C=1, V=0)
				inst := encodeCCMP(1, 2, CondEQ_CC, 0b1010, true, false)
				program := condCmpProgram(inst)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().PSTATE.N).To(BeTrue())  // nzcv bit 3
				Expect(e.RegFile().PSTATE.Z).To(BeFalse()) // nzcv bit 2
				Expect(e.RegFile().PSTATE.C).To(BeTrue())  // nzcv bit 1
				Expect(e.RegFile().PSTATE.V).To(BeFalse()) // nzcv bit 0
			})

			It("should set Z flag when operands are equal", func() {
				e.RegFile().WriteReg(1, 42)
				e.RegFile().WriteReg(2, 42)
				e.RegFile().PSTATE.Z = true // EQ condition true

				// CCMP X1, X2, #0, EQ => 42-42=0, Z=1
				inst := encodeCCMP(1, 2, CondEQ_CC, 0, true, false)
				program := condCmpProgram(inst)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().PSTATE.Z).To(BeTrue())
				Expect(e.RegFile().PSTATE.C).To(BeTrue()) // 42 >= 42
			})

			It("should set N flag for negative result", func() {
				e.RegFile().WriteReg(1, 50)
				e.RegFile().WriteReg(2, 100)
				e.RegFile().PSTATE.Z = true // EQ condition true

				// CCMP X1, X2, #0, EQ => 50-100=-50, N=1
				inst := encodeCCMP(1, 2, CondEQ_CC, 0, true, false)
				program := condCmpProgram(inst)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().PSTATE.N).To(BeTrue())
				Expect(e.RegFile().PSTATE.C).To(BeFalse()) // 50 < 100 (borrow)
			})
		})

		Context("64-bit immediate form", func() {
			It("should compare with immediate when condition is true", func() {
				e.RegFile().WriteReg(1, 100)
				e.RegFile().PSTATE.Z = true // EQ condition true

				// CCMP X1, #10, #0, EQ => 100-10=90
				inst := encodeCCMP(1, 10, CondEQ_CC, 0, true, true)
				program := condCmpProgram(inst)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().PSTATE.N).To(BeFalse())
				Expect(e.RegFile().PSTATE.Z).To(BeFalse())
				Expect(e.RegFile().PSTATE.C).To(BeTrue()) // 100 >= 10
			})
		})

		Context("32-bit form", func() {
			It("should compare 32-bit values when condition is true", func() {
				e.RegFile().WriteReg(1, 0xFFFFFFFF00000064) // Lower 32 bits = 100
				e.RegFile().WriteReg(2, 0xFFFFFFFF00000032) // Lower 32 bits = 50
				e.RegFile().PSTATE.Z = true

				// CCMP W1, W2, #0, EQ (32-bit)
				inst := encodeCCMP(1, 2, CondEQ_CC, 0, false, false)
				program := condCmpProgram(inst)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().PSTATE.N).To(BeFalse())
				Expect(e.RegFile().PSTATE.Z).To(BeFalse())
				Expect(e.RegFile().PSTATE.C).To(BeTrue())
			})

			It("should set 32-bit N flag correctly", func() {
				e.RegFile().WriteReg(1, 0x80000000) // Negative in 32-bit
				e.RegFile().WriteReg(2, 1)
				e.RegFile().PSTATE.Z = true

				// CCMP W1, W2, #0, EQ
				// 0x80000000 - 1 = 0x7FFFFFFF (positive in 32-bit)
				inst := encodeCCMP(1, 2, CondEQ_CC, 0, false, false)
				program := condCmpProgram(inst)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().PSTATE.N).To(BeFalse()) // 0x7FFFFFFF is positive
				Expect(e.RegFile().PSTATE.V).To(BeTrue())  // Signed overflow
			})
		})
	})

	Describe("CCMN (Conditional Compare Negative)", func() {
		Context("64-bit register form", func() {
			It("should add when condition is true", func() {
				e.RegFile().WriteReg(1, 100)
				e.RegFile().WriteReg(2, 50)
				e.RegFile().PSTATE.Z = true // EQ condition true

				// CCMN X1, X2, #0, EQ => 100+50=150, flags like ADDS
				inst := encodeCCMN(1, 2, CondEQ_CC, 0, true, false)
				program := condCmpProgram(inst)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().PSTATE.N).To(BeFalse())
				Expect(e.RegFile().PSTATE.Z).To(BeFalse())
				Expect(e.RegFile().PSTATE.C).To(BeFalse()) // No carry
			})

			It("should set nzcv when condition is false", func() {
				e.RegFile().WriteReg(1, 100)
				e.RegFile().WriteReg(2, 50)
				e.RegFile().PSTATE.Z = false // EQ condition false

				// CCMN X1, X2, #0b1111, EQ
				// If EQ is false: set flags to nzcv=1111
				inst := encodeCCMN(1, 2, CondEQ_CC, 0b1111, true, false)
				program := condCmpProgram(inst)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().PSTATE.N).To(BeTrue())
				Expect(e.RegFile().PSTATE.Z).To(BeTrue())
				Expect(e.RegFile().PSTATE.C).To(BeTrue())
				Expect(e.RegFile().PSTATE.V).To(BeTrue())
			})

			It("should detect unsigned overflow (carry) in addition", func() {
				e.RegFile().WriteReg(1, 0xFFFFFFFFFFFFFFFF) // Max uint64
				e.RegFile().WriteReg(2, 1)
				e.RegFile().PSTATE.Z = true

				// CCMN X1, X2, #0, EQ => 0xFFFF...FF + 1 = 0, C=1 (overflow)
				inst := encodeCCMN(1, 2, CondEQ_CC, 0, true, false)
				program := condCmpProgram(inst)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().PSTATE.Z).To(BeTrue()) // Result is 0
				Expect(e.RegFile().PSTATE.C).To(BeTrue()) // Unsigned overflow
			})
		})

		Context("64-bit immediate form", func() {
			It("should add with immediate when condition is true", func() {
				e.RegFile().WriteReg(1, 100)
				e.RegFile().PSTATE.Z = true

				// CCMN X1, #5, #0, EQ => 100+5=105
				inst := encodeCCMN(1, 5, CondEQ_CC, 0, true, true)
				program := condCmpProgram(inst)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().PSTATE.N).To(BeFalse())
				Expect(e.RegFile().PSTATE.Z).To(BeFalse())
			})
		})

		Context("32-bit form", func() {
			It("should add 32-bit values when condition is true", func() {
				e.RegFile().WriteReg(1, 100)
				e.RegFile().WriteReg(2, 50)
				e.RegFile().PSTATE.Z = true

				// CCMN W1, W2, #0, EQ (32-bit)
				inst := encodeCCMN(1, 2, CondEQ_CC, 0, false, false)
				program := condCmpProgram(inst)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().PSTATE.N).To(BeFalse())
				Expect(e.RegFile().PSTATE.Z).To(BeFalse())
			})

			It("should detect 32-bit unsigned overflow", func() {
				e.RegFile().WriteReg(1, 0xFFFFFFFF) // Max uint32
				e.RegFile().WriteReg(2, 1)
				e.RegFile().PSTATE.Z = true

				// CCMN W1, W2, #0, EQ => 0xFFFFFFFF + 1 = 0, C=1
				inst := encodeCCMN(1, 2, CondEQ_CC, 0, false, false)
				program := condCmpProgram(inst)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().PSTATE.Z).To(BeTrue()) // Result is 0
				Expect(e.RegFile().PSTATE.C).To(BeTrue()) // Unsigned overflow
			})
		})
	})

	Describe("Different conditions", func() {
		It("should handle NE condition (not equal)", func() {
			e.RegFile().WriteReg(1, 100)
			e.RegFile().WriteReg(2, 50)
			e.RegFile().PSTATE.Z = false // NE condition true (Z=0)

			// CCMP X1, X2, #0, NE
			inst := encodeCCMP(1, 2, CondNE_CC, 0, true, false)
			program := condCmpProgram(inst)
			e.LoadProgram(0x1000, program)

			result := e.Step()

			Expect(result.Err).To(BeNil())
			// NE is true (Z=0), so comparison happens
			Expect(e.RegFile().PSTATE.C).To(BeTrue()) // 100 >= 50
		})

		It("should handle GE condition (signed greater or equal)", func() {
			e.RegFile().WriteReg(1, 100)
			e.RegFile().WriteReg(2, 50)
			e.RegFile().PSTATE.N = false
			e.RegFile().PSTATE.V = false // GE is true (N==V)

			// CCMP X1, X2, #0, GE
			inst := encodeCCMP(1, 2, CondGE_CC, 0, true, false)
			program := condCmpProgram(inst)
			e.LoadProgram(0x1000, program)

			result := e.Step()

			Expect(result.Err).To(BeNil())
			// GE is true, comparison happens
			Expect(e.RegFile().PSTATE.C).To(BeTrue())
		})

		It("should handle AL condition (always)", func() {
			e.RegFile().WriteReg(1, 100)
			e.RegFile().WriteReg(2, 50)

			// CCMP X1, X2, #0b1111, AL
			// AL always evaluates to true, so comparison always happens
			inst := encodeCCMP(1, 2, CondAL_CC, 0b1111, true, false)
			program := condCmpProgram(inst)
			e.LoadProgram(0x1000, program)

			result := e.Step()

			Expect(result.Err).To(BeNil())
			// AL is true, comparison happens
			Expect(e.RegFile().PSTATE.C).To(BeTrue()) // 100 >= 50
		})
	})
})
