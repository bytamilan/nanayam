# CLI Reference

**Languages:** **English** · [தமிழ்](CLI-Reference-ta)

Every command supports `--help`. Global flags: `--config <file>` to point at an alternate config, `--verbose` for detailed output.

---

## `nanayam prerequisites`

Checks for, and optionally installs, everything Nanayam needs.

```bash
nanayam prerequisites          # report what is missing
nanayam prerequisites --auto   # install what is missing
```

Installs Docker Compose, `jq`, and the Fabric binaries into `~/.nanayam/fabric-bin`.

---

## `nanayam network`

| Command | Effect |
|---|---|
| `network up` | Start the network, generating missing artifacts first |
| `network down` | Stop containers, keep volumes and data |
| `network clean` | Stop containers and delete all data |
| `network status` | Show what is running |

```bash
nanayam network up                                  # basic profile
nanayam network up --profile complaint              # grievance network
nanayam network up --config docker/fabric-network.yaml
```

**Profiles versus `--config`.** A profile is a known network Nanayam can repair: if crypto or channel artifacts are missing, it generates them. `--config` points at one specific compose file and uses only that. Nanayam still validates the mounts, but it will not invent crypto for a topology it does not recognise, and it will not auto-attach `apps.yaml`.

---

## `nanayam crypto`

```bash
nanayam crypto generate
nanayam crypto generate --config config/crypto-config-complaint.yaml --output crypto-config
nanayam crypto generate --channel-artifacts=false   # certificates only
nanayam crypto renew --org Org1
```

`generate` runs `cryptogen` and then, unless told otherwise, `configtxgen` for the genesis block, channel transaction, and anchor peer updates. It records which crypto config produced the artifacts in `channel-artifacts/.nanayam-artifact-source`, so a later `network up` can tell a complaint-network artifact set from a basic one and regenerate when they do not match.

---

## `nanayam channel`

```bash
nanayam channel create --name mychannel --profile TwoOrgsChannel
nanayam channel join --name mychannel
nanayam channel join --name mychannel --org Org2
nanayam channel list
nanayam channel update-anchor --name mychannel --org Org1MSP
```

---

## `nanayam chaincode`

```bash
nanayam chaincode package --path ./chaincode/asset-transfer-basic --name basic --version 1.0
nanayam chaincode install --package basic.tar.gz
nanayam chaincode queryinstalled
nanayam chaincode approve --name basic --channel mychannel --package-id basic_1.0:abc123
nanayam chaincode checkcommitreadiness --name basic --channel mychannel
nanayam chaincode commit --name basic --channel mychannel
nanayam chaincode invoke --channel mychannel --name basic --function InitLedger
nanayam chaincode invoke --channel mychannel --name basic --function CreateAsset --args asset1,blue,5,Alice,300
nanayam chaincode query --channel mychannel --name basic --function GetAllAssets
```

`approve` must be run once per organisation in the endorsement policy. `checkcommitreadiness` shows which organisations have and have not approved, which is the first thing to check when `commit` fails.

---

## `nanayam node`

```bash
nanayam node init --type peer --org Org1
nanayam node init --type orderer
nanayam node init --type ca --org Org1
nanayam node start peer0.org1.nanayam.com
nanayam node stop  peer0.org1.nanayam.com
nanayam node status
```

`node init` renders a compose file and cryptogen config for a single node from the embedded templates — useful when adding an organisation to an existing consortium rather than starting a whole network.

---

## `nanayam user`

```bash
nanayam user create --id alice --secret alicepw --type client --org Org1
nanayam user enroll --id alice --secret alicepw --org Org1
nanayam user list --org Org1
```

These manage **Fabric identities** — X.509 certificates from the CA. Console accounts are separate; see [Architecture](Architecture).

---

## `nanayam consortium`

```bash
nanayam consortium connect \
  --orderer orderer.example.com:7050 \
  --tls-cert ./tls.crt \
  --org NewOrg \
  --domain neworg.example.com

nanayam consortium join-channel --name mychannel --block ./mychannel.block
```

For joining an existing multi-organisation network rather than running your own.

---

## `nanayam gateway` and `nanayam console`

```bash
nanayam gateway                    # REST :8080, gRPC :50051
nanayam gateway --http-port 9090

nanayam console                    # Next.js dev server on :3000
nanayam console --port 4000
nanayam console --docker           # build and run the container instead
```

---

## `nanayam upgrade`

```bash
nanayam upgrade --check                                   # is a newer release out?
nanayam upgrade                                           # install the latest
nanayam upgrade --refresh                                 # reinstall the current release
nanayam upgrade --dev-local --refresh --source /path/to/nanayam
```

Version comparison is semver-aware, so `upgrade` will not move you backwards onto an older tag. A binary built from source reports `dev` and treats any tagged release as newer.

---

## Make targets

| Target | Effect |
|---|---|
| `make build` | Build the CLI into `build/` |
| `make install` | Install to `~/.nanayam/bin` |
| `make test` | Run every test suite: CLI, gateway, console |
| `make test-cli` / `test-gateway` / `test-console` | Run one suite |
| `make lint` | `go vet` both modules |
| `make fmt` / `make fmt-check` | Format, or fail if unformatted |
| `make validate` | fmt-check, lint, build, and test |
| `make build-all` | Cross-compile for all platforms |
| `make release-assets` | Build packaged release archives |
| `make deploy-cloud ARGS="..."` | Run the cloud deployment script |
| `make help` | List every target |
