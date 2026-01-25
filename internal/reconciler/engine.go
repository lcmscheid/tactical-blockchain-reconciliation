/*
 * Copyright (c) 2026 Luiz Carlos Martins Scheid
 * MBA em Engenharia de Software - USP/Esalq
 *
 * Project: Algoritmo de Reconciliação de Bifurcações em Blockchain para Sistemas de Comando Tático
 */

package reconciler

import (
	"fmt"
	"sort"

	"github.com/lcmscheid/tactical-blockchain-reconciliation/internal/core"
)

type Reconciler struct {
	Strategy ReconciliationStrategy
}

func NewReconciler(strategy ReconciliationStrategy) *Reconciler {
	return &Reconciler{
		Strategy: strategy,
	}
}

// MergeChains takes two divergent chains and produces a unified list of transactions representing the reconciled state.
func (r *Reconciler) MergeChains(chainA, chainB []*core.Block) ([]core.Transaction, error) {
	// Find Common Ancestor (Simplified for simulation: initialize to -1 to determine if a mismatch was actually found)
	// In production, we would traverse backwards hash-by-hash.
	divergenceIndex := 0
	minLen := len(chainA)
	if len(chainB) < minLen {
		minLen = len(chainB)
	}

	// Compare blocks up to the length of the shorter chain
	for i := 0; i < minLen; i++ {
		if chainA[i].Hash != chainB[i].Hash {
			divergenceIndex = i
			break
		}
	}

	// Handling the "No Mismatch Found" results
	if divergenceIndex == -1 {
		if len(chainA) == len(chainB) {
			// Case A: Chains are identical. No reconciliation needed.
			fmt.Println("Reconciler: Chains are identical. No conflicts.")
			return []core.Transaction{}, nil
		}

		// Case B: One chain is a valid extension of the other (Subset).
		// The "divergence" (new data) starts at the end of the shorter chain.
		divergenceIndex = minLen
	}

	fmt.Printf("Reconciler: Fork/Extension detected starting at Block Index %d\n", divergenceIndex)

	// Collect ALL transactions from divergent branches
	txMap := make(map[string]core.Transaction)

	// Harvest Chain A
	for _, block := range chainA[divergenceIndex:] {
		for _, tx := range block.Transactions {
			txMap[tx.ID] = tx
		}
	}
	// Harvest Chain B
	for _, block := range chainB[divergenceIndex:] {
		for _, tx := range block.Transactions {
			txMap[tx.ID] = tx
		}
	}

	// Apply Scoring Strategy to ALL transactions
	type ScoredTx struct {
		Tx    core.Transaction
		Score float64
	}
	var rankedTxs []ScoredTx

	for _, tx := range txMap {
		score := r.Strategy.CalculateScore(tx)
		rankedTxs = append(rankedTxs, ScoredTx{Tx: tx, Score: score})
	}

	// Sort by Score (Descending)
	// High score = Higher priority to be included in the merged state
	sort.Slice(rankedTxs, func(i, j int) bool {
		return rankedTxs[i].Score > rankedTxs[j].Score
	})

	// Conflict Resolution (Simplified)
	// In a real C2 system, we would check for semantic conflicts (e.g. "Move to A" vs "Move to B").
	// Here, we simulate that we keep ALL valid scored transactions, ordered by importance.
	// Data loss would occur if we enforced a "One Command Per Unit" rule.

	finalTxs := make([]core.Transaction, len(rankedTxs))
	for i, rt := range rankedTxs {
		finalTxs[i] = rt.Tx
	}

	return finalTxs, nil
}
