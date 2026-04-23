/*
 * Copyright (c) 2026 Luiz Carlos Martins Scheid
 * MBA em Engenharia de Software - USP/Esalq
 *
 * Project: Algoritmo de Reconciliação de Bifurcações em Blockchain para Sistemas de Comando Tático
 */

package reconciler

import "github.com/lcmscheid/tactical-blockchain-reconciliation/internal/core"

// ReconciliationStrategy defines the contract for ranking conflicting transactions
type ReconciliationStrategy interface {
	Name() string
	CalculateScore(tx core.Transaction) float64
}

// StrategyPriorityAuthority favors chain of command above all else.
type StrategyPriorityAuthority struct{}

func (s StrategyPriorityAuthority) Name() string {
	return "Priority-Authority"
}

func (s StrategyPriorityAuthority) CalculateScore(tx core.Transaction) float64 {
	score := 0.0

	// Dimension 1: Hierarchy (Weight: 1000)
	// JFC(1) -> 1000, PDU(2) -> 100, DU(3) -> 10
	switch tx.Authority {
	case core.AuthJFC:
		score += 1000.0
	case core.AuthPDU:
		score += 100.0
	case core.AuthDU:
		score += 10.0
	}

	// Dimension 3: Mission Priority (Tie-braker, Weight: 1)
	switch tx.Priority {
	case core.PriorityCritical:
		score += 4.0
	case core.PriorityHigh:
		score += 3.0
	case core.PriorityMedium:
		score += 2.0
	case core.PriorityLow:
		score += 1.0
	}

	return score
}

// StrategyTemporalAuthority favors urgency (Critical Priority) regardless of rank.
type StrategyTemporalAuthority struct{}

func (s StrategyTemporalAuthority) Name() string {
	return "Temporal-Authority"
}

func (s StrategyTemporalAuthority) CalculateScore(tx core.Transaction) float64 {
	score := 0.0

	// Dimension 3: Mission Priority (Primary Factor)
	switch tx.Priority {
	case core.PriorityCritical:
		score += 1000.0
	case core.PriorityHigh:
		score += 500.0
	case core.PriorityMedium:
		score += 100.0
	case core.PriorityLow:
		score += 10.0
	}

	// Hierarchy acts only as a minor tie-breaker here
	if tx.Authority == core.AuthJFC {
		score += 5.0
	}

	return score
}

// StrategyHybrid combines authority and priority linearly after normalizing both
// to the interval [0, 100]. The default weights (0.6 for authority, 0.4 for
// priority) reflect the compromise documented in the thesis and can be
// overridden per configuration to match a different doctrinal profile.
//
// Normalization tables:
//
//	Authority: JFC=100, PDU=50, DU=10
//	Priority : Critical=100, High=75, Medium=50, Low=25
//
// With default weights a PDU signing a Critical transaction ties a JFC
// signing a Low transaction at score 70, a property that the draft cites to
// illustrate the intermediate behaviour of the hybrid policy.
type StrategyHybrid struct {
	WeightAuthority float64
	WeightPriority  float64
}

// NewStrategyHybrid returns a StrategyHybrid pre-configured with the default
// 0.6 / 0.4 weights.
func NewStrategyHybrid() StrategyHybrid {
	return StrategyHybrid{WeightAuthority: 0.6, WeightPriority: 0.4}
}

func (s StrategyHybrid) Name() string {
	return "Hybrid"
}

func (s StrategyHybrid) CalculateScore(tx core.Transaction) float64 {
	var auth float64
	switch tx.Authority {
	case core.AuthJFC:
		auth = 100.0
	case core.AuthPDU:
		auth = 50.0
	case core.AuthDU:
		auth = 10.0
	}

	var pri float64
	switch tx.Priority {
	case core.PriorityCritical:
		pri = 100.0
	case core.PriorityHigh:
		pri = 75.0
	case core.PriorityMedium:
		pri = 50.0
	case core.PriorityLow:
		pri = 25.0
	}

	return s.WeightAuthority*auth + s.WeightPriority*pri
}
