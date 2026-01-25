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

// NetworkController manages the simulation topology and particions
type NetworkController struct {
	Nodes map[string]*Node

	// ConnectivityMatrix determines if Node A can talk to Node B.
	// Key: "NodeID1:NodeID2" -> Value: bool (true=connected)
	ConnectivityMatrix map[string]bool
}

func NewNetworkController() *NetworkController {
	return &NetworkController{
		Nodes:              make(map[string]*Node),
		ConnectivityMatrix: make(map[string]bool),
	}
}

// AddNode registers a node and fully connects it to existing peers initially
func (nc *NetworkController) AddNode(n *Node) {
	nc.Nodes[n.ID] = n

	// Default: Connect to everyone (Mesh topology)
	for peerID := range nc.Nodes {
		if peerID == n.ID {
			continue
		}
		nc.SetConnection(n.ID, peerID, true)
	}
}

// SetConnection explicitly enables or disables the link between two nodes
func (nc *NetworkController) SetConnection(nodeA, nodeB string, connected bool) {
	key1 := fmt.Sprintf("%s:%s", nodeA, nodeB)
	key2 := fmt.Sprintf("%s:%s", nodeB, nodeA)

	nc.ConnectivityMatrix[key1] = connected
	nc.ConnectivityMatrix[key2] = connected
}

// CreatePartition isolates a list of node IDs from the rest of the network
func (nc *NetworkController) CreatePartition(isolatedGroup []string) {
	groupMap := make(map[string]bool)
	for _, id := range isolatedGroup {
		groupMap[id] = true
	}

	// Iterate all possible pairs
	for id1 := range nc.Nodes {
		for id2 := range nc.Nodes {
			if id1 == id2 {
				continue
			}

			// If one is in the group and the other is NOT, cut the connection
			inGroup1 := groupMap[id1]
			inGroup2 := groupMap[id2]

			if inGroup1 != inGroup2 {
				nc.SetConnection(id1, id2, false)
			}
		}
	}
	fmt.Printf("--- NETWORK PARTITION ACTIVE: Group %v is isolated ---\n", isolatedGroup)
}

// HealNetwork restores all connections (resolving the partition)
func (nc *NetworkController) HealNetwork() {
	for id1 := range nc.Nodes {
		for id2 := range nc.Nodes {
			if id1 == id2 {
				continue
			}

			nc.SetConnection(id1, id2, true)
		}
	}
	fmt.Println("--- NETWORK HEALED: Connectivity restored ---")
}

// BroadcastBlock attempts to propagate a mined block from sender to all peers.
// It respects the current partitions.
func (nc *NetworkController) BroadcastBlock(senderID string, block *core.Block) {
	sender, exists := nc.Nodes[senderID]
	if !exists {
		return
	}

	for peerID, peer := range nc.Nodes {
		if peerID == senderID {
			continue
		}

		// Check Connectivity
		key := fmt.Sprintf("%s:%s", senderID, peerID)
		if connected, ok := nc.ConnectivityMatrix[key]; ok && connected {
			// Delivery successful
			peer.ReceiveBlock(block)
		} else {
			// Delivery failed due to partition
			fmt.Printf("Network: Dropped packet from %s to %s (Partitioned)\n", senderID, peerID)
		}
	}

	// Note: The sender already added the block to their own chain via MineBlock
	_ = sender
}
