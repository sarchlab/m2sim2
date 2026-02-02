// Package emu provides functional ARM64 emulation.
package emu

// LoadStoreUnit implements ARM64 load and store operations.
type LoadStoreUnit struct {
	regFile *RegFile
	memory  *Memory
}

// NewLoadStoreUnit creates a new LoadStoreUnit connected to the given
// register file and memory.
func NewLoadStoreUnit(regFile *RegFile, memory *Memory) *LoadStoreUnit {
	return &LoadStoreUnit{
		regFile: regFile,
		memory:  memory,
	}
}

// LDR64 performs a 64-bit load: Xd = mem[Xn + offset]
func (lsu *LoadStoreUnit) LDR64(rd, rn uint8, offset uint64) {
	base := lsu.regFile.ReadReg(rn)
	addr := base + offset
	value := lsu.memory.Read64(addr)
	lsu.regFile.WriteReg(rd, value)
}

// LDR64SP performs a 64-bit load using SP as base: Xd = mem[SP + offset]
func (lsu *LoadStoreUnit) LDR64SP(rd uint8, offset uint64) {
	addr := lsu.regFile.SP + offset
	value := lsu.memory.Read64(addr)
	lsu.regFile.WriteReg(rd, value)
}

// LDR32 performs a 32-bit load with zero extension: Xd = zero_extend(mem[Xn + offset])
func (lsu *LoadStoreUnit) LDR32(rd, rn uint8, offset uint64) {
	base := lsu.regFile.ReadReg(rn)
	addr := base + offset
	value := lsu.memory.Read32(addr)
	// Zero-extend to 64 bits by storing as uint64
	lsu.regFile.WriteReg(rd, uint64(value))
}

// LDR32SP performs a 32-bit load using SP as base: Xd = zero_extend(mem[SP + offset])
func (lsu *LoadStoreUnit) LDR32SP(rd uint8, offset uint64) {
	addr := lsu.regFile.SP + offset
	value := lsu.memory.Read32(addr)
	lsu.regFile.WriteReg(rd, uint64(value))
}

// STR64 performs a 64-bit store: mem[Xn + offset] = Xd
func (lsu *LoadStoreUnit) STR64(rd, rn uint8, offset uint64) {
	base := lsu.regFile.ReadReg(rn)
	addr := base + offset
	value := lsu.regFile.ReadReg(rd)
	lsu.memory.Write64(addr, value)
}

// STR64SP performs a 64-bit store using SP as base: mem[SP + offset] = Xd
func (lsu *LoadStoreUnit) STR64SP(rd uint8, offset uint64) {
	addr := lsu.regFile.SP + offset
	value := lsu.regFile.ReadReg(rd)
	lsu.memory.Write64(addr, value)
}

// STR32 performs a 32-bit store: mem[Xn + offset] = Wd (lower 32 bits)
func (lsu *LoadStoreUnit) STR32(rd, rn uint8, offset uint64) {
	base := lsu.regFile.ReadReg(rn)
	addr := base + offset
	value := uint32(lsu.regFile.ReadReg(rd))
	lsu.memory.Write32(addr, value)
}

// STR32SP performs a 32-bit store using SP as base: mem[SP + offset] = Wd
func (lsu *LoadStoreUnit) STR32SP(rd uint8, offset uint64) {
	addr := lsu.regFile.SP + offset
	value := uint32(lsu.regFile.ReadReg(rd))
	lsu.memory.Write32(addr, value)
}
