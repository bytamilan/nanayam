# Nanayam Architecture

> This document describes how the Nanayam components interact to form a complete decentralized private ledger stack.

---

## High-Level Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         User / Browser                                   │
└─────────────────────────────────┬───────────────────────────────────────┘
                                  │ HTTP
                                  ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  ┌─────────────────────┐                                               │
│  │  Operator Console   │  Next.js 15 + React 19                        │
│  │  (Port 3000)        │  - Asset table view                           │
│  │                     │  - Channel management                         │
│  │                     │  - Identity login via Fabric CA               │
│  └──────────┬──────────┘                                               │
│             │ REST JSON                                                 │
│             ▼                                                           │
│  ┌─────────────────────┐                                               │
│  │  Distribution       │  Go 1.21 — gRPC + REST Gateway                │
│  │  Server             │  - Connects to Fabric via Gateway SDK         │
│  │  (Port 50051/8080)  │  - Translates REST ↔ Fabric transactions      │
│  └──────────┬──────────┘                                               │
│             │ gRPC-TLS                                                  │
│             ▼                                                           │
│  ┌─────────────────────┐    ┌─────────────────────┐                   │
│  │  Peer (Org1)        │◀──▶│  Peer (Org2)        │                   │
│  │  peer0.org1         │    │  peer0.org2         │  Gossip           │
│  │  :7051              │    │  :9051              │  sync             │
│  └──────────┬──────────┘    └─────────────────────┘                   │
│             │                                                           │
│             │ Order / Deliver                                           │
│             ▼                                                           │
│  ┌─────────────────────┐                                               │
│  │  Orderer (Raft)     │  :7050                                         │
│  │  orderer.nanayam    │  Sequences transactions into blocks           │
│  └─────────────────────┘                                               │
│                                                                         │
│  ┌─────────────────────┐    ┌─────────────────────┐                   │
│  │  CA (Org1)          │    │  CA (Org2)          │                   │
│  │  :7054              │    │  :8054              │  Issue certificates │
│  └─────────────────────┘    └─────────────────────┘                   │
│                                                                         │
│  Shared Docker Network:  nanayam                                        │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Component Breakdown

### 1. Operator Console (`apps/org-console/`)

A Next.js 15 web application that provides a visual interface for interacting with the ledger.

**Key responsibilities:**
- Display asset tables and channel status.
- Authenticate users via Fabric CA enrollment.
- Proxy API calls to the Go distribution server.

**Tech stack:**
- Next.js 15 (App Router)
- React 19 + Tailwind CSS
- `fabric-network` & `fabric-ca-client` (Node.js SDK)

**Key files:**
- `src/app/page.tsx` — Dashboard with `AssetTable` and `ChannelTable`.
- `src/lib/api.ts` — REST client for gateway endpoints.
- `src/lib/bevelHelper.ts` — CA enrollment helper.

---

### 2. Distribution Server (`services/gateway/`)

The **distribution server** is the central API layer that bridges the web UI (and any other client) with the Hyperledger Fabric network.

**Key responsibilities:**
- Maintain a persistent connection to a Fabric peer via the **Fabric Gateway SDK**.
- Expose a **gRPC API** for strongly-typed internal communication.
- Expose a **REST API** for browser and third-party integrations.
- Translate high-level asset operations into Fabric chaincode invocations.

**Tech stack:**
- Go 1.21
- `hyperledger/fabric-gateway` (Go SDK)
- `google.golang.org/grpc`

**Key files:**
- `connection.go` — Loads X.509 identity, private key, and TLS certs; establishes Gateway connection.
- `handler.go` — gRPC service implementation (`CreateAsset`, `QueryAsset`, `ListAssets`).
- `http.go` — Simple HTTP mux that wraps gRPC methods into REST endpoints.
- `main.go` — Bootstraps both gRPC and HTTP servers, handles graceful shutdown.

**Endpoints:**

| Protocol | Endpoint | Description |
|----------|----------|-------------|
| gRPC | `CreateAsset` | Submit a new asset to the ledger. |
| gRPC | `QueryAsset` | Evaluate (read-only) a single asset. |
| gRPC | `ListAssets` | Evaluate all assets, return IDs. |
| REST | `POST /v1/CreateAsset` | JSON body → chaincode submit. |
| REST | `GET /v1/QueryAsset?assetId=...` | Query single asset. |
| REST | `GET /v1/ListAssets` | List all asset IDs. |
| REST | `GET /health` | Health check. |

---

### 3. Fabric Network (`docker/fabric-network.yaml`)

A standard **Hyperledger Fabric test network** running in Docker Compose.

**Topology:**
- **1 Orderer** (`orderer.nanayam.com`) — Raft consensus, port `7050`.
- **2 Peer Organizations** — `Org1MSP` and `Org2MSP`.
- **2 Peers** — `peer0.org1.nanayam.com:7051`, `peer0.org2.nanayam.com:9051`.
- **2 CAs** — `ca_org1:7054`, `ca_org2:8054`.
- **1 CLI** container — For ad-hoc channel and chaincode operations.

**Channel:**
- Name: `mychannel`
- Members: Org1, Org2
- Chaincode: `basic` (asset-transfer-basic from Fabric Samples)

---

## Data Flow: Creating an Asset

```
1. User fills form in Operator Console
        │
        ▼ HTTP POST /api/create-asset
2. Next.js API route proxies to Go Gateway
        │
        ▼ HTTP POST /v1/CreateAsset
3. Go Gateway unmarshals JSON → gRPC CreateAssetRequest
        │
        ▼ Fabric Gateway SDK → gRPC-TLS → Peer
4. Peer simulates chaincode (asset-transfer-basic/CreateAsset)
   - Generates Read-Write set
   - Returns endorsed proposal response
        │
        ▼ Gateway SDK submits to Orderer
5. Orderer sequences transaction into a block
        │
        ▼ Block broadcast to all channel peers
6. Peers validate endorsement + MVCC, commit to ledger
        │
        ▼ Response flows back through Gateway → REST → Console
7. UI shows success / updated asset list
```

---

## Configuration Files

| File | Purpose |
|------|---------|
| `config/crypto-config.yaml` | Defines orgs, peers, orderers for `cryptogen`. |
| `config/configtx.yaml` | Channel profiles, policies, capabilities for `configtxgen`. |
| `config/connection-profile.yaml` | Client SDK discovery config (peers, orderers, CAs). |
| `docker/fabric-network.yaml` | Docker Compose for Fabric infrastructure. |
| `docker/apps.yaml` | Docker Compose for Gateway + Console. |

---

## Security Model

1. **TLS Everywhere** — All peer, orderer, and CA communications use mutual TLS.
2. **MSP Enrollment** — Only identities enrolled by the Org CA can submit transactions.
3. **Channel Isolation** — Data is only visible to members of `mychannel`.
4. **Endorsement Policy** — The `basic` chaincode requires endorsement from both Org1 and Org2 (default for 2-org setup).

---

## Deployment Patterns

### Local Development (Docker Compose)
Use the scripts in `scripts/` to spin up the full stack on a single machine. See `docs/local-setup-guide.md`.

### Production (Kubernetes + Fabric Operator)
The existing `scripts/setup.sh` and `.github/workflows/deploy-fabric-operator.yml` target a **k3d** local cluster or **GKE** with the Hyperledger Labs Fabric Operator. This pattern is kept for production deployments but is not required for local development.
