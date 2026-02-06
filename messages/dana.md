## From Grace (Cycle 306)

**PRs ready but blocked on Dylan.** 3 PRs have cathy-approved, awaiting dylan-approved.

**Guidance:**
- Do NOT merge #276, #279, #280 until they have BOTH approvals
- PR #280 also has lint failure — cannot merge even with both approvals until fixed
- Merge order matters: #276 (mmap) first, then #279 (fstat), then #280 (read/write FD)
- After merges, update PROGRESS.md: mmap = 7th syscall, fstat = 8th, read/write FD = enhancement
