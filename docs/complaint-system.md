# Anti-Corruption Complaint System

> A Hyperledger Fabric prototype demonstrating **multi-authority workflow enforcement** where no single department can suppress or alter a complaint.

---

## Architecture

```
┌─────────────┐   REST   ┌─────────────┐   gRPC-TLS   ┌─────────────┐
│   Console   │◀────────▶│   Gateway   │◀────────────▶│  ACB Peer   │
│  (Next.js)  │          │    (Go)     │              │  (Org1)     │
└─────────────┘          └─────────────┘              └──────┬──────┘
                                                             │
                         ┌─────────────┐   Gossip           │
                         │  Dept Peer  │◀───────────────────┤
                         │  (Org2)     │                    │
                         └──────┬──────┘                    │
                                │                          │
                         ┌──────┴──────┐                   │
                         │ Oversight   │◀──────────────────┤
                         │ Peer (Org3) │                   │
                         └──────┬──────┘                   │
                                │                          │
                         ┌──────┴──────┐                   │
                         │ Judiciary   │◀──────────────────┘
                         │ Peer (Org4) │
                         └─────────────┘
                              │
                              ▼
                       ┌─────────────┐
                       │   Orderer   │
                       │   (Raft)    │
                       └─────────────┘
```

---

## Organizations

| Org | MSP ID | Role |
|-----|--------|------|
| **ACB** | `ACBMSP` | Receives complaints, acknowledges, assigns investigators, requests closure |
| **Department** | `DeptMSP` | Investigated party — can update status and add evidence, **cannot close** |
| **Oversight** | `OversightMSP` | Independent body — **must approve closure** (anti-suppression control) |
| **Judiciary** | `JudiciaryMSP` | Optional escalation peer (read-only in this prototype) |

---

## Complaint State Machine

```
Draft (client-side)
    │
    ▼
Submitted ──────────────▶ Rejected (ACB only, with reason)
    │
    ▼ Acknowledge (ACB only)
Acknowledged
    │
    ▼ Assign / Start Investigation
UnderInvestigation
    │
    ▼ Action Taken
ActionTaken
    │
    ▼ Request Closure (ACB only)
PendingClosure
    │
    ▼ Approve Closure (Oversight only)
Closed
```

**Key rule:** ACB cannot close a complaint alone. Oversight must `ApproveClosure`.

---

## Chaincode Access Control

| Function | Allowed MSP | Description |
|----------|-------------|-------------|
| `SubmitComplaint` | Any | Citizen submits complaint |
| `AcknowledgeComplaint` | ACBMSP | ACB officer acknowledges receipt |
| `AssignInvestigator` | ACBMSP | Assign to a department |
| `UpdateStatus` | ACBMSP, DeptMSP | Strict transition rules enforced |
| `AddEvidence` | ACBMSP, DeptMSP | Append IPFS/S3 reference |
| `RequestClosure` | ACBMSP | Move to PendingClosure |
| `ApproveClosure` | **OversightMSP** | **Critical control — only Oversight can finalize closure** |
| `RejectComplaint` | ACBMSP | Reject with reason |
| `GetComplaint` | Any | Read public fields |
| `GetAllComplaints` | Any | List all complaints |
| `GetComplaintHistory` | Any | Full tamper-proof audit trail |
| `GetPrivateComplaintData` | ACBMSP, OversightMSP | Read PII (private data collection) |

---

## Privacy Design

| Data | Storage | Access |
|------|---------|--------|
| Complaint metadata (status, category, timestamps) | Public ledger state | All channel members |
| Citizen hash, description hash | **Private Data Collection** | ACB + Oversight only |
| Full documents / attachments | Off-chain (IPFS/S3) | Referenced by hash on-chain |

---

## Quick Start

```bash
# 1. Setup: binaries, images, crypto, channel artifacts
./scripts/setup-complaint.sh

# 2. Start the 4-org Fabric network
./scripts/start-complaint.sh

# 3. Deploy the complaint chaincode
./scripts/deploy-complaint.sh

# 4. Start gateway + console
./scripts/start-complaint-apps.sh
```

Open **http://localhost:3000**

---

## REST API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1/SubmitComplaint` | POST | Create new complaint |
| `/v1/UpdateComplaint` | POST | Action: acknowledge, assign, updateStatus, addEvidence, requestClosure, approveClosure, reject |
| `/v1/QueryComplaint?complaintId=` | GET | Read single complaint |
| `/v1/ListComplaints` | GET | List all complaints |
| `/v1/GetComplaintHistory?complaintId=` | GET | Full audit trail |

### Example: Submit
```bash
curl -X POST http://localhost:8080/v1/SubmitComplaint \
  -H "Content-Type: application/json" \
  -d '{"complaintId":"COMP-001","category":"bribery","citizenHash":"sha256:abc","descriptionHash":"sha256:def","attachmentsRef":"ipfs:QmExample"}'
```

### Example: Acknowledge
```bash
curl -X POST http://localhost:8080/v1/UpdateComplaint \
  -H "Content-Type: application/json" \
  -d '{"complaintId":"COMP-001","action":"acknowledge"}'
```

### Example: Approve Closure (Oversight)
```bash
curl -X POST http://localhost:8080/v1/UpdateComplaint \
  -H "Content-Type: application/json" \
  -d '{"complaintId":"COMP-001","action":"approveClosure"}'
```

---

## Stopping

```bash
# Stop apps only
docker-compose -f docker/complaint-apps.yaml down

# Stop Fabric network (keeps data)
docker-compose -f docker/complaint-network.yaml down

# Stop everything and wipe
docker-compose -f docker/complaint-apps.yaml down
docker-compose -f docker/complaint-network.yaml down -v
rm -rf crypto-config channel-artifacts
```

---

## Production Hardening

| Prototype (Current) | Production |
|---------------------|------------|
| Docker Compose on 1 machine | Kubernetes across 3+ data centres |
| 1 peer per org | 2+ peers per org for HA |
| 1 orderer | 5+ Raft orderers for BFT |
| Self-signed TLS | NIC-approved PKI |
| `cryptogen` static certs | Fabric CA with HSM |
| Single gateway (ACB identity) | Separate gateway per org with proper IAM |
| LevelDB | CouchDB for rich queries (already configured) |
| No backup | Automated snapshots + DR |
