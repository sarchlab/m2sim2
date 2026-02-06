/**
 * polybench.h - Minimal bare-metal PolyBench header for M2Sim
 *
 * This is a stripped-down version of polybench.h that removes all
 * libc dependencies for bare-metal execution on M2Sim.
 */

#ifndef _POLYBENCH_H
#define _POLYBENCH_H

/* Dataset sizes - using MINI for fast validation */
#ifndef MINI_DATASET
#define MINI_DATASET
#endif

/* GEMM matrix dimensions (MINI) */
#ifdef MINI_DATASET
  #define NI 16
  #define NJ 16
  #define NK 16
#endif

/* ATAX matrix dimensions (MINI) */
#ifdef MINI_DATASET
  #define NX 16
  #define NY 16
#endif

/* Integer data type for bare-metal (no FPU dependencies) */
typedef int DATA_TYPE;

/* Timing macros - stubs for bare-metal */
#define polybench_start_instruments
#define polybench_stop_instruments
#define polybench_print_instruments

/* Prevent array scalarization - use restrict qualifier */
#ifdef POLYBENCH_USE_RESTRICT
  #define POLYBENCH_RESTRICT __restrict
#else
  #define POLYBENCH_RESTRICT
#endif

/* Declare 2D array type */
#define POLYBENCH_2D_ARRAY_DECL(var, type, d1, d2, name) \
    type var[d1][d2]

/* Static allocation helpers */
#define POLYBENCH_ALLOC_2D_ARRAY(n1, n2, type) /* noop - static arrays */
#define POLYBENCH_FREE_ARRAY(x) /* noop - static arrays */

#endif /* _POLYBENCH_H */
