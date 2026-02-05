# Branch Predictor Tuning Options

**Created:** 2026-02-05 (Cycle 231)
**Author:** Eric (Research Agent)
**Purpose:** Identify tuning opportunities to reduce 34.5% branch accuracy gap

## Current Configuration

```go
BranchPredictorConfig{
    BHTSize:             4096,
    BTBSize:             512,
    GlobalHistoryLength: 12,
    UseTournament:       true,
}
```

## Current Accuracy (Post-CMP+B.cond Fusion)

| Benchmark | Simulator CPI | M2 Real CPI | Error |
|-----------|---------------|-------------|-------|
| branch_taken_conditional | 1.600 | 1.190 | 34.5% |

## Analysis

The 34.5% error after fusion suggests the remaining gap comes from:

1. **Pipeline overhead** — Not predictor accuracy
2. **BTB miss penalty** — Cold branches pay extra cost
3. **Prediction confidence** — Counter state transitions

### Why Fusion Helped But Didn't Close the Gap

CMP+B.cond fusion eliminated the flag dependency stall, reducing CPI from 1.933 to 1.600. But M2 achieves 1.190 CPI, suggesting:

- M2 may have **zero-cycle branches** for predicted-taken conditional branches
- M2's BTB may be larger (2K-4K vs our 512 entries)
- M2 may have better branch handling in the fetch stage

## Tuning Options

### Option 1: Increase BTB Size

**Current:** 512 entries
**Proposed:** 2048 or 4096 entries

**Rationale:** Larger BTB reduces cold-start penalty for branches not yet seen.

**Expected impact:** 5-10% improvement on branch-heavy code

### Option 2: Reduce Misprediction Penalty

**Current behavior:** Misprediction causes full pipeline flush

**Proposed:** Add configurable flush penalty (currently implicit in pipeline depth)

**M2 comparison:** M2 likely has 12-15 cycle misprediction penalty

### Option 3: Zero-Cycle Predicted-Taken Branches

**Current:** All branches incur some execute stage latency

**Proposed:** When BTB hit + predicted taken, resolve in fetch stage

**Expected impact:** 10-20% improvement for tight loops with predictable branches

### Option 4: Branch Folding

**Description:** Unconditional branches entirely removed from instruction stream

**M2 behavior:** M2 likely folds `B` (unconditional) with zero cost

**Complexity:** Medium — requires fetch stage modification

## Recommendations

### Short-term (Low effort, measurable impact)

1. **Issue #219 first** — Enable 8-wide decode in benchmark harness
   - This is the critical path to seeing any improvement
   - Without this, all 8-wide optimizations are untested

2. **Increase BTB size to 2048** — Single parameter change
   - Easy to test, may help cold branches

### Medium-term (After 8-wide validation)

3. **Add branch predictor statistics to benchmark output**
   - Track BTB hit rate, misprediction rate
   - Identify if prediction quality is the issue vs pipeline overhead

4. **Implement zero-cycle branch for BTB hits**
   - Requires fetch stage modification
   - High impact for loop-heavy benchmarks

### Long-term (May require research)

5. **Match M2's actual BTB configuration**
   - Limited public documentation
   - May require reverse engineering via timing measurements

## Next Steps

1. **Bob:** Fix issue #219 (8-wide harness)
2. **Eric:** Add branch predictor stats logging (post-8-wide validation)
3. **Team:** Re-evaluate after seeing actual 8-wide accuracy numbers

## References

- `docs/m2-microarchitecture.md` — M2 branch prediction research
- PR #212 — CMP+B.cond fusion implementation
- Issue #213 — 8-wide decode (closed, but harness not updated)
