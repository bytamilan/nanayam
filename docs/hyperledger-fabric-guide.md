---
layout: default
title: Hyperledger Fabric — A Practical Guide
---

# Hyperledger Fabric — A Practical Guide

> This document explains what Hyperledger Fabric is, how its core components work, and why Nanayam uses it as its ledger backbone.

---

## 1. What is Hyperledger Fabric?

**Hyperledger Fabric** is an open-source enterprise-grade permissioned blockchain platform. Unlike public blockchains (e.g., Bitcoin, Ethereum), Fabric is designed for private networks where participants are known and trusted — making it ideal for consortiums, supply chains, finance, and internal enterprise ledgers.

### Key Characteristics

| Feature | Description |
|---------|-------------|
| **Permissioned** | Only enrolled members can join the network. |
| **Modular** | Pluggable consensus, membership services, and data stores. |
| **Private Channels** | Subsets of members can transact privately on isolated channels. |
| **No Native Token** | No cryptocurrency is required to operate the network. |
| **Chaincode (Smart Contracts)** | Business logic written in Go, Java, JavaScript, or TypeScript. |
| **High Throughput** | Designed for thousands of transactions per second. |

---

## 2. Core Concepts

### 2.1 Membership Service Provider (MSP)

The **MSP** defines the rules under which identities are validated, authenticated, and allowed to participate in the network. Each organization has its own MSP (e.g., `Org1MSP`, `Org2MSP`).

An MSP contains:
- **Root CAs** — Trust anchors for certificates.
- **Intermediate CAs** — Optional delegation of certificate issuance.
- **Admin Certs** — Identities with administrative privileges.
- **Revocation Lists** — Certificates that have been revoked.

### 2.2 Certificate Authority (CA)

The **Fabric CA** is the default certificate authority for enrolling identities (users, peers, orderers). It issues X.509 certificates that prove membership in an organization.

In a typical network:
- Each organization runs its own CA.
- The CA registers users with roles: `peer`, `orderer`, `admin`, `client`.
- Enrolled identities receive a certificate + private key pair.

### 2.3 Peers

**Peers** are the nodes that host ledgers and smart contracts (chaincode). There are two types:

| Type | Role |
|------|------|
| **Committing Peer** | Maintains a copy of the ledger, validates and commits blocks. |
| **Endorsing Peer** | Simulates transactions (chaincode execution) and returns an endorsement. |

Peers communicate via **gossip** to disseminate blocks and keep ledger copies in sync.

### 2.4 Orderer

The **Orderer** (or Ordering Service Node) is responsible for:
- Receiving endorsed transactions from clients.
- Ordering transactions into a strict sequence.
- Packaging them into **blocks**.
- Distributing blocks to peers for validation and commit.

Fabric supports multiple consensus mechanisms. The default is **Raft** (a crash-fault-tolerant protocol) which runs inside the orderer nodes themselves.

### 2.5 Channels

A **channel** is a private subnet of communication between specific organizations. Each channel has:
- Its own ledger (independent of other channels).
- Its own chaincode instances.
- Its own set of policies.

This allows competing or private business interests to coexist on the same physical infrastructure without sharing data.

### 2.6 Chaincode (Smart Contracts)

**Chaincode** is the business logic that reads from and writes to the ledger. It runs inside a Docker container on the peer.

Lifecycle (Fabric 2.x):
1. **Package** — Bundle chaincode source into a `.tar.gz`.
2. **Install** — Install the package on target peers.
3. **Approve** — Organizations approve the chaincode definition.
4. **Commit** — Once enough organizations approve, the definition is committed to the channel.
5. **Invoke** — Clients submit transactions to the committed chaincode.

### 2.7 Ledger

The Fabric ledger consists of two parts:

1. **Blockchain** — An immutable, ordered log of all transactions (blocks).
2. **World State** — A database (LevelDB or CouchDB) that holds the current value of all keys. It is derived from the blockchain but optimized for fast queries.

### 2.8 Transaction Flow (Endorse → Order → Validate → Commit)

```
┌─────────┐     1. Submit        ┌─────────┐     2. Simulate
│ Client  │ ───────────────────▶ │ Endorser│ ───────────────────┐
│ (SDK)   │                      │  Peer   │                    │
└─────────┘                      └─────────┘                    │
     ▲                              │                           │
     │                              ▼                           │
     │                        3. Proposal Response               │
     │                        (RW set + signature)               │
     │                              │                           │
     │ 4. Send to Orderer           │                           │
     │                              │                           │
     │    ┌─────────┐               │                           │
     └─── │ Orderer │ ◀─────────────┘                           │
          │  (Raft) │                                           │
          └────┬────┘                                           │
               │                                                │
               ▼ 5. Order into block                            │
          ┌─────────┐                                           │
          │  Block  │ ────────────────────────────────────────▶ │
          └────┬────┘                                           │
               │                                                ▼
               ▼ 6. Distribute                              ┌─────────┐
          ┌─────────┐                                       │ Commit  │
          │  Peers  │ ────────────────────────────────────▶ │  Peer   │
          └─────────┘                                       └─────────┘
                                                              7. Validate
                                                              8. Commit to ledger
```

1. **Proposal** — Client sends a transaction proposal to endorsing peers.
2. **Simulate** — Peers execute chaincode against the world state (no writes yet).
3. **Endorse** — Peers return a signed **Read-Write (RW) set**.
4. **Submit** — Client sends the endorsed transaction to the orderer.
5. **Order** — Orderer batches transactions into a block.
6. **Distribute** — Block is sent to all channel peers.
7. **Validate** — Peers check endorsement policies and MVCC (multi-version concurrency control).
8. **Commit** — Valid transactions update the world state and are appended to the blockchain.

---

## 3. Why Nanayam Uses Fabric

Nanayam is a **private Web3 ledger** — it needs:

- **Data privacy** — Channels isolate sensitive asset data between participants.
- **No gas fees** — No cryptocurrency is needed for transactions.
- **Enterprise governance** — MSPs and policies define who can do what.
- **Proven immutability** — The blockchain provides an auditable, tamper-evident history.
- **Pluggable architecture** — Easy to swap consensus, database, or crypto backends.

Fabric’s permissioned model aligns perfectly with Nanayam’s goal of being a decentralized ledger for trusted consortia rather than a fully public blockchain.

---

## 4. Further Reading

- [Hyperledger Fabric Official Docs](https://hyperledger-fabric.readthedocs.io/)
- [Fabric Gateway Client SDK (Go)](https://pkg.go.dev/github.com/hyperledger/fabric-gateway)
- [Fabric CA Operations Guide](https://hyperledger-fabric-ca.readthedocs.io/)
- [Fabric Samples (test-network)](https://github.com/hyperledger/fabric-samples)
