---
layout: default
title: Testing
lang: en
---

# Testing

**Languages:** **English** · [தமிழ்](Testing-ta.html)

---

## Running the suites

```bash
make test            # everything: CLI, gateway, console
make test-cli
make test-gateway
make test-console
make validate        # fmt-check, lint, build, and test
```

Directly:

```bash
cd cli              && go test ./...
cd services/gateway && go test -race ./...
cd apps/org-console && pnpm test
```

---

## What is covered

| Suite | Location | Focus |
|---|---|---|
| CLI | `cli/**/*_test.go` | Compose resolution, artifact generation, binary discovery, semver, template rendering |
| Gateway | `services/gateway/*_test.go` | REST routing, auth middleware, the auth store, error propagation |
| Console | `apps/org-console/src/**/*.test.ts` | API route handlers, cookie handling, the API client |

---

## Conventions

**Never hardcode an absolute path.** A test that embeds `/Users/someone/…` passes on one machine and fails everywhere else, CI included. Use `t.TempDir()`, or `repoRoot(t)` from `cli/cmd/testsupport_test.go`, which walks up to the checked-in `config/configtx.yaml`.

**Name the behaviour, not the function.** `TestLoginFailuresAreIndistinguishable` says what is guaranteed. `TestLogin` does not.

**Comment the why, not the what.** A test asserting that login errors match for an unknown user and a wrong password should say it is preventing account enumeration — otherwise the next reader will "simplify" it away.

**Test the error paths.** Most of the gateway suite is about what happens when things fail: no token, an expired token, a token from another deployment, malformed JSON, the peer unreachable.

---

## Writing a Go test

```go
func TestValidateTLSDirRequiresAllThreeFiles(t *testing.T) {
    dir := t.TempDir()   // cleaned up automatically

    if err := os.WriteFile(filepath.Join(dir, "ca.crt"), []byte("x"), 0o644); err != nil {
        t.Fatalf("write ca.crt: %v", err)
    }

    issues := validateTLSDir("peer0", dir)
    if len(issues) != 2 {
        t.Fatalf("expected 2 issues for server.crt and server.key, got %v", issues)
    }
}
```

Table tests for anything with several cases:

```go
func TestDefaultChannelID(t *testing.T) {
    cases := map[string]string{
        "TwoOrgsChannel":     "mychannel",
        "ComplaintChannel":   "complaint-channel",
        "SupplyChainChannel": "supply-chain-channel",
    }

    for profile, want := range cases {
        if got := defaultChannelID(profile); got != want {
            t.Errorf("defaultChannelID(%q) = %q, want %q", profile, got, want)
        }
    }
}
```

### Testing code that shells out

`internal/ca` builds `fabric-ca-client` arguments and runs the binary. The tests point `BinaryPath` at a recorder script that logs its arguments, so the argument construction is verified without the real CA:

```go
binary, calls := newRecorder(t)
c := NewClient(binary, t.TempDir(), "localhost:7054")

if err := c.Enroll("alice", "alice-pw", "/msp/alice"); err != nil {
    t.Fatalf("Enroll() = %v", err)
}

if !hasFlag(calls()[0], "-u", "alice:alice-pw@localhost:7054") {
    t.Errorf("enrollment URL malformed: %v", calls()[0])
}
```

---

## Writing a console test

Route handlers are tested with a mocked `fetch`:

```ts
const mockFetch = jest.fn();
global.fetch = mockFetch as unknown as typeof fetch;

it('sets an httpOnly session cookie on a successful login', async () => {
  const { POST } = await import('./route');
  mockFetch.mockResolvedValueOnce({
    ok: true,
    status: 200,
    text: async () => JSON.stringify({ token: 'jwt-token-value' }),
  });

  const res = await POST(jsonRequest({ username: 'admin', password: 'admin' }));

  expect(res.cookies.get('nanayam_token')?.httpOnly).toBe(true);
});
```

Import the route **inside** the test with `await import('./route')`. Route modules read `process.env.GATEWAY_URL` at module load, so a top-level import freezes whatever the environment was when the file was first evaluated.

Tests needing a DOM declare it per file:

```ts
/**
 * @jest-environment jsdom
 */
```

---

## The race detector

The gateway's auth store is shared across every in-flight request, so its tests run under `-race` in CI:

```bash
cd services/gateway && go test -race ./...
```

This is slower — bcrypt dominates — but it is the only thing that catches a concurrent map write before production does.

---

## Coverage

```bash
cd cli && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
cd apps/org-console && pnpm test:coverage
```

Coverage is a hint, not a target. A package at 90% whose tests never exercise an error path is worse tested than one at 60% whose tests do.

---

## What CI runs

`.github/workflows/ci.yml` runs on every push and pull request:

| Job | Checks |
|---|---|
| `cli` | gofmt, vet, build, `go test -race` |
| `gateway` | gofmt, vet, build, `go test -race` |
| `console` | typecheck, jest |
| `scripts` | `bash -n` and shellcheck on every `.sh` |
| `manifests` | Renders `k8s/` and fails on unsubstituted placeholders or unparseable YAML |

Run the equivalent locally before pushing:

```bash
make validate
```
