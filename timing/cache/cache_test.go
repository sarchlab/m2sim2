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
			Expect(config.Size).To(Equal(16 * 1024 * 1024))
			Expect(config.Associativity).To(Equal(16))
			Expect(config.BlockSize).To(Equal(128))
			Expect(config.HitLatency).To(Equal(uint64(12)))
			Expect(config.MissLatency).To(Equal(uint64(200)))
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
