package benchmarks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// StallProfileResult holds per-kernel stall profiling data.
type StallProfileResult struct {
	Name                      string  `json:"name"`
	Cycles                    uint64  `json:"cycles"`
	Instructions              uint64  `json:"instructions"`
	CPI                       float64 `json:"cpi"`
	RAWHazardStalls           uint64  `json:"raw_hazard_stalls"`
	StructuralHazardStalls    uint64  `json:"structural_hazard_stalls"`
	ExecStalls                uint64  `json:"exec_stalls"`
	MemStalls                 uint64  `json:"mem_stalls"`
	BranchMispredictionStalls uint64  `json:"branch_misprediction_stalls"`
	BranchMispredictions      uint64  `json:"branch_mispredictions"`
	PipelineFlushes           uint64  `json:"pipeline_flushes"`
	FetchStalls               uint64  `json:"fetch_stalls"`
	EliminatedBranches        uint64  `json:"eliminated_branches"`
}

// TestStallProfileOctuple runs gemm, bicg, and atax with 8-wide + caches.
func TestStallProfileOctuple(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stall profiling in short mode")
	}

	kernels := []struct {
		name    string
		elfName string
	}{
		{"gemm", "gemm"},
		{"bicg", "bicg"},
		{"atax", "atax"},
	}

	var results []StallProfileResult

	for _, k := range kernels {
		elfPath := polybenchELFPath(k.elfName)
		if _, err := os.Stat(elfPath); os.IsNotExist(err) {
			t.Skipf("ELF not found: %s", elfPath)
		}

		config := DefaultConfig() // 8-wide + caches
		config.Output = &bytes.Buffer{}
		config.MaxCycles = 5_000_000

		harness := NewHarness(config)
		harness.AddBenchmark(BenchmarkFromELF(k.name, k.name, elfPath))

		benchResults := harness.RunAll()
		r := benchResults[0]
		if r.ExitCode == -1 {
			t.Fatalf("%s: failed to load ELF", k.name)
		}

		sp := StallProfileResult{
			Name:                      k.name + " (8-wide+cache)",
			Cycles:                    r.SimulatedCycles,
			Instructions:              r.InstructionsRetired,
			CPI:                       r.CPI,
			RAWHazardStalls:           r.RAWHazardStalls,
			StructuralHazardStalls:    r.StructuralHazardStalls,
			ExecStalls:                r.ExecStalls,
			MemStalls:                 r.MemStalls,
			BranchMispredictionStalls: r.BranchMispredictionStalls,
			BranchMispredictions:      r.BranchMispredictions,
			PipelineFlushes:           r.PipelineFlushes,
			FetchStalls:               r.StallCycles,
			EliminatedBranches:        r.EliminatedBranches,
		}
		results = append(results, sp)

		t.Logf("\n=== %s (8-wide+cache) ===\n"+
			"  Cycles: %d, Instructions: %d, CPI: %.3f\n"+
			"  RAW: %d, Structural: %d, Exec: %d, Mem: %d, BrMispred: %d, Flushes: %d, Fetch: %d\n",
			k.name, sp.Cycles, sp.Instructions, sp.CPI,
			sp.RAWHazardStalls, sp.StructuralHazardStalls,
			sp.ExecStalls, sp.MemStalls,
			sp.BranchMispredictionStalls, sp.PipelineFlushes, sp.FetchStalls)
	}

	jsonData, _ := json.MarshalIndent(results, "", "  ")
	fmt.Printf("\n%s\n", string(jsonData))
}

// TestStallProfile runs gemm, bicg, and atax with stall profiling enabled.
// Run with: go test -run TestStallProfile -v ./benchmarks/ -timeout 600s
func TestStallProfile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stall profiling in short mode")
	}

	kernels := []struct {
		name    string
		elfName string
	}{
		{"gemm", "gemm"},
		{"bicg", "bicg"},
		{"atax", "atax"},
	}

	var results []StallProfileResult

	for _, k := range kernels {
		elfPath := polybenchELFPath(k.elfName)
		if _, err := os.Stat(elfPath); os.IsNotExist(err) {
			t.Skipf("ELF not found: %s (run benchmarks/polybench/build.sh)", elfPath)
		}

		config := DefaultConfig()
		config.Output = &bytes.Buffer{}
		config.EnableICache = false
		config.EnableDCache = false
		config.EnableOctupleIssue = false
		config.EnableSextupleIssue = true
		config.MaxCycles = 10_000_000

		harness := NewHarness(config)
		harness.AddBenchmark(BenchmarkFromELF(k.name, k.name, elfPath))

		benchResults := harness.RunAll()
		if len(benchResults) != 1 {
			t.Fatalf("%s: expected 1 result, got %d", k.name, len(benchResults))
		}

		r := benchResults[0]
		if r.ExitCode == -1 {
			t.Fatalf("%s: failed to load ELF", k.name)
		}
		if r.ExitCode == -2 {
			t.Logf("%s: exceeded cycle limit (partial data from %d cycles, %d instructions)",
				k.name, r.SimulatedCycles, r.InstructionsRetired)
		}

		sp := StallProfileResult{
			Name:                      k.name,
			Cycles:                    r.SimulatedCycles,
			Instructions:              r.InstructionsRetired,
			CPI:                       r.CPI,
			RAWHazardStalls:           r.RAWHazardStalls,
			StructuralHazardStalls:    r.StructuralHazardStalls,
			ExecStalls:                r.ExecStalls,
			MemStalls:                 r.MemStalls,
			BranchMispredictionStalls: r.BranchMispredictionStalls,
			BranchMispredictions:      r.BranchMispredictions,
			PipelineFlushes:           r.PipelineFlushes,
			FetchStalls:               r.StallCycles,
			EliminatedBranches:        r.EliminatedBranches,
		}
		results = append(results, sp)

		t.Logf("\n=== %s ===\n"+
			"  Cycles:                    %d\n"+
			"  Instructions:              %d\n"+
			"  CPI:                       %.3f\n"+
			"  RAW Hazard Stalls:         %d\n"+
			"  Structural Hazard Stalls:  %d\n"+
			"  Exec Stalls:               %d\n"+
			"  Mem Stalls:                %d\n"+
			"  Branch Mispred Stalls:     %d\n"+
			"  Branch Mispredictions:     %d\n"+
			"  Pipeline Flushes:          %d\n"+
			"  Fetch/Other Stalls:        %d\n"+
			"  Eliminated Branches:       %d\n",
			k.name,
			sp.Cycles, sp.Instructions, sp.CPI,
			sp.RAWHazardStalls, sp.StructuralHazardStalls,
			sp.ExecStalls, sp.MemStalls,
			sp.BranchMispredictionStalls, sp.BranchMispredictions,
			sp.PipelineFlushes, sp.FetchStalls,
			sp.EliminatedBranches)
	}

	// Write JSON results
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}
	fmt.Printf("\n%s\n", string(jsonData))
}
