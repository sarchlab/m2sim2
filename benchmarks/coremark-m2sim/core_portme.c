/*
 * M2Sim CoreMark Port
 * Bare-metal implementation for M2Sim cycle-accurate simulation
 */
#include "coremark.h"
#include "core_portme.h"

#if VALIDATION_RUN
volatile ee_s32 seed1_volatile = 0x3415;
volatile ee_s32 seed2_volatile = 0x3415;
volatile ee_s32 seed3_volatile = 0x66;
#endif
#if PERFORMANCE_RUN
volatile ee_s32 seed1_volatile = 0x0;
volatile ee_s32 seed2_volatile = 0x0;
volatile ee_s32 seed3_volatile = 0x66;
#endif
#if PROFILE_RUN
volatile ee_s32 seed1_volatile = 0x8;
volatile ee_s32 seed2_volatile = 0x8;
volatile ee_s32 seed3_volatile = 0x8;
#endif

volatile ee_s32 seed4_volatile = ITERATIONS;
volatile ee_s32 seed5_volatile = 0;

ee_u32 default_num_contexts = 1;

/* Simple instruction counter - M2Sim will measure actual cycles */
static ee_u32 tick_count = 0;
static CORETIMETYPE start_time_val, stop_time_val;

/* Read ARM64 cycle counter - for bare-metal without counter access, use simple tick */
static inline ee_u32 read_cycle_counter(void) {
    return ++tick_count;
}

#define GETMYTIME(_t)              (*_t = read_cycle_counter())
#define MYTIMEDIFF(fin, ini)       ((fin) - (ini))
#define TIMER_RES_DIVIDER          1
#define EE_TICKS_PER_SEC           1000000  /* Arbitrary - M2Sim measures real cycles */

void start_time(void) {
    GETMYTIME(&start_time_val);
}

void stop_time(void) {
    GETMYTIME(&stop_time_val);
}

CORE_TICKS get_time(void) {
    CORE_TICKS elapsed = (CORE_TICKS)(MYTIMEDIFF(stop_time_val, start_time_val));
    return elapsed;
}

/* secs_ret is ee_u32 when HAS_FLOAT=0 */
secs_ret time_in_secs(CORE_TICKS ticks) {
    secs_ret retval = ticks / EE_TICKS_PER_SEC;
    return retval;
}

/* Memory region for CoreMark heap */
static ee_u8 static_memblk[TOTAL_DATA_SIZE];

void *portable_malloc(ee_size_t size) {
    /* Simple bump allocator - sufficient for single run */
    static ee_size_t allocated = 0;
    void *ret = NULL;
    
    if (allocated + size <= TOTAL_DATA_SIZE) {
        ret = static_memblk + allocated;
        allocated += size;
    }
    return ret;
}

void portable_free(void *p) {
    /* No-op for static allocation */
    (void)p;
}

void portable_init(core_portable *p, int *argc, char *argv[]) {
    (void)argc;
    (void)argv;
    p->portable_id = 1;
}

void portable_fini(core_portable *p) {
    p->portable_id = 0;
}

/* Minimal ee_printf for bare-metal - outputs to memory location M2Sim can read */
/* For now, stub it - M2Sim validation uses instruction count, not output */
int ee_printf(const char *fmt, ...) {
    (void)fmt;
    return 0;
}
