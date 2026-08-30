# Operator Console

The Next.js web UI for Nanayam — a working sample application built against the [distribution gateway](../../docs/wiki/API-Reference.md)'s REST API. Every screen (login, dashboard, complaints, ledger explorer) is a thin client over that API; there is no direct connection to Fabric from the browser or from this app.

Full build-and-usage documentation, including configuration, project layout, and what each route calls, lives in the **[Sample Application Guide](https://bytamilan.github.io/nanayam/sample-app-guide.html)**.

## Quick start

```bash
pnpm install
pnpm dev        # http://localhost:3000, requires a gateway on :8080
```

```bash
pnpm build && pnpm start   # production build
pnpm test                  # Jest, fetch is mocked — no gateway required
pnpm typecheck
pnpm lint
```

By default the console talks to a gateway at `http://localhost:8080`. Point it elsewhere with `GATEWAY_URL` (server-side) and `NEXT_PUBLIC_GATEWAY_URL` (client-side) — see the [Sample Application Guide](https://bytamilan.github.io/nanayam/sample-app-guide.html#configuration) for details.

## Learn more

- [Nanayam documentation](https://bytamilan.github.io/nanayam/) — full docs site
- [API Reference](https://bytamilan.github.io/nanayam/wiki/API-Reference.html) / [API Explorer](https://bytamilan.github.io/nanayam/api.html) — the gateway endpoints this console calls
- [Getting Started](https://bytamilan.github.io/nanayam/wiki/Getting-Started.html) — bring up the full stack (Fabric network + gateway + console)

This project was bootstrapped with [`create-next-app`](https://nextjs.org/docs/app/api-reference/cli/create-next-app) and uses [`next/font`](https://nextjs.org/docs/app/building-your-application/optimizing/fonts) with [Geist](https://vercel.com/font).
