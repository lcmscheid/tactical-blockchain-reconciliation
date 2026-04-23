/*
 * Copyright (c) 2026 Luiz Carlos Martins Scheid
 * MBA em Engenharia de Software - USP/Esalq
 *
 * Project: Algoritmo de Reconciliação de Bifurcações em Blockchain para Sistemas de Comando Tático
 */

// Package experiment contains the factorial trial harness, the complexity
// micro-benchmark and the 9-node tactical topology that back the experimental
// evidence reported in the thesis. It is deliberately separate from the demo
// simulator in cmd/simulator so the benchmark code can evolve without touching
// the didactic scenarios.
package experiment

import (
	"fmt"

	"github.com/lcmscheid/tactical-blockchain-reconciliation/internal/core"
	"github.com/lcmscheid/tactical-blockchain-reconciliation/internal/network"
)

// TacticalNetwork wraps a NetworkController pre-populated with the hierarchical
// topology promised by the research project: one Joint Forces Commander, three
// Parent Dispersed Units and five Dispersed Units, all sharing a common
// genesis block.
type TacticalNetwork struct {
	Controller *network.NetworkController
	Genesis    *core.Block
	JFC        *network.Node
	PDUs       []*network.Node
	DUs        []*network.Node
}

// NewTacticalNetwork builds the canonical 9-node tactical topology.
func NewTacticalNetwork() *TacticalNetwork {
	genesis := core.NewGenesisBlock()
	nc := network.NewNetworkController()

	jfc := network.NewNode("NODE-JFC-1", core.AuthJFC, genesis)
	nc.AddNode(jfc)

	pduNames := []string{"NODE-PDU-ALPHA", "NODE-PDU-BRAVO", "NODE-PDU-CHARLIE"}
	pdus := make([]*network.Node, 0, len(pduNames))
	for _, name := range pduNames {
		n := network.NewNode(name, core.AuthPDU, genesis)
		nc.AddNode(n)
		pdus = append(pdus, n)
	}

	dus := make([]*network.Node, 0, 5)
	for i := 1; i <= 5; i++ {
		n := network.NewNode(fmt.Sprintf("NODE-DU-%d", i), core.AuthDU, genesis)
		nc.AddNode(n)
		dus = append(dus, n)
	}

	return &TacticalNetwork{Controller: nc, Genesis: genesis, JFC: jfc, PDUs: pdus, DUs: dus}
}

// AllNodes returns the nodes in a stable order: JFC, PDUs, DUs. Stability
// matters: the harness relies on it to make trial execution reproducible.
func (t *TacticalNetwork) AllNodes() []*network.Node {
	out := make([]*network.Node, 0, 1+len(t.PDUs)+len(t.DUs))
	out = append(out, t.JFC)
	out = append(out, t.PDUs...)
	out = append(out, t.DUs...)
	return out
}

// PartitionInto distributes the nine nodes across the requested number of
// groups using a deterministic round-robin over the stable ordering returned
// by AllNodes. The caller is expected to pass groups >= 2; otherwise a single
// group containing every node is returned.
func (t *TacticalNetwork) PartitionInto(groups int) [][]*network.Node {
	nodes := t.AllNodes()
	if groups < 2 {
		return [][]*network.Node{nodes}
	}
	buckets := make([][]*network.Node, groups)
	for i, n := range nodes {
		buckets[i%groups] = append(buckets[i%groups], n)
	}
	return buckets
}

// ApplyPartition enforces the supplied grouping on the underlying controller:
// every pair of nodes in distinct groups has its link cut, every pair in the
// same group stays connected.
func (t *TacticalNetwork) ApplyPartition(groups [][]*network.Node) {
	// First ensure a clean mesh so that any residual partition from a previous
	// call is lifted before the new one is installed.
	t.Controller.HealNetwork()

	owner := make(map[string]int, 9)
	for idx, g := range groups {
		for _, n := range g {
			owner[n.ID] = idx
		}
	}
	for idA := range t.Controller.Nodes {
		for idB := range t.Controller.Nodes {
			if idA == idB {
				continue
			}
			if owner[idA] != owner[idB] {
				t.Controller.SetConnection(idA, idB, false)
			}
		}
	}
}

// Heal restores full connectivity across every node pair.
func (t *TacticalNetwork) Heal() {
	t.Controller.HealNetwork()
}
