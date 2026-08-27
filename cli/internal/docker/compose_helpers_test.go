package docker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestComposeConfigCandidatesForAbsolutePath(t *testing.T) {
	// An absolute path is taken literally: never rewritten into docker/.
	got := composeConfigCandidates("/project", "/etc/nanayam/custom.yaml")

	if len(got) != 1 || got[0] != "/etc/nanayam/custom.yaml" {
		t.Fatalf("candidates = %v, want just the absolute path", got)
	}
}

func TestComposeConfigCandidatesForRelativePath(t *testing.T) {
	got := composeConfigCandidates("/project", "docker/fabric-network.yaml")

	want := []string{
		filepath.Join("/project", "docker/fabric-network.yaml"),
		filepath.Join("/project", "docker", "docker/fabric-network.yaml"),
	}
	if len(got) != len(want) {
		t.Fatalf("candidates = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("candidates = %v, want %v", got, want)
		}
	}
}

// A bare name such as "complaint-network" should also be tried with .yaml so
// users do not have to type the extension.
func TestComposeConfigCandidatesAppendYAMLForExtensionlessNames(t *testing.T) {
	got := composeConfigCandidates("/project", "complaint-network")

	wantSuffixes := []string{
		filepath.Join("/project", "complaint-network"),
		filepath.Join("/project", "docker", "complaint-network"),
		filepath.Join("/project", "complaint-network.yaml"),
		filepath.Join("/project", "docker", "complaint-network.yaml"),
	}
	if len(got) != len(wantSuffixes) {
		t.Fatalf("candidates = %v, want %v", got, wantSuffixes)
	}
	for i, want := range wantSuffixes {
		if got[i] != want {
			t.Fatalf("candidate[%d] = %q, want %q", i, got[i], want)
		}
	}
}

func TestFindComposeFilesInDir(t *testing.T) {
	base := t.TempDir()
	dockerDir := filepath.Join(base, "docker")
	if err := os.MkdirAll(dockerDir, 0o755); err != nil {
		t.Fatalf("create docker dir: %v", err)
	}

	for _, name := range []string{"fabric-network.yaml", "complaint-network.yaml", "custom.yaml"} {
		if err := os.WriteFile(filepath.Join(dockerDir, name), []byte("services: {}\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	cases := map[string]string{
		"basic":     "fabric-network.yaml",
		"complaint": "complaint-network.yaml",
		"custom":    "custom.yaml",
	}

	for profile, wantFile := range cases {
		t.Run(profile, func(t *testing.T) {
			files, err := FindComposeFilesInDir(base, profile)
			if err != nil {
				t.Fatalf("FindComposeFilesInDir(%q) = %v", profile, err)
			}
			if len(files) != 1 {
				t.Fatalf("expected 1 compose file, got %v", files)
			}
			if filepath.Base(files[0]) != wantFile {
				t.Fatalf("got %q, want %q", filepath.Base(files[0]), wantFile)
			}
		})
	}
}

func TestFindComposeFilesInDirRejectsUnknownProfile(t *testing.T) {
	base := t.TempDir()
	if err := os.MkdirAll(filepath.Join(base, "docker"), 0o755); err != nil {
		t.Fatalf("create docker dir: %v", err)
	}

	_, err := FindComposeFilesInDir(base, "does-not-exist")
	if err == nil {
		t.Fatal("FindComposeFilesInDir() = nil for an unknown profile, want an error")
	}
	if !strings.Contains(err.Error(), "does-not-exist") {
		t.Errorf("error %q does not name the profile", err)
	}
}

// A named profile whose compose file was deleted must fail loudly rather than
// handing docker compose a path that does not exist.
func TestFindComposeFilesInDirReportsMissingBuiltInFile(t *testing.T) {
	base := t.TempDir()
	if err := os.MkdirAll(filepath.Join(base, "docker"), 0o755); err != nil {
		t.Fatalf("create docker dir: %v", err)
	}

	_, err := FindComposeFilesInDir(base, "basic")
	if err == nil {
		t.Fatal("FindComposeFilesInDir(basic) = nil with no compose file, want an error")
	}
	if !strings.Contains(err.Error(), "fabric-network.yaml") {
		t.Errorf("error %q does not name the missing file", err)
	}
}

func TestParseBindMount(t *testing.T) {
	cases := []struct {
		name       string
		mount      string
		wantOK     bool
		wantSource string
		wantTarget string
	}{
		{"relative source", "../crypto-config:/etc/hyperledger/crypto", true, "../crypto-config", "/etc/hyperledger/crypto"},
		{"absolute source", "/var/run/docker.sock:/host/var/run/docker.sock", true, "/var/run/docker.sock", "/host/var/run/docker.sock"},
		{"home source", "~/.nanayam:/data", true, "~/.nanayam", "/data"},
		{"read-only flag", "./msp:/etc/msp:ro", true, "./msp", "/etc/msp"},
		{"named volume", "peer0-data:/var/hyperledger/production", false, "", ""},
		{"no target", "./crypto-config", false, "", ""},
		{"empty", "", false, "", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source, target, ok := parseBindMount(tc.mount)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if !ok {
				return
			}
			if source != tc.wantSource {
				t.Errorf("source = %q, want %q", source, tc.wantSource)
			}
			if target != tc.wantTarget {
				t.Errorf("target = %q, want %q", target, tc.wantTarget)
			}
		})
	}
}

func TestLooksLikeBindSource(t *testing.T) {
	bind := []string{"./crypto", "../crypto", "/abs/path", "~/home/path"}
	named := []string{"peer0-data", "orderer_data", "nanayam-vol"}

	for _, source := range bind {
		if !looksLikeBindSource(source) {
			t.Errorf("looksLikeBindSource(%q) = false, want true", source)
		}
	}
	for _, source := range named {
		if looksLikeBindSource(source) {
			t.Errorf("looksLikeBindSource(%q) = true; named volumes are not bind mounts", source)
		}
	}
}

func TestValidateMSPDirRequiresSigncerts(t *testing.T) {
	withSigncerts := t.TempDir()
	signcerts := filepath.Join(withSigncerts, "signcerts")
	if err := os.MkdirAll(signcerts, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(signcerts, "cert.pem"), []byte("cert"), 0o644); err != nil {
		t.Fatalf("write cert: %v", err)
	}
	if issues := validateMSPDir("peer0", withSigncerts); len(issues) != 0 {
		t.Errorf("expected no issues, got %v", issues)
	}

	// An MSP directory generated but never populated is the exact failure that
	// makes a peer container exit at startup, so it must be caught up front.
	empty := t.TempDir()
	if err := os.MkdirAll(filepath.Join(empty, "signcerts"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if issues := validateMSPDir("peer0", empty); len(issues) != 1 {
		t.Errorf("expected 1 issue for an empty signcerts dir, got %v", issues)
	}

	if issues := validateMSPDir("peer0", t.TempDir()); len(issues) != 1 {
		t.Errorf("expected 1 issue when signcerts is absent, got %v", issues)
	}
}

func TestValidateTLSDirRequiresAllThreeFiles(t *testing.T) {
	complete := t.TempDir()
	for _, name := range []string{"ca.crt", "server.crt", "server.key"} {
		if err := os.WriteFile(filepath.Join(complete, name), []byte("x"), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	if issues := validateTLSDir("peer0", complete); len(issues) != 0 {
		t.Errorf("expected no issues, got %v", issues)
	}

	partial := t.TempDir()
	if err := os.WriteFile(filepath.Join(partial, "ca.crt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write ca.crt: %v", err)
	}
	issues := validateTLSDir("peer0", partial)
	if len(issues) != 2 {
		t.Fatalf("expected 2 issues for server.crt and server.key, got %v", issues)
	}
	joined := strings.Join(issues, " ")
	if !strings.Contains(joined, "server.crt") || !strings.Contains(joined, "server.key") {
		t.Errorf("issues %v do not name both missing files", issues)
	}
}

func TestValidateRegularFile(t *testing.T) {
	dir := t.TempDir()
	block := filepath.Join(dir, "genesis.block")
	if err := os.WriteFile(block, []byte("block"), 0o644); err != nil {
		t.Fatalf("write block: %v", err)
	}
	if issues := validateRegularFile("orderer", block); len(issues) != 0 {
		t.Errorf("expected no issues, got %v", issues)
	}

	if issues := validateRegularFile("orderer", filepath.Join(dir, "missing.block")); len(issues) != 1 {
		t.Errorf("expected 1 issue for a missing file, got %v", issues)
	}

	// configtxgen writing into an existing directory is a common mistake; the
	// container would then mount a directory where it expects a block file.
	if issues := validateRegularFile("orderer", dir); len(issues) != 1 {
		t.Errorf("expected 1 issue when the path is a directory, got %v", issues)
	}
}

func TestValidateBindMountDispatchesOnTarget(t *testing.T) {
	dir := t.TempDir()

	// An unrecognised target is not something we can meaningfully check.
	if issues := validateBindMount("peer0", dir, "/etc/hyperledger/anything"); len(issues) != 0 {
		t.Errorf("expected no issues for an unchecked target, got %v", issues)
	}

	// A source that does not exist at all fails regardless of target kind.
	missing := filepath.Join(dir, "nope")
	if issues := validateBindMount("peer0", missing, "/etc/hyperledger/msp"); len(issues) != 1 {
		t.Errorf("expected 1 issue for a missing source, got %v", issues)
	}
}

func TestWriteComposeFileCreatesTheDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "generated", "nodes")

	if err := WriteComposeFile(dir, "peer0.yaml", []byte("services: {}\n")); err != nil {
		t.Fatalf("WriteComposeFile() = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "peer0.yaml"))
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	if string(content) != "services: {}\n" {
		t.Errorf("content = %q", content)
	}
}

func TestValidateComposePrerequisitesReportsAllIssuesSorted(t *testing.T) {
	base := t.TempDir()
	composeFile := filepath.Join(base, "network.yaml")
	compose := `services:
  peer0:
    volumes:
      - ./missing-msp:/etc/hyperledger/fabric/msp
      - ./missing-tls:/etc/hyperledger/fabric/tls
  orderer:
    volumes:
      - ./missing-genesis.block:/var/hyperledger/orderer/genesis.block
`
	if err := os.WriteFile(composeFile, []byte(compose), 0o644); err != nil {
		t.Fatalf("write compose: %v", err)
	}

	err := ValidateComposePrerequisites([]string{composeFile})
	if err == nil {
		t.Fatal("ValidateComposePrerequisites() = nil, want issues for every missing mount")
	}

	message := err.Error()
	for _, want := range []string{"missing-msp", "missing-tls", "missing-genesis.block"} {
		if !strings.Contains(message, want) {
			t.Errorf("error does not mention %q:\n%s", want, message)
		}
	}
}

func TestValidateComposePrerequisitesReportsUnreadableFile(t *testing.T) {
	err := ValidateComposePrerequisites([]string{filepath.Join(t.TempDir(), "absent.yaml")})
	if err == nil {
		t.Fatal("ValidateComposePrerequisites() = nil for a missing compose file, want an error")
	}
}

func TestValidateComposePrerequisitesRejectsMalformedYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.yaml")
	if err := os.WriteFile(path, []byte("services:\n  peer0:\n   - bad\n  : :\n"), 0o644); err != nil {
		t.Fatalf("write compose: %v", err)
	}

	if err := ValidateComposePrerequisites([]string{path}); err == nil {
		t.Fatal("ValidateComposePrerequisites() = nil for malformed YAML, want an error")
	}
}

func TestValidateComposePrerequisitesIgnoresNamedVolumes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "network.yaml")
	compose := `services:
  peer0:
    volumes:
      - peer0-data:/var/hyperledger/production
      - orderer_data:/var/hyperledger/orderer
`
	if err := os.WriteFile(path, []byte(compose), 0o644); err != nil {
		t.Fatalf("write compose: %v", err)
	}

	// Named volumes are created by Docker on demand, so they are not a
	// prerequisite the user has to satisfy beforehand.
	if err := ValidateComposePrerequisites([]string{path}); err != nil {
		t.Fatalf("ValidateComposePrerequisites() = %v, want nil for named volumes", err)
	}
}
