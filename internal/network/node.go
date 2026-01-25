/*
 * Copyright (c) 2026 Luiz Carlos Martins Scheid
 * MBA em Engenharia de Software - USP/Esalq
 *
 * Project: Algoritmo de Reconciliação de Bifurcações em Blockchain para Sistemas de Comando Tático
 */

package network

import (
	"fmt"

	"github.com/lcmscheid/tactical-blockchain-reconciliation/internal/core"
)

type Node struct {
	ID        string
	Authority core.AuthorityLevel
	// LocalChain represents this node's view of the truth.
	// In a partition, this will diverge from other nodes.
	LocalChain []*core.Block
}

// NewNode creates a node with the Genesis block already initialized
func NewNode(id string, auth core.AuthorityLevel, genesis *core.Block) *Node {
	return &Node{
		ID:         id,
		Authority:  auth,
		LocalChain: []*core.Block{genesis},
	}
}

// GetHead returns the latest block in this node's local chain
func (n *Node) GetHead() *core.Block {
	if len(n.LocalChain) == 0 {
		return nil
	}
	return n.LocalChain[len(n.LocalChain)-1]
}

// MineBlock simulates the node creating a new block based on its current state.
// It links the new block to its own specific Head (PreviousHash).
func (n *Node) MineBlock(txs []core.Transaction) (*core.Block, error) {
	head := n.GetHead()

	// Create the block pointing to the local head
	newBlock, err := core.NewBlock(head.Index+1, txs, head.Hash, n.ID)
	if err != nil {
		return nil, fmt.Errorf("node %s failed to mine: %w", n.ID, err)
	}

	// Append to local chain immediately (valid in this simplified sim)
	n.LocalChain = append(n.LocalChain, newBlock)
	return newBlock, nil
}

// ReceiveBlock is called when the network delivers a block from a peer.
// In a real blockchain, extensive validation happens here.
func (n *Node) ReceiveBlock(b *core.Block) {
	head := n.GetHead()

	// Simple chain continuity check
	if b.PreviousHash == head.Hash {
		n.LocalChain = append(n.LocalChain, b)
		fmt.Printf("[%s] Synced block %d from %s\n", n.ID, b.Index, b.ValidatorID)
	} else {
		// If hashes don't match, we have a FORK or a missing block.
		// For this simulation, we log it as a divergence.
		fmt.Printf("[%s] REJECTED block %d from %s (Hash Mismatch - Fork Detected)\n", n.ID, b.Index, b.ValidatorID)
	}
}
