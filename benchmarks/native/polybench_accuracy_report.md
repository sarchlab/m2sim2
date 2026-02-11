# H5 PolyBench Accuracy Report

## Summary

- **Average Error:** 470943.4%
- **Benchmarks Analyzed:** 7
- **H5 Target:** <20% average error
- **Status:** NOT ACHIEVED

## Individual Benchmark Results

| Benchmark | Description | HW CPI | Sim CPI | HW ns/inst | Sim ns/inst | Error % |
|-----------|-------------|--------|---------|------------|-------------|---------|
| gemm | GEMM: General matrix-matrix mu... | 0.0 | 1.2 | 956.6 | 0.343 | 278917.7% |
| atax | ATAX: Matrix transpose and vec... | 0.0 | 4.0 | 7632.4 | 1.143 | 667733.7% |
| 2mm | 2MM: 2 matrix multiplications ... | 0.0 | 1.5 | 608.4 | 0.429 | 141864.1% |
| mvt | MVT: Matrix vector product and... | 0.0 | 4.1 | 7705.9 | 1.171 | 657724.4% |
| jacobi-1d | Jacobi-1D: 1D Jacobi stencil c... | 0.0 | 3.8 | 7620.2 | 1.086 | 701763.9% |
| 3mm | 3MM: 3 matrix multiplications ... | 0.0 | 1.8 | 406.8 | 0.514 | 79000.9% |
| bicg | BiCG: Bi-conjugate gradient su... | 0.0 | 4.2 | 9236.4 | 1.200 | 769598.9% |

## Analysis

This report analyzes 7 PolyBench intermediate complexity benchmarks against M2 hardware baselines.
The average prediction error is 470943.4%, which does not meet the H5 milestone target of <20%.

## Next Steps

- ✅ PolyBench hardware baselines collected and validated
- ✅ Accuracy framework extended to support intermediate benchmarks
- ❌ H5 milestone accuracy target requires improvement

**H5 Status:** REQUIRES CALIBRATION IMPROVEMENTS