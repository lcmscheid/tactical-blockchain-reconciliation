/*
 * Copyright (c) 2026 Luiz Carlos Martins Scheid
 * MBA em Engenharia de Software - USP/Esalq
 *
 * Project: Algoritmo de Reconciliação de Bifurcações em Blockchain para Sistemas de Comando Tático
 */

package core

import (
	"testing"
)

func TestNewGenesisBlock(t *testing.T) {
	genesis := NewGenesisBlock()

	if genesis.Index != 0 {
		t.Errorf("Genesis block index should be 0, got %d", genesis.Index)
	}

	if genesis.PreviousHash == "" {
		t.Error("Genesis block should have a placeholder previous hash")
	}

	if len(genesis.Transactions) == 0 {
		t.Error("Genesis block should contain at least the initialization transaction")
	}
}

func TestNewBlock_Chaining(t *testing.T) {
	genesis := NewGenesisBlock()

	tx := Transaction{ID: "TEST-TX", Authority: AuthPDU}
	block1, err := NewBlock(1, []Transaction{tx}, genesis.Hash, "VALIDATOR-1")
	if err != nil {
		t.Fatalf("Failed to create Block 1: %v", err)
	}

	// Verify the Link
	if block1.PreviousHash != genesis.Hash {
		t.Errorf("Chain broken. Block 1 PreviousHash (%s) != Genesis Hash (%s)",
			block1.PreviousHash, genesis.Hash)
	}

	// Verify Self-Hash
	if block1.Hash == "" {
		t.Error("Block 1 was not hashed")
	}
}
