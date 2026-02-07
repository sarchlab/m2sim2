// Package benchmarks provides SPEC CPU 2017 benchmark runner infrastructure.
package benchmarks

import (
	"fmt"
	"os"
	"path/filepath"
)

// SPECBenchmark represents a SPEC CPU 2017 benchmark configuration.
type SPECBenchmark struct {
	// Name is the SPEC benchmark name (e.g., "557.xz_r")
	Name string

	// Binary is the relative path to the ARM64 executable within SPEC
	Binary string

	// TestArgs are the command-line arguments for the 'test' input size
	TestArgs []string

	// TestInputFiles are required input files for the 'test' workload
	TestInputFiles []string

	// WorkingDir is the relative directory to run from (within SPEC)
	WorkingDir string

	// Description explains what the benchmark does
	Description string
}

// GetSPECBenchmarks returns the list of supported SPEC CPU 2017 benchmarks.
// These are integer rate benchmarks suitable for M2Sim validation.
func GetSPECBenchmarks() []SPECBenchmark {
	return []SPECBenchmark{
		{
			Name:        "557.xz_r",
			Binary:      "benchspec/CPU/557.xz_r/exe/xz_r_base.arm64",
			Description: "Data compression using LZMA2 algorithm",
			TestArgs: []string{
				"cpu2006docs.tar.xz", "4", "055ce243071129412e9dd0b3b69a21654033a9b723d874b2015c774fac1553d9713be561ca86f74e4f16f22e664fc17a79f30caa5ad2c04fbc447549c2e96f63",
				"1548636", "1555348", "0",
			},
			TestInputFiles: []string{"cpu2006docs.tar.xz"},
			WorkingDir:     "benchspec/CPU/557.xz_r/run/run_base_test_arm64.0000",
		},
		{
			Name:           "505.mcf_r",
			Binary:         "benchspec/CPU/505.mcf_r/exe/mcf_r_base.arm64",
			Description:    "Vehicle scheduling using network simplex algorithm",
			TestArgs:       []string{"inp.in"},
			TestInputFiles: []string{"inp.in"},
			WorkingDir:     "benchspec/CPU/505.mcf_r/run/run_base_test_arm64.0000",
		},
		{
			Name:           "531.deepsjeng_r",
			Binary:         "benchspec/CPU/531.deepsjeng_r/exe/deepsjeng_r_base.arm64",
			Description:    "Chess engine with alpha-beta search",
			TestArgs:       []string{"test.txt"},
			TestInputFiles: []string{"test.txt"},
			WorkingDir:     "benchspec/CPU/531.deepsjeng_r/run/run_base_test_arm64.0000",
		},
		{
			Name:           "548.exchange2_r",
			Binary:         "benchspec/CPU/548.exchange2_r/exe/exchange2_r_base.arm64",
			Description:    "Sudoku puzzle solver using Fortran 95 array operations",
			TestArgs:       []string{"0"},
			TestInputFiles: []string{"puzzles.txt"},
			WorkingDir:     "benchspec/CPU/548.exchange2_r/run/run_base_test_arm64.0000",
		},
	}
}

// SPECRunner manages SPEC benchmark execution.
type SPECRunner struct {
	// SPECRoot is the path to the SPEC installation
	SPECRoot string
}

// NewSPECRunner creates a new SPEC benchmark runner.
// It looks for SPEC at benchmarks/spec (symlink) relative to the repository root.
func NewSPECRunner(repoRoot string) (*SPECRunner, error) {
	specPath := filepath.Join(repoRoot, "benchmarks", "spec")

	// Check if SPEC directory exists
	info, err := os.Stat(specPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("SPEC not found at %s - see docs/spec-integration.md for setup", specPath)
	}
	if err != nil {
		return nil, fmt.Errorf("error checking SPEC path: %w", err)
	}
	if !info.IsDir() {
		// Follow symlink
		realPath, err := filepath.EvalSymlinks(specPath)
		if err != nil {
			return nil, fmt.Errorf("error resolving SPEC symlink: %w", err)
		}
		info, err = os.Stat(realPath)
		if err != nil || !info.IsDir() {
			return nil, fmt.Errorf("SPEC symlink target is not a directory: %s", realPath)
		}
	}

	return &SPECRunner{SPECRoot: specPath}, nil
}

// GetBinaryPath returns the full path to a benchmark's ARM64 binary.
func (r *SPECRunner) GetBinaryPath(bench SPECBenchmark) string {
	return filepath.Join(r.SPECRoot, bench.Binary)
}

// BinaryExists checks if the ARM64 binary for a benchmark has been built.
func (r *SPECRunner) BinaryExists(bench SPECBenchmark) bool {
	path := r.GetBinaryPath(bench)
	_, err := os.Stat(path)
	return err == nil
}

// GetWorkingDir returns the full path to a benchmark's run directory.
func (r *SPECRunner) GetWorkingDir(bench SPECBenchmark) string {
	return filepath.Join(r.SPECRoot, bench.WorkingDir)
}

// ValidateSetup checks if SPEC is properly set up for running benchmarks.
func (r *SPECRunner) ValidateSetup() error {
	// Check for required SPEC structure
	required := []string{
		"shrc",
		"benchspec/CPU",
	}

	for _, path := range required {
		fullPath := filepath.Join(r.SPECRoot, path)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return fmt.Errorf("SPEC structure incomplete: missing %s", path)
		}
	}

	return nil
}

// ListAvailableBenchmarks returns benchmarks that have ARM64 binaries built.
func (r *SPECRunner) ListAvailableBenchmarks() []SPECBenchmark {
	var available []SPECBenchmark
	for _, bench := range GetSPECBenchmarks() {
		if r.BinaryExists(bench) {
			available = append(available, bench)
		}
	}
	return available
}

// ListMissingBenchmarks returns benchmarks that need ARM64 binaries built.
func (r *SPECRunner) ListMissingBenchmarks() []SPECBenchmark {
	var missing []SPECBenchmark
	for _, bench := range GetSPECBenchmarks() {
		if !r.BinaryExists(bench) {
			missing = append(missing, bench)
		}
	}
	return missing
}
