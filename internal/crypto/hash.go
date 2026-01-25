/*
 * Copyright (c) 2026 Luiz Carlos Martins Scheid
 * MBA em Engenharia de Software - USP/Esalq
 *
 * Project: Algoritmo de Reconciliação de Bifurcações em Blockchain para Sistemas de Comando Tático
 */

// Package crypto contains cryptographic helpers (SHA-256 hashing)
package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// CalculateHash computes the SHA-256 hash of any interface (typically a Block or Transaction).
// It serializes the object to JSON first to ensure deterministic output.
func CalculateHash(o interface{}) (string, error) {
	// Serialize the interface to JSON bytes
	bytes, err := json.Marshal(o)
	if err != nil {
		return "", fmt.Errorf("failed to serialize object of hashing: %w", err)
	}

	h := sha256.New()
	h.Write(bytes)
	hashed := h.Sum(nil)

	return hex.EncodeToString(hashed), nil
}
