# Makefile for M2Sim Native Calibration Benchmarks
# Compiles ARM64 assembly benchmarks for macOS

# Detect architecture
ARCH := $(shell uname -m)

# Compiler settings
AS = as
LD = ld

# macOS-specific linker flags
LDFLAGS = -lSystem -L$(shell xcrun --show-sdk-path)/usr/lib -syslibroot $(shell xcrun --show-sdk-path) -e _main

# Assembly sources (short benchmarks)
SRCS = arithmetic_sequential.s \
       dependency_chain.s \
       memory_sequential.s \
       function_calls.s \
       branch_taken.s \
       mixed_operations.s

# Long-running benchmarks (10M iterations)
LONG_SRCS = arithmetic_sequential_long.s \
            dependency_chain_long.s \
            memory_sequential_long.s \
            memory_strided_long.s \
            memory_random_long.s \
            branch_taken_long.s \
            mixed_operations_long.s

# Output binaries
BINS = $(SRCS:.s=)
LONG_BINS = $(LONG_SRCS:.s=)

# Object files
OBJS = $(SRCS:.s=.o)
LONG_OBJS = $(LONG_SRCS:.s=.o)

.PHONY: all clean run check tools long

all: check_arch $(BINS) measure

# Build long-running benchmarks (1M iterations)
long: check_arch $(LONG_BINS)

tools: measure

check_arch:
ifneq ($(ARCH),arm64)
	$(error This Makefile requires ARM64/Apple Silicon. Current arch: $(ARCH))
endif

# Pattern rule: .s -> .o
%.o: %.s
	$(AS) -o $@ $<

# Pattern rule: .o -> binary
%: %.o
	$(LD) $(LDFLAGS) -o $@ $<

# Build all binaries (short and long)
$(BINS): %: %.o
$(LONG_BINS): %: %.o

# Build measurement tool
measure: measure.c
	clang -O2 -o $@ $<

clean:
	rm -f $(OBJS) $(LONG_OBJS) $(BINS) $(LONG_BINS) measure

# Run all benchmarks and verify exit codes
run: all
	@echo "Running native calibration benchmarks..."
	@echo ""
	@for bin in $(BINS); do \
		printf "%-25s: " "$$bin"; \
		./$$bin; \
		echo "exit=$$?"; \
	done

# Run long-running benchmarks with timing
run-long: long
	@echo "Running long-running benchmarks (10M iterations each)..."
	@echo "Benchmark execution time should dominate process startup overhead (~18ms)"
	@echo ""
	@for bin in $(LONG_BINS); do \
		printf "%-30s: " "$$bin"; \
		start=$$(python3 -c 'import time; print(time.time())'); \
		./$$bin; \
		exit_code=$$?; \
		end=$$(python3 -c 'import time; print(time.time())'); \
		elapsed=$$(python3 -c "print(f'{($$end - $$start)*1000:.2f}')"); \
		echo "exit=$$exit_code  time=$${elapsed}ms"; \
	done

# Verify expected exit codes match
verify: all
	@echo "Verifying benchmark exit codes..."
	@failed=0; \
	./arithmetic_sequential; test $$? -eq 4 || { echo "FAIL: arithmetic_sequential (expected 4)"; failed=1; }; \
	./dependency_chain; test $$? -eq 20 || { echo "FAIL: dependency_chain (expected 20)"; failed=1; }; \
	./memory_sequential; test $$? -eq 42 || { echo "FAIL: memory_sequential (expected 42)"; failed=1; }; \
	./function_calls; test $$? -eq 5 || { echo "FAIL: function_calls (expected 5)"; failed=1; }; \
	./branch_taken; test $$? -eq 5 || { echo "FAIL: branch_taken (expected 5)"; failed=1; }; \
	./mixed_operations; test $$? -eq 100 || { echo "FAIL: mixed_operations (expected 100)"; failed=1; }; \
	if [ $$failed -eq 0 ]; then echo "All benchmarks verified!"; else exit 1; fi

# Verify long benchmark exit codes (10M iterations)
# Exit codes: (total increments) mod 256
# - arithmetic_sequential_long: 40M increments (4 per iter * 10M), 40000000 mod 256 = 0
# - dependency_chain_long: 200M increments (20 per iter * 10M), 200000000 mod 256 = 0
# - memory_sequential_long: 50M increments (5 per iter * 10M), 50000000 mod 256 = 128
# - memory_strided_long: 50M increments (5 per iter * 10M), 50000000 mod 256 = 128
# - memory_random_long: 50M increments (5 per iter * 10M), 50000000 mod 256 = 128
# - branch_taken_long: 50M increments (5 per iter * 10M), 50000000 mod 256 = 128
# - mixed_operations_long: 10M increments (1 per iter * 10M), 10000000 mod 256 = 128
verify-long: long
	@echo "Verifying long benchmark exit codes..."
	@failed=0; \
	./arithmetic_sequential_long; test $$? -eq 0 || { echo "FAIL: arithmetic_sequential_long (expected 0)"; failed=1; }; \
	./dependency_chain_long; test $$? -eq 0 || { echo "FAIL: dependency_chain_long (expected 0)"; failed=1; }; \
	./memory_sequential_long; test $$? -eq 128 || { echo "FAIL: memory_sequential_long (expected 128)"; failed=1; }; \
	./memory_strided_long; test $$? -eq 128 || { echo "FAIL: memory_strided_long (expected 128)"; failed=1; }; \
	./memory_random_long; test $$? -eq 128 || { echo "FAIL: memory_random_long (expected 128)"; failed=1; }; \
	./branch_taken_long; test $$? -eq 128 || { echo "FAIL: branch_taken_long (expected 128)"; failed=1; }; \
	./mixed_operations_long; test $$? -eq 128 || { echo "FAIL: mixed_operations_long (expected 128)"; failed=1; }; \
	if [ $$failed -eq 0 ]; then echo "All long benchmarks verified!"; else exit 1; fi

# Time a single benchmark with /usr/bin/time
time-%: %
	@echo "Timing $<..."
	/usr/bin/time -l ./$<

# Run benchmark N times for statistical sampling
bench-%: %
	@echo "Running $< 1000 times..."
	@for i in $$(seq 1 1000); do ./$<; done
