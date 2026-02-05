# Apple M2 Microarchitecture Notes

**Last updated:** 2026-02-05 (Cycle 213)
**Purpose:** Guide accuracy tuning for M2Sim

## Branch Prediction

### Key Findings (from reflexive.space research)

The M2 "Avalanche" P-cores use a sophisticated branch predictor:

1. **Local Prediction with Saturating Counters**
   - Uses 2-bit saturating counters for per-branch history
   - States: Strongly not-taken (00), Weakly not-taken (01), Weakly taken (10), Strongly taken (11)
   - First branch at new address defaults to predict "not-taken"

2. **Training Behavior**
   - Counter updates after each branch resolution
   - Correct prediction strengthens confidence
   - Misprediction weakens confidence (may flip direction)

3. **Implications for M2Sim**
   - Our branch predictor should default to "not-taken" for cold branches
   - Need proper 2-bit counter implementation
   - Single misprediction shouldn't immediately flip prediction direction

### Branch Target Buffer (BTB)
- M2 has complex BTB organization (research paper: MDPI Electronics 2025)
- Heterogeneous behavior between P-cores and E-cores
- Limited public documentation due to macOS PMU restrictions

## Execution Width

### P-cores (Avalanche)
- **Issue width:** 6-wide (reported)
- **ALU units:** Multiple integer ALUs
- **Current M2Sim:** 4-wide superscalar implemented

### Accuracy Gap Analysis
Current error rates suggest:
- **Arithmetic (49.3% error):** M2 likely has more parallelism than modeled
- **Branch (51.3% error):** Branch predictor training may not match M2 behavior
- **Dependency (18.9% error):** Forwarding paths reasonably accurate

## Recommendations for Accuracy Improvement

1. **Verify branch predictor training**
   - Add debug logging to confirm predictor updates on outcomes
   - Check initial state for cold branches

2. **Consider 6-wide issue**
   - Current 4-wide may explain arithmetic throughput gap
   - M2 Avalanche cores are 6-wide decode/issue

3. **Review ALU resources**
   - M2 may have more functional units than modeled
   - Check for bottlenecks in execute stage

## References
- https://reflexive.space/apple-m2-bp/ (Branch prediction research)
- https://www.mdpi.com/2079-9292/14/23/4686 (BTB organization paper)
- https://semianalysis.com/2022/06/10/apple-m2-die-shot-and-architecture/

## Branch Handling Deep Dive (Added Cycle 214)

### Zero-Cycle Branches

Modern high-performance processors often optimize unconditional branches:

1. **Direct Unconditional Branches (b label)**
   - Target is known at decode time
   - Can be resolved without waiting for ALU
   - Some processors "fold" these with zero penalty

2. **Conditional Branches with Predicted Targets**
   - BTB provides predicted target speculatively
   - Fetch continues at predicted target immediately
   - Misprediction penalty varies by pipeline depth

3. **M2 Speculation Depth**
   - M2 likely has deep speculation capability
   - Misprediction penalty may be 10-15 cycles (estimated)
   - Our simulator flush penalty may differ

### BTB Configuration Analysis

From MDPI Electronics 2025 paper observations:

1. **BTB Organization**
   - Multi-level BTB hierarchy likely (small/fast + large/slower)
   - First-level BTB: ~256-512 entries (speculation)
   - Larger BTB: 2K-4K entries

2. **BTB Miss Handling**
   - Cold branches (not in BTB): larger fetch penalty
   - This may explain our 51.3% branch error
   - First-time branch execution has extra latency

3. **Return Stack Buffer (RSB)**
   - Specialized predictor for function returns
   - M2 likely has 32-64 entry RSB
   - Our `call`/`ret` handling may differ

### Recommendations for Bob (#199)

Based on this research:

1. **Change default prediction to "not-taken"** (matches M2)
2. **Measure BTB miss rate** — first encounter of branch needs extra cycles
3. **Consider zero-penalty for unconditional branches** — if target is in fetch buffer
4. **Check fetch unit modeling** — M2 may handle sequential branches more efficiently

