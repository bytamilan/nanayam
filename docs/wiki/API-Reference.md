---
layout: default
title: API Reference
lang: en
---

# API Reference

**Languages:** **English** · [தமிழ்](API-Reference-ta.html)

The gateway exposes the same operations over REST (`:8080`) and gRPC (`:50051`).

---

## Authentication

Every endpoint except `/health`, `/v1/Config`, `/v1/Login`, and `/v1/Register` requires a bearer token.

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/v1/Login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin"}' | jq -r .token)

curl http://localhost:8080/v1/ListAssets -H "Authorization: Bearer $TOKEN"
```

Tokens are HS256 JWTs carrying `sub`, `usr`, `org`, `role`, `iat`, and `exp`. They expire after `AUTH_SESSION_HOURS` (default 24).

Missing, malformed, expired, or foreign-signed tokens all return **401**.

---

## Public endpoints

### `GET /health`

```json
{ "status": "ok" }
```

Used by the Kubernetes readiness and liveness probes.

### `GET /v1/Config`

```json
{ "signupEnabled": false, "channel": "mychannel", "chaincode": "basic" }
```

The console reads this to decide whether to show the signup link. If the gateway is unreachable, the console treats `signupEnabled` as `false` — it fails closed.

### `POST /v1/Login`

```json
{ "username": "admin", "password": "admin" }
```

**200** `{ "token": "eyJhbGci…" }` · **401** on bad credentials.

The error is identical whether the username is unknown or the password is wrong, so the endpoint cannot be used to enumerate accounts.

### `POST /v1/Register`

```json
{ "username": "alice", "password": "s3cret", "org": "Org1MSP" }
```

**201** with the created user · **403** when `AUTH_SIGNUP_ENABLED` is not `true` · **400** on a duplicate or empty username.

The password hash is never included in the response.

---

## Authenticated endpoints

### `GET /v1/Me`

Returns the authenticated user: `id`, `username`, `org`, `role`, `createdAt`.

### `GET /v1/ChannelInfo`

Channel name, chaincode name, MSP id, the organisations on the channel and their roles, and the orderer endpoints.

### Assets

| Endpoint | Method | Body / Query | Returns |
|---|---|---|---|
| `/v1/CreateAsset` | POST | `assetId`, `color`, `size`, `owner`, `appraisedValue` | `{ "success": true }` |
| `/v1/QueryAsset` | GET | `?assetId=` | `{ "data": "…json…" }` |
| `/v1/ListAssets` | GET | — | `{ "assetIds": [...] }` |

```bash
curl -X POST http://localhost:8080/v1/CreateAsset \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"assetId":"asset1","color":"blue","size":5,"owner":"Alice","appraisedValue":300}'
```

`CreateAsset` submits a transaction and waits for it to commit, so it takes roughly one block time. The query endpoints evaluate against a single peer and return in milliseconds.

### Complaints

| Endpoint | Method | Body / Query | Returns |
|---|---|---|---|
| `/v1/SubmitComplaint` | POST | `complaintId`, `category`, `citizenHash`, `descriptionHash`, `attachmentsRef` | `{ "success": true }` |
| `/v1/UpdateComplaint` | POST | `complaintId`, `action`, `value` | `{ "success": true }` |
| `/v1/QueryComplaint` | GET | `?complaintId=` | `{ "data": "…json…" }` |
| `/v1/ListComplaints` | GET | — | `{ "complaintIds": [...] }` |
| `/v1/GetComplaintHistory` | GET | `?complaintId=` | `{ "data": "…json array…" }` |

Only hashes go on the ledger — `citizenHash` and `descriptionHash`, not names or complaint text. Personally identifying data stays off-chain; the ledger proves the record has not changed without publishing its contents to every organisation on the channel.

`GetComplaintHistory` returns every version of the record with its transaction id and timestamp. That history cannot be edited, which is the entire point of putting a grievance process on a ledger.

### Ledger explorer

| Endpoint | Method | Query | Returns |
|---|---|---|---|
| `/v1/LedgerBlocks` | GET | `?start=0&end=10` | Block summaries |
| `/v1/LedgerActivity` | GET | — | Chain height and counts |

---

## Status codes

| Code | Meaning |
|---|---|
| 200 | Success |
| 201 | User created |
| 400 | Malformed JSON, or a required parameter is missing |
| 401 | Missing, malformed, expired, or invalid token |
| 403 | Registration attempted while signup is disabled |
| 405 | Wrong HTTP method for that endpoint |
| 500 | The chaincode or the peer returned an error |
| 502 | Gateway returned something the console could not parse |

---

## gRPC

The service is defined in `proto/fabric.proto`:

```protobuf
service FabricService {
  rpc CreateAsset(CreateAssetRequest) returns (CreateAssetResponse);
  rpc QueryAsset(QueryAssetRequest) returns (QueryAssetResponse);
  rpc ListAssets(ListAssetsRequest) returns (ListAssetsResponse);
  rpc SubmitComplaint(SubmitComplaintRequest) returns (SubmitComplaintResponse);
  rpc UpdateComplaint(UpdateComplaintRequest) returns (UpdateComplaintResponse);
  rpc QueryComplaint(QueryComplaintRequest) returns (QueryComplaintResponse);
  rpc ListComplaints(ListComplaintsRequest) returns (ListComplaintsResponse);
  rpc GetComplaintHistory(GetComplaintHistoryRequest) returns (GetComplaintHistoryResponse);
}
```

```bash
grpcurl -plaintext localhost:50051 list
grpcurl -plaintext -d '{}' localhost:50051 FabricService/ListAssets
```

---

## CORS

The gateway allows all origins, which suits development. For a deployment reachable by others, terminate at an ingress that sets a real origin policy, or narrow `corsMiddleware` in `services/gateway/http.go`.
