/*
 * Copyright (c) 2026 Luiz Carlos Martins Scheid
 * MBA em Engenharia de Software - USP/Esalq
 *
 * Project: Algoritmo de Reconciliação de Bifurcações em Blockchain para Sistemas de Comando Tático
 */

package network

import (
	"testing"

	"github.com/lcmscheid/tactical-blockchain-reconciliation/internal/core"
)

func TestPartitionScenario(t *testing.T) {
	// Setup
	genesis := core.NewGenesisBlock
	nc := NewNetworkController()

	nodeA := NewNode("JFC-1", core.AuthJFC, genesis())
	nodeB := NewNode("DU-1", core.AuthDU, genesis())

	nc.AddNode(nodeA)
	nc.AddNode(nodeB)

	// Partition: Isolate Node A
	nc.CreatePartition([]string{"JFC-1"})

	// Both mine a block simultaneosly
	txA := core.Transaction{ID: "TX-A", CommandContent: "Attack"}
	txB := core.Transaction{ID: "TX-B", CommandContent: "Retreat"}

	blockA, _ := nodeA.MineBlock([]core.Transaction{txA})
	blockB, _ := nodeB.MineBlock([]core.Transaction{txB})

	// Try to broadcast (should fail due to partition)
	nc.BroadcastBlock(nodeA.ID, blockA)
	nc.BroadcastBlock(nodeB.ID, blockB)

	//  Assert Divergence
	// Node A should only have Block A
	if nodeA.GetHead().Hash != blockA.Hash {
		t.Error("Node A head should be Block A")
	}
	// Node B should only have Block B (it never received A)
	if nodeB.GetHead().Hash != blockB.Hash {
		t.Error("Node B head should be Block B")
	}

	// Assert Fork
	if nodeA.GetHead().Hash == nodeB.GetHead().Hash {
		t.Error("Simulated Fork Failed! Nodes have the same head despite partition.")
	}
}
