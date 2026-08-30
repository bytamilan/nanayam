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

Check for updates later with:

```bash
nanayam upgrade --check
nanayam upgrade
```

Then run:
```bash
nanayam prerequisites --auto    # Install Docker, Fabric binaries, etc.
nanayam network up              # Start the basic Fabric network
nanayam channel create --name mychannel --profile TwoOrgsChannel
nanayam chaincode package --path ./chaincode/asset-transfer-basic --name basic
```

For the complaint workflow, use:

```bash
nanayam network up --profile complaint
```

`nanayam network up` now detects missing built-in Fabric crypto/channel artifacts and runs the matching setup script automatically:

- `fabric-network.yaml` / `apps.yaml` → `./scripts/setup-fabric.sh`
- `complaint-network.yaml` / `complaint-apps.yaml` → `./scripts/setup-complaint.sh`

When you use `--config`, Nanayam uses only the compose file you specify. Automatic recovery is available for the built-in network configs above; custom compose files must provide their own crypto and channel artifacts.

### One Script: Server Only

> Requires [Docker](https://docs.docker.com/get-docker/) and Docker Compose.
> Nothing else to install — no Go, no CLI, no Node.

For the *server* side only — the Fabric network and the gateway, no
operator console, no client apps — one command brings everything up, and
is safe to re-run (each stage is skipped if it's already done):

```bash
# macOS / Linux
./scripts/start-server.sh

# Windows (PowerShell or double-click)
.\scripts\start-server.ps1
```

```bash
./scripts/start-server.sh --down    # stop it, keep generated crypto/data
./scripts/start-server.sh --clean   # stop it and wipe crypto/channel data
```

The Windows script delegates to the same `start-server.sh` through WSL or
Git Bash's `bash.exe` (Docker Desktop on Windows runs Linux containers
through one of those anyway, so this isn't an extra dependency). See
[`docker container images`](#container-images) below if you'd rather run a
pre-built gateway image instead of building it locally.

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
nanayam network up                       # basic network, auto-recovers built-in artifacts if needed
nanayam network up --profile complaint   # complaint network, auto-recovers built-in artifacts if needed
nanayam network up --config docker/fabric-network.yaml
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

# Reinstall the current release explicitly
curl -fsSL https://raw.githubusercontent.com/bytamilan/nanayam/main/install.sh | bash -s -- --refresh

# Or build from source
make build
make install

# Or run the full local installer against the current checkout
make local

# Or package release archives locally
make release-assets
```

If a release archive is not available yet, `install.sh` now falls back to building from source.
It prefers the current checkout when available, and otherwise clones the repository temporarily.

### Network profiles and config files

```bash
# Built-in profiles
nanayam network up --profile basic
nanayam network up --profile complaint

# Explicit compose file (uses only this compose file)
nanayam network up --config docker/fabric-network.yaml
nanayam network up --config docker/complaint-network.yaml
```

Notes:

- Built-in networks can auto-generate missing certificates and channel artifacts during `network up`.
- Explicit `--config` does **not** auto-attach `apps.yaml` / `complaint-apps.yaml`; start app stacks separately when needed.
- Custom compose files are validated before startup, but Nanayam does not invent crypto for unknown topologies.

### Upgrade and Local Refresh

```bash
# Check whether a newer release exists
nanayam upgrade --check

# Download and install the latest release
nanayam upgrade

# Reinstall the current release
nanayam upgrade --refresh

# Build from a local checkout and refresh the installed CLI
nanayam upgrade --dev-local --refresh --source /path/to/nanayam
```

---

## Project Structure

```
nanayam/
├── apps/
│   └── org-console/           # Next.js web UI
├── flutter/                   # Melos-managed Flutter/Dart workspace
│   ├── packages/               # Reusable gateway client, models, UI kit, voucher domain
│   └── apps/voucher_wallet/    # Example app: voucher provisioning & usage
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
│   ├── start-distribution.sh  # Start gateway + console
│   ├── start-server.sh        # One-shot server-only start (Mac/Linux)
│   ├── start-server.ps1       # One-shot server-only start (Windows)
│   └── start-server.cmd       # CMD wrapper for start-server.ps1
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

## Container Images

`.github/workflows/release-containers.yml` builds and publishes the
gateway's container image to the GitHub Container Registry on every version
tag (`v*`) push, and on demand via `workflow_dispatch`:

```
ghcr.io/bytamilan/nanayam-gateway:latest
ghcr.io/bytamilan/nanayam-gateway:v1.2.3
ghcr.io/bytamilan/nanayam-gateway:sha-abcdef0
```

Images are built for `linux/amd64` and `linux/arm64`. Only the gateway is
published this way — it's the one server-side piece this repo owns and
builds; `docker/apps.yaml` (and `scripts/start-server.sh`) already build it
locally with the same Dockerfile, so publishing it separately just saves a
build the next time you (or CI, or a Kubernetes deployment) need the image
and don't want to build from source. The operator console and any Flutter
app are clients, not servers, and aren't published by this workflow.

```bash
docker pull ghcr.io/bytamilan/nanayam-gateway:latest
docker run -d --name nanayam-gateway --network nanayam \
  -p 8080:8080 -p 50051:50051 \
  -e FABRIC_CHANNEL=mychannel -e FABRIC_CHAINCODE=basic -e MSP_ID=Org1MSP \
  -v "$(pwd)/crypto-config/peerOrganizations/org1.nanayam.com:/app/crypto:ro" \
  ghcr.io/bytamilan/nanayam-gateway:latest
```

(A freshly published package on GHCR is private by default — an org/repo
admin makes it public once under the package's own Settings → Danger Zone
→ Change visibility, the same one-time step any GHCR image needs.)

---

## Flutter Voucher Example

[`flutter/`](flutter) is a [Melos](https://melos.invertase.dev/)-managed
Flutter/Dart monorepo demonstrating how to build a mobile client against the
gateway's REST API. It ships reusable packages first —
[`nanayam_ledger_models`](flutter/packages/nanayam_ledger_models),
[`nanayam_ledger_client`](flutter/packages/nanayam_ledger_client), and
[`nanayam_ui_kit`](flutter/packages/nanayam_ui_kit) have nothing to do with
vouchers and are meant for any future Nanayam Flutter app — plus one
voucher-specific domain package,
[`nanayam_voucher_core`](flutter/packages/nanayam_voucher_core), and an
example app, [`voucher_wallet`](flutter/apps/voucher_wallet), that provisions
CDC-voucher-style vouchers to citizens and lets businesses redeem them, with
every transaction recorded on the Nanayam sample ledger.

Every package and the app itself have their own test suites — unit tests
against hand-rolled fake gateways for the packages, widget tests driving the
real screens for the app — none of which need a running Fabric network:

```bash
cd flutter
dart pub global activate melos
melos bootstrap
melos run test
```

See [`docs/flutter-voucher-example.md`](docs/flutter-voucher-example.md#quickstart)
for a step-by-step guide covering the Fabric network, the gateway, and the
app together, and the rest of that document for the full design write-up.

---

## Documentation

Full documentation lives in the **[Nanayam Wiki](https://github.com/bytamilan/nanayam/wiki)** and on the **[documentation site](https://bytamilan.github.io/nanayam/)**, available in **English** and **தமிழ்**. The source is in [`docs/`](docs) and both surfaces publish automatically on every push to `main` via GitHub Actions ([`.github/workflows/deploy-docs.yml`](.github/workflows/deploy-docs.yml) for the Pages site, [`.github/workflows/publish-wiki.yml`](.github/workflows/publish-wiki.yml) for the wiki).

| | |
|---|---|
| [API Explorer](https://bytamilan.github.io/nanayam/api.html) | Interactive Swagger UI generated from [`docs/openapi.yaml`](docs/openapi.yaml) — try every REST endpoint against a local gateway |
| [Sample Application Guide](https://bytamilan.github.io/nanayam/sample-app-guide.html) | Build, configure, and use the Operator Console (`apps/org-console`) — the working sample client for the gateway API |
| [Flutter Voucher Example](https://bytamilan.github.io/nanayam/flutter-voucher-example.html) | The voucher provisioning & redemption app (`flutter/`) and the design write-up behind it |

| Page | English | தமிழ் |
|---|---|---|
| Getting Started | [EN](https://github.com/bytamilan/nanayam/wiki/Getting-Started) | [TA](https://github.com/bytamilan/nanayam/wiki/Getting-Started-ta) |
| Architecture | [EN](https://github.com/bytamilan/nanayam/wiki/Architecture) | [TA](https://github.com/bytamilan/nanayam/wiki/Architecture-ta) |
| CLI Reference | [EN](https://github.com/bytamilan/nanayam/wiki/CLI-Reference) | [TA](https://github.com/bytamilan/nanayam/wiki/CLI-Reference-ta) |
| Cloud Deployment | [EN](https://github.com/bytamilan/nanayam/wiki/Cloud-Deployment) | [TA](https://github.com/bytamilan/nanayam/wiki/Cloud-Deployment-ta) |
| API Reference | [EN](https://github.com/bytamilan/nanayam/wiki/API-Reference) | [TA](https://github.com/bytamilan/nanayam/wiki/API-Reference-ta) |
| Testing | [EN](https://github.com/bytamilan/nanayam/wiki/Testing) | [TA](https://github.com/bytamilan/nanayam/wiki/Testing-ta) |
| Troubleshooting | [EN](https://github.com/bytamilan/nanayam/wiki/Troubleshooting) | [TA](https://github.com/bytamilan/nanayam/wiki/Troubleshooting-ta) |
| Contributing | [EN](https://github.com/bytamilan/nanayam/wiki/Contributing) | [TA](https://github.com/bytamilan/nanayam/wiki/Contributing-ta) |

Reference material also lives alongside the code:

| Document | Description |
|----------|-------------|
| [`docs/local-setup-guide.md`](docs/local-setup-guide.md) | Complete local development setup |
| [`docs/hyperledger-fabric-guide.md`](docs/hyperledger-fabric-guide.md) | Hyperledger Fabric concepts & internals |
| [`docs/nanayam-architecture.md`](docs/nanayam-architecture.md) | How Nanayam components interact |
| [`docs/complaint-system.md`](docs/complaint-system.md) | The grievance workflow in detail |
| [`docs/flutter-voucher-example.md`](docs/flutter-voucher-example.md) | Flutter voucher provisioning & usage example: packages, app, and ledger design |

---

## Stopping the Stack

```bash
# If you started with ./scripts/start-server.sh
./scripts/start-server.sh --down     # Stop the server stack (preserves data)
./scripts/start-server.sh --clean    # Stop it and wipe all data

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

## Cloud Deployment

One command deploys the gateway and console to any Kubernetes cluster your `kubectl` points at — GKE, EKS, AKS, k3d, kind, or self-managed:

```bash
./scripts/deploy-cloud.sh --registry ghcr.io/bytamilan --domain nanayam.example.com
```

It builds and pushes both images, uploads the Fabric crypto material and a generated JWT secret, applies the manifests in [`k8s/`](k8s), and waits for the rollout.

```bash
./scripts/deploy-cloud.sh --help                      # every flag
./scripts/deploy-cloud.sh --registry … --dry-run      # preview, no cluster needed
./scripts/deploy-cloud.sh --registry … --profile complaint
./scripts/deploy-cloud.sh --destroy                   # tear it all down
```

Full guide: [Cloud Deployment](https://github.com/bytamilan/nanayam/wiki/Cloud-Deployment) ([தமிழ்](https://github.com/bytamilan/nanayam/wiki/Cloud-Deployment-ta)).

Also available:

- `.github/workflows/ci.yml` — tests, linting, and manifest validation on every push
- `.github/workflows/deploy-fabric-operator.yml` — CD to GKE via Workload Identity Federation
- `scripts/setup.sh` — k3d local cluster with the HLF Operator, Istio, and Helm

### Testing

```bash
make test        # CLI, gateway, and console suites
make validate    # formatting, linting, build, and tests
```

See [Testing](https://github.com/bytamilan/nanayam/wiki/Testing) ([தமிழ்](https://github.com/bytamilan/nanayam/wiki/Testing-ta)).

---
## Notebook LLM
[link to learn HLF](https://notebooklm.google.com/notebook/915d332c-3f08-430f-98cd-0f393116c54a)
---

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
