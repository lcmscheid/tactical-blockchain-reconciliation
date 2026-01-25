/*
 * Copyright (c) 2026 Luiz Carlos Martins Scheid
 * MBA em Engenharia de Software - USP/Esalq
 *
 * Project: Algoritmo de Reconciliação de Bifurcações em Blockchain para Sistemas de Comando Tático
 */

package crypto

import (
	"testing"
)

func TestCalculateHash_Determinism(t *testing.T) {
	data := "Mission Critical Payload"

	// Calculate hash twice for the same data
	hash1, err1 := CalculateHash(data)
	hash2, err2 := CalculateHash(data)

	if err1 != nil || err2 != nil {
		t.Fatalf("Hashing failed: %v", err1)
	}

	if hash1 != hash2 {
		t.Errorf("Hash is not deterministic. \nGot A: %s\nGot B: %s", hash1, hash2)
	}
}

func TestCalculateHash_Distinctness(t *testing.T) {
	hash1, _ := CalculateHash("Order A")
	hash2, _ := CalculateHash("Order B")

	if hash1 == hash2 {
		t.Error("Collision detected! Different inputs produced same hash.")
	}
}
