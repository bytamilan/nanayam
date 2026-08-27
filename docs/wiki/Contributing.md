# Contributing

**Languages:** **English** · [தமிழ்](Contributing-ta)

---

## Setting up

```bash
git clone https://github.com/bytamilan/nanayam.git
cd nanayam

nanayam prerequisites --auto              # or install Go, Node, Docker yourself
cd apps/org-console && pnpm install && cd ../..

make build
make test
```

If `make test` is green on a fresh clone, your environment is ready.

---

## Before you push

```bash
make validate
```

That runs formatting, vet, build, and every test suite — the same checks CI runs. Running it locally is faster than waiting for a red build.

---

## Repository layout

```
nanayam/
├── apps/org-console/     Next.js console
├── cli/                  Nanayam CLI (Go, Cobra)
│   ├── cmd/              One file per subcommand
│   ├── internal/         ca, config, docker, fabric
│   └── templates/        Embedded node templates
├── services/gateway/     Go gRPC + REST gateway
├── config/               cryptogen and configtx
├── docker/               Compose stacks
├── k8s/                  Kubernetes manifests
├── scripts/              Setup and deployment scripts
└── docs/wiki/            This wiki
```

---

## Style

**Go.** Standard `gofmt`; `make fmt` applies it. Wrap errors with context: `fmt.Errorf("read configtx source %s: %w", path, err)`. Exported identifiers get doc comments. Keep `cmd/` files to command definition and orchestration; real logic belongs in `internal/`.

**TypeScript.** Match the surrounding file. Route handlers return `NextResponse.json` with an explicit status. Never let a token or secret reach the response body.

**Shell.** `set -euo pipefail` at the top. Quote every expansion. Every script must pass `bash -n` and shellcheck at warning level.

---

## Commits

Conventional Commits:

```
feat: add complaint escalation to the judiciary org
fix: resolve compose path when --config is relative
docs: document the cloud deployment flags
test: cover the auth store under concurrent access
refactor: extract compose validation into internal/docker
```

Explain **why** in the body, not just what. The diff already shows what changed.

---

## Pull requests

1. Branch from `main`.
2. Make the change, with tests.
3. Run `make validate`.
4. Update the docs if behaviour changed — including the Tamil page.
5. Open the PR describing the problem and how you solved it.

A change to a documented behaviour that leaves the docs stale is not finished.

---

## Tests

Every bug fix should come with a test that fails without it. See [Testing](Testing) for conventions — the short version:

- No hardcoded absolute paths. Use `t.TempDir()` or `repoRoot(t)`.
- Name tests after the behaviour they guarantee.
- Comment why an assertion matters, not what it does.
- Cover the error paths, not just the happy one.

---

## Documentation

The wiki lives in `docs/wiki/` and is published to the GitHub Wiki automatically on push to `main`.

**Pages come in pairs.** `Architecture.md` and `Architecture-ta.md` are the same page in English and Tamil. When you change one, change the other. If you cannot write the Tamil, say so in the PR and someone will help — a Tamil page that silently drifts out of date is worse than an obviously missing one.

Both versions link to each other at the top:

```markdown
**Languages:** **English** · [தமிழ்](Page-Name-ta)
```

Diagrams are [Mermaid](https://mermaid.js.org/) in fenced ```mermaid blocks, which GitHub renders natively.

---

## Adding a wiki page

1. Write `New-Page.md` and `New-Page-ta.md` in `docs/wiki/`.
2. Add both to `_Sidebar.md`.
3. Link them from `Home.md` and `Home-ta.md` if they belong in the main table.
4. Cross-link the two at the top of each.

---

## Reporting bugs

Include what you ran, what you expected, the full error, and your versions (`nanayam version`, `docker --version`, `go version`, OS and architecture). A report someone can reproduce is worth several that they cannot.

---

## Code of conduct

Be decent to each other. Assume good faith. Review the change, not the person.

---

## License

Contributions are licensed under the [MIT License](https://github.com/bytamilan/nanayam/blob/main/LICENSE).
