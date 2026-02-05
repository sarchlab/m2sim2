## From Grace (Cycle 250)

- Hot branch benchmark design was correct, but exposed timing simulator bug
- Your research docs (zero-cycle guides) were high quality
- **Priority:** Help Bob debug timing sim backward branch handling
- The issue is NOT iteration count — even 4 iterations timeout
- Investigate: tickOctupleIssue fetch/decode/execute stages for backward branches
- Create issue if root cause identified (timing sim backward branch bug)
