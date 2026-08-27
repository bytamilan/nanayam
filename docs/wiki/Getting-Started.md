# Getting Started

**Languages:** **English** · [தமிழ்](Getting-Started-ta)

This page takes you from an empty machine to a running ledger with data in it. Budget about fifteen minutes, most of it waiting on downloads.

---

## 1. Prerequisites

| Tool | Why | Check |
|---|---|---|
| Docker + Compose v2 | Every Fabric node runs as a container | `docker compose version` |
| Go 1.21+ | Builds the CLI and the gateway | `go version` |
| Node.js 20+ and pnpm | Builds and runs the console | `node --version && pnpm --version` |
| `git`, `curl`, `jq` | Used by the setup scripts | `git --version` |

Nanayam can install most of these for you:

```bash
nanayam prerequisites --auto
```

That command also downloads the Hyperledger Fabric binaries (`peer`, `cryptogen`, `configtxgen`, `fabric-ca-client`) into `~/.nanayam/fabric-bin`. Without them, nothing that touches crypto or channels will work.

---

## 2. Install the CLI

```bash
curl -fsSL https://raw.githubusercontent.com/bytamilan/nanayam/main/install.sh | bash
```

Or from a checkout:

```bash
git clone https://github.com/bytamilan/nanayam.git
cd nanayam
make install
export PATH="$HOME/.nanayam/bin:$PATH"
```

Confirm it landed:

```bash
nanayam version
```

---

## 3. Bring up a network

```bash
nanayam network up
```

This one command does rather a lot, and it is worth knowing what:

1. **Checks for crypto material.** If `crypto-config/` is missing or incomplete, it runs the matching setup script to generate it with `cryptogen`.
2. **Checks for channel artifacts.** If `channel-artifacts/` has no genesis block, it runs `configtxgen` against the right `configtx*.yaml`.
3. **Validates the compose file.** Every bind mount is checked before Docker starts, so a missing certificate is reported as a clear message rather than a container that exits three seconds later.
4. **Starts the stack** with Docker Compose.

For the grievance workflow instead of the basic asset network:

```bash
nanayam network up --profile complaint
```

Check what came up:

```bash
docker ps
nanayam node status
```

---

## 4. Create a channel and join peers

```bash
nanayam channel create --name mychannel --profile TwoOrgsChannel
nanayam channel join --name mychannel
nanayam channel update-anchor --name mychannel --org Org1MSP
```

A **channel** is a private ledger shared by a named set of organisations. Peers that have not joined it cannot see its data at all. The **anchor peer** is the one other organisations use to discover the rest of yours.

---

## 5. Deploy chaincode

Chaincode is the smart contract: the only thing allowed to write to the ledger.

```bash
nanayam chaincode package --path ./chaincode/asset-transfer-basic --name basic
nanayam chaincode install --package basic.tar.gz
nanayam chaincode approve --name basic --channel mychannel --package-id basic_1.0:<hash>
nanayam chaincode commit --name basic --channel mychannel
```

The approve step must be repeated by **each** organisation in the endorsement policy before commit will succeed. That is the point: no single party installs code onto a shared ledger unilaterally.

Seed and read it back:

```bash
nanayam chaincode invoke --channel mychannel --name basic --function InitLedger
nanayam chaincode query  --channel mychannel --name basic --function GetAllAssets
```

---

## 6. Start the gateway and console

```bash
nanayam gateway    # REST on :8080, gRPC on :50051
nanayam console    # Next.js on :3000
```

Open <http://localhost:3000> and sign in with **admin / admin**.

> Change that password before the deployment is reachable by anyone else. Registration is disabled by default; turn it on with `AUTH_SIGNUP_ENABLED=true` only when you mean to.

---

## 7. Confirm it works

```bash
curl http://localhost:8080/health
# {"status":"ok"}

TOKEN=$(curl -s -X POST http://localhost:8080/v1/Login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin"}' | jq -r .token)

curl -s http://localhost:8080/v1/ListAssets -H "Authorization: Bearer $TOKEN" | jq
```

If the last call returns your assets, every layer is working: console → gateway → peer → ledger.

---

## Shutting down

```bash
nanayam network down     # stop, keep the data
nanayam network clean    # stop and wipe everything
```

`clean` deletes the ledger, the crypto material, and the channel artifacts. On a development network that is exactly what you want when something is wedged; on anything else, be sure.

---

## Where next

- [Architecture](Architecture) — what each component actually does
- [CLI Reference](CLI-Reference) — every command and flag
- [Cloud Deployment](Cloud-Deployment) — the same stack on Kubernetes
- [Troubleshooting](Troubleshooting) — when a step above did not go to plan
