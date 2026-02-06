package pipeline_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sarchlab/m2sim/insts"
	"github.com/sarchlab/m2sim/timing/pipeline"
)

var _ = Describe("Stage Helper Functions", func() {
	Describe("IsCMP", func() {
		It("should return true for CMP instructions (SUB with SetFlags and Rd=31)", func() {
			inst := &insts.Instruction{
				Op:       insts.OpSUB,
				SetFlags: true,
				Rd:       31,
			}
			Expect(pipeline.IsCMP(inst)).To(BeTrue())
		})

		It("should return false for nil instruction", func() {
			Expect(pipeline.IsCMP(nil)).To(BeFalse())
		})

		It("should return false for non-SUB instruction", func() {
			inst := &insts.Instruction{
				Op:       insts.OpADD,
				SetFlags: true,
				Rd:       31,
			}
			Expect(pipeline.IsCMP(inst)).To(BeFalse())
		})

		It("should return false for SUB without SetFlags", func() {
			inst := &insts.Instruction{
				Op:       insts.OpSUB,
				SetFlags: false,
				Rd:       31,
			}
			Expect(pipeline.IsCMP(inst)).To(BeFalse())
		})

		It("should return false for SUB with SetFlags but Rd!=31 (SUBS)", func() {
			inst := &insts.Instruction{
				Op:       insts.OpSUB,
				SetFlags: true,
				Rd:       5,
			}
			Expect(pipeline.IsCMP(inst)).To(BeFalse())
		})
	})

	Describe("IsBCond", func() {
		It("should return true for B.cond instructions", func() {
			inst := &insts.Instruction{Op: insts.OpBCond}
			Expect(pipeline.IsBCond(inst)).To(BeTrue())
		})

		It("should return false for other branch instructions", func() {
			inst := &insts.Instruction{Op: insts.OpB}
			Expect(pipeline.IsBCond(inst)).To(BeFalse())
		})

		It("should return false for BL instruction", func() {
			inst := &insts.Instruction{Op: insts.OpBL}
			Expect(pipeline.IsBCond(inst)).To(BeFalse())
		})

		It("should return false for CBZ instruction", func() {
			inst := &insts.Instruction{Op: insts.OpCBZ}
			Expect(pipeline.IsBCond(inst)).To(BeFalse())
		})

		It("should return false for non-branch instructions", func() {
			inst := &insts.Instruction{Op: insts.OpADD}
			Expect(pipeline.IsBCond(inst)).To(BeFalse())
		})

		It("should return false for nil instruction", func() {
			Expect(pipeline.IsBCond(nil)).To(BeFalse())
		})
	})

	Describe("ComputeSubFlags", func() {
		Context("64-bit operations", func() {
			It("should set Z flag when result is zero", func() {
				n, z, c, v := pipeline.ComputeSubFlags(10, 10, true)
				Expect(z).To(BeTrue(), "Z flag should be set when result is zero")
				Expect(n).To(BeFalse(), "N flag should be clear for zero result")
				Expect(c).To(BeTrue(), "C flag should be set (no borrow) when op1 >= op2")
				Expect(v).To(BeFalse(), "V flag should be clear (no overflow)")
			})

			It("should set N flag when result is negative", func() {
				n, z, c, v := pipeline.ComputeSubFlags(5, 10, true)
				Expect(n).To(BeTrue(), "N flag should be set when result is negative")
				Expect(z).To(BeFalse(), "Z flag should be clear when result is non-zero")
				Expect(c).To(BeFalse(), "C flag should be clear (borrow) when op1 < op2")
				Expect(v).To(BeFalse(), "V flag should be clear (no signed overflow)")
			})

			It("should set C flag when no borrow", func() {
				n, z, c, v := pipeline.ComputeSubFlags(20, 10, true)
				Expect(c).To(BeTrue(), "C flag should be set (no borrow)")
				Expect(n).To(BeFalse(), "N flag should be clear for positive result")
				Expect(z).To(BeFalse(), "Z flag should be clear for non-zero result")
				Expect(v).To(BeFalse(), "V flag should be clear")
			})

			It("should set V flag on signed overflow (positive - negative = negative)", func() {
				// 0x7FFFFFFFFFFFFFFF - 0xFFFFFFFFFFFFFFFF (-1)
				// = 0x7FFFFFFFFFFFFFFF + 1 = 0x8000000000000000 (negative, overflow!)
				op1 := uint64(0x7FFFFFFFFFFFFFFF) // max positive
				op2 := uint64(0xFFFFFFFFFFFFFFFF) // -1 in two's complement
				n, z, _, v := pipeline.ComputeSubFlags(op1, op2, true)
				Expect(v).To(BeTrue(), "V flag should be set on signed overflow")
				Expect(n).To(BeTrue(), "Result should be negative (overflow)")
				Expect(z).To(BeFalse())
			})

			It("should not set V flag when no signed overflow", func() {
				n, _, c, v := pipeline.ComputeSubFlags(100, 50, true)
				Expect(v).To(BeFalse(), "V flag should be clear when no overflow")
				Expect(n).To(BeFalse())
				Expect(c).To(BeTrue())
			})
		})

		Context("32-bit operations", func() {
			It("should set Z flag when result is zero (32-bit)", func() {
				n, z, c, _ := pipeline.ComputeSubFlags(10, 10, false)
				Expect(z).To(BeTrue())
				Expect(n).To(BeFalse())
				Expect(c).To(BeTrue())
			})

			It("should set N flag for negative 32-bit result", func() {
				n, z, c, v := pipeline.ComputeSubFlags(5, 10, false)
				Expect(n).To(BeTrue())
				Expect(z).To(BeFalse())
				Expect(c).To(BeFalse())
				Expect(v).To(BeFalse())
			})

			It("should handle 32-bit overflow correctly", func() {
				// 0x7FFFFFFF - 0xFFFFFFFF = 0x80000000 (overflow)
				op1 := uint64(0x7FFFFFFF) // max positive 32-bit
				op2 := uint64(0xFFFFFFFF) // -1 in 32-bit two's complement
				n, _, _, v := pipeline.ComputeSubFlags(op1, op2, false)
				Expect(v).To(BeTrue(), "V flag should be set on 32-bit signed overflow")
				Expect(n).To(BeTrue())
			})
		})
	})

	Describe("EvaluateConditionWithFlags", func() {
		Context("EQ/NE conditions (based on Z flag)", func() {
			It("should evaluate EQ correctly when Z=1", func() {
				result := pipeline.EvaluateConditionWithFlags(insts.CondEQ, false, true, false, false)
				Expect(result).To(BeTrue(), "EQ should be true when Z=1")
			})

			It("should evaluate EQ correctly when Z=0", func() {
				result := pipeline.EvaluateConditionWithFlags(insts.CondEQ, false, false, false, false)
				Expect(result).To(BeFalse(), "EQ should be false when Z=0")
			})

			It("should evaluate NE correctly when Z=0", func() {
				result := pipeline.EvaluateConditionWithFlags(insts.CondNE, false, false, false, false)
				Expect(result).To(BeTrue(), "NE should be true when Z=0")
			})

			It("should evaluate NE correctly when Z=1", func() {
				result := pipeline.EvaluateConditionWithFlags(insts.CondNE, false, true, false, false)
				Expect(result).To(BeFalse(), "NE should be false when Z=1")
			})
		})

		Context("CS/CC conditions (based on C flag)", func() {
			It("should evaluate CS correctly when C=1", func() {
				result := pipeline.EvaluateConditionWithFlags(insts.CondCS, false, false, true, false)
				Expect(result).To(BeTrue(), "CS should be true when C=1")
			})

			It("should evaluate CC correctly when C=0", func() {
				result := pipeline.EvaluateConditionWithFlags(insts.CondCC, false, false, false, false)
				Expect(result).To(BeTrue(), "CC should be true when C=0")
			})
		})

		Context("MI/PL conditions (based on N flag)", func() {
			It("should evaluate MI correctly when N=1", func() {
				result := pipeline.EvaluateConditionWithFlags(insts.CondMI, true, false, false, false)
				Expect(result).To(BeTrue(), "MI should be true when N=1")
			})

			It("should evaluate PL correctly when N=0", func() {
				result := pipeline.EvaluateConditionWithFlags(insts.CondPL, false, false, false, false)
				Expect(result).To(BeTrue(), "PL should be true when N=0")
			})
		})

		Context("VS/VC conditions (based on V flag)", func() {
			It("should evaluate VS correctly when V=1", func() {
				result := pipeline.EvaluateConditionWithFlags(insts.CondVS, false, false, false, true)
				Expect(result).To(BeTrue(), "VS should be true when V=1")
			})

			It("should evaluate VC correctly when V=0", func() {
				result := pipeline.EvaluateConditionWithFlags(insts.CondVC, false, false, false, false)
				Expect(result).To(BeTrue(), "VC should be true when V=0")
			})
		})

		Context("HI/LS conditions (C && !Z / !C || Z)", func() {
			It("should evaluate HI correctly when C=1 and Z=0", func() {
				result := pipeline.EvaluateConditionWithFlags(insts.CondHI, false, false, true, false)
				Expect(result).To(BeTrue(), "HI should be true when C=1 and Z=0")
			})

			It("should evaluate HI correctly when C=1 and Z=1", func() {
				result := pipeline.EvaluateConditionWithFlags(insts.CondHI, false, true, true, false)
				Expect(result).To(BeFalse(), "HI should be false when Z=1")
			})

			It("should evaluate LS correctly when C=0", func() {
				result := pipeline.EvaluateConditionWithFlags(insts.CondLS, false, false, false, false)
				Expect(result).To(BeTrue(), "LS should be true when C=0")
			})

			It("should evaluate LS correctly when Z=1", func() {
				result := pipeline.EvaluateConditionWithFlags(insts.CondLS, false, true, true, false)
				Expect(result).To(BeTrue(), "LS should be true when Z=1")
			})
		})

		Context("GE/LT conditions (N == V / N != V)", func() {
			It("should evaluate GE correctly when N=V=0", func() {
				result := pipeline.EvaluateConditionWithFlags(insts.CondGE, false, false, false, false)
				Expect(result).To(BeTrue(), "GE should be true when N=V")
			})

			It("should evaluate GE correctly when N=V=1", func() {
				result := pipeline.EvaluateConditionWithFlags(insts.CondGE, true, false, false, true)
				Expect(result).To(BeTrue(), "GE should be true when N=V")
			})

			It("should evaluate LT correctly when N!=V", func() {
				result := pipeline.EvaluateConditionWithFlags(insts.CondLT, true, false, false, false)
				Expect(result).To(BeTrue(), "LT should be true when N!=V")
			})
		})

		Context("GT/LE conditions (!Z && N==V / Z || N!=V)", func() {
			It("should evaluate GT correctly when Z=0 and N=V", func() {
				result := pipeline.EvaluateConditionWithFlags(insts.CondGT, false, false, false, false)
				Expect(result).To(BeTrue(), "GT should be true when Z=0 and N=V")
			})

			It("should evaluate GT correctly when Z=1", func() {
				result := pipeline.EvaluateConditionWithFlags(insts.CondGT, false, true, false, false)
				Expect(result).To(BeFalse(), "GT should be false when Z=1")
			})

			It("should evaluate LE correctly when Z=1", func() {
				result := pipeline.EvaluateConditionWithFlags(insts.CondLE, false, true, false, false)
				Expect(result).To(BeTrue(), "LE should be true when Z=1")
			})

			It("should evaluate LE correctly when N!=V", func() {
				result := pipeline.EvaluateConditionWithFlags(insts.CondLE, true, false, false, false)
				Expect(result).To(BeTrue(), "LE should be true when N!=V")
			})
		})

		Context("AL/NV conditions (always true)", func() {
			It("should evaluate AL as always true", func() {
				result := pipeline.EvaluateConditionWithFlags(insts.CondAL, false, false, false, false)
				Expect(result).To(BeTrue(), "AL should always be true")
			})

			It("should evaluate NV as always true", func() {
				result := pipeline.EvaluateConditionWithFlags(insts.CondNV, true, true, true, true)
				Expect(result).To(BeTrue(), "NV should always be true")
			})
		})
	})
})
