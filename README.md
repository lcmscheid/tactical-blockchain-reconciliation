# Tactical Blockchain Reconciliation (TCC Project)

**Algoritmo de Reconciliação de Bifurcações em Blockchain para Sistemas de Comando Tático**

> **Note:** This repository contains the source code for the Master's Thesis (TCC) submitted to MBA USP/ESALQ.

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![Go Version](https://img.shields.io/badge/go-1.25%2B-blue)
![Status](https://img.shields.io/badge/status-Prototype%20Implementation-orange)

## 📄 Abstract
This project implements a context-aware blockchain reconciliation algorithm designed for decentralized Command and Control (C2) systems. Unlike traditional consensus mechanisms that discard shorter chains or require a majority quorum, this solution merges divergent blockchain branches created during network partitions.

**Key Distinction:** This is a **discrete-event simulation**. It models the behavior of a blockchain network under programmable partition conditions to validate the reconciliation logic without the overhead of physical network infrastructure.

The goal is to preserve the integrity of tactical decisions made by disconnected units, ensuring that "operational intentions" are fused rather than overwritten.

## 🎯 Objectives
* **Primary:** Develop and validate a reconciliation algorithm capable of merging multiple blockchain branches while preserving tactical consistency.
* **Secondary:**
    1. Implement a custom blockchain simulator in Go to emulate hierarchical structures and network partitions.
    2. Evaluate three merging strategies: Priority-Authority, Temporal-Priority, and Hybrid.
    3. Measure performance metrics such as reconciliation time (target < 10s) and data loss.

## 🚦 Current Status: Preliminary Results Phase
The project is currently in the **Implementation Phase** (Sprint 1), focusing on:
- [x] Definition of Core Data Structures (Block, Transaction, Context Enums).
- [ ] Implementation of Cryptographic Helpers (SHA-256 Hashing).
- [ ] Development of the Network Partition Engine.
- [ ] Prototyping the "Priority-Authority" Reconciliation Strategy.

## 🧪 Simulation Scenarios
The algorithm will be validated against three complexity levels of network partitioning:
1.  **Single Partition:** Two isolated groups operating independently.
2.  **Multi-Partition:** Three or more groups operating simultaneously.
3.  **Cascading Partitions:** Fragmentation occurring within an already isolated group.

## 🛠 Technology Stack
* **Language:** Go (v1.25+) 
    * *Chosen for concurrency support and alignment with Hyperledger Fabric ecosystem.*
* **Architecture:** Custom simulator mimicking a hierarchical military network (9 nodes: 1 JFC, 3 PDU, 5 DU).
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
