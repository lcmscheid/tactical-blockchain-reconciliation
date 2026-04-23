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

func TestMergeChains_TemporalStrategy_UrgencyWins(t *testing.T) {
	// 1. Setup Strategy: Temporal (Prioriza Critical/High sobre Patente)
	strategy := StrategyTemporalAuthority{}
	reconciler := NewReconciler(strategy)

	// 2. Cenario: "O Cabo Gritando Fogo" vs "O General Pedindo Café"
	// Tx A: General (JFC) com prioridade Baixa/Media
	txGeneral := core.Transaction{
		ID:        "TX-GENERAL",
		Authority: core.AuthJFC,        // Rank Máximo
		Priority:  core.PriorityMedium, // Peso na StrategyTemporal: 100.0 + 5 (Bonus JFC) = 105.0
	}

	// Tx B: Unidade Dispersa (DU) com prioridade Crítica
	txSoldier := core.Transaction{
		ID:        "TX-SOLDIER",
		Authority: core.AuthDU,           // Rank Mínimo
		Priority:  core.PriorityCritical, // Peso na StrategyTemporal: 1000.0
	}

	// 3. Criar Cadeias
	chainA := []*core.Block{{Index: 1, Transactions: []core.Transaction{txGeneral}}}
	chainB := []*core.Block{{Index: 1, Transactions: []core.Transaction{txSoldier}}}

	// 4. Merge
	mergedTxs, _ := reconciler.MergeChains(chainA, chainB)

	// 5. Validação: Soldado deve vencer General neste modo
	if mergedTxs[0].ID != txSoldier.ID {
		t.Errorf("Falha na Estratégia Temporal!\nEsperado 1º: %s (Score ~1000)\nObtido 1º:   %s (Score ~105)",
			txSoldier.ID, mergedTxs[0].ID)
	}
}

func TestMergeChains_TemporalStrategy_TieBreaker(t *testing.T) {
	// Testar se o desempate temporal funciona mesmo com a estratégia de Urgência
	strategy := StrategyTemporalAuthority{}
	reconciler := NewReconciler(strategy)
	baseTime := int64(5000)

	// Duas transações de MESMA urgência (Critical) e MESMO Rank
	tx1 := core.Transaction{
		ID: "TX-CRITICAL-EARLY", Timestamp: baseTime,
		Authority: core.AuthPDU, Priority: core.PriorityCritical,
	}
	tx2 := core.Transaction{
		ID: "TX-CRITICAL-LATE", Timestamp: baseTime + 10,
		Authority: core.AuthPDU, Priority: core.PriorityCritical,
	}

	chainA := []*core.Block{{Transactions: []core.Transaction{tx2}}}
	chainB := []*core.Block{{Transactions: []core.Transaction{tx1}}}

	mergedTxs, _ := reconciler.MergeChains(chainA, chainB)

	// Deve ordenar pelo tempo (Early -> Late)
	if mergedTxs[0].ID != tx1.ID {
		t.Error("O motor falhou em desempatar por tempo na estratégia Temporal.")
	}
}

func TestMergeChains_TemporalTieBreaker(t *testing.T) {
	// 1. Setup Strategy that allows ties (Priority-Authority)
	strategy := StrategyPriorityAuthority{}
	reconciler := NewReconciler(strategy)

	// 2. Setup Base Time
	baseTime := int64(100000)

	// 3. Create Two Transactions with IDENTICAL Scores but DIFFERENT Times
	// Both are PDU (Score 100) + High Priority (Score 3) = Total 103
	txEarly := core.Transaction{
		ID:        "TX-EARLY",
		Timestamp: baseTime, // T=100000
		Authority: core.AuthPDU,
		Priority:  core.PriorityHigh,
	}

	txLate := core.Transaction{
		ID:        "TX-LATE",
		Timestamp: baseTime + 50, // T=100050
		Authority: core.AuthPDU,
		Priority:  core.PriorityHigh,
	}

	// 4. Create Divergent Chains
	// Chain A has the LATE transaction
	blockA := &core.Block{
		Index:        1,
		Hash:         "hashA",
		Transactions: []core.Transaction{txLate},
	}
	chainA := []*core.Block{blockA}

	// Chain B has the EARLY transaction
	blockB := &core.Block{
		Index:        1,
		Hash:         "hashB",
		Transactions: []core.Transaction{txEarly},
	}
	chainB := []*core.Block{blockB}

	// 5. Execute Merge
	// We pass chainA first to ensure "First Found" logic doesn't accidentally pass the test.
	// The engine must explicitly sort them.
	mergedTxs, err := reconciler.MergeChains(chainA, chainB)
	if err != nil {
		t.Fatalf("MergeChains failed: %v", err)
	}

	// 6. Assertions
	if len(mergedTxs) != 2 {
		t.Fatalf("Expected 2 transactions, got %d", len(mergedTxs))
	}

	// Verification 1: Order must be Early -> Late
	if mergedTxs[0].ID != txEarly.ID {
		t.Errorf("Ordering Mismatch!\nExpected First: %s (Time: %d)\nGot First:      %s (Time: %d)\nNote: Engine should sort equal scores by timestamp ascending.",
			txEarly.ID, txEarly.Timestamp, mergedTxs[0].ID, mergedTxs[0].Timestamp)
	}

	if mergedTxs[1].ID != txLate.ID {
		t.Errorf("Ordering Mismatch!\nExpected Second: %s\nGot Second:      %s",
			txLate.ID, mergedTxs[1].ID)
	}
}

func TestStrategyHybrid_DefaultWeights_ScoresMatchTable(t *testing.T) {
	s := NewStrategyHybrid()

	cases := []struct {
		name string
		tx   core.Transaction
		want float64
	}{
		{"JFC-Critical", core.Transaction{Authority: core.AuthJFC, Priority: core.PriorityCritical}, 100},
		{"JFC-Low", core.Transaction{Authority: core.AuthJFC, Priority: core.PriorityLow}, 70},
		{"PDU-Critical", core.Transaction{Authority: core.AuthPDU, Priority: core.PriorityCritical}, 70},
		{"PDU-Medium", core.Transaction{Authority: core.AuthPDU, Priority: core.PriorityMedium}, 50},
		{"DU-Critical", core.Transaction{Authority: core.AuthDU, Priority: core.PriorityCritical}, 46},
		{"DU-Low", core.Transaction{Authority: core.AuthDU, Priority: core.PriorityLow}, 16},
	}
	for _, tc := range cases {
		got := s.CalculateScore(tc.tx)
		if got != tc.want {
			t.Errorf("%s: got %.4f, want %.4f", tc.name, got, tc.want)
		}
	}
}

func TestStrategyHybrid_JFCOutranksDUOnTie(t *testing.T) {
	// JFC+Medium = 0.6*100 + 0.4*50 = 80; DU+Critical = 0.6*10 + 0.4*100 = 46.
	// Hybrid must place the JFC tx first despite the DU tx being Critical.
	r := NewReconciler(NewStrategyHybrid())
	txJFC := core.Transaction{ID: "TX-JFC", Authority: core.AuthJFC, Priority: core.PriorityMedium}
	txDU := core.Transaction{ID: "TX-DU", Authority: core.AuthDU, Priority: core.PriorityCritical}

	chainA := []*core.Block{{Transactions: []core.Transaction{txDU}}}
	chainB := []*core.Block{{Transactions: []core.Transaction{txJFC}}}
	merged, _ := r.MergeChains(chainA, chainB)

	if merged[0].ID != "TX-JFC" {
		t.Fatalf("Hybrid strategy: expected TX-JFC first, got %s", merged[0].ID)
	}
}

func TestStrategyHybrid_PDUCriticalTiesJFCLow_BreaksByTimestamp(t *testing.T) {
	// Both score 70 under the default weights, so the timestamp key decides.
	r := NewReconciler(NewStrategyHybrid())
	txEarly := core.Transaction{
		ID: "TX-PDU-EARLY", Timestamp: 1000,
		SignerID: "PDU-ALPHA", Authority: core.AuthPDU, Priority: core.PriorityCritical,
	}
	txLate := core.Transaction{
		ID: "TX-JFC-LATE", Timestamp: 2000,
		SignerID: "NODE-JFC-1", Authority: core.AuthJFC, Priority: core.PriorityLow,
	}

	chainA := []*core.Block{{Transactions: []core.Transaction{txLate}}}
	chainB := []*core.Block{{Transactions: []core.Transaction{txEarly}}}
	merged, _ := r.MergeChains(chainA, chainB)

	if merged[0].ID != txEarly.ID {
		t.Fatalf("Hybrid tie: expected earlier timestamp first, got %s", merged[0].ID)
	}
}

func TestStrategyHybrid_CustomWeights_PriorityDominant(t *testing.T) {
	// Invert the default compromise so priority dominates authority.
	s := StrategyHybrid{WeightAuthority: 0.2, WeightPriority: 0.8}
	r := NewReconciler(s)
	txJFCLow := core.Transaction{
		ID: "TX-JFC-LOW", SignerID: "JFC", Authority: core.AuthJFC, Priority: core.PriorityLow,
	}
	txDUCrit := core.Transaction{
		ID: "TX-DU-CRIT", SignerID: "DU", Authority: core.AuthDU, Priority: core.PriorityCritical,
	}

	// With these weights JFC+Low = 0.2*100 + 0.8*25 = 40;
	// DU+Critical = 0.2*10 + 0.8*100 = 82 — DU must win.
	chainA := []*core.Block{{Transactions: []core.Transaction{txJFCLow}}}
	chainB := []*core.Block{{Transactions: []core.Transaction{txDUCrit}}}
	merged, _ := r.MergeChains(chainA, chainB)

	if merged[0].ID != txDUCrit.ID {
		t.Fatalf("Hybrid with priority-dominant weights: expected DU first, got %s", merged[0].ID)
	}
}

func TestSort_FullTie_BreaksByTxID(t *testing.T) {
	// Identical score, timestamp, and signer: only the transaction id can decide.
	r := NewReconciler(StrategyPriorityAuthority{})
	common := core.Transaction{
		Timestamp: 500, SignerID: "PDU-ALPHA",
		Authority: core.AuthPDU, Priority: core.PriorityHigh,
	}
	txFirst := common
	txFirst.ID = "TX-AAAA"
	txSecond := common
	txSecond.ID = "TX-ZZZZ"

	chainA := []*core.Block{{Transactions: []core.Transaction{txSecond}}}
	chainB := []*core.Block{{Transactions: []core.Transaction{txFirst}}}
	merged, _ := r.MergeChains(chainA, chainB)

	if merged[0].ID != "TX-AAAA" || merged[1].ID != "TX-ZZZZ" {
		t.Fatalf("Four-key sort failed on full tie: got %s, %s", merged[0].ID, merged[1].ID)
	}
}

// ---------- Dependency tests (Dimension 4) ----------

func TestDeps_BasicOrdering(t *testing.T) {
	// TX-A (JFC, score 1001) depends on TX-B (DU, score 14).
	// Without deps TX-A would be first; with deps TX-B must come first.
	r := NewReconciler(StrategyPriorityAuthority{})

	txA := core.Transaction{
		ID: "TX-A", Authority: core.AuthJFC, Priority: core.PriorityLow,
		Dependencies: []string{"TX-B"},
	}
	txB := core.Transaction{
		ID: "TX-B", Authority: core.AuthDU, Priority: core.PriorityCritical,
	}

	chainA := []*core.Block{{Transactions: []core.Transaction{txA}}}
	chainB := []*core.Block{{Transactions: []core.Transaction{txB}}}
	merged, _ := r.MergeChains(chainA, chainB)

	if len(merged) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(merged))
	}
	if merged[0].ID != "TX-B" {
		t.Errorf("dependency ordering failed: expected TX-B first (dependency), got %s", merged[0].ID)
	}
	if merged[1].ID != "TX-A" {
		t.Errorf("dependency ordering failed: expected TX-A second (dependent), got %s", merged[1].ID)
	}
}

func TestDeps_ChainOfThree(t *testing.T) {
	// TX-C depends on TX-B, TX-B depends on TX-A. All have descending scores.
	r := NewReconciler(StrategyPriorityAuthority{})

	txA := core.Transaction{ID: "TX-A", Authority: core.AuthDU, Priority: core.PriorityLow}
	txB := core.Transaction{
		ID: "TX-B", Authority: core.AuthPDU, Priority: core.PriorityMedium,
		Dependencies: []string{"TX-A"},
	}
	txC := core.Transaction{
		ID: "TX-C", Authority: core.AuthJFC, Priority: core.PriorityCritical,
		Dependencies: []string{"TX-B"},
	}

	chain := []*core.Block{{Transactions: []core.Transaction{txC, txB, txA}}}
	chainEmpty := []*core.Block{{Transactions: []core.Transaction{}}}
	merged, _ := r.MergeChains(chain, chainEmpty)

	if len(merged) != 3 {
		t.Fatalf("expected 3 transactions, got %d", len(merged))
	}
	if merged[0].ID != "TX-A" || merged[1].ID != "TX-B" || merged[2].ID != "TX-C" {
		t.Errorf("chain ordering failed: got %s -> %s -> %s, want TX-A -> TX-B -> TX-C",
			merged[0].ID, merged[1].ID, merged[2].ID)
	}
}

func TestDeps_IndependentGroups(t *testing.T) {
	// Two independent dependency chains. Within each chain, order by deps.
	// Between chains, the higher-scored chain's root goes first.
	r := NewReconciler(StrategyPriorityAuthority{})

	// Chain 1: JFC -> PDU (JFC depends on PDU)
	tx1Root := core.Transaction{ID: "TX-PDU-ROOT", Authority: core.AuthPDU, Priority: core.PriorityHigh}
	tx1Dep := core.Transaction{
		ID: "TX-JFC-DEP", Authority: core.AuthJFC, Priority: core.PriorityLow,
		Dependencies: []string{"TX-PDU-ROOT"},
	}

	// Chain 2: independent DU transaction (no deps, score 14)
	tx2 := core.Transaction{ID: "TX-DU-INDEP", Authority: core.AuthDU, Priority: core.PriorityCritical}

	chain := []*core.Block{{Transactions: []core.Transaction{tx1Dep, tx1Root, tx2}}}
	chainEmpty := []*core.Block{{Transactions: []core.Transaction{}}}
	merged, _ := r.MergeChains(chain, chainEmpty)

	if len(merged) != 3 {
		t.Fatalf("expected 3 transactions, got %d", len(merged))
	}

	// TX-PDU-ROOT (score 103) must come before TX-JFC-DEP (score 1001) due to dep.
	// TX-DU-INDEP (score 14) has no deps so it's sorted by score among zero-in-degree txs.
	// Zero-in-degree set: TX-PDU-ROOT (103) and TX-DU-INDEP (14).
	// TX-PDU-ROOT emitted first (higher score), then TX-JFC-DEP (dep satisfied), then TX-DU-INDEP.
	// Wait — after TX-PDU-ROOT is emitted, TX-JFC-DEP becomes zero-in-degree (score 1001)
	// and TX-DU-INDEP is already zero (score 14). So TX-JFC-DEP goes next.
	if merged[0].ID != "TX-PDU-ROOT" {
		t.Errorf("expected TX-PDU-ROOT first, got %s", merged[0].ID)
	}
	if merged[1].ID != "TX-JFC-DEP" {
		t.Errorf("expected TX-JFC-DEP second (dep satisfied), got %s", merged[1].ID)
	}
	if merged[2].ID != "TX-DU-INDEP" {
		t.Errorf("expected TX-DU-INDEP third, got %s", merged[2].ID)
	}
}

func TestDeps_MissingDep(t *testing.T) {
	// TX-A depends on "TX-COMMITTED" which is not in the merge set.
	// The dependency should be treated as satisfied (already in common prefix).
	r := NewReconciler(StrategyPriorityAuthority{})

	txA := core.Transaction{
		ID: "TX-A", Authority: core.AuthJFC, Priority: core.PriorityLow,
		Dependencies: []string{"TX-COMMITTED"},
	}
	txB := core.Transaction{
		ID: "TX-B", Authority: core.AuthDU, Priority: core.PriorityCritical,
	}

	chainA := []*core.Block{{Transactions: []core.Transaction{txA}}}
	chainB := []*core.Block{{Transactions: []core.Transaction{txB}}}
	merged, _ := r.MergeChains(chainA, chainB)

	// TX-A (score 1001) should be first since its dep is outside the merge set.
	if merged[0].ID != "TX-A" {
		t.Errorf("missing dep should be treated as satisfied: expected TX-A first, got %s", merged[0].ID)
	}
}

func TestDeps_CycleDetection(t *testing.T) {
	// TX-A depends on TX-B, TX-B depends on TX-A — a cycle.
	// Both should still appear in output (appended by score), not lost.
	r := NewReconciler(StrategyPriorityAuthority{})

	txA := core.Transaction{
		ID: "TX-A", Authority: core.AuthJFC, Priority: core.PriorityLow,
		Dependencies: []string{"TX-B"},
	}
	txB := core.Transaction{
		ID: "TX-B", Authority: core.AuthDU, Priority: core.PriorityCritical,
		Dependencies: []string{"TX-A"},
	}

	chainA := []*core.Block{{Transactions: []core.Transaction{txA}}}
	chainB := []*core.Block{{Transactions: []core.Transaction{txB}}}
	merged, _ := r.MergeChains(chainA, chainB)

	if len(merged) != 2 {
		t.Fatalf("cycle should not lose transactions: expected 2, got %d", len(merged))
	}
	// Both are in a cycle, so they're appended by score. TX-A (JFC, 1001) > TX-B (DU, 14).
	if merged[0].ID != "TX-A" {
		t.Errorf("cycle fallback: expected TX-A first (higher score), got %s", merged[0].ID)
	}
}

func TestDeps_NoDeps_BackwardCompat(t *testing.T) {
	// Re-run the original Priority-Authority test to confirm backward compatibility.
	r := NewReconciler(StrategyPriorityAuthority{})

	txA := core.Transaction{ID: "TX-DU", Authority: core.AuthDU, Priority: core.PriorityCritical}
	txB := core.Transaction{ID: "TX-JFC", Authority: core.AuthJFC, Priority: core.PriorityLow}

	blockA, _ := core.NewBlock(1, []core.Transaction{txA}, "genesis_hash", "ValidatorA")
	blockB, _ := core.NewBlock(1, []core.Transaction{txB}, "genesis_hash", "ValidatorB")

	merged, _ := r.MergeChains([]*core.Block{blockA}, []*core.Block{blockB})

	if len(merged) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(merged))
	}
	if merged[0].ID != "TX-JFC" {
		t.Errorf("backward compat failed: expected TX-JFC first, got %s", merged[0].ID)
	}
}

func TestDeps_SelfDependency_Ignored(t *testing.T) {
	// A transaction that depends on itself should not deadlock.
	r := NewReconciler(StrategyPriorityAuthority{})

	txA := core.Transaction{
		ID: "TX-A", Authority: core.AuthJFC, Priority: core.PriorityLow,
		Dependencies: []string{"TX-A"},
	}
	txB := core.Transaction{ID: "TX-B", Authority: core.AuthDU, Priority: core.PriorityCritical}

	chainA := []*core.Block{{Transactions: []core.Transaction{txA}}}
	chainB := []*core.Block{{Transactions: []core.Transaction{txB}}}
	merged, _ := r.MergeChains(chainA, chainB)

	if len(merged) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(merged))
	}
	if merged[0].ID != "TX-A" {
		t.Errorf("self-dep should be ignored: expected TX-A first (score 1001), got %s", merged[0].ID)
	}
}
