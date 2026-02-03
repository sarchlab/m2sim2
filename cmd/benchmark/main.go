// Command benchmark runs the M2Sim timing benchmark harness.
//
// Usage:
//
//	go run ./cmd/benchmark [flags]
//
// Flags:
//
//	-format     Output format: text, csv, or json (default: text)
//	-core       Run only the 3 core benchmarks (loop, matrix, branch)
//	-no-icache  Disable instruction cache simulation
//	-no-dcache  Disable data cache simulation
//	-o          Output file (default: stdout)
//	-v          Verbose output with validation checks
//
// Example:
//
//	# Run all benchmarks with human-readable output
//	go run ./cmd/benchmark
//
//	# Output JSON for automated comparison with real M2 measurements
//	go run ./cmd/benchmark -format=json > results.json
//
//	# Quick validation with 3 core benchmarks
//	go run ./cmd/benchmark -core -v
//
// The benchmark results can be compared against real M2 hardware measurements
// to calibrate the simulator's timing model. JSON output is designed for
// automated comparison with Issue #96 (M2 timing baseline).
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sarchlab/m2sim/benchmarks"
)

var (
	format     = flag.String("format", "text", "Output format: text, csv, or json")
	coreOnly   = flag.Bool("core", false, "Run only the 3 core benchmarks (loop, matrix, branch)")
	noICache   = flag.Bool("no-icache", false, "Disable instruction cache simulation")
	noDCache   = flag.Bool("no-dcache", false, "Disable data cache simulation")
	outputFile = flag.String("o", "", "Output file (default: stdout)")
	verbose    = flag.Bool("v", false, "Verbose output with validation checks")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "M2Sim Benchmark Runner - Validation Infrastructure\n\n")
		fmt.Fprintf(os.Stderr, "Usage: benchmark [options]\n\n")
		fmt.Fprintf(os.Stderr, "Runs timing benchmarks and outputs results for comparison with real M2.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  benchmark                     # Run all benchmarks, text output\n")
		fmt.Fprintf(os.Stderr, "  benchmark -format=json        # JSON output for automated comparison\n")
		fmt.Fprintf(os.Stderr, "  benchmark -format=csv -o=results.csv  # CSV to file\n")
		fmt.Fprintf(os.Stderr, "  benchmark -core -v            # Quick validation with verbose checks\n")
	}
	flag.Parse()

	// Set up output
	output := os.Stdout
	if *outputFile != "" {
		f, err := os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer func() {
			if cerr := f.Close(); cerr != nil {
				fmt.Fprintf(os.Stderr, "Error closing output file: %v\n", cerr)
			}
		}()
		output = f
	}

	// Configure harness
	config := benchmarks.DefaultConfig()
	config.EnableICache = !*noICache
	config.EnableDCache = !*noDCache
	config.Output = output
	config.Verbose = *verbose

	harness := benchmarks.NewHarness(config)

	// Select benchmarks
	var benchmarkList []benchmarks.Benchmark
	if *coreOnly {
		benchmarkList = benchmarks.GetCoreBenchmarks()
		if *verbose {
			fmt.Fprintf(os.Stderr, "Running 3 core benchmarks (loop, matrix, branch)...\n")
		}
	} else {
		benchmarkList = benchmarks.GetMicrobenchmarks()
		if *verbose {
			fmt.Fprintf(os.Stderr, "Running %d benchmarks...\n", len(benchmarkList))
		}
	}
	harness.AddBenchmarks(benchmarkList)

	// Run benchmarks
	results := harness.RunAll()

	// Output results
	switch *format {
	case "json":
		if err := harness.PrintJSON(results); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing JSON: %v\n", err)
			os.Exit(1)
		}
	case "csv":
		harness.PrintCSV(results)
	case "text":
		// Print configuration header for text output
		if *outputFile == "" {
			fmt.Println("M2Sim Timing Benchmark Harness")
			fmt.Println("==============================")
			fmt.Printf("I-Cache: %v\n", config.EnableICache)
			fmt.Printf("D-Cache: %v\n", config.EnableDCache)
			fmt.Println("")
		}
		harness.PrintResults(results)

		// Print summary for text output
		if *outputFile == "" {
			fmt.Println("=== Summary ===")
			fmt.Println("")
			fmt.Println("To compare with real M2 hardware:")
			fmt.Println("1. Run equivalent benchmarks on M2 (see Issue #96)")
			fmt.Println("2. Compare cycle counts and CPI")
			fmt.Println("3. Target: <2% error vs real M2")
			fmt.Println("")
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown format: %s (use text, csv, or json)\n", *format)
		os.Exit(1)
	}

	// Validate exit codes if verbose
	if *verbose {
		failed := 0
		for i, r := range results {
			if r.ExitCode != benchmarkList[i].ExpectedExit {
				fmt.Fprintf(os.Stderr, "WARN: %s returned %d, expected %d\n",
					r.Name, r.ExitCode, benchmarkList[i].ExpectedExit)
				failed++
			}
		}
		if failed == 0 {
			fmt.Fprintf(os.Stderr, "All benchmarks passed validation.\n")
		} else {
			fmt.Fprintf(os.Stderr, "%d benchmark(s) failed validation.\n", failed)
		}
	}
}
