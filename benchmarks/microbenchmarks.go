// Package benchmarks provides timing benchmark infrastructure for M2Sim calibration.
package benchmarks

import "github.com/sarchlab/m2sim/emu"

// GetMicrobenchmarks returns the standard set of microbenchmarks for M2 calibration.
// Each benchmark targets a specific CPU characteristic.
//
// NOTE: The timing pipeline now supports PSTATE flags (PR #205). The
// branchTakenConditional benchmark uses CMP + B.GE to match native benchmarks.
// Other benchmarks still use unrolled code for simplicity.
func GetMicrobenchmarks() []Benchmark {
	return []Benchmark{
		arithmeticSequential(),
		arithmetic6Wide(),
		arithmetic8Wide(),
		dependencyChain(),
		memorySequential(),
		memoryStrided(),
		memorySequentialScaled(),
		memoryStridedScaled(),
		memoryRandomAccess(),
		loadHeavy(),
		loadHeavyScaled(),
		storeHeavy(),
		storeHeavyScaled(),
		branchTaken(),
		branchTakenConditional(),
		branchHotLoop(),
		branchHeavy(),
		functionCalls(),
		mixedOperations(),
		matrixMultiply2x2(),
		loopSimulation(),
		vectorSum(),
		vectorAdd(),
		reductionTree(),
		strideIndirect(),
	}
}

// GetCoreBenchmarks returns a minimal set of 3 core benchmarks for quick validation.
// These correspond to the acceptance criteria: loop, matrix multiply, branch-heavy code.
// Uses branchTakenConditional to match native benchmark pattern (CMP + B.GE).
func GetCoreBenchmarks() []Benchmark {
	return []Benchmark{
		loopSimulation(),
		matrixMultiply2x2(),
		branchTakenConditional(),
	}
}

// 1. Arithmetic Sequential - Tests ALU throughput with independent operations
func arithmeticSequential() Benchmark {
	const numInstructions = 200
	const numRegisters = 5
	return Benchmark{
		Name:        "arithmetic_sequential",
		Description: "200 independent ADDs (5 registers) - measures ALU throughput",
		Setup: func(regFile *emu.RegFile, memory *emu.Memory) {
			regFile.WriteReg(8, 93) // X8 = 93 (exit syscall)
		},
		Program:      buildArithmeticSequential(numInstructions, numRegisters),
		ExpectedExit: int64(numInstructions / numRegisters), // X0 incremented once per register cycle
	}
}

func buildArithmeticSequential(n, numRegs int) []byte {
	instrs := make([]uint32, 0, n+1)
	for i := 0; i < n; i++ {
		reg := uint8(i % numRegs)
		instrs = append(instrs, EncodeADDImm(reg, reg, 1, false))
	}
	instrs = append(instrs, EncodeSVC(0))
	return BuildProgram(instrs...)
}

// 1b. Arithmetic 6-Wide - Tests full 6-wide superscalar throughput
// Uses 6 different registers to avoid RAW hazards between consecutive groups
func arithmetic6Wide() Benchmark {
	return Benchmark{
		Name:        "arithmetic_6wide",
		Description: "24 independent ADDs using 6 registers - tests full 6-wide issue",
		Setup: func(regFile *emu.RegFile, memory *emu.Memory) {
			regFile.WriteReg(8, 93) // X8 = 93 (exit syscall)
		},
		Program: BuildProgram(
			// 24 ADDs using 6 registers (X0-X5) - allows full 6-wide issue
			// Group 1: instructions 0-5 (all independent, can issue together)
			EncodeADDImm(0, 0, 1, false),
			EncodeADDImm(1, 1, 1, false),
			EncodeADDImm(2, 2, 1, false),
			EncodeADDImm(3, 3, 1, false),
			EncodeADDImm(4, 4, 1, false),
			EncodeADDImm(5, 5, 1, false),
			// Group 2: instructions 6-11 (RAW hazards with group 1, but forwarding allows issue)
			EncodeADDImm(0, 0, 1, false),
			EncodeADDImm(1, 1, 1, false),
			EncodeADDImm(2, 2, 1, false),
			EncodeADDImm(3, 3, 1, false),
			EncodeADDImm(4, 4, 1, false),
			EncodeADDImm(5, 5, 1, false),
			// Group 3
			EncodeADDImm(0, 0, 1, false),
			EncodeADDImm(1, 1, 1, false),
			EncodeADDImm(2, 2, 1, false),
			EncodeADDImm(3, 3, 1, false),
			EncodeADDImm(4, 4, 1, false),
			EncodeADDImm(5, 5, 1, false),
			// Group 4
			EncodeADDImm(0, 0, 1, false),
			EncodeADDImm(1, 1, 1, false),
			EncodeADDImm(2, 2, 1, false),
			EncodeADDImm(3, 3, 1, false),
			EncodeADDImm(4, 4, 1, false),
			EncodeADDImm(5, 5, 1, false),
			EncodeSVC(0),
		),
		ExpectedExit: 4, // X0 = 0 + 4*1 = 4
	}
}

// 1c. Arithmetic 8-Wide - Tests full 8-wide superscalar throughput (M2 P-core)
// Uses 8 different registers (X0-X7) to maximize parallelism without WAW hazards.
// This benchmark is designed to measure the true benefit of 8-wide decode (PR #215).
func arithmetic8Wide() Benchmark {
	return Benchmark{
		Name:        "arithmetic_8wide",
		Description: "32 independent ADDs using 8 registers - tests full 8-wide issue",
		Setup: func(regFile *emu.RegFile, memory *emu.Memory) {
			// Note: X8 is used as syscall number in ARM64 Linux convention
			regFile.WriteReg(8, 93) // X8 = 93 (exit syscall)
		},
		Program: BuildProgram(
			// 32 ADDs using 8 registers (X0-X7) - allows full 8-wide issue
			// Group 1: instructions 0-7 (all independent, can issue together)
			EncodeADDImm(0, 0, 1, false),
			EncodeADDImm(1, 1, 1, false),
			EncodeADDImm(2, 2, 1, false),
			EncodeADDImm(3, 3, 1, false),
			EncodeADDImm(4, 4, 1, false),
			EncodeADDImm(5, 5, 1, false),
			EncodeADDImm(6, 6, 1, false),
			EncodeADDImm(7, 7, 1, false),
			// Group 2: instructions 8-15 (RAW hazards with group 1, but forwarding allows issue)
			EncodeADDImm(0, 0, 1, false),
			EncodeADDImm(1, 1, 1, false),
			EncodeADDImm(2, 2, 1, false),
			EncodeADDImm(3, 3, 1, false),
			EncodeADDImm(4, 4, 1, false),
			EncodeADDImm(5, 5, 1, false),
			EncodeADDImm(6, 6, 1, false),
			EncodeADDImm(7, 7, 1, false),
			// Group 3
			EncodeADDImm(0, 0, 1, false),
			EncodeADDImm(1, 1, 1, false),
			EncodeADDImm(2, 2, 1, false),
			EncodeADDImm(3, 3, 1, false),
			EncodeADDImm(4, 4, 1, false),
			EncodeADDImm(5, 5, 1, false),
			EncodeADDImm(6, 6, 1, false),
			EncodeADDImm(7, 7, 1, false),
			// Group 4
			EncodeADDImm(0, 0, 1, false),
			EncodeADDImm(1, 1, 1, false),
			EncodeADDImm(2, 2, 1, false),
			EncodeADDImm(3, 3, 1, false),
			EncodeADDImm(4, 4, 1, false),
			EncodeADDImm(5, 5, 1, false),
			EncodeADDImm(6, 6, 1, false),
			EncodeADDImm(7, 7, 1, false),
			EncodeSVC(0),
		),
		ExpectedExit: 4, // X0 = 0 + 4*1 = 4
	}
}

// 2. Dependency Chain - Tests instruction latency with RAW hazards
// Uses 200 instructions to amortize pipeline fill/drain overhead (~2% vs ~20%
// with 20 instructions). Real M2 measures steady-state CPI via regression over
// millions of iterations, so our benchmark needs enough instructions to make
// pipeline startup cost negligible.
func dependencyChain() Benchmark {
	return Benchmark{
		Name:        "dependency_chain",
		Description: "200 dependent ADDs (X0 = X0 + 1) - measures forwarding latency",
		Setup: func(regFile *emu.RegFile, memory *emu.Memory) {
			regFile.WriteReg(8, 93) // X8 = 93 (exit syscall)
			regFile.WriteReg(0, 0)  // X0 = 0 (start value)
		},
		Program:      buildDependencyChain(200),
		ExpectedExit: 200, // X0 = 0 + 200*1 = 200
	}
}

func buildDependencyChain(n int) []byte {
	instrs := make([]uint32, 0, n+1)
	for i := 0; i < n; i++ {
		instrs = append(instrs, EncodeADDImm(0, 0, 1, false))
	}
	instrs = append(instrs, EncodeSVC(0))
	return BuildProgram(instrs...)
}

// 3. Memory Sequential - Tests cache/memory performance
func memorySequential() Benchmark {
	return Benchmark{
		Name:        "memory_sequential",
		Description: "10 store/load pairs to sequential addresses - measures memory latency",
		Setup: func(regFile *emu.RegFile, memory *emu.Memory) {
			regFile.WriteReg(8, 93)     // X8 = 93 (exit syscall)
			regFile.WriteReg(1, 0x8000) // X1 = base address
			regFile.WriteReg(0, 42)     // X0 = value to store/load
		},
		Program: BuildProgram(
			// Store X0 to [X1], load back, repeat at different offsets
			// Note: Between pairs (e.g., LDR X0 then STR X0), there's a load-use
			// hazard that requires a stall to ensure correct behavior.
			EncodeSTR64(0, 1, 0), EncodeLDR64(0, 1, 0),
			EncodeSTR64(0, 1, 1), EncodeLDR64(0, 1, 1), // offset = 8 bytes
			EncodeSTR64(0, 1, 2), EncodeLDR64(0, 1, 2),
			EncodeSTR64(0, 1, 3), EncodeLDR64(0, 1, 3),
			EncodeSTR64(0, 1, 4), EncodeLDR64(0, 1, 4),
			EncodeSTR64(0, 1, 5), EncodeLDR64(0, 1, 5),
			EncodeSTR64(0, 1, 6), EncodeLDR64(0, 1, 6),
			EncodeSTR64(0, 1, 7), EncodeLDR64(0, 1, 7),
			EncodeSTR64(0, 1, 8), EncodeLDR64(0, 1, 8),
			EncodeSTR64(0, 1, 9), EncodeLDR64(0, 1, 9),
			EncodeSVC(0),
		),
		// X0 starts at 42, and with proper load-use hazard handling for stores,
		// the value is preserved through all store-load pairs.
		ExpectedExit: 42,
	}
}

// 4. Function Calls - Tests BL/RET overhead
func functionCalls() Benchmark {
	return Benchmark{
		Name:        "function_calls",
		Description: "5 function calls (BL + RET pairs) - measures call overhead",
		Setup: func(regFile *emu.RegFile, memory *emu.Memory) {
			regFile.WriteReg(8, 93) // X8 = 93 (exit syscall)
			regFile.WriteReg(0, 0)  // X0 = 0 (result)
			regFile.SP = 0x10000    // Stack pointer
		},
		Program: BuildProgram(
			// main: call add_one 5 times
			EncodeBL(24), // BL add_one (offset = 6 instrs = 24 bytes)
			EncodeBL(20), // BL add_one
			EncodeBL(16), // BL add_one
			EncodeBL(12), // BL add_one
			EncodeBL(8),  // BL add_one
			EncodeSVC(0), // exit with X0

			// add_one function (at offset 24)
			EncodeADDImm(0, 0, 1, false), // X0 += 1
			EncodeRET(),                  // return
		),
		ExpectedExit: 5, // 5 calls * 1 add = 5
	}
}

// 5. Branch Taken - Tests unconditional branch overhead
func branchTaken() Benchmark {
	return Benchmark{
		Name:        "branch_taken",
		Description: "5 unconditional branches (B forward) - measures branch overhead",
		Setup: func(regFile *emu.RegFile, memory *emu.Memory) {
			regFile.WriteReg(8, 93) // X8 = 93 (exit syscall)
			regFile.WriteReg(0, 0)  // X0 = 0 (result)
		},
		Program: BuildProgram(
			// Jump over NOP-like instructions
			EncodeB(8),                    // B +8 (skip next instr)
			EncodeADDImm(1, 1, 99, false), // skipped
			EncodeADDImm(0, 0, 1, false),  // X0 += 1

			EncodeB(8),                    // B +8
			EncodeADDImm(1, 1, 99, false), // skipped
			EncodeADDImm(0, 0, 1, false),  // X0 += 1

			EncodeB(8),                    // B +8
			EncodeADDImm(1, 1, 99, false), // skipped
			EncodeADDImm(0, 0, 1, false),  // X0 += 1

			EncodeB(8),                    // B +8
			EncodeADDImm(1, 1, 99, false), // skipped
			EncodeADDImm(0, 0, 1, false),  // X0 += 1

			EncodeB(8),                    // B +8
			EncodeADDImm(1, 1, 99, false), // skipped
			EncodeADDImm(0, 0, 1, false),  // X0 += 1

			EncodeSVC(0), // exit with X0 = 5
		),
		ExpectedExit: 5,
	}
}

// 5b. Branch Taken Conditional - Uses CMP + B.GE to match native benchmark pattern
// This aligns with branch_taken_long.s which uses conditional branches
func branchTakenConditional() Benchmark {
	const numBranches = 50
	return Benchmark{
		Name:        "branch_taken_conditional",
		Description: "50 conditional branches (CMP + B.GE) - measures branch prediction",
		Setup: func(regFile *emu.RegFile, memory *emu.Memory) {
			regFile.WriteReg(8, 93) // X8 = 93 (exit syscall)
			regFile.WriteReg(0, 0)  // X0 = 0 (result, always >= 0)
		},
		Program:      buildBranchConditionalChain(numBranches),
		ExpectedExit: int64(numBranches),
	}
}

func buildBranchConditionalChain(n int) []byte {
	// Each branch pattern: CMP X0, #0; B.GE +8 (skip 1); ADD X1 (skipped); ADD X0 += 1
	instrs := make([]uint32, 0, n*4+1)
	for i := 0; i < n; i++ {
		instrs = append(instrs,
			EncodeCMPImm(0, 0),            // CMP X0, #0
			EncodeBCond(8, 10),            // B.GE +8 (CondGE = 10, always taken)
			EncodeADDImm(1, 1, 99, false), // skipped
			EncodeADDImm(0, 0, 1, false),  // X0 += 1
		)
	}
	instrs = append(instrs, EncodeSVC(0))
	return BuildProgram(instrs...)
}

// 5c. Branch Hot Loop - Tests zero-cycle branch folding with hot branches
// This benchmark uses a real loop where the same branch PC is executed multiple times.
// Zero-cycle folding requires:
// 1. BTB hit (target known from previous execution)
// 2. Predicted taken
// 3. High confidence (counter >= 3, trained by 3+ executions)
//
// With 4 iterations:
// - Iterations 1-3: Normal branch penalty (training phase)
// - Iteration 4: Zero-cycle folding (1 folded branch expected)
// Note: Reduced from 16 to 4 to avoid CI timeout (timing sim runs slowly on loops)
func branchHotLoop() Benchmark {
	return Benchmark{
		Name:        "branch_hot_loop",
		Description: "4-iteration loop with single hot branch - validates zero-cycle folding",
		Setup: func(regFile *emu.RegFile, memory *emu.Memory) {
			regFile.WriteReg(8, 93) // X8 = 93 (exit syscall)
			regFile.WriteReg(0, 4)  // X0 = 4 (loop counter, reduced for CI)
		},
		Program: BuildProgram(
			// loop:
			EncodeSUBImm(0, 0, 1, false), // X0 = X0 - 1
			EncodeCMPImm(0, 0),           // CMP X0, #0
			EncodeBCond(-8, 1),           // B.NE loop (-8 bytes = -2 instructions), CondNE = 1

			// exit: After loop, X0 = 0
			EncodeSVC(0), // exit with X0 = 0
		),
		ExpectedExit: 0,
	}
}

// 6. Mixed Operations - Combination of ALU, memory, and branches
func mixedOperations() Benchmark {
	return Benchmark{
		Name:        "mixed_operations",
		Description: "Mix of ADD, STR/LDR, and BL - realistic workload characteristics",
		Setup: func(regFile *emu.RegFile, memory *emu.Memory) {
			regFile.WriteReg(8, 93)     // X8 = 93 (exit syscall)
			regFile.WriteReg(0, 0)      // X0 = sum
			regFile.WriteReg(1, 0x8000) // X1 = buffer address
			regFile.SP = 0x10000
		},
		Program: BuildProgram(
			// Iteration 1: compute, store, load, call
			EncodeADDImm(2, 0, 10, false), // X2 = X0 + 10 = 10
			EncodeSTR64(2, 1, 0),          // [X1] = X2
			EncodeLDR64(3, 1, 0),          // X3 = [X1]
			EncodeADDReg(0, 0, 3, false),  // X0 += X3
			EncodeBL(44),                  // BL add_five

			// Iteration 2
			EncodeADDImm(2, 0, 10, false),
			EncodeSTR64(2, 1, 1),
			EncodeLDR64(3, 1, 1),
			EncodeADDReg(0, 0, 3, false),
			EncodeBL(24),

			// Iteration 3
			EncodeADDImm(2, 0, 10, false),
			EncodeSTR64(2, 1, 2),
			EncodeLDR64(3, 1, 2),
			EncodeADDReg(0, 0, 3, false),

			EncodeSVC(0), // exit with X0

			// add_five function
			EncodeADDImm(0, 0, 5, false),
			EncodeRET(),
		),
		// iter1: X0=0, X2=10, X3=10, X0=10, call +5 → X0=15
		// iter2: X0=15, X2=25, X3=25, X0=40, call +5 → X0=45
		// iter3: X0=45, X2=55, X3=55, X0=100
		ExpectedExit: 100,
	}
}

// EncodeB encodes unconditional branch: B offset
func EncodeB(offset int32) uint32 {
	var inst uint32 = 0
	inst |= 0b000101 << 26 // B opcode
	imm26 := uint32(offset/4) & 0x3FFFFFF
	inst |= imm26
	return inst
}

// 7. Matrix Operations - Tests computation with memory access pattern
// Loads values from memory, performs computations, stores results
// Note: Uses ADD instead of MUL since scalar MUL isn't implemented yet
func matrixMultiply2x2() Benchmark {
	return Benchmark{
		Name:        "matrix_operations",
		Description: "Matrix-style load/compute/store pattern - tests memory access",
		Setup: func(regFile *emu.RegFile, memory *emu.Memory) {
			regFile.WriteReg(8, 93) // X8 = 93 (exit syscall)
			// Array A at 0x8000: [10, 20, 30, 40]
			regFile.WriteReg(1, 0x8000)
			memory.Write64(0x8000, 10)
			memory.Write64(0x8008, 20)
			memory.Write64(0x8010, 30)
			memory.Write64(0x8018, 40)

			// Array B at 0x8100: [1, 2, 3, 4]
			regFile.WriteReg(2, 0x8100)
			memory.Write64(0x8100, 1)
			memory.Write64(0x8108, 2)
			memory.Write64(0x8110, 3)
			memory.Write64(0x8118, 4)

			// Array C at 0x8200 (result)
			regFile.WriteReg(3, 0x8200)
		},
		// Compute C[i] = A[i] + B[i] for i = 0..3
		// C = [11, 22, 33, 44]
		// Return sum of C = 11 + 22 + 33 + 44 = 110
		Program: BuildProgram(
			// Load A array into X10-X13
			EncodeLDR64(10, 1, 0), // X10 = A[0] = 10
			EncodeLDR64(11, 1, 1), // X11 = A[1] = 20
			EncodeLDR64(12, 1, 2), // X12 = A[2] = 30
			EncodeLDR64(13, 1, 3), // X13 = A[3] = 40

			// Load B array into X14-X17
			EncodeLDR64(14, 2, 0), // X14 = B[0] = 1
			EncodeLDR64(15, 2, 1), // X15 = B[1] = 2
			EncodeLDR64(16, 2, 2), // X16 = B[2] = 3
			EncodeLDR64(17, 2, 3), // X17 = B[3] = 4

			// Compute C[i] = A[i] + B[i]
			EncodeADDReg(20, 10, 14, false), // X20 = 10 + 1 = 11
			EncodeADDReg(21, 11, 15, false), // X21 = 20 + 2 = 22
			EncodeADDReg(22, 12, 16, false), // X22 = 30 + 3 = 33
			EncodeADDReg(23, 13, 17, false), // X23 = 40 + 4 = 44

			// Store C array
			EncodeSTR64(20, 3, 0), // C[0] = 11
			EncodeSTR64(21, 3, 1), // C[1] = 22
			EncodeSTR64(22, 3, 2), // C[2] = 33
			EncodeSTR64(23, 3, 3), // C[3] = 44

			// Sum all C elements for exit code: 11 + 22 + 33 + 44 = 110
			EncodeADDReg(0, 20, 21, false), // X0 = 11 + 22 = 33
			EncodeADDReg(0, 0, 22, false),  // X0 = 33 + 33 = 66
			EncodeADDReg(0, 0, 23, false),  // X0 = 66 + 44 = 110

			EncodeSVC(0),
		),
		ExpectedExit: 110,
	}
}

// 8. Loop Simulation - Simulates a counted loop (unrolled)
// This is what a "for i := 0; i < 10; i++" loop would look like
func loopSimulation() Benchmark {
	return Benchmark{
		Name:        "loop_simulation",
		Description: "Simulated 10-iteration loop (unrolled) - tests loop-like patterns",
		Setup: func(regFile *emu.RegFile, memory *emu.Memory) {
			regFile.WriteReg(8, 93) // X8 = 93 (exit syscall)
			regFile.WriteReg(0, 0)  // X0 = sum = 0
			regFile.WriteReg(1, 0)  // X1 = i = 0
		},
		// Simulate: for i := 0; i < 10; i++ { sum += i }
		// Result: 0 + 1 + 2 + 3 + 4 + 5 + 6 + 7 + 8 + 9 = 45
		Program: BuildProgram(
			// Iteration 0: sum += 0, i++
			EncodeADDReg(0, 0, 1, false), // sum += i
			EncodeADDImm(1, 1, 1, false), // i++

			// Iteration 1: sum += 1, i++
			EncodeADDReg(0, 0, 1, false),
			EncodeADDImm(1, 1, 1, false),

			// Iteration 2
			EncodeADDReg(0, 0, 1, false),
			EncodeADDImm(1, 1, 1, false),

			// Iteration 3
			EncodeADDReg(0, 0, 1, false),
			EncodeADDImm(1, 1, 1, false),

			// Iteration 4
			EncodeADDReg(0, 0, 1, false),
			EncodeADDImm(1, 1, 1, false),

			// Iteration 5
			EncodeADDReg(0, 0, 1, false),
			EncodeADDImm(1, 1, 1, false),

			// Iteration 6
			EncodeADDReg(0, 0, 1, false),
			EncodeADDImm(1, 1, 1, false),

			// Iteration 7
			EncodeADDReg(0, 0, 1, false),
			EncodeADDImm(1, 1, 1, false),

			// Iteration 8
			EncodeADDReg(0, 0, 1, false),
			EncodeADDImm(1, 1, 1, false),

			// Iteration 9
			EncodeADDReg(0, 0, 1, false),
			EncodeADDImm(1, 1, 1, false),

			EncodeSVC(0),
		),
		ExpectedExit: 45,
	}
}

// 9. Memory Strided - Tests strided access pattern (stride = 4 elements = 32 bytes)
// Matches native memory_strided_long.s: STR X0 → LDR X2 → ADD X0,X2,#1 chains
// with stride-4 offsets (32 bytes between accesses).
// This creates a serial dependency chain matching real hardware measurements.
func memoryStrided() Benchmark {
	return Benchmark{
		Name:        "memory_strided",
		Description: "5 store/load/add chains with stride-4 access - measures strided memory latency",
		Setup: func(regFile *emu.RegFile, memory *emu.Memory) {
			regFile.WriteReg(8, 93)     // X8 = 93 (exit syscall)
			regFile.WriteReg(1, 0x8000) // X1 = base address
			regFile.WriteReg(0, 0)      // X0 = initial value
		},
		Program: BuildProgram(
			// 5 chains of STR X0 → LDR X2 → ADD X0, X2, #1
			// Matches native: str x0,[sp,#off]; ldr x2,[sp,#off]; add x0,x2,#1
			EncodeSTR64(0, 1, 0), EncodeLDR64(2, 1, 0), EncodeADDImm(0, 2, 1, false),
			EncodeSTR64(0, 1, 4), EncodeLDR64(2, 1, 4), EncodeADDImm(0, 2, 1, false),
			EncodeSTR64(0, 1, 8), EncodeLDR64(2, 1, 8), EncodeADDImm(0, 2, 1, false),
			EncodeSTR64(0, 1, 12), EncodeLDR64(2, 1, 12), EncodeADDImm(0, 2, 1, false),
			EncodeSTR64(0, 1, 16), EncodeLDR64(2, 1, 16), EncodeADDImm(0, 2, 1, false),
			EncodeSVC(0),
		),
		ExpectedExit: 5, // X0 = 0 + 5 increments = 5
	}
}

// 9b. Memory Sequential Scaled - 200 sequential store/load pairs
// Enough instructions to amortize pipeline fill/drain overhead (~2%).
func memorySequentialScaled() Benchmark {
	const numPairs = 200
	return Benchmark{
		Name:        "memory_sequential_scaled",
		Description: "200 sequential store/load pairs - measures memory latency at scale",
		Setup: func(regFile *emu.RegFile, memory *emu.Memory) {
			regFile.WriteReg(8, 93)     // X8 = 93 (exit syscall)
			regFile.WriteReg(1, 0x8000) // X1 = base address
			regFile.WriteReg(0, 42)     // X0 = value to store/load
		},
		Program:      buildMemorySequentialScaled(numPairs),
		ExpectedExit: 42,
	}
}

func buildMemorySequentialScaled(numPairs int) []byte {
	instrs := make([]uint32, 0, numPairs*2+1)
	for i := 0; i < numPairs; i++ {
		offset := uint16(i) // sequential offsets: 0, 1, 2, ... (each 8 bytes)
		instrs = append(instrs,
			EncodeSTR64(0, 1, offset),
			EncodeLDR64(0, 1, offset),
		)
	}
	instrs = append(instrs, EncodeSVC(0))
	return BuildProgram(instrs...)
}

// 9c. Memory Strided Scaled - 200 strided store/load/add chains (stride = 4 elements = 32 bytes)
// Matches native memory_strided_long.s structure at scale.
func memoryStridedScaled() Benchmark {
	const numChains = 200
	return Benchmark{
		Name:        "memory_strided_scaled",
		Description: "200 strided store/load/add chains (32-byte stride) - measures strided memory latency at scale",
		Setup: func(regFile *emu.RegFile, memory *emu.Memory) {
			regFile.WriteReg(8, 93)     // X8 = 93 (exit syscall)
			regFile.WriteReg(1, 0x8000) // X1 = base address
			regFile.WriteReg(0, 0)      // X0 = initial value
		},
		Program:      buildMemoryStridedScaled(numChains),
		ExpectedExit: int64(numChains), // X0 incremented once per chain
	}
}

func buildMemoryStridedScaled(numChains int) []byte {
	instrs := make([]uint32, 0, numChains*3+1)
	for i := 0; i < numChains; i++ {
		offset := uint16(i * 4) // stride-4 offsets: 0, 4, 8, 12, ... (each unit = 8 bytes)
		instrs = append(instrs,
			EncodeSTR64(0, 1, offset),    // STR X0, [X1, #offset]
			EncodeLDR64(2, 1, offset),    // LDR X2, [X1, #offset]
			EncodeADDImm(0, 2, 1, false), // ADD X0, X2, #1
		)
	}
	instrs = append(instrs, EncodeSVC(0))
	return BuildProgram(instrs...)
}

// 9d. Memory Random Access - 200 store/load pairs with pseudo-random offsets
// Uses a deterministic permutation to create non-sequential access patterns
// that stress cache line utilization without exceeding the addressable range.
func memoryRandomAccess() Benchmark {
	const numPairs = 200
	return Benchmark{
		Name:        "memory_random_access",
		Description: "200 store/load pairs with pseudo-random offsets - measures random memory latency",
		Setup: func(regFile *emu.RegFile, memory *emu.Memory) {
			regFile.WriteReg(8, 93)     // X8 = 93 (exit syscall)
			regFile.WriteReg(1, 0x8000) // X1 = base address
			regFile.WriteReg(0, 13)     // X0 = value to store/load
		},
		Program:      buildMemoryRandomAccess(numPairs),
		ExpectedExit: 13,
	}
}

func buildMemoryRandomAccess(numPairs int) []byte {
	// Generate pseudo-random offsets using a simple LCG permutation.
	// Offsets span a 3200-element range (25600 bytes) to create
	// scattered access across multiple cache lines.
	offsets := make([]uint16, numPairs)
	x := uint32(7) // seed
	for i := 0; i < numPairs; i++ {
		x = (x*1103515245 + 12345) & 0x7FFFFFFF
		offsets[i] = uint16(x % 3200) // max offset 3200 * 8 = 25600 bytes
	}

	instrs := make([]uint32, 0, numPairs*2+1)
	for i := 0; i < numPairs; i++ {
		instrs = append(instrs,
			EncodeSTR64(0, 1, offsets[i]),
			EncodeLDR64(0, 1, offsets[i]),
		)
	}
	instrs = append(instrs, EncodeSVC(0))
	return BuildProgram(instrs...)
}

// 10. Load Heavy - Instruction mix dominated by loads
// Tests load unit throughput and memory subsystem pressure.
// Runs in a 10-iteration loop to amortize pipeline startup/drain,
// matching native calibration methodology. Each iteration:
// 20 LDR + 3-instruction loop overhead, instructions_per_iter=23.
// Total retired: 10*23 = 230 instructions (SVC terminates, not retired).
func loadHeavy() Benchmark {
	return Benchmark{
		Name:        "load_heavy",
		Description: "20 loads from sequential addresses - measures load throughput",
		Setup: func(regFile *emu.RegFile, memory *emu.Memory) {
			regFile.WriteReg(8, 93)     // X8 = 93 (exit syscall)
			regFile.WriteReg(1, 0x8000) // X1 = base address
			regFile.WriteReg(21, 10)    // X21 = loop counter (10 iterations)
			// Pre-fill memory with known values
			for i := uint64(0); i < 20; i++ {
				memory.Write64(0x8000+i*8, i+1)
			}
		},
		Program: BuildProgram(
			// loop: 20 loads to distinct registers (no WAW/RAW hazards)
			EncodeLDR64(22, 1, 0),
			EncodeLDR64(2, 1, 1),
			EncodeLDR64(3, 1, 2),
			EncodeLDR64(4, 1, 3),
			EncodeLDR64(5, 1, 4),
			EncodeLDR64(6, 1, 5),
			EncodeLDR64(7, 1, 6),
			EncodeLDR64(9, 1, 7),
			EncodeLDR64(10, 1, 8),
			EncodeLDR64(11, 1, 9),
			EncodeLDR64(12, 1, 10),
			EncodeLDR64(13, 1, 11),
			EncodeLDR64(14, 1, 12),
			EncodeLDR64(15, 1, 13),
			EncodeLDR64(16, 1, 14),
			EncodeLDR64(17, 1, 15),
			EncodeLDR64(18, 1, 16),
			EncodeLDR64(19, 1, 17),
			EncodeLDR64(20, 1, 18),
			EncodeLDR64(0, 1, 19),          // X0 = 20 (exit code after last iter)
			EncodeSUBImm(21, 21, 1, false), // X21 = X21 - 1
			EncodeCMPImm(21, 0),            // CMP X21, #0
			EncodeBCond(-88, 1),            // B.NE loop (-88 = -22 instructions * 4)
			EncodeSVC(0),
		),
		ExpectedExit: 20,
	}
}

// 11. Store Heavy - Instruction mix dominated by stores
// Tests store unit throughput and write buffer behavior.
// Runs in a 10-iteration loop to amortize pipeline startup/drain,
// matching native calibration methodology. Each iteration:
// 20 STR + 3-instruction loop overhead, instructions_per_iter=23.
// Total retired: 10*23 = 230 instructions (SVC terminates, not retired).
func storeHeavy() Benchmark {
	return Benchmark{
		Name:        "store_heavy",
		Description: "20 stores to sequential addresses - measures store throughput",
		Setup: func(regFile *emu.RegFile, memory *emu.Memory) {
			regFile.WriteReg(8, 93)     // X8 = 93 (exit syscall)
			regFile.WriteReg(0, 3)      // X0 = exit code (also stored)
			regFile.WriteReg(1, 0x8000) // X1 = base address
			regFile.WriteReg(2, 99)     // X2 = value to store
			regFile.WriteReg(21, 10)    // X21 = loop counter (10 iterations)
		},
		Program: BuildProgram(
			// loop: 20 stores to sequential addresses (no data dependencies)
			EncodeSTR64(2, 1, 0),
			EncodeSTR64(2, 1, 1),
			EncodeSTR64(2, 1, 2),
			EncodeSTR64(2, 1, 3),
			EncodeSTR64(2, 1, 4),
			EncodeSTR64(2, 1, 5),
			EncodeSTR64(2, 1, 6),
			EncodeSTR64(2, 1, 7),
			EncodeSTR64(2, 1, 8),
			EncodeSTR64(2, 1, 9),
			EncodeSTR64(2, 1, 10),
			EncodeSTR64(2, 1, 11),
			EncodeSTR64(2, 1, 12),
			EncodeSTR64(2, 1, 13),
			EncodeSTR64(2, 1, 14),
			EncodeSTR64(2, 1, 15),
			EncodeSTR64(2, 1, 16),
			EncodeSTR64(2, 1, 17),
			EncodeSTR64(2, 1, 18),
			EncodeSTR64(2, 1, 19),
			EncodeSUBImm(21, 21, 1, false), // X21 = X21 - 1
			EncodeCMPImm(21, 0),            // CMP X21, #0
			EncodeBCond(-88, 1),            // B.NE loop (-88 = -22 instructions * 4)
			EncodeSVC(0),
		),
		ExpectedExit: 3,
	}
}

// 10b. Load Heavy Scaled - 200 independent loads to amortize cold miss overhead
func loadHeavyScaled() Benchmark {
	const numLoads = 200
	return Benchmark{
		Name:        "load_heavy_scaled",
		Description: "200 loads from sequential addresses - measures load throughput at scale",
		Setup: func(regFile *emu.RegFile, memory *emu.Memory) {
			regFile.WriteReg(8, 93)     // X8 = 93 (exit syscall)
			regFile.WriteReg(1, 0x8000) // X1 = base address
			for i := uint64(0); i < numLoads; i++ {
				memory.Write64(0x8000+i*8, i+1)
			}
		},
		Program:      buildLoadHeavyScaled(numLoads),
		ExpectedExit: int64(numLoads),
	}
}

func buildLoadHeavyScaled(n int) []byte {
	// Use registers X0, X2-X7, X9-X20 (avoid X1=base, X8=syscall)
	destRegs := []uint8{0, 2, 3, 4, 5, 6, 7, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	instrs := make([]uint32, 0, n+1)
	for i := 0; i < n; i++ {
		reg := destRegs[i%len(destRegs)]
		instrs = append(instrs, EncodeLDR64(reg, 1, uint16(i)))
	}
	// Final load into X0 for exit code
	instrs[n-1] = EncodeLDR64(0, 1, uint16(n-1))
	instrs = append(instrs, EncodeSVC(0))
	return BuildProgram(instrs...)
}

// 11b. Store Heavy Scaled - 200 independent stores to amortize cold miss overhead
func storeHeavyScaled() Benchmark {
	const numStores = 200
	return Benchmark{
		Name:        "store_heavy_scaled",
		Description: "200 stores to sequential addresses - measures store throughput at scale",
		Setup: func(regFile *emu.RegFile, memory *emu.Memory) {
			regFile.WriteReg(8, 93)     // X8 = 93 (exit syscall)
			regFile.WriteReg(0, 3)      // X0 = exit code
			regFile.WriteReg(1, 0x8000) // X1 = base address
			regFile.WriteReg(2, 99)     // X2 = value to store
		},
		Program:      buildStoreHeavyScaled(numStores),
		ExpectedExit: 3,
	}
}

func buildStoreHeavyScaled(n int) []byte {
	instrs := make([]uint32, 0, n+1)
	for i := 0; i < n; i++ {
		instrs = append(instrs, EncodeSTR64(2, 1, uint16(i)))
	}
	instrs = append(instrs, EncodeSVC(0))
	return BuildProgram(instrs...)
}

// 12. Branch Heavy - High branch density to stress branch prediction
// Alternating taken/not-taken conditional branches.
func branchHeavy() Benchmark {
	return Benchmark{
		Name:        "branch_heavy",
		Description: "10 conditional branches (alternating taken/not-taken) - stresses branch predictor",
		Setup: func(regFile *emu.RegFile, memory *emu.Memory) {
			regFile.WriteReg(8, 93) // X8 = 93 (exit syscall)
			regFile.WriteReg(0, 0)  // X0 = 0 (result counter)
			regFile.WriteReg(1, 5)  // X1 = 5 (comparison value)
		},
		Program: BuildProgram(
			// Pattern: CMP X0, X1; B.LT +8 (taken while X0 < 5)
			// Then increment X0, so first 5 branches taken, last 5 not taken

			// Branch 1: X0=0 < 5, taken (skip ADD X1)
			EncodeCMPReg(0, 1),            // CMP X0, X1
			EncodeBCond(8, 11),            // B.LT +8 (CondLT = 11)
			EncodeADDImm(1, 1, 99, false), // skipped (would corrupt X1)
			EncodeADDImm(0, 0, 1, false),  // X0 += 1

			// Branch 2: X0=1 < 5, taken
			EncodeCMPReg(0, 1),
			EncodeBCond(8, 11),
			EncodeADDImm(1, 1, 99, false),
			EncodeADDImm(0, 0, 1, false),

			// Branch 3: X0=2 < 5, taken
			EncodeCMPReg(0, 1),
			EncodeBCond(8, 11),
			EncodeADDImm(1, 1, 99, false),
			EncodeADDImm(0, 0, 1, false),

			// Branch 4: X0=3 < 5, taken
			EncodeCMPReg(0, 1),
			EncodeBCond(8, 11),
			EncodeADDImm(1, 1, 99, false),
			EncodeADDImm(0, 0, 1, false),

			// Branch 5: X0=4 < 5, taken
			EncodeCMPReg(0, 1),
			EncodeBCond(8, 11),
			EncodeADDImm(1, 1, 99, false),
			EncodeADDImm(0, 0, 1, false),

			// Branch 6: X0=5 >= 5, NOT taken (falls through to corrupt + add)
			EncodeCMPReg(0, 1),
			EncodeBCond(8, 11),
			EncodeADDImm(3, 3, 1, false), // X3 += 1 (not-taken counter)
			EncodeADDImm(0, 0, 1, false), // X0 += 1

			// Branch 7: X0=6 >= 5, NOT taken
			EncodeCMPReg(0, 1),
			EncodeBCond(8, 11),
			EncodeADDImm(3, 3, 1, false),
			EncodeADDImm(0, 0, 1, false),

			// Branch 8: X0=7 >= 5, NOT taken
			EncodeCMPReg(0, 1),
			EncodeBCond(8, 11),
			EncodeADDImm(3, 3, 1, false),
			EncodeADDImm(0, 0, 1, false),

			// Branch 9: X0=8 >= 5, NOT taken
			EncodeCMPReg(0, 1),
			EncodeBCond(8, 11),
			EncodeADDImm(3, 3, 1, false),
			EncodeADDImm(0, 0, 1, false),

			// Branch 10: X0=9 >= 5, NOT taken
			EncodeCMPReg(0, 1),
			EncodeBCond(8, 11),
			EncodeADDImm(3, 3, 1, false),
			EncodeADDImm(0, 0, 1, false),

			EncodeSVC(0), // exit with X0 = 10
		),
		ExpectedExit: 10,
	}
}

// 13. Vector Sum - Loop summing array elements
// Tests a realistic reduction loop: load from array, accumulate, branch back.
// Pattern: for i := 0; i < N; i++ { sum += A[i] }
func vectorSum() Benchmark {
	const n = 16
	return Benchmark{
		Name:        "vector_sum",
		Description: "16-iteration loop summing array elements - tests load+accumulate loop",
		Setup: func(regFile *emu.RegFile, memory *emu.Memory) {
			regFile.WriteReg(8, 93)     // X8 = 93 (exit syscall)
			regFile.WriteReg(0, 0)      // X0 = sum = 0
			regFile.WriteReg(1, 0x8000) // X1 = base address of array
			regFile.WriteReg(2, 0)      // X2 = i = 0
			regFile.WriteReg(3, n)      // X3 = N = 16
			// Fill array: A[i] = i + 1
			for i := uint64(0); i < n; i++ {
				memory.Write64(0x8000+i*8, i+1)
			}
		},
		// sum = 1+2+3+...+16 = 136; exit code = 136
		Program: BuildProgram(
			// loop:
			EncodeLDR64(4, 1, 0),         // X4 = A[X1] (load current element)
			EncodeADDReg(0, 0, 4, false), // sum += X4
			EncodeADDImm(1, 1, 8, false), // X1 += 8 (advance pointer)
			EncodeADDImm(2, 2, 1, false), // i++
			EncodeCMPReg(2, 3),           // CMP i, N
			EncodeBCond(-20, 11),         // B.LT loop (-20 bytes = -5 instructions)
			// exit:
			EncodeSVC(0), // exit with X0 = 136
		),
		ExpectedExit: 136, // sum(1..16) = 136
	}
}

// 14. Vector Add - Loop adding two arrays element-by-element
// Tests load+load+ALU+store loop, common in scientific computing.
// Pattern: for i := 0; i < N; i++ { C[i] = A[i] + B[i] }
func vectorAdd() Benchmark {
	const n = 16
	return Benchmark{
		Name:        "vector_add",
		Description: "16-iteration vector add loop (C=A+B) - tests load+ALU+store loop",
		Setup: func(regFile *emu.RegFile, memory *emu.Memory) {
			regFile.WriteReg(8, 93)     // X8 = 93 (exit syscall)
			regFile.WriteReg(1, 0x8000) // X1 = A base
			regFile.WriteReg(2, 0x8100) // X2 = B base
			regFile.WriteReg(3, 0x8200) // X3 = C base
			regFile.WriteReg(4, 0)      // X4 = i = 0
			regFile.WriteReg(5, n)      // X5 = N = 16
			// A[i] = i + 1, B[i] = 2*(i+1)
			for i := uint64(0); i < n; i++ {
				memory.Write64(0x8000+i*8, i+1)
				memory.Write64(0x8100+i*8, 2*(i+1))
			}
		},
		// C[i] = A[i]+B[i] = 3*(i+1)
		// Verify: load C[0] = 3 as exit code
		Program: BuildProgram(
			// loop:
			EncodeLDR64(6, 1, 0),         // X6 = A[i]
			EncodeLDR64(7, 2, 0),         // X7 = B[i]
			EncodeADDReg(9, 6, 7, false), // X9 = A[i] + B[i]
			EncodeSTR64(9, 3, 0),         // C[i] = X9
			EncodeADDImm(1, 1, 8, false), // A ptr += 8
			EncodeADDImm(2, 2, 8, false), // B ptr += 8
			EncodeADDImm(3, 3, 8, false), // C ptr += 8
			EncodeADDImm(4, 4, 1, false), // i++
			EncodeCMPReg(4, 5),           // CMP i, N
			EncodeBCond(-36, 11),         // B.LT loop (-36 bytes = -9 instructions)
			// Verify: load C[0] as exit code
			EncodeSUBImm(3, 3, 128, false), // X3 -= 128 (back to C base: 16*8=128)
			EncodeLDR64(0, 3, 0),           // X0 = C[0] = 3
			EncodeSVC(0),
		),
		ExpectedExit: 3, // C[0] = A[0]+B[0] = 1+2 = 3
	}
}

// 15. Reduction Tree - Parallel reduction over 16 values
// Tests ILP: independent partial sums that merge in a tree pattern.
// Unlike vector_sum's serial chain, this has log2(N) dependency depth.
func reductionTree() Benchmark {
	return Benchmark{
		Name:        "reduction_tree",
		Description: "16-element parallel reduction tree - tests ILP in reduction",
		Setup: func(regFile *emu.RegFile, memory *emu.Memory) {
			regFile.WriteReg(8, 93)     // X8 = 93 (exit syscall)
			regFile.WriteReg(1, 0x8000) // X1 = array base
			// Fill 16 elements: values 1..16
			for i := uint64(0); i < 16; i++ {
				memory.Write64(0x8000+i*8, i+1)
			}
		},
		// Tree reduction: sum(1..16) = 136
		Program: BuildProgram(
			// Load all 16 elements into registers (level 0)
			EncodeLDR64(0, 1, 0),   // r0 = 1
			EncodeLDR64(2, 1, 1),   // r2 = 2
			EncodeLDR64(3, 1, 2),   // r3 = 3
			EncodeLDR64(4, 1, 3),   // r4 = 4
			EncodeLDR64(5, 1, 4),   // r5 = 5
			EncodeLDR64(6, 1, 5),   // r6 = 6
			EncodeLDR64(7, 1, 6),   // r7 = 7
			EncodeLDR64(9, 1, 7),   // r9 = 8
			EncodeLDR64(10, 1, 8),  // r10 = 9
			EncodeLDR64(11, 1, 9),  // r11 = 10
			EncodeLDR64(12, 1, 10), // r12 = 11
			EncodeLDR64(13, 1, 11), // r13 = 12
			EncodeLDR64(14, 1, 12), // r14 = 13
			EncodeLDR64(15, 1, 13), // r15 = 14
			EncodeLDR64(16, 1, 14), // r16 = 15
			EncodeLDR64(17, 1, 15), // r17 = 16

			// Level 1: 8 independent pairwise sums
			EncodeADDReg(0, 0, 2, false),    // r0 = 1+2 = 3
			EncodeADDReg(3, 3, 4, false),    // r3 = 3+4 = 7
			EncodeADDReg(5, 5, 6, false),    // r5 = 5+6 = 11
			EncodeADDReg(7, 7, 9, false),    // r7 = 7+8 = 15
			EncodeADDReg(10, 10, 11, false), // r10 = 9+10 = 19
			EncodeADDReg(12, 12, 13, false), // r12 = 11+12 = 23
			EncodeADDReg(14, 14, 15, false), // r14 = 13+14 = 27
			EncodeADDReg(16, 16, 17, false), // r16 = 15+16 = 31

			// Level 2: 4 independent sums
			EncodeADDReg(0, 0, 3, false),    // r0 = 3+7 = 10
			EncodeADDReg(5, 5, 7, false),    // r5 = 11+15 = 26
			EncodeADDReg(10, 10, 12, false), // r10 = 19+23 = 42
			EncodeADDReg(14, 14, 16, false), // r14 = 27+31 = 58

			// Level 3: 2 independent sums
			EncodeADDReg(0, 0, 5, false),    // r0 = 10+26 = 36
			EncodeADDReg(10, 10, 14, false), // r10 = 42+58 = 100

			// Level 4: final sum
			EncodeADDReg(0, 0, 10, false), // r0 = 36+100 = 136

			EncodeSVC(0), // exit with X0 = 136
		),
		ExpectedExit: 136, // sum(1..16) = 136
	}
}

// 16. Stride Indirect - Pointer chasing through array
// Each load uses the result of the previous load as the address offset.
// Tests memory-level parallelism limits: loads are serialized by data dependency.
// Pattern: addr = base + A[addr]; repeat
func strideIndirect() Benchmark {
	const n = 8
	return Benchmark{
		Name:        "stride_indirect",
		Description: "8-element pointer chase (dependent loads) - tests load-to-use latency chain",
		Setup: func(regFile *emu.RegFile, memory *emu.Memory) {
			regFile.WriteReg(8, 93)     // X8 = 93 (exit syscall)
			regFile.WriteReg(1, 0x8000) // X1 = base address
			regFile.WriteReg(2, 0)      // X2 = current offset (in 8-byte units)
			regFile.WriteReg(3, 0)      // X3 = hop counter
			regFile.WriteReg(4, n)      // X4 = max hops
			// Build a chain: A[0]=3, A[3]=1, A[1]=5, A[5]=2, A[2]=7, A[7]=4, A[4]=6, A[6]=0
			// Chain: 0→3→1→5→2→7→4→6→0 (8 hops, visits all)
			memory.Write64(0x8000+0*8, 3)
			memory.Write64(0x8000+3*8, 1)
			memory.Write64(0x8000+1*8, 5)
			memory.Write64(0x8000+5*8, 2)
			memory.Write64(0x8000+2*8, 7)
			memory.Write64(0x8000+7*8, 4)
			memory.Write64(0x8000+4*8, 6)
			memory.Write64(0x8000+6*8, 0)
		},
		Program: buildStrideIndirect(n),
		// After 8 hops: 0→3→1→5→2→7→4→6, X2 ends at 6; but final LDR loads A[6]=0
		// Actually, after 8 hops: hop 1: load A[0]=3, hop 2: load A[3]=1, ...
		// hop 8: load A[6]=0. So X2 = 0 at exit. Exit code = hop count = 8.
		ExpectedExit: int64(n),
	}
}

func buildStrideIndirect(n int) []byte {
	// loop body (6 instructions per hop, matching native LSL encoding):
	//   X5 = X2 << 3 (shift left by 3 for 8-byte elements) — ADD X5, XZR, X2, LSL #3
	//   X5 = X1 + X5 (compute address)
	//   X2 = [X5] (load next offset, encoded as LDR X2, [X5, #0])
	//   X3++ (hop count)
	//   CMP X3, X4
	//   B.LT loop
	instrs := []uint32{
		// loop:
		EncodeADDRegShifted(5, 31, 2, 3, 0, false), // X5 = XZR + (X2 << 3) = X2 * 8
		EncodeADDReg(5, 1, 5, false),               // X5 = base + offset*8
		EncodeLDR64(2, 5, 0),                       // X2 = [X5] (next index)
		EncodeADDImm(3, 3, 1, false),               // X3++ (hop count)
		EncodeCMPReg(3, 4),                         // CMP X3, N
		EncodeBCond(-20, 11),                       // B.LT loop (-20 bytes = -5 instructions)
		// exit:
		EncodeADDReg(0, 3, 31, false), // X0 = X3 (hop count), XZR add
	}
	// Actually, ADD Rd, Rn, XZR is equivalent to MOV Rd, Rn
	// X31 = XZR = 0 in register reads, so EncodeADDReg(0, 3, 31) = X0 = X3 + 0 = X3
	instrs = append(instrs, EncodeSVC(0))
	return BuildProgram(instrs...)
}

// EncodeCMPReg encodes compare register: CMP Xn, Xm
// This is an alias for SUBS XZR, Xn, Xm (sets flags, discards result)
func EncodeCMPReg(rn, rm uint8) uint32 {
	return EncodeSUBReg(31, rn, rm, true)
}

// Note: encodeMUL removed - scalar MUL/MADD not yet implemented in simulator
