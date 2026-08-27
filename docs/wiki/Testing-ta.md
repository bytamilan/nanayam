# சோதனை

**மொழிகள்:** [English](Testing) · **தமிழ்**

---

## தொகுப்புகளை இயக்குதல்

```bash
make test            # அனைத்தும்: CLI, gateway, கன்சோல்
make test-cli
make test-gateway
make test-console
make validate        # fmt-check, lint, build, test
```

நேரடியாக:

```bash
cd cli              && go test ./...
cd services/gateway && go test -race ./...
cd apps/org-console && pnpm test
```

---

## என்ன உள்ளடக்கப்பட்டுள்ளது

| தொகுப்பு | இடம் | கவனம் |
|---|---|---|
| CLI | `cli/**/*_test.go` | Compose தீர்வு, artifact உருவாக்கம், பைனரி கண்டுபிடிப்பு, semver, டெம்ப்ளேட் rendering |
| Gateway | `services/gateway/*_test.go` | REST routing, அங்கீகார middleware, auth store, பிழைப் பரவல் |
| கன்சோல் | `apps/org-console/src/**/*.test.ts` | API route handlers, குக்கீ கையாளுதல், API client |

---

## மரபுகள்

**முழுமையான பாதையை (absolute path) ஒருபோதும் நேரடியாக எழுத வேண்டாம்.** `/Users/someone/…` உள்ளடக்கிய சோதனை ஒரு கணினியில் வெற்றி பெற்று, மற்ற எல்லா இடங்களிலும் — CI உட்பட — தோல்வியடையும். `t.TempDir()` பயன்படுத்துங்கள், அல்லது `cli/cmd/testsupport_test.go`-இல் உள்ள `repoRoot(t)` — இது களஞ்சியத்தில் உள்ள `config/configtx.yaml` வரை மேலே நடந்து செல்கிறது.

**செயல்பாட்டைக் குறிப்பிடாமல், நடத்தையைக் குறிப்பிடுங்கள்.** `TestLoginFailuresAreIndistinguishable` எது உறுதி செய்யப்படுகிறது என்று சொல்கிறது. `TestLogin` சொல்வதில்லை.

**என்ன என்பதற்குப் பதிலாக ஏன் என்பதைக் குறிப்பிடுங்கள்.** தெரியாத பயனருக்கும் தவறான கடவுச்சொல்லுக்கும் உள்நுழைவுப் பிழைகள் ஒன்றேதான் என்று உறுதிப்படுத்தும் சோதனை, இது கணக்குகளைக் கண்டறிவதைத் தடுக்கிறது என்று சொல்ல வேண்டும் — இல்லையெனில் அடுத்து படிப்பவர் அதை "எளிமையாக்கி" நீக்கிவிடுவார்.

**பிழைப் பாதைகளைச் சோதியுங்கள்.** Gateway தொகுப்பின் பெரும்பகுதி, விஷயங்கள் தோல்வியடையும்போது என்ன நடக்கிறது என்பது பற்றியதே: டோக்கன் இல்லாதது, காலாவதியான டோக்கன், வேறு நிறுவலின் டோக்கன், தவறான JSON, பியரை அடைய முடியாதது.

---

## Go சோதனை எழுதுதல்

```go
func TestValidateTLSDirRequiresAllThreeFiles(t *testing.T) {
    dir := t.TempDir()   // தானாகவே சுத்தம் செய்யப்படுகிறது

    if err := os.WriteFile(filepath.Join(dir, "ca.crt"), []byte("x"), 0o644); err != nil {
        t.Fatalf("write ca.crt: %v", err)
    }

    issues := validateTLSDir("peer0", dir)
    if len(issues) != 2 {
        t.Fatalf("expected 2 issues for server.crt and server.key, got %v", issues)
    }
}
```

பல வழக்குகள் உள்ள எதற்கும் table சோதனைகள்:

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

### வெளிக் கட்டளைகளை இயக்கும் நிரலைச் சோதித்தல்

`internal/ca` `fabric-ca-client` அளவுருக்களை உருவாக்கி பைனரியை இயக்குகிறது. சோதனைகள் `BinaryPath`-ஐ, தன் அளவுருக்களைப் பதிவு செய்யும் recorder ஸ்கிரிப்டில் சுட்டுகின்றன. இதனால் உண்மையான CA இல்லாமலேயே அளவுரு உருவாக்கம் சரிபார்க்கப்படுகிறது:

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

## கன்சோல் சோதனை எழுதுதல்

Route handlers, mock செய்யப்பட்ட `fetch` மூலம் சோதிக்கப்படுகின்றன:

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

Route-ஐ சோதனைக்கு **உள்ளே** `await import('./route')` மூலம் இறக்குமதி செய்யுங்கள். Route modules `process.env.GATEWAY_URL`-ஐ module ஏற்றப்படும்போது படிக்கின்றன; எனவே மேல்நிலை இறக்குமதி, கோப்பு முதலில் மதிப்பிடப்பட்டபோது இருந்த சூழலை நிலைநிறுத்திவிடும்.

DOM தேவைப்படும் சோதனைகள் ஒவ்வொரு கோப்பிலும் அதை அறிவிக்கின்றன:

```ts
/**
 * @jest-environment jsdom
 */
```

---

## Race detector

Gateway-இன் auth store ஒவ்வொரு நடப்புக் கோரிக்கையாலும் பகிரப்படுகிறது; எனவே அதன் சோதனைகள் CI-இல் `-race` உடன் இயங்குகின்றன:

```bash
cd services/gateway && go test -race ./...
```

இது மெதுவானது — bcrypt-தான் பெரும்பாலான நேரத்தை எடுக்கிறது — ஆனால் உற்பத்திக்கு முன்பே இணையான map எழுத்தைப் பிடிக்கும் ஒரே வழி இதுவே.

---

## Coverage

```bash
cd cli && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
cd apps/org-console && pnpm test:coverage
```

Coverage ஒரு குறிப்பு, இலக்கு அல்ல. பிழைப் பாதையை ஒருபோதும் இயக்காத சோதனைகளுடன் 90%-இல் உள்ள தொகுப்பு, அதை இயக்கும் சோதனைகளுடன் 60%-இல் உள்ள தொகுப்பைவிட மோசமாகச் சோதிக்கப்பட்டது.

---

## CI என்ன இயக்குகிறது

`.github/workflows/ci.yml` ஒவ்வொரு push மற்றும் pull request-இலும் இயங்குகிறது:

| வேலை | சரிபார்ப்புகள் |
|---|---|
| `cli` | gofmt, vet, build, `go test -race` |
| `gateway` | gofmt, vet, build, `go test -race` |
| `console` | typecheck, jest |
| `scripts` | ஒவ்வொரு `.sh`-க்கும் `bash -n` மற்றும் shellcheck |
| `manifests` | `k8s/`-ஐ உருவாக்கி, மாற்றப்படாத placeholders அல்லது பகுக்க முடியாத YAML இருந்தால் தோல்வியடைகிறது |

Push செய்வதற்கு முன் இதற்கு இணையானதை உள்ளூரில் இயக்குங்கள்:

```bash
make validate
```
