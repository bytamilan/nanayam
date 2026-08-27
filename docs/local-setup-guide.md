# Local Setup Guide

> Get the complete Nanayam stack running on your local machine with Docker Compose.

---

## Prerequisites

| Tool | Minimum Version | Install Link |
|------|-----------------|--------------|
| Docker | 24.x | https://docs.docker.com/get-docker/ |
| Docker Compose | 2.x | Included with Docker Desktop |
| curl | any | Pre-installed on most systems |
| bash | 4.x+ | Pre-installed on macOS/Linux |

**macOS users:** Docker Desktop includes everything you need.

**Linux users:** Ensure your user is in the `docker` group:
```bash
sudo usermod -aG docker $USER
```

---

## Quick Start (One-Liner Style)

### CLI-first workflow

```bash
# 1. Clone (if you haven't already)
git clone https://github.com/bytamilan/nanayam.git
cd nanayam

# 2. Install prerequisites and Fabric binaries
nanayam prerequisites --auto

# 3. Start the built-in basic network
nanayam network up

# 4. Or start the complaint network
nanayam network up --profile complaint

# 5. Or use an explicit compose file
nanayam network up --config docker/fabric-network.yaml
```

For the built-in basic and complaint networks, `nanayam network up` automatically runs the matching setup script if required certificates or channel artifacts are missing. For custom compose files, you must generate those artifacts yourself.

### Classic script workflow

```bash
# 1. Clone (if you haven't already)
git clone https://github.com/bytamilan/nanayam.git
cd nanayam

# 2. Pull Fabric binaries, images, and generate crypto
./scripts/setup-fabric.sh

# 3. Start the Fabric network (orderer, peers, CA, channel)
./scripts/start-fabric.sh

# 4. Deploy the asset-transfer chaincode
./scripts/deploy-chaincode.sh

# 5. Start the distribution server + operator console
./scripts/start-distribution.sh
```

After step 5, open your browser: **http://localhost:3000**

---

## Step-by-Step Explanation

### Step 1 — `setup-fabric.sh`

This script does the heavy lifting of downloading and preparing everything:

- **Pulls Docker images:** `fabric-peer`, `fabric-orderer`, `fabric-tools`, `fabric-ca`.
- **Downloads binaries:** `cryptogen`, `configtxgen`, `peer` into `./bin/`.
- **Generates crypto material:** Uses `cryptogen` to create MSP certificates for Org1, Org2, and the Orderer.
- **Creates channel artifacts:** Uses `configtxgen` to create:
  - `genesis.block` — The orderer's initial configuration.
  - `channel.tx` — Transaction to create `mychannel`.
  - `Org1MSPanchors.tx` & `Org2MSPanchors.tx` — Anchor peer updates.

**Output directories created:**
```
./bin/              # Fabric CLI binaries
./crypto-config/    # X.509 certificates and keys
./channel-artifacts/# Genesis block and channel transactions
```

> ⏱️ This step takes ~3-5 minutes depending on your internet speed.

---

### Step 2 — `start-fabric.sh`

This script starts the actual Fabric infrastructure:

1. **Launches containers** via `docker-compose -f docker/fabric-network.yaml up -d`:
   - `ca_org1` — Certificate Authority for Org1
   - `ca_org2` — Certificate Authority for Org2
   - `orderer.nanayam.com` — Raft orderer
   - `peer0.org1.nanayam.com` — Peer for Org1
   - `peer0.org2.nanayam.com` — Peer for Org2
   - `cli` — Administrative CLI container

2. **Creates the channel** `mychannel` using the `cli` container.

3. **Joins peers** to `mychannel`.

4. **Updates anchor peers** for both organizations.

CLI equivalent:

```bash
nanayam network up
# or
nanayam network up --config docker/fabric-network.yaml
```

If the built-in Fabric certificates or channel artifacts are missing, the CLI will run `./scripts/setup-fabric.sh` automatically before starting the containers.

**Verify it's running:**
```bash
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
```

You should see all 6 containers in `Up` status.

---

### Step 3 — `deploy-chaincode.sh`

This script deploys the **asset-transfer-basic** chaincode (from Hyperledger Fabric Samples) to `mychannel`.

Lifecycle steps automated:
1. **Downloads** the chaincode source if not present.
2. **Packages** it into `basic.tar.gz`.
3. **Installs** the package on `peer0.org1` and `peer0.org2`.
4. **Queries** installed packages to get the `package ID`.
5. **Approves** the chaincode definition for both Org1 and Org2.
6. **Commits** the definition to the channel.
7. **Initializes** the ledger with sample assets (`InitLedger`).

**Verify chaincode:**
```bash
docker exec cli peer chaincode query -C mychannel -n basic -c '{"Args":["GetAllAssets"]}'
```

You should see a JSON array of sample assets.

---

### Step 4 — `start-distribution.sh`

This script builds and starts the application layer:

1. **Builds** the Go gateway Docker image (`services/gateway/Dockerfile`).
2. **Builds** the Next.js console Docker image (`apps/org-console/Dockerfile`).
3. **Starts** both services on the shared `nanayam` Docker network.

**Services available:**

| Service | URL | Protocol |
|---------|-----|----------|
| Operator Console | http://localhost:3000 | HTTP |
| REST Gateway | http://localhost:8080 | HTTP |
| gRPC Gateway | grpc://localhost:50051 | gRPC |

**Test the REST API directly:**
```bash
# Health check
curl http://localhost:8080/health

# List all assets
curl http://localhost:8080/v1/ListAssets

# Query a specific asset
curl "http://localhost:8080/v1/QueryAsset?assetId=asset1"

# Create a new asset
curl -X POST http://localhost:8080/v1/CreateAsset \
  -H "Content-Type: application/json" \
  -d '{"assetId":"asset100","color":"blue","size":20,"owner":"Alice","appraisedValue":500}'
```

---

## Stopping Everything

### Stop only application services (gateway + console)
```bash
docker-compose -f docker/apps.yaml down
```

### Stop the Fabric network (keep data)
```bash
./scripts/stop-fabric.sh
```

### Stop everything and wipe all data
```bash
./scripts/stop-fabric.sh --clean
```

> ⚠️ `--clean` removes `crypto-config/`, `channel-artifacts/`, and `bin/`. You will need to run `setup-fabric.sh` again to restart.

---

## Troubleshooting

### `nanayam network up --config docker/fabric-network.yaml` reports missing artifacts

For the built-in Fabric network, Nanayam now auto-runs `./scripts/setup-fabric.sh` when possible. If that setup step fails, fix the underlying problem (usually Docker, missing Fabric binaries, or a failed `cryptogen/configtxgen` run) and retry.

For custom compose files, generate the mounted crypto/channel artifacts manually before starting the network.

### "Docker network 'nanayam' not found"
Run `./scripts/start-fabric.sh` first. The `nanayam` network is created by the Fabric compose file.

### "Connection refused" when calling REST API
Wait 10-15 seconds after `start-distribution.sh` for the containers to finish initializing. Check logs:
```bash
docker logs nanayam-gateway
docker logs nanayam-console
```

### Chaincode container fails to build
Ensure Docker has enough disk space and memory (minimum 4GB RAM allocated to Docker Desktop).

### Port conflicts
If ports `3000`, `50051`, `7050-7054`, `8080`, `9051` are already in use, modify the port mappings in the respective `docker/*.yaml` files.

### macOS M1/M2 (Apple Silicon)
The setup uses `arm64` Fabric images automatically. If you see architecture warnings, ensure Rosetta 2 is enabled or use Docker Desktop's virtualization settings.

---

## Next Steps

- Explore the **Operator Console** at http://localhost:3000
- Read `docs/hyperledger-fabric-guide.md` to understand Fabric internals.
- Read `docs/nanayam-architecture.md` to understand how the components connect.
- Modify `services/gateway/handler.go` to add new chaincode functions.
- Extend the Operator Console UI in `apps/org-console/src/components/`.
