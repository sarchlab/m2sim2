# Implementation Guide: jacobi-1d and 3mm Benchmarks

**Author:** Eric (AI Researcher)  
**Created:** 2026-02-06 (Cycle 273)  
**Purpose:** Detailed implementation guide for next two PolyBench benchmarks

## Overview

This guide provides step-by-step implementation details for the next two PolyBench kernels needed to reach our 15+ benchmark target:

| Benchmark | Type | Effort | New Total |
|-----------|------|--------|-----------|
| jacobi-1d | Stencil | Low | 12 |
| 3mm | Matrix chain | Medium | 13 |

Both benchmarks follow the established PolyBench/M2Sim pattern used by gemm, atax, 2mm, and mvt.

---

## jacobi-1d Implementation

### Algorithm Overview

Jacobi 1D iterative stencil computation - smooths an array by averaging neighboring values.

```c
// Pseudocode
for (t = 0; t < TSTEPS; t++) {
    for (i = 1; i < N - 1; i++)
        B[i] = 0.33333 * (A[i-1] + A[i] + A[i+1]);
    for (i = 1; i < N - 1; i++)
        A[i] = B[i];
}
```

### Integer Adaptation

Since M2Sim doesn't support FP, use integer averaging:

```c
// Integer version - divide by 3
B[i] = (A[i-1] + A[i] + A[i+1]) / 3;
```

### Required Constants (polybench.h additions)

```c
/* JACOBI-1D dimensions (MINI) */
#ifdef MINI_DATASET
  #define TSTEPS 8    /* Number of iterations */
  #define N_SIZE 32   /* Array size (avoid conflict with NK) */
#endif
```

### Full Implementation: jacobi-1d.c

```c
/**
 * jacobi-1d.c - 1D Jacobi Stencil for M2Sim
 *
 * Computes iterative stencil smoothing:
 *   B[i] = (A[i-1] + A[i] + A[i+1]) / 3
 *
 * This is a bare-metal adaptation of the PolyBench jacobi-1d kernel,
 * using integer arithmetic for M2Sim validation.
 *
 * Original: PolyBench/C 4.2.1 (stencils/jacobi-1d)
 */

#include "../common/polybench.h"

/* Array dimensions - MINI dataset */
#ifndef TSTEPS
#define TSTEPS 8
#endif
#ifndef N_SIZE
#define N_SIZE 32
#endif

/* Static arrays */
static DATA_TYPE A[N_SIZE];
static DATA_TYPE B[N_SIZE];

/**
 * Initialize arrays with deterministic values
 */
static void init_array(void) {
    int i;
    for (i = 0; i < N_SIZE; i++) {
        A[i] = (i * 3) % 256;
        B[i] = (i * 2) % 256;
    }
}

/**
 * Jacobi 1D kernel
 * Iterative stencil: B[i] = (A[i-1] + A[i] + A[i+1]) / 3
 */
static void kernel_jacobi_1d(void) {
    int t, i;
    
    polybench_start_instruments;
    
    for (t = 0; t < TSTEPS; t++) {
        /* Compute B from A */
        for (i = 1; i < N_SIZE - 1; i++) {
            B[i] = (A[i-1] + A[i] + A[i+1]) / 3;
        }
        /* Copy B back to A */
        for (i = 1; i < N_SIZE - 1; i++) {
            A[i] = B[i];
        }
    }
    
    polybench_stop_instruments;
}

/**
 * Compute checksum
 */
static int compute_checksum(void) {
    int i;
    int sum = 0;
    
    for (i = 0; i < N_SIZE; i++) {
        sum += A[i];
    }
    
    return sum & 0xFF;
}

int main(void) {
    init_array();
    kernel_jacobi_1d();
    return compute_checksum();
}
```

### Directory Structure

```
benchmarks/polybench/jacobi-1d/
└── jacobi-1d.c
```

### Build Script Update

Add to `BENCHMARKS` in build.sh:
```bash
BENCHMARKS="gemm atax 2mm mvt jacobi-1d"
```

### Expected Characteristics

| Metric | Estimate |
|--------|----------|
| Instructions | ~3-5K (MINI dataset) |
| Memory access pattern | Streaming/strided |
| Loop iterations | TSTEPS × (N_SIZE - 2) |

---

## 3mm Implementation

### Algorithm Overview

Three matrix multiplications chained together:

```c
// E := A x B  (NI x NK) × (NK x NJ) = (NI x NJ)
// F := C x D  (NJ x NL) × (NL x NM) = (NJ x NM)
// G := E x F  (NI x NJ) × (NJ x NM) = (NI x NM)
```

### Required Constants (polybench.h additions)

```c
/* 3MM matrix dimensions (MINI) */
#ifdef MINI_DATASET
  #define NI_3MM 16
  #define NJ_3MM 16
  #define NK_3MM 16
  #define NL_3MM 16
  #define NM_3MM 16
#endif
```

### Full Implementation: 3mm.c

```c
/**
 * 3mm.c - Three Matrix Multiplications for M2Sim
 *
 * Computes:
 *   E := A x B
 *   F := C x D
 *   G := E x F
 *
 * This is a bare-metal adaptation of the PolyBench 3mm kernel,
 * using integer arithmetic for M2Sim validation.
 *
 * Original: PolyBench/C 4.2.1 (linear-algebra/kernels/3mm)
 */

#include "../common/polybench.h"

/* Matrix dimensions - MINI dataset */
#ifndef NI_3MM
#define NI_3MM 16
#endif
#ifndef NJ_3MM
#define NJ_3MM 16
#endif
#ifndef NK_3MM
#define NK_3MM 16
#endif
#ifndef NL_3MM
#define NL_3MM 16
#endif
#ifndef NM_3MM
#define NM_3MM 16
#endif

/* Static arrays */
static DATA_TYPE A[NI_3MM][NK_3MM];
static DATA_TYPE B[NK_3MM][NJ_3MM];
static DATA_TYPE C[NJ_3MM][NL_3MM];
static DATA_TYPE D[NL_3MM][NM_3MM];
static DATA_TYPE E[NI_3MM][NJ_3MM];  /* E := A x B */
static DATA_TYPE F[NJ_3MM][NM_3MM];  /* F := C x D */
static DATA_TYPE G[NI_3MM][NM_3MM];  /* G := E x F */

/**
 * Initialize arrays with deterministic values
 */
static void init_array(void) {
    int i, j;
    
    for (i = 0; i < NI_3MM; i++) {
        for (j = 0; j < NK_3MM; j++) {
            A[i][j] = (i * NK_3MM + j) % 256;
        }
    }
    
    for (i = 0; i < NK_3MM; i++) {
        for (j = 0; j < NJ_3MM; j++) {
            B[i][j] = (i * NJ_3MM + j + 1) % 256;
        }
    }
    
    for (i = 0; i < NJ_3MM; i++) {
        for (j = 0; j < NL_3MM; j++) {
            C[i][j] = (i * NL_3MM + j + 2) % 256;
        }
    }
    
    for (i = 0; i < NL_3MM; i++) {
        for (j = 0; j < NM_3MM; j++) {
            D[i][j] = (i * NM_3MM + j + 3) % 256;
        }
    }
    
    /* Zero output matrices */
    for (i = 0; i < NI_3MM; i++) {
        for (j = 0; j < NJ_3MM; j++) {
            E[i][j] = 0;
        }
    }
    
    for (i = 0; i < NJ_3MM; i++) {
        for (j = 0; j < NM_3MM; j++) {
            F[i][j] = 0;
        }
    }
    
    for (i = 0; i < NI_3MM; i++) {
        for (j = 0; j < NM_3MM; j++) {
            G[i][j] = 0;
        }
    }
}

/**
 * 3MM kernel: Three matrix multiplications
 * E := A x B
 * F := C x D
 * G := E x F
 */
static void kernel_3mm(void) {
    int i, j, k;
    
    polybench_start_instruments;
    
    /* E := A x B */
    for (i = 0; i < NI_3MM; i++) {
        for (k = 0; k < NK_3MM; k++) {
            for (j = 0; j < NJ_3MM; j++) {
                E[i][j] += A[i][k] * B[k][j];
            }
        }
    }
    
    /* F := C x D */
    for (i = 0; i < NJ_3MM; i++) {
        for (k = 0; k < NL_3MM; k++) {
            for (j = 0; j < NM_3MM; j++) {
                F[i][j] += C[i][k] * D[k][j];
            }
        }
    }
    
    /* G := E x F */
    for (i = 0; i < NI_3MM; i++) {
        for (k = 0; k < NJ_3MM; k++) {
            for (j = 0; j < NM_3MM; j++) {
                G[i][j] += E[i][k] * F[k][j];
            }
        }
    }
    
    polybench_stop_instruments;
}

/**
 * Compute checksum of result matrix G
 */
static int compute_checksum(void) {
    int i, j;
    int sum = 0;
    
    for (i = 0; i < NI_3MM; i++) {
        for (j = 0; j < NM_3MM; j++) {
            sum += G[i][j];
        }
    }
    
    return sum & 0xFF;
}

int main(void) {
    init_array();
    kernel_3mm();
    return compute_checksum();
}
```

### Directory Structure

```
benchmarks/polybench/3mm/
└── 3mm.c
```

### Build Script Update

Add to `BENCHMARKS` in build.sh:
```bash
BENCHMARKS="gemm atax 2mm mvt jacobi-1d 3mm"
```

### Expected Characteristics

| Metric | Estimate |
|--------|----------|
| Instructions | ~90-120K (3× gemm-like loops) |
| Memory footprint | 7 matrices × 256 × 4 bytes = ~7KB |
| Complexity | O(n³) per multiplication × 3 |

---

## Implementation Checklist for Bob

### jacobi-1d (Lower effort)

1. [ ] Create `benchmarks/polybench/jacobi-1d/` directory
2. [ ] Create `jacobi-1d.c` per template above
3. [ ] Update `BENCHMARKS` in `build.sh`
4. [ ] Run `./build.sh jacobi-1d`
5. [ ] Test with emulator: `go run ./cmd/m2sim -elf benchmarks/polybench/jacobi-1d_m2sim.elf`
6. [ ] Verify exit code and instruction count
7. [ ] Create PR

### 3mm (Medium effort)

1. [ ] Create `benchmarks/polybench/3mm/` directory
2. [ ] Create `3mm.c` per template above
3. [ ] Update `BENCHMARKS` in `build.sh`
4. [ ] Run `./build.sh 3mm`
5. [ ] Test with emulator
6. [ ] Verify exit code and instruction count
7. [ ] Create PR

---

## Workload Diversity

After implementing these benchmarks:

| Category | Benchmarks | Count |
|----------|------------|-------|
| Matrix/Linear Algebra | gemm, atax, 2mm, mvt, 3mm | 5 |
| Stencil | **jacobi-1d** | 1 |
| Integer/Crypto | aha-mont64, crc32 | 2 |
| Signal Processing | edn | 1 |
| Control/State | primecount, statemate | 2 |
| Compression | huffbench | 1 |

**Total after these: 13 benchmarks** (need 2 more for 15)

---

## Notes

- Both kernels use i-k-j loop order for better cache behavior
- Integer arithmetic avoids FP instruction requirements
- MINI dataset keeps instruction counts reasonable for emulation
- Checksum validates correct execution without stdout

*This document supports the Path to 15+ Benchmarks roadmap.*
