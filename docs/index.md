---
layout: default
title: Documentation
description: Nanayam — a private, permissioned Web3 ledger on Hyperledger Fabric.
---

# Nanayam · நாணயம்

> **நாணயம்** — a Tamil word meaning both *coin* and *integrity*. A ledger should carry both.

Nanayam is a **private, permissioned Web3 ledger** built on [Hyperledger Fabric](https://www.hyperledger.org/use/fabric). It ships a complete stack: a Fabric network, a Go gateway speaking gRPC and REST, and a Next.js console — plus one CLI that drives all of it.

All documentation is available in **English** and **தமிழ்**.

---

## Documentation

| | English | தமிழ் |
|---|---|---|
| Overview and navigation | [Home](wiki/Home.html) | [முகப்பு](wiki/Home-ta.html) |
| Run it locally in ten minutes | [Getting Started](wiki/Getting-Started.html) | [தொடங்குதல்](wiki/Getting-Started-ta.html) |
| How the pieces fit together | [Architecture](wiki/Architecture.html) | [கட்டமைப்பு](wiki/Architecture-ta.html) |
| Every command and flag | [CLI Reference](wiki/CLI-Reference.html) | [CLI கையேடு](wiki/CLI-Reference-ta.html) |
| Deploy to Kubernetes | [Cloud Deployment](wiki/Cloud-Deployment.html) | [கிளவுட் நிறுவல்](wiki/Cloud-Deployment-ta.html) |
| REST and gRPC endpoints | [API Reference](wiki/API-Reference.html) | [API கையேடு](wiki/API-Reference-ta.html) |
| Running and writing tests | [Testing](wiki/Testing.html) | [சோதனை](wiki/Testing-ta.html) |
| When something breaks | [Troubleshooting](wiki/Troubleshooting.html) | [சிக்கல் தீர்வு](wiki/Troubleshooting-ta.html) |
| How to contribute | [Contributing](wiki/Contributing.html) | [பங்களிப்பு](wiki/Contributing-ta.html) |

### API and client examples

- [API Explorer](api.html) — interactive Swagger UI generated from [`openapi.yaml`](openapi.yaml); try every endpoint against a local gateway
- [Sample Application Guide](sample-app-guide.html) — build, configure, and use the Operator Console (the Next.js app in `apps/org-console`)
- [Flutter Voucher Example](flutter-voucher-example.html) — a voucher provisioning & redemption app built against the same gateway API, plus the design write-up behind it

### Background reading

- [Hyperledger Fabric guide](hyperledger-fabric-guide.html) — Fabric concepts from first principles
- [Nanayam architecture](nanayam-architecture.html) — how the components interact
- [Complaint system](complaint-system.html) — the grievance workflow in detail
- [Local setup guide](local-setup-guide.html) — step-by-step development setup
- [How it helps e-governance](how-it-helps-e-governance.html) — the case for a permissioned ledger

---

## Quick start

```bash
curl -fsSL https://raw.githubusercontent.com/bytamilan/nanayam/main/install.sh | bash

nanayam prerequisites --auto    # Docker, Fabric binaries, and friends
nanayam network up              # start the Fabric network
```

Then open <http://localhost:3000>. The full walkthrough is in [Getting Started](wiki/Getting-Started.html).

### Deploy to a cluster

```bash
./scripts/deploy-cloud.sh --registry ghcr.io/bytamilan --domain nanayam.example.com
```

One command to GKE, EKS, AKS, k3d, or kind. Use `--dry-run` to preview without a cluster and `--destroy` to tear it down. See [Cloud Deployment](wiki/Cloud-Deployment.html).

---

## What Nanayam is for

A public blockchain makes every transaction visible to everyone. That is the wrong shape for a government department, a hospital, or a consortium of banks: they need a shared, tamper-evident record **and** control over who can read it.

- **Permissioned.** Every participant holds an issued certificate. No anonymous writers.
- **Private.** Data lives on channels. An organisation outside a channel cannot read it.
- **Tamper-evident.** Each block carries the hash of the one before it.
- **No mining.** Ordering is done by a Raft cluster — no proof-of-work, no cryptocurrency.

The worked example that ships with Nanayam is a **public grievance system**: a citizen files a complaint, an anti-corruption bureau acknowledges it, a department investigates, and an oversight body must co-sign before it can be closed. No single organisation can quietly close a case, because the endorsement policy requires several signatures.

---

## Project

- [Repository](https://github.com/bytamilan/nanayam)
- [Issues](https://github.com/bytamilan/nanayam/issues)
- [MIT License](https://github.com/bytamilan/nanayam/blob/main/LICENSE)
