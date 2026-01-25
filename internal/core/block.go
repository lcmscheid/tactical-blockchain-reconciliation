/*
 * Copyright (c) 2026 Luiz Carlos Martins Scheid
 * MBA em Engenharia de Software - USP/Esalq
 *
 * Project: Algoritmo de Reconciliação de Bifurcações em Blockchain para Sistemas de Comando Tático
 */

package core

import (
	"time"

	"github.com/lcmscheid/tactical-blockchain-reconciliation/internal/crypto"
)

// NewBlock creates a new block in the chain.
// It requires the previous block's hash to maintain the chain integrity.
func NewBlock(index int, transactions []Transaction, previousHash string, validatorID string) (*Block, error) {
	block := &Block{
		Index:        index,
		Timestamp:    time.Now().Unix(),
		Transactions: transactions,
		PreviousHash: previousHash,
		ValidatorID:  validatorID,
	}

	// Calculate the hash for this new block
	// Note: the Hash field itself is empty during calculation to avoid circular dependency
	hash, err := crypto.CalculateHash(block)
	if err != nil {
		return nil, err
	}
	block.Hash = hash

	return block, nil
}

// NewGenesisBlock creates the first block of the blockchain.
// It has index 0 and an arbitrary previous hash (usually all zeros).
func NewGenesisBlock() *Block {
	// Create a dummy transaction for the genesis block
	genesisTx := Transaction{
		ID:             "GENESIS-TX",
		Timestamp:      time.Now().Unix(),
		CommandContent: "SYSTEM INITIALIZATION - JFC ROOT",
		SignerID:       "SYSTEM",
		Authority:      AuthJFC,
		Priority:       PriorityCritical,
		Dependencies:   []string{},
	}

	block, _ := NewBlock(0, []Transaction{genesisTx}, "0000000000000000000000000000000000000000000000000000000000000000", "SYSTEM")
	return block
}
