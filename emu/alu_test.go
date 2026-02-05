package emu_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sarchlab/m2sim/emu"
)

var _ = Describe("ALU", func() {
	var (
		regFile *emu.RegFile
		alu     *emu.ALU
	)

	BeforeEach(func() {
		regFile = &emu.RegFile{}
		alu = emu.NewALU(regFile)
	})

	Describe("ADD (64-bit)", func() {
		Context("register form", func() {
			It("should add two registers", func() {
				regFile.WriteReg(1, 10)
				regFile.WriteReg(2, 20)

				alu.ADD64(0, 1, 2, false)

				Expect(regFile.ReadReg(0)).To(Equal(uint64(30)))
			})

			It("should handle XZR as source", func() {
				regFile.WriteReg(1, 100)

				alu.ADD64(0, 1, 31, false)

				Expect(regFile.ReadReg(0)).To(Equal(uint64(100)))
			})

			It("should handle XZR as destination (discard)", func() {
				regFile.WriteReg(1, 10)
				regFile.WriteReg(2, 20)

				alu.ADD64(31, 1, 2, false)

				Expect(regFile.ReadReg(31)).To(Equal(uint64(0)))
			})

			It("should handle overflow wrapping", func() {
				regFile.WriteReg(1, 0xFFFFFFFFFFFFFFFF)
				regFile.WriteReg(2, 1)

				alu.ADD64(0, 1, 2, false)

				Expect(regFile.ReadReg(0)).To(Equal(uint64(0)))
			})
		})

		Context("immediate form", func() {
			It("should add register and immediate", func() {
				regFile.WriteReg(1, 100)

				alu.ADD64Imm(0, 1, 50, false)

				Expect(regFile.ReadReg(0)).To(Equal(uint64(150)))
			})

			It("should add with shifted immediate (LSL #12)", func() {
				regFile.WriteReg(1, 0)

				alu.ADD64ImmShifted(0, 1, 1, 12, false)

				Expect(regFile.ReadReg(0)).To(Equal(uint64(0x1000)))
			})
		})

		Context("with flag setting (ADDS)", func() {
			It("should set Z flag when result is zero", func() {
				regFile.WriteReg(1, 0)
				regFile.WriteReg(2, 0)

				alu.ADD64(0, 1, 2, true)

				Expect(regFile.PSTATE.Z).To(BeTrue())
				Expect(regFile.PSTATE.N).To(BeFalse())
			})

			It("should set N flag when result is negative", func() {
				regFile.WriteReg(1, 0x8000000000000000)
				regFile.WriteReg(2, 0)

				alu.ADD64(0, 1, 2, true)

				Expect(regFile.PSTATE.N).To(BeTrue())
				Expect(regFile.PSTATE.Z).To(BeFalse())
			})

			It("should set C flag on unsigned overflow", func() {
				regFile.WriteReg(1, 0xFFFFFFFFFFFFFFFF)
				regFile.WriteReg(2, 1)

				alu.ADD64(0, 1, 2, true)

				Expect(regFile.PSTATE.C).To(BeTrue())
			})

			It("should set V flag on signed overflow", func() {
				regFile.WriteReg(1, 0x7FFFFFFFFFFFFFFF) // max positive
				regFile.WriteReg(2, 1)

				alu.ADD64(0, 1, 2, true)

				Expect(regFile.PSTATE.V).To(BeTrue())
			})
		})
	})

	Describe("ADD (32-bit)", func() {
		It("should add two 32-bit values and zero-extend", func() {
			regFile.WriteReg(1, 0xFFFFFFFF00000064) // only lower 32 bits matter
			regFile.WriteReg(2, 0x00000000000000C8)

			alu.ADD32(0, 1, 2, false)

			// Result should be zero-extended 32-bit value
			Expect(regFile.ReadReg(0)).To(Equal(uint64(0x64 + 0xC8)))
		})

		It("should wrap at 32 bits", func() {
			regFile.WriteReg(1, 0xFFFFFFFF)
			regFile.WriteReg(2, 1)

			alu.ADD32(0, 1, 2, false)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(0)))
		})
	})

	Describe("SUB (64-bit)", func() {
		Context("register form", func() {
			It("should subtract two registers", func() {
				regFile.WriteReg(1, 100)
				regFile.WriteReg(2, 30)

				alu.SUB64(0, 1, 2, false)

				Expect(regFile.ReadReg(0)).To(Equal(uint64(70)))
			})

			It("should handle underflow wrapping", func() {
				regFile.WriteReg(1, 0)
				regFile.WriteReg(2, 1)

				alu.SUB64(0, 1, 2, false)

				Expect(regFile.ReadReg(0)).To(Equal(uint64(0xFFFFFFFFFFFFFFFF)))
			})
		})

		Context("immediate form", func() {
			It("should subtract immediate from register", func() {
				regFile.WriteReg(1, 100)

				alu.SUB64Imm(0, 1, 30, false)

				Expect(regFile.ReadReg(0)).To(Equal(uint64(70)))
			})
		})

		Context("with flag setting (SUBS)", func() {
			It("should set Z flag when result is zero", func() {
				regFile.WriteReg(1, 50)
				regFile.WriteReg(2, 50)

				alu.SUB64(0, 1, 2, true)

				Expect(regFile.PSTATE.Z).To(BeTrue())
			})

			It("should set N flag when result is negative", func() {
				regFile.WriteReg(1, 0)
				regFile.WriteReg(2, 1)

				alu.SUB64(0, 1, 2, true)

				Expect(regFile.PSTATE.N).To(BeTrue())
			})

			It("should set C flag when no borrow occurs", func() {
				regFile.WriteReg(1, 100)
				regFile.WriteReg(2, 50)

				alu.SUB64(0, 1, 2, true)

				Expect(regFile.PSTATE.C).To(BeTrue())
			})

			It("should clear C flag when borrow occurs", func() {
				regFile.WriteReg(1, 50)
				regFile.WriteReg(2, 100)

				alu.SUB64(0, 1, 2, true)

				Expect(regFile.PSTATE.C).To(BeFalse())
			})

			It("should set V flag on signed overflow", func() {
				regFile.WriteReg(1, 0x8000000000000000) // min negative
				regFile.WriteReg(2, 1)

				alu.SUB64(0, 1, 2, true)

				Expect(regFile.PSTATE.V).To(BeTrue())
			})
		})
	})

	Describe("SUB (32-bit)", func() {
		It("should subtract two 32-bit values and zero-extend", func() {
			regFile.WriteReg(1, 100)
			regFile.WriteReg(2, 30)

			alu.SUB32(0, 1, 2, false)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(70)))
		})
	})

	Describe("AND (64-bit)", func() {
		It("should perform bitwise AND", func() {
			regFile.WriteReg(1, 0xFF00FF00FF00FF00)
			regFile.WriteReg(2, 0x0F0F0F0F0F0F0F0F)

			alu.AND64(0, 1, 2, false)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(0x0F000F000F000F00)))
		})

		Context("with flag setting (ANDS)", func() {
			It("should set Z flag when result is zero", func() {
				regFile.WriteReg(1, 0xF0F0F0F0F0F0F0F0)
				regFile.WriteReg(2, 0x0F0F0F0F0F0F0F0F)

				alu.AND64(0, 1, 2, true)

				Expect(regFile.PSTATE.Z).To(BeTrue())
			})

			It("should set N flag when MSB is set", func() {
				regFile.WriteReg(1, 0xFFFFFFFFFFFFFFFF)
				regFile.WriteReg(2, 0x8000000000000000)

				alu.AND64(0, 1, 2, true)

				Expect(regFile.PSTATE.N).To(BeTrue())
			})

			It("should clear C and V flags", func() {
				regFile.PSTATE.C = true
				regFile.PSTATE.V = true
				regFile.WriteReg(1, 0xFFFFFFFFFFFFFFFF)
				regFile.WriteReg(2, 0xFFFFFFFFFFFFFFFF)

				alu.AND64(0, 1, 2, true)

				Expect(regFile.PSTATE.C).To(BeFalse())
				Expect(regFile.PSTATE.V).To(BeFalse())
			})
		})
	})

	Describe("AND (32-bit)", func() {
		It("should perform bitwise AND and zero-extend", func() {
			regFile.WriteReg(1, 0xFFFFFFFFFF00FF00)
			regFile.WriteReg(2, 0x0F0F0F0F)

			alu.AND32(0, 1, 2, false)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(0x0F000F00)))
		})
	})

	Describe("ORR (64-bit)", func() {
		It("should perform bitwise OR", func() {
			regFile.WriteReg(1, 0xF0F0F0F0F0F0F0F0)
			regFile.WriteReg(2, 0x0F0F0F0F0F0F0F0F)

			alu.ORR64(0, 1, 2)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(0xFFFFFFFFFFFFFFFF)))
		})

		It("should handle zero operand", func() {
			regFile.WriteReg(1, 0x1234567890ABCDEF)
			regFile.WriteReg(2, 0)

			alu.ORR64(0, 1, 2)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(0x1234567890ABCDEF)))
		})
	})

	Describe("ORR (32-bit)", func() {
		It("should perform bitwise OR and zero-extend", func() {
			regFile.WriteReg(1, 0xF0F0F0F0)
			regFile.WriteReg(2, 0x0F0F0F0F)

			alu.ORR32(0, 1, 2)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(0xFFFFFFFF)))
		})
	})

	Describe("EOR (64-bit)", func() {
		It("should perform bitwise XOR", func() {
			regFile.WriteReg(1, 0xFFFFFFFF00000000)
			regFile.WriteReg(2, 0xFFFF0000FFFF0000)

			alu.EOR64(0, 1, 2)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(0x0000FFFFFFFF0000)))
		})

		It("should produce zero when XORing same values", func() {
			regFile.WriteReg(1, 0x1234567890ABCDEF)
			regFile.WriteReg(2, 0x1234567890ABCDEF)

			alu.EOR64(0, 1, 2)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(0)))
		})
	})

	Describe("EOR (32-bit)", func() {
		It("should perform bitwise XOR and zero-extend", func() {
			regFile.WriteReg(1, 0xFFFF0000)
			regFile.WriteReg(2, 0xFF00FF00)

			alu.EOR32(0, 1, 2)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(0x00FFFF00)))
		})
	})

	Describe("MOV (alias)", func() {
		It("should move value using ORR with XZR", func() {
			regFile.WriteReg(1, 0x1234567890ABCDEF)

			// MOV is typically ORR Xd, XZR, Xm
			alu.ORR64(0, 31, 1)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(0x1234567890ABCDEF)))
		})
	})

	// 32-bit immediate operations (coverage improvement)

	Describe("ADD32Imm", func() {
		It("should add 32-bit register and immediate", func() {
			regFile.WriteReg(1, 100)

			alu.ADD32Imm(0, 1, 50, false)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(150)))
		})

		It("should wrap at 32 bits", func() {
			regFile.WriteReg(1, 0xFFFFFFFF)

			alu.ADD32Imm(0, 1, 1, false)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(0)))
		})

		It("should zero-extend result", func() {
			regFile.WriteReg(1, 0x7FFFFFFF)

			alu.ADD32Imm(0, 1, 1, false)

			// 0x7FFFFFFF + 1 = 0x80000000, zero-extended
			Expect(regFile.ReadReg(0)).To(Equal(uint64(0x80000000)))
		})

		Context("with flag setting", func() {
			It("should set Z flag when result is zero", func() {
				regFile.WriteReg(1, 0xFFFFFFFF)

				alu.ADD32Imm(0, 1, 1, true)

				Expect(regFile.PSTATE.Z).To(BeTrue())
				Expect(regFile.PSTATE.N).To(BeFalse())
			})

			It("should set N flag when MSB of 32-bit result is set", func() {
				regFile.WriteReg(1, 0x7FFFFFFF)

				alu.ADD32Imm(0, 1, 1, true)

				Expect(regFile.PSTATE.N).To(BeTrue())
			})

			It("should set C flag on unsigned overflow", func() {
				regFile.WriteReg(1, 0xFFFFFFFF)

				alu.ADD32Imm(0, 1, 1, true)

				Expect(regFile.PSTATE.C).To(BeTrue())
			})

			It("should set V flag on signed overflow", func() {
				regFile.WriteReg(1, 0x7FFFFFFF) // max positive 32-bit

				alu.ADD32Imm(0, 1, 1, true)

				Expect(regFile.PSTATE.V).To(BeTrue())
			})
		})
	})

	Describe("SUB32Imm", func() {
		It("should subtract immediate from 32-bit register", func() {
			regFile.WriteReg(1, 100)

			alu.SUB32Imm(0, 1, 30, false)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(70)))
		})

		It("should wrap on underflow", func() {
			regFile.WriteReg(1, 0)

			alu.SUB32Imm(0, 1, 1, false)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(0xFFFFFFFF)))
		})

		Context("with flag setting", func() {
			It("should set Z flag when result is zero", func() {
				regFile.WriteReg(1, 50)

				alu.SUB32Imm(0, 1, 50, true)

				Expect(regFile.PSTATE.Z).To(BeTrue())
			})

			It("should set N flag when result is negative", func() {
				regFile.WriteReg(1, 0)

				alu.SUB32Imm(0, 1, 1, true)

				Expect(regFile.PSTATE.N).To(BeTrue())
			})

			It("should set C flag when no borrow occurs", func() {
				regFile.WriteReg(1, 100)

				alu.SUB32Imm(0, 1, 50, true)

				Expect(regFile.PSTATE.C).To(BeTrue())
			})

			It("should clear C flag when borrow occurs", func() {
				regFile.WriteReg(1, 50)

				alu.SUB32Imm(0, 1, 100, true)

				Expect(regFile.PSTATE.C).To(BeFalse())
			})

			It("should set V flag on signed overflow", func() {
				regFile.WriteReg(1, 0x80000000) // min negative 32-bit

				alu.SUB32Imm(0, 1, 1, true)

				Expect(regFile.PSTATE.V).To(BeTrue())
			})
		})
	})

	Describe("AND64Imm", func() {
		It("should perform bitwise AND with immediate", func() {
			regFile.WriteReg(1, 0xFF00FF00FF00FF00)

			alu.AND64Imm(0, 1, 0x0F0F0F0F0F0F0F0F, false)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(0x0F000F000F000F00)))
		})

		It("should mask to lower bits", func() {
			regFile.WriteReg(1, 0xFFFFFFFFFFFFFFFF)

			alu.AND64Imm(0, 1, 0x00000000FFFFFFFF, false)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(0x00000000FFFFFFFF)))
		})

		Context("with flag setting", func() {
			It("should set Z flag when result is zero", func() {
				regFile.WriteReg(1, 0xF0F0F0F0F0F0F0F0)

				alu.AND64Imm(0, 1, 0x0F0F0F0F0F0F0F0F, true)

				Expect(regFile.PSTATE.Z).To(BeTrue())
			})

			It("should set N flag when MSB is set", func() {
				regFile.WriteReg(1, 0xFFFFFFFFFFFFFFFF)

				alu.AND64Imm(0, 1, 0x8000000000000000, true)

				Expect(regFile.PSTATE.N).To(BeTrue())
			})
		})
	})

	Describe("AND32Imm", func() {
		It("should perform bitwise AND with immediate and zero-extend", func() {
			regFile.WriteReg(1, 0xFF00FF00)

			alu.AND32Imm(0, 1, 0x0F0F0F0F, false)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(0x0F000F00)))
		})

		Context("with flag setting", func() {
			It("should set Z flag when result is zero", func() {
				regFile.WriteReg(1, 0xF0F0F0F0)

				alu.AND32Imm(0, 1, 0x0F0F0F0F, true)

				Expect(regFile.PSTATE.Z).To(BeTrue())
			})

			It("should set N flag when MSB of 32-bit result is set", func() {
				regFile.WriteReg(1, 0xFFFFFFFF)

				alu.AND32Imm(0, 1, 0x80000000, true)

				Expect(regFile.PSTATE.N).To(BeTrue())
			})

			It("should clear C and V flags", func() {
				regFile.PSTATE.C = true
				regFile.PSTATE.V = true
				regFile.WriteReg(1, 0xFFFFFFFF)

				alu.AND32Imm(0, 1, 0xFFFFFFFF, true)

				Expect(regFile.PSTATE.C).To(BeFalse())
				Expect(regFile.PSTATE.V).To(BeFalse())
			})
		})
	})

	Describe("ORR64Imm", func() {
		It("should perform bitwise OR with immediate", func() {
			regFile.WriteReg(1, 0xF0F0F0F0F0F0F0F0)

			alu.ORR64Imm(0, 1, 0x0F0F0F0F0F0F0F0F)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(0xFFFFFFFFFFFFFFFF)))
		})

		It("should set bits from immediate", func() {
			regFile.WriteReg(1, 0)

			alu.ORR64Imm(0, 1, 0x00FF00FF00FF00FF)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(0x00FF00FF00FF00FF)))
		})
	})

	Describe("ORR32Imm", func() {
		It("should perform bitwise OR with immediate and zero-extend", func() {
			regFile.WriteReg(1, 0xF0F0F0F0)

			alu.ORR32Imm(0, 1, 0x0F0F0F0F)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(0xFFFFFFFF)))
		})

		It("should set bits from immediate", func() {
			regFile.WriteReg(1, 0)

			alu.ORR32Imm(0, 1, 0x12345678)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(0x12345678)))
		})
	})

	Describe("EOR64Imm", func() {
		It("should perform bitwise XOR with immediate", func() {
			regFile.WriteReg(1, 0xFFFFFFFF00000000)

			alu.EOR64Imm(0, 1, 0xFFFF0000FFFF0000)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(0x0000FFFFFFFF0000)))
		})

		It("should flip all bits when XORing with all ones", func() {
			regFile.WriteReg(1, 0x123456789ABCDEF0)

			alu.EOR64Imm(0, 1, 0xFFFFFFFFFFFFFFFF)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(0xEDCBA9876543210F)))
		})
	})

	Describe("EOR32Imm", func() {
		It("should perform bitwise XOR with immediate and zero-extend", func() {
			regFile.WriteReg(1, 0xFFFF0000)

			alu.EOR32Imm(0, 1, 0xFF00FF00)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(0x00FFFF00)))
		})

		It("should flip bits when XORing with all ones", func() {
			regFile.WriteReg(1, 0x12345678)

			alu.EOR32Imm(0, 1, 0xFFFFFFFF)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(0xEDCBA987)))
		})
	})
})
