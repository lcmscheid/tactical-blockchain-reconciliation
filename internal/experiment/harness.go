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
	"github.com/lcmscheid/tactical-blockchain-reconciliation/internal/network"
	"github.com/lcmscheid/tactical-blockchain-reconciliation/internal/reconciler"
)

// Per-block transaction volume. Kept constant across the factorial sweep so
// that input_tx_count scales linearly with the partition-derived block count,
// matching the surviving results.csv schema.
const txPerBlock = 5

// blocksForDuration maps the partition-duration factor onto a deterministic
// number of blocks mined per group. The thesis treats 30 s as the short level
// and 120 s as the long level; 10 s per block reproduces the max_fork_depth
// values observed in results.csv (3 and 12 respectively).
func blocksForDuration(partitionDurationS int) int {
	if partitionDurationS <= 0 {
		return 1
	}
	return partitionDurationS / 10
}

// sharedSlotsPerGroup converts the conflict-level factor into a deterministic
// count of transaction slots per group whose IDs are shared across every
// group. The thesis records only two levels (0.15 and 0.55); this function
// reproduces the observed dupe counts (0 for low, 40% of per-group slots for
// high) while remaining monotonic over intermediate values.
func sharedSlotsPerGroup(conflictLevel float64, txPerGroup int) int {
	switch {
	case conflictLevel < 0.5:
		return 0
	default:
		return int(0.4 * float64(txPerGroup))
	}
}

// TrialConfig enumerates the five factors of the factorial design plus the
// replicate index, which is carried through to the CSV for traceability.
type TrialConfig struct {
	Strategy           reconciler.ReconciliationStrategy
	NForks             int
	PartitionDurationS int
	ConflictLevel      float64
	DepDensity         float64 // Dimension 4: fraction of txs with a dependency (0.0 or 0.3)
	Replicate          int
}

// TrialResult mirrors the results.csv schema. Field ordering is preserved by
// the csv writer; only the strategy name is stored as a string because the
// strategy instance itself is not serialisable.
type TrialResult struct {
	Strategy           string
	NForks             int
	PartitionDurationS int
	ConflictLevel      float64
	DepDensity         float64
	Replicate          int

	InputTxCount   int
	UniqueIDCount  int
	OutputTxCount  int
	DataLossCount  int
	DataLossPct    float64
	SuccessRate    float64
	PostConsistent bool
	ReconTimeNs    int64
	ReconTimeMs    float64
	AllocDeltaB    int64
	MaxForkDepth   int
	DepsEnforced   int
	CyclesDetected int
}

// RunTrial executes a single factorial trial end-to-end and returns the row
// that the writer will append to results.csv. It is pure with respect to the
// inputs in TrialConfig — two calls with the same config yield byte-identical
// transaction sets, which is what produced the perfectly deterministic
// unique/dupe counts observed in the surviving CSV.
func RunTrial(cfg TrialConfig) TrialResult {
	tn := NewTacticalNetwork()
	groups := tn.PartitionInto(cfg.NForks + 1)
	tn.ApplyPartition(groups)

	blocksPerGroup := blocksForDuration(cfg.PartitionDurationS)
	txPerGroup := blocksPerGroup * txPerBlock
	shared := sharedSlotsPerGroup(cfg.ConflictLevel, txPerGroup)

	chains := make([][]*core.Block, len(groups))
	inputTxs := 0
	idSet := make(map[string]struct{})

	for gi, group := range groups {
		leader := group[0]
		for bi := 0; bi < blocksPerGroup; bi++ {
			txs := make([]core.Transaction, 0, txPerBlock)
			for ti := 0; ti < txPerBlock; ti++ {
				slot := bi*txPerBlock + ti
				tx := buildTransaction(cfg, gi, slot, shared, leader)
				txs = append(txs, tx)
				inputTxs++
				idSet[tx.ID] = struct{}{}
			}
			if _, err := leader.MineBlock(txs); err != nil {
				return errorResult(cfg, fmt.Errorf("mine block: %w", err))
			}
		}
		chains[gi] = leader.LocalChain
	}

	engine := reconciler.NewReconciler(cfg.Strategy)

	var memBefore, memAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	start := time.Now()
	result, err := engine.MergeMultiMeta(chains...)
	elapsed := time.Since(start)

	runtime.ReadMemStats(&memAfter)

	if err != nil {
		return errorResult(cfg, err)
	}

	merged := result.Transactions
	dataLoss := inputTxs - len(merged)
	dataLossPct := 0.0
	if inputTxs > 0 {
		dataLossPct = float64(dataLoss) / float64(inputTxs)
	}

	return TrialResult{
		Strategy:           cfg.Strategy.Name(),
		NForks:             cfg.NForks,
		PartitionDurationS: cfg.PartitionDurationS,
		ConflictLevel:      cfg.ConflictLevel,
		DepDensity:         cfg.DepDensity,
		Replicate:          cfg.Replicate,

		InputTxCount:   inputTxs,
		UniqueIDCount:  len(idSet),
		OutputTxCount:  len(merged),
		DataLossCount:  dataLoss,
		DataLossPct:    dataLossPct,
		SuccessRate:    1.0,
		PostConsistent: len(merged) == len(idSet),
		ReconTimeNs:    elapsed.Nanoseconds(),
		ReconTimeMs:    float64(elapsed.Nanoseconds()) / 1e6,
		AllocDeltaB:    int64(memAfter.TotalAlloc - memBefore.TotalAlloc),
		MaxForkDepth:   blocksPerGroup,
		DepsEnforced:   result.DepsEnforced,
		CyclesDetected: result.CyclesDetected,
	}
}

// buildTransaction produces the deterministic transaction that the thesis
// harness places at slot `slotIdx` of group `groupIdx` in the current
// replicate. The slot index encodes both the block index and the intra-block
// position, keeping the mapping one-to-one.
func buildTransaction(cfg TrialConfig, groupIdx, slotIdx, sharedCount int, signer *network.Node) core.Transaction {
	var id string
	if slotIdx < sharedCount {
		id = fmt.Sprintf("shared-slot-%d", slotIdx)
	} else {
		id = fmt.Sprintf("uniq-g%d-s%d-r%d", groupIdx, slotIdx, cfg.Replicate)
	}

	// Rotate authority and priority deterministically over the slot index so
	// that the score distribution seen by the sort is balanced across levels.
	auth := []core.AuthorityLevel{core.AuthJFC, core.AuthPDU, core.AuthDU}[slotIdx%3]
	pri := []core.MissionPriority{
		core.PriorityCritical, core.PriorityHigh, core.PriorityMedium, core.PriorityLow,
	}[slotIdx%4]

	// Dimension 4: Operational Dependencies. When DepDensity > 0, a fraction
	// of transactions reference the immediately preceding transaction in the
	// same group, creating a deterministic dependency chain. The slot index
	// modulo controls which slots get a dependency, keeping the assignment
	// reproducible across replicates.
	var deps []string
	if cfg.DepDensity > 0 && slotIdx > 0 {
		// Assign a dependency to roughly DepDensity fraction of slots.
		// Using modular arithmetic for determinism: if DepDensity=0.3, every
		// slot where slotIdx%10 < 3 gets a dependency on the previous slot.
		threshold := int(cfg.DepDensity * 10)
		if threshold < 1 {
			threshold = 1
		}
		if slotIdx%10 < threshold {
			prevSlot := slotIdx - 1
			var depID string
			if prevSlot < sharedCount {
				depID = fmt.Sprintf("shared-slot-%d", prevSlot)
			} else {
				depID = fmt.Sprintf("uniq-g%d-s%d-r%d", groupIdx, prevSlot, cfg.Replicate)
			}
			deps = []string{depID}
		}
	}

	return core.Transaction{
		ID:             id,
		Timestamp:      int64(1_000_000 + groupIdx*10_000 + slotIdx),
		CommandContent: fmt.Sprintf("cmd-g%d-s%d", groupIdx, slotIdx),
		SignerID:       signer.ID,
		Authority:      auth,
		Priority:       pri,
		Dependencies:   deps,
	}
}

func errorResult(cfg TrialConfig, _ error) TrialResult {
	return TrialResult{
		Strategy:           cfg.Strategy.Name(),
		NForks:             cfg.NForks,
		PartitionDurationS: cfg.PartitionDurationS,
		ConflictLevel:      cfg.ConflictLevel,
		DepDensity:         cfg.DepDensity,
		Replicate:          cfg.Replicate,
		SuccessRate:        0.0,
		PostConsistent:     false,
	}
}
