/*
 * Copyright (c) 2026 Luiz Carlos Martins Scheid
 * MBA em Engenharia de Software - USP/Esalq
 *
 * Project: Algoritmo de Reconciliação de Bifurcações em Blockchain para Sistemas de Comando Tático
 *
 * cmd/experiment drives the factorial sweep and the complexity micro-benchmark
 * used to produce thesis/experiments/results.csv and complexity.csv. The CSVs
 * committed in the thesis tree are the data-of-record; this binary exists so
 * the experiment remains reproducible, not to regenerate numbers casually.
 */

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/lcmscheid/tactical-blockchain-reconciliation/internal/experiment"
	"github.com/lcmscheid/tactical-blockchain-reconciliation/internal/reconciler"
)

func main() {
	var (
		outResults    = flag.String("out-results", "results.csv", "path for factorial CSV output")
		outComplexity = flag.String("out-complexity", "complexity.csv", "path for complexity micro-benchmark CSV output")
		mode          = flag.String("mode", "all", "what to run: all | factorial | complexity")
		replicates    = flag.Int("replicates", 10, "replicates per factorial configuration")
		quiet         = flag.Bool("quiet", false, "suppress progress output")
	)
	flag.Parse()

	switch *mode {
	case "all", "factorial", "complexity":
	default:
		fmt.Fprintf(os.Stderr, "unknown mode %q\n", *mode)
		os.Exit(2)
	}

	if *mode == "all" || *mode == "factorial" {
		rows := runFactorial(*replicates, *quiet)
		if err := experiment.WriteResults(*outResults, rows); err != nil {
			log.Fatalf("write results: %v", err)
		}
		if !*quiet {
			fmt.Printf("Wrote %d factorial rows to %s\n", len(rows), *outResults)
		}
	}

	if *mode == "all" || *mode == "complexity" {
		rows := experiment.RunComplexitySweep(experiment.DefaultComplexityConfig())
		if err := experiment.WriteComplexity(*outComplexity, rows); err != nil {
			log.Fatalf("write complexity: %v", err)
		}
		if !*quiet {
			fmt.Printf("Wrote %d complexity rows to %s\n", len(rows), *outComplexity)
		}
	}
}

func runFactorial(replicates int, quiet bool) []experiment.TrialResult {
	strategies := []reconciler.ReconciliationStrategy{
		reconciler.StrategyPriorityAuthority{},
		reconciler.StrategyTemporalAuthority{},
		reconciler.NewStrategyHybrid(),
	}
	forks := []int{2, 3, 4}
	durations := []int{30, 120}
	conflicts := []float64{0.15, 0.55}
	depDensities := []float64{0.0, 0.3}

	cap := len(strategies) * len(forks) * len(durations) * len(conflicts) * len(depDensities) * replicates
	rows := make([]experiment.TrialResult, 0, cap)
	total := 0
	for _, s := range strategies {
		for _, nf := range forks {
			for _, d := range durations {
				for _, c := range conflicts {
					for _, dd := range depDensities {
						for r := 1; r <= replicates; r++ {
							total++
							if !quiet && total%120 == 0 {
								fmt.Printf("  trial %d …\n", total)
							}
							rows = append(rows, experiment.RunTrial(experiment.TrialConfig{
								Strategy:           s,
								NForks:             nf,
								PartitionDurationS: d,
								ConflictLevel:      c,
								DepDensity:         dd,
								Replicate:          r,
							}))
						}
					}
				}
			}
		}
	}
	return rows
}
