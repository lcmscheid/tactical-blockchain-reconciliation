# Tactical Blockchain Reconciliation (TCC Project)

**Algoritmo de Reconciliação de Bifurcações em Blockchain para Sistemas de Comando Tático**

> **Note:** This repository contains the source code for the Master's Thesis (TCC) submitted to MBAUSP ESALO.

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![Go Version](https://img.shields.io/badge/go-1.25%2B-blue)
![Status](https://img.shields.io/badge/status-Research%20Prototype-orange)

## 📄 Abstract
This project implements a context-aware blockchain reconciliation algorithm designed for decentralized Command and Control (C2) systems. Unlike traditional consensus mechanisms that discard shorter chains (e.g., Nakamoto Consensus) or require a majority quorum (BFT), this solution merges divergent blockchain branches created during network partitions.

The goal is to preserve the integrity of tactical decisions made by disconnected units, ensuring that "operational intentions" are fused rather than overwritten.

## 🎯 Objectives
* **Primary:** Develop and validate a reconciliation algorithm capable of merging multiple blockchain branches while preserving tactical consistency.
* **Secondary:**
    1. Implement a custom blockchain simulator in Go.
    2. Evaluate three merging strategies: Priority-Authority, Temporal-Priority, and Hybrid.
    3. Measure performance metrics such as reconciliation time and data loss.

## 🛠 Technology Stack
* **Language:** Go (v1.25+) 
    * *Chosen for concurrency support and alignment with Hyperledger Fabric ecosystem.*
* **Architecture:** Custom simulator mimicking a hierarchical military network (JFC, PDU, DU).
* **Crypto:** SHA-256 for block hashing.
* **Analysis:** Python 3.13+ (Pandas, Scipy) used for post-experiment statistical analysis.

## 🧩 Context-Aware Reconciliation
The core algorithm resolves conflicts based on four operational dimensions:

1.  **Hierarchy of Authority:** JFC > PDU > DU.
2.  **Temporal Precedence:** Transaction timestamps.
3.  **Mission Priority:** Critical, High, Medium, Low.
4.  **Operational Dependencies:** A graph of related tactical decisions.

### Merging Strategies
The system supports three configurable strategies:
* **Priority-Authority:** Favors higher command levels (Chain of Command focus).
* **Temporal-Priority:** Favors urgency and recent timestamps.
* **Hybrid:** A weighted approach combining all context factors.
