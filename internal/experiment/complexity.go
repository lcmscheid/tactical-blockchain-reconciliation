/*
 * Copyright (c) 2026 Luiz Carlos Martins Scheid
 * MBA em Engenharia de Software - USP/Esalq
 *
 * Project: Algoritmo de Reconciliação de Bifurcações em Blockchain para Sistemas de Comando Tático
 */

package experiment

import (
	"fmt"
	"runtime"
	"time"

	"github.com/lcmscheid/tactical-blockchain-reconciliation/internal/core"
	"github.com/lcmscheid/tactical-blockchain-reconciliation/internal/reconciler"
)

// ComplexityConfig tunes the micro-benchmark. The defaults match the 5×30 grid
// recorded in thesis/experiments/complexity.csv.
type ComplexityConfig struct {
	// Ns is the list of n values swept. For each n the harness measures the
	// MergeChains latency across Replicates runs after a warm-up pass.
	Ns []int
	// Replicates per n value.
	Replicates int
	// Strategy controls which scoring function the benchmark uses. The thesis
	// reports the Priority-Authority baseline, so that is the default.
	Strategy reconciler.ReconciliationStrategy
}

// DefaultComplexityConfig returns the config used to produce the committed
// complexity.csv: n ∈ {10, 100, 1000, 5000, 10000}, 30 replicates per point.
func DefaultComplexityConfig() ComplexityConfig {
	return ComplexityConfig{
		Ns:         []int{10, 100, 1000, 5000, 10000},
		Replicates: 30,
		Strategy:   reconciler.StrategyPriorityAuthority{},
	}
}

// RunComplexitySweep executes the micro-benchmark and returns one row per
// (n, replicate) pair, ready to be written with WriteComplexity.
func RunComplexitySweep(cfg ComplexityConfig) []ComplexityRow {
	if cfg.Strategy == nil {
		cfg.Strategy = reconciler.StrategyPriorityAuthority{}
	}
	rows := make([]ComplexityRow, 0, len(cfg.Ns)*cfg.Replicates)
	engine := reconciler.NewReconciler(cfg.Strategy)

	for _, n := range cfg.Ns {
		// Pre-build the two branches once per n. They are reused across
		// replicates so the timing isolates MergeChains itself, not the
		// construction of the inputs.
		chainA, chainB := buildConflictingChains(n)

		// A single warm-up pass primes caches and the GC.
		_, _ = engine.MergeChains(chainA, chainB)
		runtime.GC()

		for r := 1; r <= cfg.Replicates; r++ {
			start := time.Now()
			_, _ = engine.MergeChains(chainA, chainB)
			elapsed := time.Since(start)

			rows = append(rows, ComplexityRow{
				N:         n,
				Replicate: r,
				TimeNs:    elapsed.Nanoseconds(),
				TimeUs:    float64(elapsed.Nanoseconds()) / 1e3,
			})
		}
	}
	return rows
}

// buildConflictingChains produces two branches holding `n` conflicting
// transactions each. Every pair of transactions shares the same ID (so the
// reconciler deduplicates into exactly n unique events) while differing in
// authority and priority so the comparator exercises every key.
func buildConflictingChains(n int) ([]*core.Block, []*core.Block) {
	txsA := make([]core.Transaction, n)
	txsB := make([]core.Transaction, n)
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("tx-%06d", i)
		txsA[i] = core.Transaction{
			ID:        id,
			Timestamp: int64(10_000 + i),
			SignerID:  "NODE-A",
			Authority: []core.AuthorityLevel{core.AuthJFC, core.AuthPDU, core.AuthDU}[i%3],
			Priority:  []core.MissionPriority{core.PriorityCritical, core.PriorityHigh, core.PriorityMedium, core.PriorityLow}[i%4],
		}
		txsB[i] = core.Transaction{
			ID:        id,
			Timestamp: int64(20_000 + i),
			SignerID:  "NODE-B",
			Authority: []core.AuthorityLevel{core.AuthDU, core.AuthPDU, core.AuthJFC}[i%3],
			Priority:  []core.MissionPriority{core.PriorityLow, core.PriorityMedium, core.PriorityHigh, core.PriorityCritical}[i%4],
		}
	}
	blockA := &core.Block{Index: 1, Timestamp: 1, Transactions: txsA}
	blockB := &core.Block{Index: 1, Timestamp: 2, Transactions: txsB}
	return []*core.Block{blockA}, []*core.Block{blockB}
}
