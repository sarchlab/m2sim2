// Package emu provides functional ARM64 emulation.
package emu

import "encoding/binary"

// Memory provides a simple byte-addressable memory model for emulation.
type Memory struct {
	data map[uint64]byte
}

// NewMemory creates a new memory instance.
func NewMemory() *Memory {
	return &Memory{
		data: make(map[uint64]byte),
	}
}

// Read8 reads a single byte from memory.
func (m *Memory) Read8(addr uint64) byte {
	return m.data[addr]
}

// Write8 writes a single byte to memory.
func (m *Memory) Write8(addr uint64, value byte) {
	m.data[addr] = value
}

// Read32 reads a 32-bit little-endian value from memory.
func (m *Memory) Read32(addr uint64) uint32 {
	var buf [4]byte
	for i := uint64(0); i < 4; i++ {
		buf[i] = m.data[addr+i]
	}
	return binary.LittleEndian.Uint32(buf[:])
}

// Write32 writes a 32-bit little-endian value to memory.
func (m *Memory) Write32(addr uint64, value uint32) {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], value)
	for i := uint64(0); i < 4; i++ {
		m.data[addr+i] = buf[i]
	}
}

// Read64 reads a 64-bit little-endian value from memory.
func (m *Memory) Read64(addr uint64) uint64 {
	var buf [8]byte
	for i := uint64(0); i < 8; i++ {
		buf[i] = m.data[addr+i]
	}
	return binary.LittleEndian.Uint64(buf[:])
}

// Write64 writes a 64-bit little-endian value to memory.
func (m *Memory) Write64(addr uint64, value uint64) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], value)
	for i := uint64(0); i < 8; i++ {
		m.data[addr+i] = buf[i]
	}
}
