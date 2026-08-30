---
layout: default
title: Sample Application Guide
---

# Sample Application Guide

> Build, run, and use the **Operator Console** — the Next.js sample application that ships with Nanayam.

The console at [`apps/org-console`](https://github.com/bytamilan/nanayam/tree/main/apps/org-console) is a working example of a client built against the [distribution gateway](wiki/API-Reference.html): it authenticates against `/v1/Login`, and every ledger action a browser can trigger routes through the gateway's REST API. Read this page to build it, run it standalone, and understand what each screen does; read the [API Reference](wiki/API-Reference.html) or the [API Explorer](api.html) for the endpoints it calls.

---

## Prerequisites

| Tool | Version | Check |
|---|---|---|
| Node.js | 20+ | `node --version` |
| pnpm | 9+ | `pnpm --version` |
| A running gateway | — | `curl http://localhost:8080/health` |

The console is a plain client: it needs a gateway to talk to, but does not need a Fabric network of its own if you only want to work on the UI against a gateway someone else is running (or a mocked one — see [Testing](#testing)).

Install pnpm if you do not have it:

```bash
corepack enable
corepack prepare pnpm@latest --activate
```

---

## Build and run

```bash
cd apps/org-console
pnpm install
```

### Development server

```bash
pnpm dev
```

Opens on <http://localhost:3000> with hot reload (Turbopack). Requires a gateway reachable at the URL in `GATEWAY_URL` (defaults to `http://localhost:8080` — start one with `nanayam gateway` or `./scripts/start-distribution.sh`).

### Production build

```bash
pnpm build   # compiles the app into .next/
pnpm start   # serves the production build on :3000
```

### Via the Nanayam CLI

From the repository root, the CLI wraps the same commands and is the path used by `nanayam network up`:

```bash
nanayam console
```

### Via Docker

The console ships its own [`Dockerfile`](https://github.com/bytamilan/nanayam/blob/main/apps/org-console/Dockerfile) (multi-stage: pnpm build, then a slim runtime image). It is normally started together with the gateway:

```bash
docker-compose -f docker/apps.yaml up -d --build console
```

---

## Configuration

The console reads its gateway address from the environment. There is no config file — only these variables:

| Variable | Used by | Purpose |
|---|---|---|
| `GATEWAY_URL` | Server-side route handlers (`src/app/api/**`) and server components | Where the console's own backend forwards requests. Defaults to `http://localhost:8080`. |
| `NEXT_PUBLIC_GATEWAY_URL` | Client bundle | Set when the browser needs to reach the gateway directly rather than through the console's own API routes. In the shipped Docker Compose setup this points at the `gateway` service on the internal Docker network. |

```bash
GATEWAY_URL=http://localhost:8080 pnpm dev
```

The console never talks to Fabric directly — every request goes through the gateway, which is the only thing holding a peer connection and signing identity.

---

## Project structure

```
apps/org-console/
├── src/
│   ├── app/
│   │   ├── (auth)/            # /login, /signup — unauthenticated routes
│   │   ├── (console)/         # /dashboard, /channels, /complaints, /ledger — session required
│   │   └── api/                # Server route handlers that proxy to the gateway
│   ├── components/
│   │   ├── auth/               # LoginForm, SignupForm
│   │   └── layout/              # Shell chrome shared by the (console) routes
│   ├── actions/                 # Server actions used by forms
│   └── lib/
│       ├── api.ts               # Typed client for assets and complaints
│       └── auth.ts              # Session cookie handling
├── Dockerfile
├── jest.config.js
└── package.json
```

`src/lib/api.ts` is the one place that knows the shape of gateway responses — every page and component goes through it rather than calling `fetch` directly, so a change to the gateway's JSON only needs updating in one file.

---

## What each screen does

| Route | Requires session | What it calls |
|---|---|---|
| `/login` | No | `POST /v1/Login`; also reads `GET /v1/Config` to decide whether to show the signup link |
| `/signup` | No | `POST /v1/Register` (only rendered when `Config.signupEnabled` is true) |
| `/dashboard` | Yes | `GET /v1/ListAssets`, `POST /v1/CreateAsset` |
| `/channels` | Yes | `GET /v1/ChannelInfo` |
| `/complaints` | Yes | `GET /v1/ListComplaints`, `POST /v1/SubmitComplaint`, `POST /v1/UpdateComplaint`, `GET /v1/GetComplaintHistory` |
| `/ledger` | Yes | `GET /v1/LedgerBlocks`, `GET /v1/LedgerActivity` |

Session state is a cookie set by the console's own `/api/auth/login` route after it exchanges credentials with the gateway — the JWT itself never reaches client-side JavaScript.

---

## Testing

```bash
pnpm test           # Jest, with fetch mocked — no gateway or network required
pnpm test:coverage
pnpm typecheck
pnpm lint
```

The test suite mocks `fetch` (see `src/lib/api.test.ts`), so `pnpm test` needs neither a running gateway nor a Fabric network. This is also what `make test` and CI (`.github/workflows/ci.yml`, the `console` job) run.

---

## Where next

- [Getting Started](wiki/Getting-Started.html) — bring up the full stack this console talks to
- [API Reference](wiki/API-Reference.html) / [API Explorer](api.html) — every endpoint the console calls
- [Architecture](wiki/Architecture.html) — how the console, gateway, and Fabric network fit together
- [Cloud Deployment](wiki/Cloud-Deployment.html) — running the console on Kubernetes
