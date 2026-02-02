package emu_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sarchlab/m2sim/emu"
)

var _ = Describe("LoadStoreUnit", func() {
	var (
		regFile *emu.RegFile
		memory  *emu.Memory
		lsu     *emu.LoadStoreUnit
	)

	BeforeEach(func() {
		regFile = &emu.RegFile{}
		memory = emu.NewMemory()
		lsu = emu.NewLoadStoreUnit(regFile, memory)
	})

	Describe("LDR (64-bit)", func() {
		Context("base register only", func() {
			It("should load 64-bit value from memory", func() {
				// Set up base address in X1
				regFile.WriteReg(1, 0x1000)
				// Store value in memory
				memory.Write64(0x1000, 0xDEADBEEF12345678)

				// LDR X0, [X1]
				lsu.LDR64(0, 1, 0)

				Expect(regFile.ReadReg(0)).To(Equal(uint64(0xDEADBEEF12345678)))
			})

			It("should handle XZR as destination (discard)", func() {
				regFile.WriteReg(1, 0x1000)
				memory.Write64(0x1000, 0x1234567890ABCDEF)

				lsu.LDR64(31, 1, 0)

				Expect(regFile.ReadReg(31)).To(Equal(uint64(0)))
			})
		})

		Context("with immediate offset", func() {
			It("should load from base + positive offset", func() {
				regFile.WriteReg(1, 0x1000)
				memory.Write64(0x1010, 0xCAFEBABE00000000)

				// LDR X0, [X1, #16]
				lsu.LDR64(0, 1, 16)

				Expect(regFile.ReadReg(0)).To(Equal(uint64(0xCAFEBABE00000000)))
			})

			It("should load from base + large offset", func() {
				regFile.WriteReg(1, 0x1000)
				memory.Write64(0x1100, 0xAAAABBBBCCCCDDDD)

				// LDR X0, [X1, #256]
				lsu.LDR64(0, 1, 256)

				Expect(regFile.ReadReg(0)).To(Equal(uint64(0xAAAABBBBCCCCDDDD)))
			})
		})

		Context("using SP as base", func() {
			It("should load using stack pointer", func() {
				regFile.SP = 0x2000
				memory.Write64(0x2008, 0x1111222233334444)

				// LDR X0, [SP, #8]
				lsu.LDR64SP(0, 8)

				Expect(regFile.ReadReg(0)).To(Equal(uint64(0x1111222233334444)))
			})
		})
	})

	Describe("LDR (32-bit)", func() {
		Context("base register only", func() {
			It("should load 32-bit value and zero-extend", func() {
				regFile.WriteReg(1, 0x1000)
				memory.Write32(0x1000, 0xDEADBEEF)

				// LDR W0, [X1]
				lsu.LDR32(0, 1, 0)

				// Result should be zero-extended
				Expect(regFile.ReadReg(0)).To(Equal(uint64(0xDEADBEEF)))
			})

			It("should properly zero-extend (clear upper bits)", func() {
				regFile.WriteReg(1, 0x1000)
				// Pre-set X0 with garbage in upper bits
				regFile.WriteReg(0, 0xFFFFFFFFFFFFFFFF)
				memory.Write32(0x1000, 0x12345678)

				lsu.LDR32(0, 1, 0)

				// Upper 32 bits should be zeroed
				Expect(regFile.ReadReg(0)).To(Equal(uint64(0x12345678)))
			})
		})

		Context("with immediate offset", func() {
			It("should load from base + offset", func() {
				regFile.WriteReg(1, 0x1000)
				memory.Write32(0x1008, 0xABCDEF00)

				// LDR W0, [X1, #8]
				lsu.LDR32(0, 1, 8)

				Expect(regFile.ReadReg(0)).To(Equal(uint64(0xABCDEF00)))
			})
		})
	})

	Describe("STR (64-bit)", func() {
		Context("base register only", func() {
			It("should store 64-bit value to memory", func() {
				regFile.WriteReg(0, 0xDEADBEEF12345678)
				regFile.WriteReg(1, 0x1000)

				// STR X0, [X1]
				lsu.STR64(0, 1, 0)

				Expect(memory.Read64(0x1000)).To(Equal(uint64(0xDEADBEEF12345678)))
			})

			It("should store XZR as zero", func() {
				regFile.WriteReg(1, 0x1000)
				// Pre-fill memory with garbage
				memory.Write64(0x1000, 0xFFFFFFFFFFFFFFFF)

				// STR XZR, [X1]
				lsu.STR64(31, 1, 0)

				Expect(memory.Read64(0x1000)).To(Equal(uint64(0)))
			})
		})

		Context("with immediate offset", func() {
			It("should store to base + positive offset", func() {
				regFile.WriteReg(0, 0xCAFEBABE00000000)
				regFile.WriteReg(1, 0x1000)

				// STR X0, [X1, #24]
				lsu.STR64(0, 1, 24)

				Expect(memory.Read64(0x1018)).To(Equal(uint64(0xCAFEBABE00000000)))
			})
		})

		Context("using SP as base", func() {
			It("should store using stack pointer", func() {
				regFile.WriteReg(0, 0x9999888877776666)
				regFile.SP = 0x2000

				// STR X0, [SP, #16]
				lsu.STR64SP(0, 16)

				Expect(memory.Read64(0x2010)).To(Equal(uint64(0x9999888877776666)))
			})
		})
	})

	Describe("STR (32-bit)", func() {
		Context("base register only", func() {
			It("should store lower 32 bits to memory", func() {
				regFile.WriteReg(0, 0xFFFFFFFFDEADBEEF)
				regFile.WriteReg(1, 0x1000)

				// STR W0, [X1]
				lsu.STR32(0, 1, 0)

				// Only lower 32 bits should be stored
				Expect(memory.Read32(0x1000)).To(Equal(uint32(0xDEADBEEF)))
			})
		})

		Context("with immediate offset", func() {
			It("should store to base + offset", func() {
				regFile.WriteReg(0, 0x12345678)
				regFile.WriteReg(1, 0x1000)

				// STR W0, [X1, #12]
				lsu.STR32(0, 1, 12)

				Expect(memory.Read32(0x100C)).To(Equal(uint32(0x12345678)))
			})
		})
	})

	Describe("Memory alignment", func() {
		It("should handle unaligned 64-bit access", func() {
			regFile.WriteReg(1, 0x1001) // Unaligned address
			memory.Write64(0x1001, 0x123456789ABCDEF0)

			lsu.LDR64(0, 1, 0)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(0x123456789ABCDEF0)))
		})

		It("should handle unaligned 32-bit access", func() {
			regFile.WriteReg(1, 0x1003) // Unaligned address
			memory.Write32(0x1003, 0xAABBCCDD)

			lsu.LDR32(0, 1, 0)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(0xAABBCCDD)))
		})
	})

	Describe("Edge cases", func() {
		It("should handle zero offset", func() {
			regFile.WriteReg(1, 0x1000)
			memory.Write64(0x1000, 0x5555666677778888)

			lsu.LDR64(0, 1, 0)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(0x5555666677778888)))
		})

		It("should handle maximum valid offset", func() {
			regFile.WriteReg(1, 0x1000)
			memory.Write64(0x8FF8, 0xAAAABBBBCCCCDDDD) // offset 0x7FF8

			lsu.LDR64(0, 1, 0x7FF8)

			Expect(regFile.ReadReg(0)).To(Equal(uint64(0xAAAABBBBCCCCDDDD)))
		})
	})
})
