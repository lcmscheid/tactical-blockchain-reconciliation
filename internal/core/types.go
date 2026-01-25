/*
 * Copyright (c) 2026 Luiz Carlos Martins Scheid
 * MBA em Engenharia de Software - USP/Esalq
 *
 * Project: Algoritmo de Reconciliação de Bifurcações em Blockchain para Sistemas de Comando Tático
 */

// Package core contains domain types
package core

// AuthorityLevel represents the command hierarchy level of the node
type AuthorityLevel int

const (
	// JFC - Joint Forces Commander (Highest Authority)
	AuthJFC AuthorityLevel = iota + 1
	// PDU - Parent Dispersed Unit (Intermediate)
	AuthPDU
	// DU - Dispersed Unit (Leaf node / Lowest Authority)
	AuthDU
)

// String returns the string representation of the AuthorityLevel
func (a AuthorityLevel) String() string {
	return [...]string{"Unknown", "JFC", "PDU", "DU"}[a]
}

type MissionPriority int

const (
	// Critical - Immediate impact on strategic success or survival
	PriorityCritical MissionPriority = iota + 1
	// High - Significant operational impact
	PriorityHigh
	// Medium - Routine tactical operations
	PriorityMedium
	// Low - Administrative or logistic updates
	PriorityLow
)

// String returns the string representation of MissionPriority
func (p MissionPriority) String() string {
	return [...]string{"Unknown", "Critical", "High", "Medium", "Low"}[p]
}

// Transaction represents a tactical military decision
type Transaction struct {
	ID        string `json:"id"`
	Timestamp int64  `json:"timestamp"` // Dimension 2: Temporal Precedence

	// Payload
	CommandContent string `json:"command_content"` // e.g., "Attack Grid A", "Resupply Unit B"

	// Context Awareness Metadata
	SignerID     string          `json:"signer_id"`
	Authority    AuthorityLevel  `json:"authority"`        // Dimension 1
	Priority     MissionPriority `json:"mission_priority"` // Dimension 3
	Dependencies []string        `json:"dependencies"`     // Dimension 4: Operational Dependencies
}

// Block represents a unit of storage in  the chain
type Block struct {
	Index        int           `json:"index"`
	Timestamp    int64         `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`

	// Cryptography link
	Hash         string `json:"hash"`
	PreviousHash string `json:"previous_hash"`

	// Simulator Metadata
	ValidatorID string `json:"validator_id"`
}

// Chain represents the ledger held by a node
type Chain struct {
	Blocks []Block
}
