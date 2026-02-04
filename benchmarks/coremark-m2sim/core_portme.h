/*
 * M2Sim CoreMark Port
 * Cross-compiled for aarch64-elf bare-metal execution in M2Sim
 */
#ifndef CORE_PORTME_H
#define CORE_PORTME_H

#include <stddef.h>  /* For size_t, NULL */
#include <stdint.h>  /* For fixed-width types */
#include <stdarg.h>  /* For va_list */

/* No floating point needed for CoreMark core */
#define HAS_FLOAT 0

/* No time.h in bare-metal */
#define HAS_TIME_H 0
#define USE_CLOCK 0

/* No stdio in bare-metal - use ee_printf */
#define HAS_STDIO 0

/* Data type definitions for 64-bit ARM */
typedef int16_t   ee_s16;
typedef uint16_t  ee_u16;
typedef int32_t   ee_s32;
typedef double    ee_f32;
typedef uint8_t   ee_u8;
typedef uint32_t  ee_u32;
typedef uintptr_t ee_ptr_int;
typedef size_t    ee_size_t;

/* Seed method and memory configuration */
#define SEED_METHOD    SEED_VOLATILE
#define MEM_METHOD     MEM_STATIC
#define MULTITHREAD    1
#define USE_PTHREAD    0

/* CoreMark configuration */
#define MAIN_HAS_NOARGC 1
#define MAIN_HAS_NORETURN 0

/* Time type - use simple counter */
typedef ee_u32 CORE_TICKS;
typedef ee_u32 CORETIMETYPE;

/* secs_ret is defined in coremark.h based on HAS_FLOAT */

/* Iterations - small for simulator validation */
#ifndef ITERATIONS
#define ITERATIONS 10
#endif

/* Memory - must be defined before coremark.h includes it */
#ifndef TOTAL_DATA_SIZE
#define TOTAL_DATA_SIZE (2000 + 4 * 1000)
#endif

/* Compiler info for reporting */
#define COMPILER_VERSION "aarch64-elf-gcc 15.2.0"
#define COMPILER_FLAGS   "-O2 -static -nostdlib -ffreestanding"
#define MEM_LOCATION     "STATIC"

/* Memory alignment helper - align to 8-byte boundary for ARM64 */
#define align_mem(x) (void *)(8 + (((ee_ptr_int)(x)-1) & ~7))

/* Core portable structure */
typedef struct CORE_PORTABLE_S {
    ee_u8 portable_id;
} core_portable;

/* Compiler flags will be set in Makefile */
extern ee_u32 default_num_contexts;

/* Portable functions */
void portable_init(core_portable *p, int *argc, char *argv[]);
void portable_fini(core_portable *p);

/* printf replacement */
int ee_printf(const char *fmt, ...);

#endif /* CORE_PORTME_H */
