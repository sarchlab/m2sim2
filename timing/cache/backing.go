// Package cache provides cache hierarchy modeling using Akita cache components.
package cache

import (
	"github.com/sarchlab/m2sim/emu"
)

// MemoryBacking wraps emu.Memory as a BackingStore.
type MemoryBacking struct {
	memory *emu.Memory
}

// NewMemoryBacking creates a new MemoryBacking adapter.
func NewMemoryBacking(memory *emu.Memory) *MemoryBacking {
	return &MemoryBacking{memory: memory}
}

// Read fetches data from the backing memory.
func (m *MemoryBacking) Read(addr uint64, size int) []byte {
	data := make([]byte, size)
	for i := 0; i < size; i++ {
		data[i] = m.memory.Read8(addr + uint64(i))
	}
	return data
}

// Write stores data to the backing memory.
func (m *MemoryBacking) Write(addr uint64, data []byte) {
	for i, b := range data {
		m.memory.Write8(addr+uint64(i), b)
	}
}

// CacheBacking wraps a Cache as a BackingStore.
// This enables hierarchical cache configurations (e.g., L1 → L2 → Memory).
type CacheBacking struct {
	cache *Cache
}

// NewCacheBacking creates a new CacheBacking adapter.
func NewCacheBacking(cache *Cache) *CacheBacking {
	return &CacheBacking{cache: cache}
}

// Read fetches data from the backing cache.
func (c *CacheBacking) Read(addr uint64, size int) []byte {
	data := make([]byte, size)

	// Aggregate reads to avoid one cache access per byte.
	// We read up to 8 bytes at a time and unpack them from the uint64 Data field.
	offset := 0
	for offset < size {
		chunkSize := size - offset
		if chunkSize > 8 {
			chunkSize = 8
		}

		result := c.cache.Read(addr+uint64(offset), chunkSize)
		word := result.Data

		for i := 0; i < chunkSize; i++ {
			data[offset+i] = byte(word & 0xff)
			word >>= 8
		}

		offset += chunkSize
	}

	return data
}

// Write stores data to the backing cache.
func (c *CacheBacking) Write(addr uint64, data []byte) {
	size := len(data)
	offset := 0

	// Aggregate writes to avoid one cache access per byte.
	// We pack up to 8 bytes into a uint64 and write them in a single cache access.
	for offset < size {
		chunkSize := size - offset
		if chunkSize > 8 {
			chunkSize = 8
		}

		var word uint64
		// Construct the word so that the least significant byte corresponds
		// to the lowest address (little-endian layout).
		for i := chunkSize - 1; i >= 0; i-- {
			word <<= 8
			word |= uint64(data[offset+i])
		}

		c.cache.Write(addr+uint64(offset), chunkSize, word)
		offset += chunkSize
	}
}

// Cache returns the underlying cache for statistics access.
func (c *CacheBacking) Cache() *Cache {
	return c.cache
}
