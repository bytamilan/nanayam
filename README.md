# Nanayam – A Decentralized Private Web3 Ledger

![CodeRabbit Pull Request Reviews](https://img.shields.io/coderabbit/prs/github/bytamilan/nanayam?utm_source=oss&utm_medium=github&utm_campaign=bytamilan%2Fnanayam&labelColor=171717&color=FF570A&link=https%3A%2F%2Fcoderabbit.ai&label=CodeRabbit+Reviews)

Nanayam is a **private, permissioned Web3 ledger** built on [Hyperledger Fabric](https://www.hyperledger.org/use/fabric). It provides a complete local development stack including a Fabric test network, a Go-based distribution gateway, and a Next.js operator console.

---

## Architecture

```
┌─────────────────┐      REST       ┌──────────────────┐      gRPC-TLS      ┌─────────────┐
│ Operator Console│ ◀─────────────▶ │ Distribution     │ ◀───────────────▶ │ Fabric Peer │
│ (Next.js 3000)  │                 │ Server (Go 8080) │                  │ (Org1)      |
└─────────────────┘                 └──────────────────┘                  └──────┬──────┘
                                                                                  │
                                                                         Order / Deliver
                                                                                  │
                                                                           ┌──────┴──────┐
                                                                           │  Orderer    │
                                                                           │  (Raft)     │
                                                                           └─────────────┘
```

- **Fabric Network** — Docker Compose stack with Orderer, 2 Peers, 2 CAs, and a CLI.
- **Distribution Server** — Go gRPC/REST gateway connecting to Fabric via the Gateway SDK.
- **Operator Console** — Next.js web UI for visualizing and managing ledger assets.

---

## Quick Start

### One-Liner Install (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/bytamilan/nanayam/main/install.sh | bash
```

Then run:
```bash
nanayam prerequisites --auto    # Install Docker, Fabric binaries, etc.
nanayam network up              # Start the full Fabric network
nanayam channel create --name mychannel --profile TwoOrgsChannel
nanayam chaincode package --path ./chaincode/asset-transfer-basic --name basic
```

### Classic Script Setup

> Requires [Docker](https://docs.docker.com/get-docker/) and Docker Compose.

```bash
# 1. Setup: pull images, download binaries, generate crypto
./scripts/setup-fabric.sh

# 2. Start the Fabric network (orderer, peers, channel)
./scripts/start-fabric.sh

# 3. Deploy the sample chaincode
./scripts/deploy-chaincode.sh

# 4. Start the gateway + console
./scripts/start-distribution.sh
```

Open **http://localhost:3000** in your browser.

For detailed instructions, see [`docs/local-setup-guide.md`](docs/local-setup-guide.md).

---

## Nanayam CLI

Nanayam includes a unified CLI for managing the entire Fabric stack:

```bash
# Install prerequisites and Fabric binaries
nanayam prerequisites --auto

# Node lifecycle
nanayam node init --type peer --org Org1
nanayam node start peer0.org1.nanayam.com
nanayam node stop peer0.org1.nanayam.com
nanayam node status

# Network orchestration
nanayam network up --profile basic       # or --profile complaint
nanayam network down
nanayam network clean                    # Wipe all data

# Crypto & channels
nanayam crypto generate
nanayam channel create --name mychannel --profile TwoOrgsChannel
nanayam channel join --name mychannel
nanayam channel update-anchor --name mychannel --org Org1MSP

# Chaincode lifecycle
nanayam chaincode package --path ./chaincode/asset-transfer-basic --name basic
nanayam chaincode install --package basic.tar.gz
nanayam chaincode approve --name basic --channel mychannel --package-id basic_1.0:...
nanayam chaincode commit --name basic --channel mychannel
nanayam chaincode invoke --channel mychannel --name basic --function InitLedger
nanayam chaincode query --channel mychannel --name basic --function GetAllAssets

# User identity management
nanayam user create --id alice --secret alicepw --type client --org Org1
nanayam user enroll --id alice --secret alicepw --org Org1

# Consortium connectivity
nanayam consortium connect --orderer orderer.example.com:7050 --tls-cert ./tls.crt --org NewOrg --domain neworg.example.com
nanayam consortium join-channel --name mychannel --block ./mychannel.block

# Application services
nanayam gateway                          # Start Go REST/gRPC gateway
nanayam console                          # Start Next.js operator console
```

### Install the CLI

```bash
# macOS & Linux
curl -fsSL https://raw.githubusercontent.com/bytamilan/nanayam/main/install.sh | bash

# With Fabric binaries + prerequisites
curl -fsSL https://raw.githubusercontent.com/bytamilan/nanayam/main/install.sh | bash -s -- --with-fabric --setup

# Or build from source
make build
make install
```

---

## Project Structure

```
nanayam/
├── apps/
│   └── operator-console/      # Next.js web UI
├── cli/                       # Nanayam CLI (Go + Cobra)
│   ├── cmd/                   # Command implementations
│   ├── internal/              # Internal packages
│   ├── templates/             # Embedded Docker/Config templates
│   └── main.go                # CLI entry point
├── config/
│   ├── crypto-config.yaml     # Cryptogen configuration
│   ├── configtx.yaml          # Channel configuration
│   └── connection-profile.yaml# SDK connection profile
├── docker/
│   ├── fabric-network.yaml    # Fabric infrastructure (Compose)
│   └── apps.yaml              # Gateway + Console (Compose)
├── docs/
│   ├── local-setup-guide.md   # Step-by-step setup
│   ├── hyperledger-fabric-guide.md  # Fabric concepts
│   └── nanayam-architecture.md      # Component architecture
├── proto/
│   └── fabric.proto           # gRPC service definitions
├── scripts/
│   ├── setup-fabric.sh        # Pull images, generate crypto
│   ├── start-fabric.sh        # Start Fabric network
│   ├── stop-fabric.sh         # Stop Fabric network
│   ├── deploy-chaincode.sh    # Deploy asset-transfer chaincode
│   └── start-distribution.sh  # Start gateway + console
├── services/
│   └── gateway/               # Go gRPC/REST distribution server
├── install.sh                 # One-liner installer (macOS/Linux)
├── install.ps1                # PowerShell installer (Windows)
├── install.cmd                # CMD installer (Windows)
├── Makefile                   # Build system
└── README.md
```

---

## API Endpoints

The distribution server exposes both **gRPC** and **REST** APIs:

| Method | gRPC | REST | Description |
|--------|------|------|-------------|
| Create | `CreateAsset` | `POST /v1/CreateAsset` | Submit a new asset |
| Read | `QueryAsset` | `GET /v1/QueryAsset?assetId=` | Read a single asset |
| List | `ListAssets` | `GET /v1/ListAssets` | List all asset IDs |

---

## Documentation

| Document | Description |
|----------|-------------|
| [`docs/local-setup-guide.md`](docs/local-setup-guide.md) | Complete local development setup |
| [`docs/hyperledger-fabric-guide.md`](docs/hyperledger-fabric-guide.md) | Hyperledger Fabric concepts & internals |
| [`docs/nanayam-architecture.md`](docs/nanayam-architecture.md) | How Nanayam components interact |

---

## Stopping the Stack

```bash
# Using the CLI
nanayam network down          # Stop Fabric network
nanayam network clean         # Stop and wipe all data

# Or using Docker Compose directly
# Stop apps only (gateway + console)
docker-compose -f docker/apps.yaml down

# Stop Fabric network (preserves data)
./scripts/stop-fabric.sh

# Stop everything and wipe all data
./scripts/stop-fabric.sh --clean
```

---

## Production Deployment

For Kubernetes / GKE deployments, the project includes a Fabric Operator workflow:

- `.github/workflows/deploy-fabric-operator.yml` — CI/CD for GKE
- `scripts/setup.sh` — k3d local cluster with HLF Operator, Istio, and Helm

See the existing operator setup for production-grade orchestration.

---
## Notebook LLM
[link to learn HLF](https://notebooklm.google.com/notebook/915d332c-3f08-430f-98cd-0f393116c54a)
---

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
