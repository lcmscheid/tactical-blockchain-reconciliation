/*
 * Copyright (c) 2026 Luiz Carlos Martins Scheid
 * MBA em Engenharia de Software - USP/Esalq
 *
 * Project: Context-Aware Blockchain Reconciliation Algorithm for Tactical C2
 */
package main

import (
	"fmt"
	"time"

	"github.com/lcmscheid/tactical-blockchain-reconciliation/internal/core"
	"github.com/lcmscheid/tactical-blockchain-reconciliation/internal/network"
	"github.com/lcmscheid/tactical-blockchain-reconciliation/internal/reconciler"
)

func main() {
	fmt.Println("===============================================================")
	fmt.Println("   TACTICAL BLOCKCHAIN RECONCILIATION SIMULATOR v0.1 (PoC)   ")
	fmt.Println("===============================================================")
	fmt.Println("")

	// Execute Test Scenarios
	runScenarioPriorityAuthority()
	fmt.Println("")
	runScenarioTemporalPriority()
	fmt.Println("")
	runScenarioTemporalTieBreaker()
	fmt.Println("")
	runScenarioHybrid()
}

// ============================================================================
// SCENARIO 4: HYBRID STRATEGY (Authority + Priority, normalized linear blend)
// Description: A PDU-Critical transaction meets a JFC-Low one; under the
// default 0.6/0.4 Hybrid weights both score 70, so the four-key sort falls
// back to the timestamp, then to the signer id.
// ============================================================================
func runScenarioHybrid() {
	fmt.Println(">>> SCENARIO 4: Hybrid Blend (Strategy: Hybrid, default 0.6/0.4)")

	genesis := core.NewGenesisBlock()
	nc := network.NewNetworkController()

	nodeJFC := network.NewNode("NODE-JFC-1", core.AuthJFC, genesis)
	nodePDU := network.NewNode("NODE-PDU-ALPHA", core.AuthPDU, genesis)
	nc.AddNode(nodeJFC)
	nc.AddNode(nodePDU)

	fmt.Println(" [1] Network Partition Active: JFC isolated from PDU.")
	nc.CreatePartition([]string{nodeJFC.ID})

	baseTime := time.Now().Unix()

	txPDU := core.Transaction{
		ID:             "TX-PDU-CRIT",
		Timestamp:      baseTime,
		CommandContent: "Local Contact Report (Critical)",
		SignerID:       nodePDU.ID,
		Authority:      core.AuthPDU,
		Priority:       core.PriorityCritical,
	}
	txJFC := core.Transaction{
		ID:             "TX-JFC-LOW",
		Timestamp:      baseTime + 5,
		CommandContent: "Routine Logistical Directive",
		SignerID:       nodeJFC.ID,
		Authority:      core.AuthJFC,
		Priority:       core.PriorityLow,
	}

	nodePDU.MineBlock([]core.Transaction{txPDU})
	nodeJFC.MineBlock([]core.Transaction{txJFC})

	fmt.Println(" [2] Blocks Mined in Isolation (Fork Created).")
	fmt.Println(" [3] Network Healed. Initiating Reconciliation under Hybrid policy...")

	engine := reconciler.NewReconciler(reconciler.NewStrategyHybrid())
	mergedTxs, _ := engine.MergeChains(nodeJFC.LocalChain, nodePDU.LocalChain)

	printResultTable(mergedTxs)
}

// ============================================================================
// SCENARIO 1: HIERARCHY CONFLICT (Priority-Authority Strategy)
// Description: A Joint Forces Commander (JFC) conflicts with a Dispersed Unit (DU).
// Expectation: The JFC transaction should appear first in the reconciled chain.
// ============================================================================
func runScenarioPriorityAuthority() {
	fmt.Println(">>> SCENARIO 1: Hierarchy Conflict (Strategy: Priority-Authority)")

	// 1. Setup Network
	genesis := core.NewGenesisBlock()
	nc := network.NewNetworkController()

	nodeJFC := network.NewNode("NODE-JFC-1", core.AuthJFC, genesis)
	nodeDU := network.NewNode("NODE-DU-5", core.AuthDU, genesis)

	nc.AddNode(nodeJFC)
	nc.AddNode(nodeDU)

	// 2. Create Partition
	fmt.Println(" [1] Network Partition Active: JFC isolated from DU.")
	nc.CreatePartition([]string{nodeJFC.ID})

	// 3. Simulate Divergent Operations (Fork)
	// Transaction A: High Rank (JFC), Low Priority
	txA := core.Transaction{
		ID:             "TX-A (JFC Order)",
		Timestamp:      time.Now().Unix(),
		CommandContent: "Strategic Redeployment (Routine)",
		SignerID:       nodeJFC.ID,
		Authority:      core.AuthJFC,
		Priority:       core.PriorityLow,
	}

	// Transaction B: Low Rank (DU), Critical Priority
	txB := core.Transaction{
		ID:             "TX-B (DU Alert)",
		Timestamp:      time.Now().Unix(),
		CommandContent: "Contact Report (Under Fire)",
		SignerID:       nodeDU.ID,
		Authority:      core.AuthDU,
		Priority:       core.PriorityCritical,
	}

	// Nodes mine on their local, disconnected chains
	nodeJFC.MineBlock([]core.Transaction{txA})
	nodeDU.MineBlock([]core.Transaction{txB})

	fmt.Println(" [2] Blocks Mined in Isolation (Fork Created).")

	// 4. Reconciliation
	fmt.Println(" [3] Network Healed. Initiating Reconciliation...")

	// Select Strategy
	strategy := reconciler.StrategyPriorityAuthority{}
	engine := reconciler.NewReconciler(strategy)

	// Merge
	mergedTxs, _ := engine.MergeChains(nodeJFC.LocalChain, nodeDU.LocalChain)

	// 5. Report Results
	printResultTable(mergedTxs)
}

// ============================================================================
// SCENARIO 2: URGENCY CONFLICT (Temporal-Priority Strategy)
// Description: Two units of equal rank conflict, but one message is CRITICAL.
// Expectation: The Critical transaction should outweigh the others.
// ============================================================================
func runScenarioTemporalPriority() {
	fmt.Println(">>> SCENARIO 2: Urgency Conflict (Strategy: Temporal-Priority)")

	// 1. Setup Network
	genesis := core.NewGenesisBlock()
	nodePDU1 := network.NewNode("PDU-ALPHA", core.AuthPDU, genesis)
	nodePDU2 := network.NewNode("PDU-BRAVO", core.AuthPDU, genesis)

	// 2. Simulate Fork (Skipping network controller boilerplate for brevity)

	// Tx C: Medium Priority
	txC := core.Transaction{
		ID:        "TX-C (Routine)",
		Authority: core.AuthPDU,
		Priority:  core.PriorityMedium,
	}
	// Tx D: Critical Priority
	txD := core.Transaction{
		ID:        "TX-D (CRITICAL)",
		Authority: core.AuthPDU,
		Priority:  core.PriorityCritical,
	}

	nodePDU1.MineBlock([]core.Transaction{txC})
	nodePDU2.MineBlock([]core.Transaction{txD})

	// 3. Reconciliation
	strategy := reconciler.StrategyTemporalAuthority{}
	engine := reconciler.NewReconciler(strategy)

	mergedTxs, _ := engine.MergeChains(nodePDU1.LocalChain, nodePDU2.LocalChain)

	// 4. Report
	printResultTable(mergedTxs)
}

// ============================================================================
// SCENARIO 3: TEMPORAL TIE-BREAKER (Auditability Check)
// Description: Two units of SAME RANK emit orders of SAME PRIORITY.
// Challenge: The system must order them chronologically (Time A < Time B).
// ============================================================================
func runScenarioTemporalTieBreaker() {
	fmt.Println(">>> SCENARIO 3: Temporal Tie-Breaker (Same Rank/Priority)")
	fmt.Println("    [Testing if Engine respects chronological order for semantic ties]")

	// 1. Setup Network
	genesis := core.NewGenesisBlock()
	// Two nodes of equal Authority (PDU)
	nodeAlpha := network.NewNode("PDU-ALPHA", core.AuthPDU, genesis)
	nodeBravo := network.NewNode("PDU-BRAVO", core.AuthPDU, genesis)

	// 2. Simulate Fork
	// We simulate that Alpha acted BEFORE Bravo.
	baseTime := time.Now().Unix()

	// Tx 1: Occurred at T+0s
	txEarly := core.Transaction{
		ID:             "TX-EARLY (Alpha)",
		Timestamp:      baseTime, // EARLIER
		CommandContent: "Initial Spot Report",
		SignerID:       nodeAlpha.ID,
		Authority:      core.AuthPDU,      // Same Authority
		Priority:       core.PriorityHigh, // Same Priority
	}

	// Tx 2: Occurred at T+5s
	txLate := core.Transaction{
		ID:             "TX-LATE (Bravo)",
		Timestamp:      baseTime + 5, // LATER
		CommandContent: "Confirmation Update",
		SignerID:       nodeBravo.ID,
		Authority:      core.AuthPDU,      // Same Authority
		Priority:       core.PriorityHigh, // Same Priority
	}

	// Mine separate blocks (Fork)
	nodeAlpha.MineBlock([]core.Transaction{txEarly})
	nodeBravo.MineBlock([]core.Transaction{txLate})

	fmt.Printf(" [1] Fork Created: TX-EARLY (Time: %d) vs TX-LATE (Time: %d)\n", txEarly.Timestamp, txLate.Timestamp)

	// 3. Reconciliation
	// We use the Priority-Authority strategy (which calculates identical scores for both)
	strategy := reconciler.StrategyPriorityAuthority{}
	engine := reconciler.NewReconciler(strategy)

	fmt.Println(" [2] Reconciling...")
	mergedTxs, _ := engine.MergeChains(nodeAlpha.LocalChain, nodeBravo.LocalChain)

	// 4. Report
	printResultTable(mergedTxs)
}

// Helper to print the "Preliminary Results" table
func printResultTable(txs []core.Transaction) {
	fmt.Printf("\n   %-20s | %-10s | %-10s | %s\n", "Transaction ID", "Authority", "Priority", "Status")
	fmt.Println("   ---------------------|------------|------------|-----------")

	for i, tx := range txs {
		status := fmt.Sprintf("Pos: %d", i+1)
		fmt.Printf("   %-20s | %-10s | %-10s | %s\n",
			tx.ID, tx.Authority, tx.Priority, status)
	}
	fmt.Println("   -----------------------------------------------------------")
}
