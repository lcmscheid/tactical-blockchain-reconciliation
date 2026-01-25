/*
 * Copyright (c) 2026 Luiz Carlos Martins Scheid
 * MBA em Engenharia de Software - USP/Esalq
 *
 * Project: Context-Aware Blockchain Reconciliation Algorithm for Tactical C2
 */

package reconciler

import (
	"testing"

	"github.com/lcmscheid/tactical-blockchain-reconciliation/internal/core"
)

func TestMerge_PriorityAuthority(t *testing.T) {
	// Setup Strategy
	strategy := StrategyPriorityAuthority{}
	reconciler := NewReconciler(strategy)

	// Create conflicting transactions
	// TxA: Low Rank (DU) but Critical Priority (e.g., "I'm under fire")
	txA := core.Transaction{
		ID: "TX-DU", Authority: core.AuthDU, Priority: core.PriorityCritical,
	}
	// TxB: High Rank (JFC) but Low Priority (e.g., "Routine Report")
	txB := core.Transaction{
		ID: "TX-JFC", Authority: core.AuthJFC, Priority: core.PriorityLow,
	}

	// Calculate Scores Manually to verify expectation
	// JFC(1000) + Low(1) = 1001
	// DU(10) + Critical(4) = 14
	// Expectation: JFC wins despite low priority because Authority weight is massive in this strategy.

	// Mock Blocks
	blockA, _ := core.NewBlock(1, []core.Transaction{txA}, "genesis_hash", "ValidatorA")
	blockB, _ := core.NewBlock(1, []core.Transaction{txB}, "genesis_hash", "ValidatorB")

	chainA := []*core.Block{blockA}
	chainB := []*core.Block{blockB}

	// Merge
	merged, _ := reconciler.MergeChains(chainA, chainB)

	if len(merged) != 2 {
		t.Fatalf("Expected 2 transactions merged, got %d", len(merged))
	}

	// Assert Order
	if merged[0].ID != "TX-JFC" {
		t.Errorf("Priority-Authority Strategy failed. Expected JFC first, got %s", merged[0].ID)
	}
}
