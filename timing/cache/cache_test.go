package cache_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sarchlab/m2sim/emu"
	"github.com/sarchlab/m2sim/timing/cache"
)

var _ = Describe("Cache", func() {
	var (
		c       *cache.Cache
		memory  *emu.Memory
		backing *cache.MemoryBacking
	)

	BeforeEach(func() {
		memory = emu.NewMemory()
		backing = cache.NewMemoryBacking(memory)
		// Small cache for testing: 4KB, 4-way, 64B lines
		config := cache.Config{
			Size:          4 * 1024,
			Associativity: 4,
			BlockSize:     64,
			HitLatency:    1,
			MissLatency:   10,
		}
		c = cache.New(config, backing)
	})

	Describe("Read operations", func() {
		It("should miss on cold cache", func() {
			// Write data to memory first
			memory.Write64(0x1000, 0xDEADBEEF)

			result := c.Read(0x1000, 8)
			Expect(result.Hit).To(BeFalse())
			Expect(result.Latency).To(Equal(uint64(10)))
			Expect(result.Data).To(Equal(uint64(0xDEADBEEF)))

			stats := c.Stats()
			Expect(stats.Reads).To(Equal(uint64(1)))
			Expect(stats.Misses).To(Equal(uint64(1)))
			Expect(stats.Hits).To(Equal(uint64(0)))
		})

		It("should hit on cached data", func() {
			memory.Write64(0x1000, 0xCAFEBABE)

			// First read - miss
			c.Read(0x1000, 8)

			// Second read - should hit
			result := c.Read(0x1000, 8)
			Expect(result.Hit).To(BeTrue())
			Expect(result.Latency).To(Equal(uint64(1)))
			Expect(result.Data).To(Equal(uint64(0xCAFEBABE)))

			stats := c.Stats()
			Expect(stats.Reads).To(Equal(uint64(2)))
			Expect(stats.Misses).To(Equal(uint64(1)))
			Expect(stats.Hits).To(Equal(uint64(1)))
		})

		It("should hit on different addresses in same cache line", func() {
			memory.Write32(0x1000, 0x11111111)
			memory.Write32(0x1004, 0x22222222)

			// First read at 0x1000 - miss, loads entire cache line
			c.Read(0x1000, 4)

			// Read at 0x1004 - should hit (same cache line)
			result := c.Read(0x1004, 4)
			Expect(result.Hit).To(BeTrue())
			Expect(result.Data).To(Equal(uint64(0x22222222)))
		})
	})

	Describe("Write operations", func() {
		It("should write-allocate on miss", func() {
			result := c.Write(0x1000, 8, 0x12345678)
			Expect(result.Hit).To(BeFalse())
			Expect(result.Latency).To(Equal(uint64(10)))

			// Subsequent read should hit
			readResult := c.Read(0x1000, 8)
			Expect(readResult.Hit).To(BeTrue())
			Expect(readResult.Data).To(Equal(uint64(0x12345678)))
		})

		It("should hit on cached data", func() {
			// First write - miss
			c.Write(0x1000, 8, 0x11111111)

			// Second write - should hit
			result := c.Write(0x1000, 8, 0x22222222)
			Expect(result.Hit).To(BeTrue())
			Expect(result.Latency).To(Equal(uint64(1)))

			// Verify data
			readResult := c.Read(0x1000, 8)
			Expect(readResult.Data).To(Equal(uint64(0x22222222)))
		})
	})

	Describe("Eviction", func() {
		It("should evict when cache is full", func() {
			// 4KB cache, 64B lines, 4-way = 16 sets
			// Fill one set completely (4 ways), then access one more
			// Set 0 addresses: 0, 1024, 2048, 3072, 4096 (all map to set 0)
			// (assuming sets = 4KB / (4 * 64) = 16 sets)

			// Fill set 0 with 4 blocks
			c.Write(0x0000, 8, 0x11111111) // Set 0, way 0
			c.Write(0x0400, 8, 0x22222222) // Set 0, way 1
			c.Write(0x0800, 8, 0x33333333) // Set 0, way 2
			c.Write(0x0C00, 8, 0x44444444) // Set 0, way 3

			// All should hit now
			Expect(c.Read(0x0000, 8).Hit).To(BeTrue())
			Expect(c.Read(0x0400, 8).Hit).To(BeTrue())
			Expect(c.Read(0x0800, 8).Hit).To(BeTrue())
			Expect(c.Read(0x0C00, 8).Hit).To(BeTrue())

			// Access 5th address in same set - should evict LRU
			result := c.Write(0x1000, 8, 0x55555555)
			Expect(result.Hit).To(BeFalse())
			Expect(result.Evicted).To(BeTrue())

			stats := c.Stats()
			Expect(stats.Evictions).To(Equal(uint64(1)))
		})

		It("should writeback dirty evicted blocks", func() {
			// Fill set 0 completely
			c.Write(0x0000, 8, 0x11111111)
			c.Write(0x0400, 8, 0x22222222)
			c.Write(0x0800, 8, 0x33333333)
			c.Write(0x0C00, 8, 0x44444444)

			// Access the first three to make 0x0000 the LRU
			c.Read(0x0400, 8)
			c.Read(0x0800, 8)
			c.Read(0x0C00, 8)

			// Evict - should write back 0x0000
			c.Write(0x1000, 8, 0x55555555)

			// Check memory was written back
			Expect(memory.Read64(0x0000)).To(Equal(uint64(0x11111111)))

			stats := c.Stats()
			Expect(stats.Writebacks).To(Equal(uint64(1)))
		})
	})

	Describe("Flush", func() {
		It("should write back all dirty blocks", func() {
			c.Write(0x0000, 8, 0x11111111)
			c.Write(0x1000, 8, 0x22222222)

			// Data not yet in memory (only in cache)
			Expect(memory.Read64(0x0000)).To(Equal(uint64(0)))
			Expect(memory.Read64(0x1000)).To(Equal(uint64(0)))

			c.Flush()

			// After flush, data should be in memory
			Expect(memory.Read64(0x0000)).To(Equal(uint64(0x11111111)))
			Expect(memory.Read64(0x1000)).To(Equal(uint64(0x22222222)))

			stats := c.Stats()
			Expect(stats.Writebacks).To(Equal(uint64(2)))
		})
	})

	Describe("Default configurations", func() {
		It("should create L1I config", func() {
			config := cache.DefaultL1IConfig()
			Expect(config.Size).To(Equal(192 * 1024))
			Expect(config.Associativity).To(Equal(6))
			Expect(config.BlockSize).To(Equal(64))
		})

		It("should create L1D config", func() {
			config := cache.DefaultL1DConfig()
			Expect(config.Size).To(Equal(128 * 1024))
			Expect(config.Associativity).To(Equal(8))
			Expect(config.BlockSize).To(Equal(64))
		})

		It("should create L2 config", func() {
			config := cache.DefaultL2Config()
			Expect(config.Size).To(Equal(24 * 1024 * 1024)) // M2 has 24MB shared L2
			Expect(config.Associativity).To(Equal(16))
			Expect(config.BlockSize).To(Equal(128))
			Expect(config.HitLatency).To(Equal(uint64(12)))
			Expect(config.MissLatency).To(Equal(uint64(150))) // Unified memory is faster
		})

		It("should create L2 per-core config", func() {
			config := cache.DefaultL2PerCoreConfig()
			Expect(config.Size).To(Equal(512 * 1024))
			Expect(config.Associativity).To(Equal(8))
			Expect(config.BlockSize).To(Equal(128))
		})
	})
})

var _ = Describe("Cache Hierarchy", func() {
	var (
		l1     *cache.Cache
		l2     *cache.Cache
		memory *emu.Memory
	)

	BeforeEach(func() {
		memory = emu.NewMemory()
		memBacking := cache.NewMemoryBacking(memory)

		// L2 cache: 8KB, 4-way, 128B lines
		l2Config := cache.Config{
			Size:          8 * 1024,
			Associativity: 4,
			BlockSize:     128,
			HitLatency:    10,
			MissLatency:   100,
		}
		l2 = cache.New(l2Config, memBacking)

		// L1 cache: 2KB, 2-way, 64B lines, backed by L2
		l2Backing := cache.NewCacheBacking(l2)
		l1Config := cache.Config{
			Size:          2 * 1024,
			Associativity: 2,
			BlockSize:     64,
			HitLatency:    1,
			MissLatency:   10,
		}
		l1 = cache.New(l1Config, l2Backing)
	})

	Describe("Hierarchical reads", func() {
		It("should propagate read through hierarchy on cold miss", func() {
			// Write data to memory
			memory.Write64(0x1000, 0xDEADBEEF)

			// Read through L1 - should miss L1, miss L2, hit memory
			result := l1.Read(0x1000, 8)
			Expect(result.Hit).To(BeFalse())
			Expect(result.Data).To(Equal(uint64(0xDEADBEEF)))

			// L1 and L2 should both have misses
			l1Stats := l1.Stats()
			l2Stats := l2.Stats()
			Expect(l1Stats.Misses).To(Equal(uint64(1)))
			Expect(l2Stats.Misses).To(BeNumerically(">", 0)) // L2 accessed during L1 miss
		})

		It("should hit L1 on repeated access", func() {
			memory.Write64(0x1000, 0xCAFEBABE)

			// First read - misses
			l1.Read(0x1000, 8)

			// Second read - should hit L1
			result := l1.Read(0x1000, 8)
			Expect(result.Hit).To(BeTrue())
			Expect(result.Data).To(Equal(uint64(0xCAFEBABE)))
			Expect(result.Latency).To(Equal(uint64(1)))
		})

		It("should hit L2 after L1 eviction", func() {
			memory.Write64(0x1000, 0x11111111)

			// Fill L1 set causing eviction (L1 is 2KB, 2-way, 64B = 16 sets)
			// Addresses that map to same set: 0x1000, 0x1400 (1024 apart for 16 sets)
			l1.Read(0x1000, 8) // Fill L1
			l1.Read(0x1400, 8) // Fill L1, same set
			l1.Read(0x1800, 8) // Evict from L1, but data still in L2

			// Record L2 stats before accessing the evicted address
			l2StatsBefore := l2.Stats()

			// Access evicted address - should miss L1 but data comes from L2
			result := l1.Read(0x1000, 8)
			Expect(result.Hit).To(BeFalse()) // L1 miss
			Expect(result.Data).To(Equal(uint64(0x11111111)))

			// L2 should have had hits (not misses) for this access
			// Note: Multiple L2 accesses occur because CacheBacking reads in 8-byte chunks
			l2StatsAfter := l2.Stats()
			Expect(l2StatsAfter.Hits - l2StatsBefore.Hits).To(BeNumerically(">", 0))
			Expect(l2StatsAfter.Misses - l2StatsBefore.Misses).To(Equal(uint64(0)))
		})
	})

	Describe("Hierarchical writes", func() {
		It("should write through hierarchy", func() {
			// Write through L1
			l1.Write(0x2000, 8, 0x12345678)

			// Flush L1 to push dirty data to L2
			l1.Flush()

			// Flush L2 to push to memory
			l2.Flush()

			// Verify data reached memory
			Expect(memory.Read64(0x2000)).To(Equal(uint64(0x12345678)))
		})

		It("should maintain coherence on read-after-write", func() {
			// Write data
			l1.Write(0x3000, 8, 0xABCDABCD)

			// Read it back
			result := l1.Read(0x3000, 8)
			Expect(result.Hit).To(BeTrue())
			Expect(result.Data).To(Equal(uint64(0xABCDABCD)))
		})
	})

	Describe("CacheBacking interface", func() {
		It("should expose underlying cache for statistics", func() {
			memBacking := cache.NewMemoryBacking(memory)
			l2Config := cache.Config{
				Size:          4 * 1024,
				Associativity: 4,
				BlockSize:     64,
				HitLatency:    10,
				MissLatency:   100,
			}
			l2Cache := cache.New(l2Config, memBacking)
			cacheBacking := cache.NewCacheBacking(l2Cache)

			// Verify we can access the underlying cache
			Expect(cacheBacking.Cache()).To(Equal(l2Cache))
		})
	})
})

var _ = Describe("Akita Cache Verification", func() {
	var (
		memory  *emu.Memory
		backing *cache.MemoryBacking
	)

	BeforeEach(func() {
		memory = emu.NewMemory()
		backing = cache.NewMemoryBacking(memory)
	})

	Describe("LRU ordering", func() {
		It("should evict the least recently used block", func() {
			// 4KB, 4-way, 64B lines = 16 sets. Set stride = 1024 (0x400).
			config := cache.Config{
				Size:          4 * 1024,
				Associativity: 4,
				BlockSize:     64,
				HitLatency:    1,
				MissLatency:   10,
			}
			c := cache.New(config, backing)

			// Fill set 0 with blocks A, B, C, D (all map to set 0)
			addrA := uint64(0x0000) // set 0
			addrB := uint64(0x0400) // set 0
			addrC := uint64(0x0800) // set 0
			addrD := uint64(0x0C00) // set 0

			c.Write(addrA, 8, 0xAAAA)
			c.Write(addrB, 8, 0xBBBB)
			c.Write(addrC, 8, 0xCCCC)
			c.Write(addrD, 8, 0xDDDD)

			// Access B and D to make them recently used.
			// LRU order (oldest first): A, C, B, D
			c.Read(addrB, 8)
			c.Read(addrD, 8)

			// 5th block should evict A (the LRU)
			addrE := uint64(0x1000) // set 0
			result := c.Write(addrE, 8, 0xEEEE)
			Expect(result.Evicted).To(BeTrue())
			Expect(result.EvictedAddr).To(Equal(addrA))

			// B/C/D/E should still be present (don't read A — it would
			// cause another eviction and disturb the set)
			Expect(c.Read(addrB, 8).Hit).To(BeTrue())
			Expect(c.Read(addrC, 8).Hit).To(BeTrue())
			Expect(c.Read(addrD, 8).Hit).To(BeTrue())
			Expect(c.Read(addrE, 8).Hit).To(BeTrue())
		})

		It("should track LRU order across reads and writes", func() {
			config := cache.Config{
				Size:          4 * 1024,
				Associativity: 4,
				BlockSize:     64,
				HitLatency:    1,
				MissLatency:   10,
			}
			c := cache.New(config, backing)

			addrA := uint64(0x0000)
			addrB := uint64(0x0400)
			addrC := uint64(0x0800)
			addrD := uint64(0x0C00)

			// Fill set 0: A, B, C, D
			c.Write(addrA, 8, 0xAA)
			c.Write(addrB, 8, 0xBB)
			c.Write(addrC, 8, 0xCC)
			c.Write(addrD, 8, 0xDD)

			// Touch A via write, making B the new LRU
			c.Write(addrA, 8, 0xA2)

			// Evict should pick B
			addrE := uint64(0x1000)
			result := c.Write(addrE, 8, 0xEE)
			Expect(result.Evicted).To(BeTrue())
			Expect(result.EvictedAddr).To(Equal(addrB))
		})
	})

	Describe("Default M2 config behavior", func() {
		It("should produce correct latencies with DefaultL1DConfig", func() {
			l1dConfig := cache.DefaultL1DConfig()
			c := cache.New(l1dConfig, backing)

			memory.Write64(0x2000, 0xFACE)

			// Cold miss — latency = MissLatency (12)
			result := c.Read(0x2000, 8)
			Expect(result.Hit).To(BeFalse())
			Expect(result.Latency).To(Equal(uint64(12)))
			Expect(result.Data).To(Equal(uint64(0xFACE)))

			// Warm hit — latency = HitLatency (4)
			result = c.Read(0x2000, 8)
			Expect(result.Hit).To(BeTrue())
			Expect(result.Latency).To(Equal(uint64(4)))
			Expect(result.Data).To(Equal(uint64(0xFACE)))
		})

		It("should produce correct latencies with DefaultL1IConfig", func() {
			l1iConfig := cache.DefaultL1IConfig()
			c := cache.New(l1iConfig, backing)

			memory.Write64(0x3000, 0xC0DE)

			// Cold miss — latency = MissLatency (12)
			result := c.Read(0x3000, 8)
			Expect(result.Hit).To(BeFalse())
			Expect(result.Latency).To(Equal(uint64(12)))

			// Warm hit — latency = HitLatency (1)
			result = c.Read(0x3000, 8)
			Expect(result.Hit).To(BeTrue())
			Expect(result.Latency).To(Equal(uint64(1)))
		})

		It("should correctly configure set count from M2 configs", func() {
			// L1D: 128KB / (8 ways * 64B) = 256 sets
			l1d := cache.DefaultL1DConfig()
			expectedSetsL1D := l1d.Size / (l1d.Associativity * l1d.BlockSize)
			Expect(expectedSetsL1D).To(Equal(256))

			// L1I: 192KB / (6 ways * 64B) = 512 sets
			l1i := cache.DefaultL1IConfig()
			expectedSetsL1I := l1i.Size / (l1i.Associativity * l1i.BlockSize)
			Expect(expectedSetsL1I).To(Equal(512))
		})
	})

	Describe("Associativity enforcement", func() {
		It("should hold exactly N blocks in an N-way set without eviction", func() {
			// 4-way cache: 4 blocks can coexist in the same set
			config := cache.Config{
				Size:          4 * 1024,
				Associativity: 4,
				BlockSize:     64,
				HitLatency:    1,
				MissLatency:   10,
			}
			c := cache.New(config, backing)

			// Set stride = 16 sets * 64B = 1024
			addrs := []uint64{0x0000, 0x0400, 0x0800, 0x0C00}

			for i, addr := range addrs {
				result := c.Write(addr, 8, uint64(i+1))
				Expect(result.Evicted).To(BeFalse(),
					"block %d at 0x%X should not evict", i, addr)
			}

			// All 4 should be present
			for _, addr := range addrs {
				Expect(c.Read(addr, 8).Hit).To(BeTrue())
			}

			stats := c.Stats()
			Expect(stats.Evictions).To(Equal(uint64(0)))
		})

		It("should evict on the (N+1)th block in the same set", func() {
			config := cache.Config{
				Size:          4 * 1024,
				Associativity: 4,
				BlockSize:     64,
				HitLatency:    1,
				MissLatency:   10,
			}
			c := cache.New(config, backing)

			// Fill 4 ways
			addrs := []uint64{0x0000, 0x0400, 0x0800, 0x0C00}
			for i, addr := range addrs {
				c.Write(addr, 8, uint64(i+1))
			}

			// 5th block triggers eviction
			result := c.Write(0x1000, 8, 0x55)
			Expect(result.Evicted).To(BeTrue())

			stats := c.Stats()
			Expect(stats.Evictions).To(Equal(uint64(1)))
		})

		It("should enforce associativity with 2-way cache", func() {
			// 2-way cache: 1KB, 2-way, 64B = 8 sets; stride = 512
			config := cache.Config{
				Size:          1024,
				Associativity: 2,
				BlockSize:     64,
				HitLatency:    1,
				MissLatency:   10,
			}
			c := cache.New(config, backing)

			// Fill 2 ways in set 0
			c.Write(0x0000, 8, 0x11) // set 0, way 0
			c.Write(0x0200, 8, 0x22) // set 0, way 1

			Expect(c.Read(0x0000, 8).Hit).To(BeTrue())
			Expect(c.Read(0x0200, 8).Hit).To(BeTrue())
			Expect(c.Stats().Evictions).To(Equal(uint64(0)))

			// 3rd in same set triggers eviction
			result := c.Write(0x0400, 8, 0x33)
			Expect(result.Evicted).To(BeTrue())
			Expect(c.Stats().Evictions).To(Equal(uint64(1)))
		})
	})

	Describe("Cache-line boundary behavior", func() {
		It("should treat last byte and first byte of adjacent lines as separate lines", func() {
			config := cache.Config{
				Size:          4 * 1024,
				Associativity: 4,
				BlockSize:     64,
				HitLatency:    1,
				MissLatency:   10,
			}
			c := cache.New(config, backing)

			// Last byte of cache line starting at 0x0000 is at offset 63 (0x003F)
			lastByteAddr := uint64(0x003F)
			// First byte of the next cache line is at 0x0040
			firstByteNextLine := uint64(0x0040)

			memory.Write8(lastByteAddr, 0xAA)
			memory.Write8(firstByteNextLine, 0xBB)

			// Access last byte — brings in cache line [0x0000, 0x003F]
			r1 := c.Read(lastByteAddr, 1)
			Expect(r1.Hit).To(BeFalse())
			Expect(byte(r1.Data)).To(Equal(byte(0xAA)))

			// Access first byte of next line — should be a separate miss
			r2 := c.Read(firstByteNextLine, 1)
			Expect(r2.Hit).To(BeFalse())
			Expect(byte(r2.Data)).To(Equal(byte(0xBB)))

			// Both are now cached — re-access should hit
			Expect(c.Read(lastByteAddr, 1).Hit).To(BeTrue())
			Expect(c.Read(firstByteNextLine, 1).Hit).To(BeTrue())

			// Exactly 2 misses, 2 hits on the re-reads
			stats := c.Stats()
			Expect(stats.Misses).To(Equal(uint64(2)))
			Expect(stats.Hits).To(Equal(uint64(2)))
		})

		It("should treat accesses within the same cache line as one line", func() {
			config := cache.Config{
				Size:          4 * 1024,
				Associativity: 4,
				BlockSize:     64,
				HitLatency:    1,
				MissLatency:   10,
			}
			c := cache.New(config, backing)

			// Byte 0 and byte 63 are in the same 64-byte cache line
			memory.Write8(0x1000, 0x11)
			memory.Write8(0x103F, 0xFF)

			c.Read(0x1000, 1) // Miss, fetches line [0x1000, 0x103F]
			r := c.Read(0x103F, 1)
			Expect(r.Hit).To(BeTrue())
			Expect(byte(r.Data)).To(Equal(byte(0xFF)))

			stats := c.Stats()
			Expect(stats.Misses).To(Equal(uint64(1)))
			Expect(stats.Hits).To(Equal(uint64(1)))
		})
	})
})
