# Troubleshooting

**Languages:** **English** · [தமிழ்](Troubleshooting-ta)

---

## Diagnose first

```bash
docker ps -a                                    # what is running, what exited
docker logs peer0.org1.nanayam.com --tail 100   # why it exited
nanayam node status
curl http://localhost:8080/health
```

Most Fabric failures are a missing or wrong certificate. The log line that matters is usually the first error, not the last.

---

## `missing or incomplete Fabric artifacts`

Nanayam validates every bind mount before starting Docker, so this is the check working as intended. The message names the file:

```
missing or incomplete Fabric artifacts:
  - peer0.org1.nanayam.com: MSP directory /path/msp is missing signcerts
```

**Fix:**

```bash
nanayam crypto generate
nanayam network up
```

If it persists, the artifacts are stale — regenerate from scratch:

```bash
nanayam network clean
nanayam crypto generate
nanayam network up
```

---

## `fabric binaries not found`

The CLI looks in `./bin`, then `~/.nanayam/fabric-bin`, then `PATH`.

```bash
nanayam prerequisites --auto
ls ~/.nanayam/fabric-bin      # peer, cryptogen, configtxgen, fabric-ca-client
```

All four must be present. If some are missing, the download was interrupted; delete the directory and re-run.

---

## `docker compose is not available`

Nanayam accepts either Compose v2 (`docker compose`) or the standalone `docker-compose`.

```bash
docker compose version || docker-compose version
```

If neither works, install Compose v2 and confirm the Docker daemon is running: `docker info`.

---

## A container starts and immediately exits

```bash
docker logs <container> --tail 50
```

| Log says | Cause | Fix |
|---|---|---|
| `cannot find MSP` | Crypto material missing or mounted wrong | `nanayam crypto generate` |
| `failed to create ledger` | Stale volumes from a previous topology | `nanayam network clean` |
| `TLS handshake failed` | Certificates regenerated for one side only | Regenerate everything, restart all |
| `bind: address already in use` | Another process holds the port | See below |

---

## Port already in use

```bash
lsof -i :7051     # or :8080, :3000, :7050
```

Stop the process, or move Nanayam:

```bash
nanayam gateway --http-port 9090
nanayam console --port 4000
```

---

## Channel creation fails

```
Error: got unexpected status: BAD_REQUEST -- error validating channel creation transaction
```

Usually the genesis block was generated from a different `configtx.yaml` than the channel transaction. Regenerate both together:

```bash
rm -rf channel-artifacts
nanayam crypto generate
nanayam channel create --name mychannel --profile TwoOrgsChannel
```

Nanayam records which crypto config produced the artifacts in `channel-artifacts/.nanayam-artifact-source`, so a later `network up` can detect a mismatch — but a hand-edited config can still get out of step.

---

## Chaincode commit fails

```bash
nanayam chaincode checkcommitreadiness --name basic --channel mychannel
```

This lists which organisations have approved. Every organisation in the endorsement policy must run `approve` with the **same** package id before `commit` succeeds.

```bash
nanayam chaincode queryinstalled    # shows the real package id
```

A mismatched package id is the most common cause: the id includes a hash of the package, so rebuilding the chaincode changes it.

---

## Gateway cannot connect to the peer

```
Failed to connect to Fabric gateway: connection error
```

| Check | Command |
|---|---|
| Is the peer running? | `docker ps \| grep peer0` |
| Is the endpoint right? | `echo $PEER_ENDPOINT` |
| Does the TLS cert exist? | `ls $TLS_CERT_PATH` |
| Same Docker network? | `docker network inspect nanayam` |

Inside Docker, use the service name (`peer0.org1.nanayam.com:7051`), not `localhost`.

---

## Console shows "Gateway down"

```bash
curl http://localhost:8080/health
```

If that fails, the gateway is not running or not reachable. If it succeeds but the console still errors, the console's `GATEWAY_URL` points somewhere else — check the environment it was started with. In Docker or Kubernetes it must be a service name, not `localhost`.

---

## 401 on every request

| Cause | Fix |
|---|---|
| No token | Log in first; check the `nanayam_token` cookie is set |
| Token expired | Log in again; raise `AUTH_SESSION_HOURS` if 24h is too short |
| Gateway restarted with a different `AUTH_JWT_SECRET` | Existing tokens are invalid — log in again |
| Gateway restarted at all | The user store is in memory, so accounts are lost on restart |

That last one is worth stating plainly: the auth store is **in-memory**. Restarting the gateway drops every registered user and reseeds only `admin`. This is fine for development; a durable store is needed before anything else.

---

## Registration returns 403

Signup is disabled by default. Enable it deliberately:

```bash
AUTH_SIGNUP_ENABLED=true nanayam gateway
```

Or with the deploy script: `--enable-signup`.

---

## Tests fail with a path error

If a test fails with a path under `/Users/…` or `/home/…` that is not yours, it hardcodes someone's machine. Use `t.TempDir()` or `repoRoot(t)` instead — see [Testing](Testing).

---

## Kubernetes pods will not start

```bash
kubectl -n nanayam get pods
kubectl -n nanayam describe pod <pod>
kubectl -n nanayam logs <pod>
```

| Pod status | Cause | Fix |
|---|---|---|
| `ImagePullBackOff` | Cluster cannot pull the image | Check the registry ref and pull secrets |
| `CreateContainerConfigError` | A referenced Secret is missing | Re-run the deploy script |
| `CrashLoopBackOff` | The container starts and dies | Read the logs; usually crypto material |
| `Pending` | No node has room | `kubectl describe node`; check resource requests |

---

## Complete reset

When something is wedged and you want a clean slate:

```bash
nanayam network clean
rm -rf crypto-config channel-artifacts
docker system prune -f
nanayam network up
```

This destroys all ledger data. On a development network that is the point.

---

## Still stuck

Open an issue at <https://github.com/bytamilan/nanayam/issues> with:

- what you ran and what you expected
- the full error output
- `nanayam version`, `docker --version`, `go version`
- your OS and architecture
