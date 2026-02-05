package emu_test

import (
	"encoding/binary"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sarchlab/m2sim/emu"
)

// Encoder helpers for data processing instructions.

// encodeUDIV encodes a UDIV instruction.
// Format: sf | 0 | S | 11010110 | Rm | opcode | Rn | Rd
// sf[31]: 0=32-bit, 1=64-bit
// bits [29:21] = 0b011010110 (S=0)
// opcode[15:10] = 000010 for UDIV
func encodeUDIV(rd, rn, rm uint8, is64Bit bool) uint32 {
	var sf uint32
	if is64Bit {
		sf = 1
	}
	// bits 29:21 = 011010110
	return (sf << 31) | (0b011010110 << 21) | (uint32(rm) << 16) | (0b000010 << 10) | (uint32(rn) << 5) | uint32(rd)
}

// encodeSDIV encodes a SDIV instruction.
func encodeSDIV(rd, rn, rm uint8, is64Bit bool) uint32 {
	var sf uint32
	if is64Bit {
		sf = 1
	}
	// bits 29:21 = 011010110
	return (sf << 31) | (0b011010110 << 21) | (uint32(rm) << 16) | (0b000011 << 10) | (uint32(rn) << 5) | uint32(rd)
}

// encodeMADD encodes a MADD instruction.
// Format: sf | op54 | 11011 | op31 | Rm | o0 | Ra | Rn | Rd
// MADD Rd, Rn, Rm, Ra => Rd = Ra + (Rn * Rm)
// bits [28:24] = 0b11011, o0[15] = 0 for MADD
func encodeMADD(rd, rn, rm, ra uint8, is64Bit bool) uint32 {
	var sf uint32
	if is64Bit {
		sf = 1
	}
	return (sf << 31) | (0b11011 << 24) | (uint32(rm) << 16) | (0 << 15) | (uint32(ra) << 10) | (uint32(rn) << 5) | uint32(rd)
}

// encodeMSUB encodes a MSUB instruction.
// MSUB Rd, Rn, Rm, Ra => Rd = Ra - (Rn * Rm)
func encodeMSUB(rd, rn, rm, ra uint8, is64Bit bool) uint32 {
	var sf uint32
	if is64Bit {
		sf = 1
	}
	return (sf << 31) | (0b11011 << 24) | (uint32(rm) << 16) | (1 << 15) | (uint32(ra) << 10) | (uint32(rn) << 5) | uint32(rd)
}

// encodeLSLV encodes a LSLV instruction (logical shift left by register).
// Format: sf | 0 | S | 11010110 | Rm | opcode | Rn | Rd
// opcode = 001000
func encodeLSLV(rd, rn, rm uint8, is64Bit bool) uint32 {
	var sf uint32
	if is64Bit {
		sf = 1
	}
	return (sf << 31) | (0b011010110 << 21) | (uint32(rm) << 16) | (0b001000 << 10) | (uint32(rn) << 5) | uint32(rd)
}

// encodeLSRV encodes a LSRV instruction (logical shift right by register).
func encodeLSRV(rd, rn, rm uint8, is64Bit bool) uint32 {
	var sf uint32
	if is64Bit {
		sf = 1
	}
	return (sf << 31) | (0b011010110 << 21) | (uint32(rm) << 16) | (0b001001 << 10) | (uint32(rn) << 5) | uint32(rd)
}

// encodeASRV encodes an ASRV instruction (arithmetic shift right by register).
func encodeASRV(rd, rn, rm uint8, is64Bit bool) uint32 {
	var sf uint32
	if is64Bit {
		sf = 1
	}
	return (sf << 31) | (0b011010110 << 21) | (uint32(rm) << 16) | (0b001010 << 10) | (uint32(rn) << 5) | uint32(rd)
}

// encodeRORV encodes a RORV instruction (rotate right by register).
func encodeRORV(rd, rn, rm uint8, is64Bit bool) uint32 {
	var sf uint32
	if is64Bit {
		sf = 1
	}
	return (sf << 31) | (0b011010110 << 21) | (uint32(rm) << 16) | (0b001011 << 10) | (uint32(rn) << 5) | uint32(rd)
}

func dataProcProgram(inst uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, inst)
	return buf
}

var _ = Describe("Data Processing Instructions", func() {
	var e *emu.Emulator

	BeforeEach(func() {
		e = emu.NewEmulator()
	})

	Describe("UDIV - Unsigned Division", func() {
		Context("64-bit", func() {
			It("should divide correctly", func() {
				// UDIV X0, X1, X2 => X0 = X1 / X2
				inst := encodeUDIV(0, 1, 2, true)
				program := dataProcProgram(inst)

				e.RegFile().WriteReg(1, 100)
				e.RegFile().WriteReg(2, 7)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().ReadReg(0)).To(Equal(uint64(14))) // 100 / 7 = 14
			})

			It("should return 0 for division by zero", func() {
				inst := encodeUDIV(0, 1, 2, true)
				program := dataProcProgram(inst)

				e.RegFile().WriteReg(1, 100)
				e.RegFile().WriteReg(2, 0)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().ReadReg(0)).To(Equal(uint64(0)))
			})
		})

		Context("32-bit", func() {
			It("should divide correctly using W registers", func() {
				// UDIV W0, W1, W2
				inst := encodeUDIV(0, 1, 2, false)
				program := dataProcProgram(inst)

				e.RegFile().WriteReg(1, 50)
				e.RegFile().WriteReg(2, 5)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().ReadReg(0)).To(Equal(uint64(10)))
			})

			It("should return 0 for division by zero in 32-bit", func() {
				inst := encodeUDIV(0, 1, 2, false)
				program := dataProcProgram(inst)

				e.RegFile().WriteReg(1, 50)
				e.RegFile().WriteReg(2, 0)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().ReadReg(0)).To(Equal(uint64(0)))
			})
		})
	})

	Describe("SDIV - Signed Division", func() {
		Context("64-bit", func() {
			It("should divide positive numbers", func() {
				// SDIV X0, X1, X2
				inst := encodeSDIV(0, 1, 2, true)
				program := dataProcProgram(inst)

				e.RegFile().WriteReg(1, 100)
				e.RegFile().WriteReg(2, 7)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().ReadReg(0)).To(Equal(uint64(14)))
			})

			It("should handle negative dividend", func() {
				inst := encodeSDIV(0, 1, 2, true)
				program := dataProcProgram(inst)

				neg100 := ^uint64(99) // Two's complement of -100
				e.RegFile().WriteReg(1, neg100)
				e.RegFile().WriteReg(2, 7)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(int64(e.RegFile().ReadReg(0))).To(Equal(int64(-14)))
			})

			It("should return 0 for division by zero", func() {
				inst := encodeSDIV(0, 1, 2, true)
				program := dataProcProgram(inst)

				e.RegFile().WriteReg(1, 100)
				e.RegFile().WriteReg(2, 0)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().ReadReg(0)).To(Equal(uint64(0)))
			})
		})

		Context("32-bit", func() {
			It("should handle signed division in 32-bit", func() {
				inst := encodeSDIV(0, 1, 2, false)
				program := dataProcProgram(inst)

				neg50 := uint64(^uint32(49)) // Two's complement of -50 in 32-bit
				e.RegFile().WriteReg(1, neg50)
				e.RegFile().WriteReg(2, 5)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(int32(uint32(e.RegFile().ReadReg(0)))).To(Equal(int32(-10)))
			})

			It("should return 0 for division by zero in 32-bit", func() {
				inst := encodeSDIV(0, 1, 2, false)
				program := dataProcProgram(inst)

				e.RegFile().WriteReg(1, 50)
				e.RegFile().WriteReg(2, 0)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().ReadReg(0)).To(Equal(uint64(0)))
			})
		})
	})

	Describe("MADD - Multiply-Add", func() {
		Context("64-bit", func() {
			It("should compute Ra + Rn * Rm", func() {
				// MADD X0, X1, X2, X3 => X0 = X3 + (X1 * X2)
				inst := encodeMADD(0, 1, 2, 3, true)
				program := dataProcProgram(inst)

				e.RegFile().WriteReg(1, 5)  // Rn
				e.RegFile().WriteReg(2, 7)  // Rm
				e.RegFile().WriteReg(3, 10) // Ra
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().ReadReg(0)).To(Equal(uint64(10 + 5*7))) // 45
			})

			It("should work with zero addend (MUL alias)", func() {
				// MUL is MADD with Ra = XZR (X31 or 0)
				inst := encodeMADD(0, 1, 2, 31, true)
				program := dataProcProgram(inst)

				e.RegFile().WriteReg(1, 6)
				e.RegFile().WriteReg(2, 8)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().ReadReg(0)).To(Equal(uint64(48))) // 6 * 8
			})
		})

		Context("32-bit", func() {
			It("should compute 32-bit multiply-add", func() {
				inst := encodeMADD(0, 1, 2, 3, false)
				program := dataProcProgram(inst)

				e.RegFile().WriteReg(1, 4)
				e.RegFile().WriteReg(2, 5)
				e.RegFile().WriteReg(3, 3)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().ReadReg(0)).To(Equal(uint64(3 + 4*5))) // 23
			})
		})
	})

	Describe("MSUB - Multiply-Subtract", func() {
		Context("64-bit", func() {
			It("should compute Ra - Rn * Rm", func() {
				// MSUB X0, X1, X2, X3 => X0 = X3 - (X1 * X2)
				inst := encodeMSUB(0, 1, 2, 3, true)
				program := dataProcProgram(inst)

				e.RegFile().WriteReg(1, 5)   // Rn
				e.RegFile().WriteReg(2, 7)   // Rm
				e.RegFile().WriteReg(3, 100) // Ra
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().ReadReg(0)).To(Equal(uint64(100 - 5*7))) // 65
			})

			It("should work with zero subtrahend (MNEG alias)", func() {
				// MNEG is MSUB with Ra = XZR
				inst := encodeMSUB(0, 1, 2, 31, true)
				program := dataProcProgram(inst)

				e.RegFile().WriteReg(1, 6)
				e.RegFile().WriteReg(2, 8)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				// 0 - 48 = -48 in unsigned is large number
				expected := ^uint64(47) // Two's complement of -48
				Expect(e.RegFile().ReadReg(0)).To(Equal(expected))
			})
		})

		Context("32-bit", func() {
			It("should compute 32-bit multiply-subtract", func() {
				inst := encodeMSUB(0, 1, 2, 3, false)
				program := dataProcProgram(inst)

				e.RegFile().WriteReg(1, 3)
				e.RegFile().WriteReg(2, 4)
				e.RegFile().WriteReg(3, 50)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().ReadReg(0)).To(Equal(uint64(50 - 3*4))) // 38
			})
		})
	})

	Describe("Variable Shift Instructions", func() {
		Describe("LSLV - Logical Shift Left Variable", func() {
			It("should shift left by register amount (64-bit)", func() {
				inst := encodeLSLV(0, 1, 2, true)
				program := dataProcProgram(inst)

				e.RegFile().WriteReg(1, 0x1)
				e.RegFile().WriteReg(2, 4) // shift by 4
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().ReadReg(0)).To(Equal(uint64(0x10)))
			})

			It("should mask shift amount to 6 bits (64-bit)", func() {
				inst := encodeLSLV(0, 1, 2, true)
				program := dataProcProgram(inst)

				e.RegFile().WriteReg(1, 0x1)
				e.RegFile().WriteReg(2, 64) // should be masked to 0
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().ReadReg(0)).To(Equal(uint64(0x1))) // no shift
			})

			It("should shift left by register amount (32-bit)", func() {
				inst := encodeLSLV(0, 1, 2, false)
				program := dataProcProgram(inst)

				e.RegFile().WriteReg(1, 0x1)
				e.RegFile().WriteReg(2, 8)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().ReadReg(0)).To(Equal(uint64(0x100)))
			})
		})

		Describe("LSRV - Logical Shift Right Variable", func() {
			It("should shift right by register amount (64-bit)", func() {
				inst := encodeLSRV(0, 1, 2, true)
				program := dataProcProgram(inst)

				e.RegFile().WriteReg(1, 0x100)
				e.RegFile().WriteReg(2, 4)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().ReadReg(0)).To(Equal(uint64(0x10)))
			})

			It("should shift right (32-bit)", func() {
				inst := encodeLSRV(0, 1, 2, false)
				program := dataProcProgram(inst)

				e.RegFile().WriteReg(1, 0x80)
				e.RegFile().WriteReg(2, 3)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().ReadReg(0)).To(Equal(uint64(0x10)))
			})
		})

		Describe("ASRV - Arithmetic Shift Right Variable", func() {
			It("should preserve sign bit (64-bit)", func() {
				inst := encodeASRV(0, 1, 2, true)
				program := dataProcProgram(inst)

				e.RegFile().WriteReg(1, 0x8000000000000000) // negative
				e.RegFile().WriteReg(2, 4)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				// Sign bit should be preserved
				Expect(e.RegFile().ReadReg(0) >> 63).To(Equal(uint64(1)))
			})

			It("should preserve sign bit (32-bit)", func() {
				inst := encodeASRV(0, 1, 2, false)
				program := dataProcProgram(inst)

				e.RegFile().WriteReg(1, 0x80000000) // negative in 32-bit
				e.RegFile().WriteReg(2, 4)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				// Sign bit should be preserved in low 32 bits
				Expect((e.RegFile().ReadReg(0) >> 31) & 1).To(Equal(uint64(1)))
			})
		})

		Describe("RORV - Rotate Right Variable", func() {
			It("should rotate right (64-bit)", func() {
				inst := encodeRORV(0, 1, 2, true)
				program := dataProcProgram(inst)

				e.RegFile().WriteReg(1, 0x1)
				e.RegFile().WriteReg(2, 1) // rotate by 1
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().ReadReg(0)).To(Equal(uint64(0x8000000000000000)))
			})

			It("should rotate right (32-bit)", func() {
				inst := encodeRORV(0, 1, 2, false)
				program := dataProcProgram(inst)

				e.RegFile().WriteReg(1, 0x1)
				e.RegFile().WriteReg(2, 1)
				e.LoadProgram(0x1000, program)

				result := e.Step()

				Expect(result.Err).To(BeNil())
				Expect(e.RegFile().ReadReg(0)).To(Equal(uint64(0x80000000)))
			})
		})
	})
})
