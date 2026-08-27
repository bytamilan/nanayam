package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestArtifactVariantFromFile(t *testing.T) {
	cases := map[string]string{
		"crypto-config.yaml":           "",
		"crypto-config-complaint.yaml": "complaint",
		"crypto-config-basic.yaml":     "basic",
		"crypto-config-my-net.yaml":    "my-net",
		"configtx.yaml":                "",
	}

	for file, want := range cases {
		prefix := "crypto-config"
		if strings.HasPrefix(file, "configtx") {
			prefix = "configtx"
		}
		if got := artifactVariantFromFile(file, prefix); got != want {
			t.Errorf("artifactVariantFromFile(%q, %q) = %q, want %q", file, prefix, got, want)
		}
	}
}

func TestConfigtxCandidatesForVariant(t *testing.T) {
	// The default variant tries several historical names so an older checkout
	// still resolves; a named variant must not silently fall back to another
	// network's configtx, which would generate the wrong genesis block.
	for _, variant := range []string{"", "basic", "fabric"} {
		got := configtxCandidatesForVariant(variant)
		if len(got) < 2 {
			t.Errorf("variant %q: expected fallbacks, got %v", variant, got)
		}
		if !contains(got, "configtx.yaml") {
			t.Errorf("variant %q: candidates %v omit configtx.yaml", variant, got)
		}
	}

	got := configtxCandidatesForVariant("complaint")
	want := []string{"configtx-complaint.yaml"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("configtxCandidatesForVariant(complaint) = %v, want %v", got, want)
	}
}

func TestConfigtxCandidatesAreDeduplicated(t *testing.T) {
	got := configtxCandidatesForVariant("basic")

	seen := make(map[string]bool, len(got))
	for _, candidate := range got {
		if seen[candidate] {
			t.Fatalf("duplicate candidate %q in %v", candidate, got)
		}
		seen[candidate] = true
	}
}

func TestResolveConfigtxSource(t *testing.T) {
	base := t.TempDir()
	configDir := filepath.Join(base, "config")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for _, name := range []string{"configtx.yaml", "configtx-complaint.yaml"} {
		if err := os.WriteFile(filepath.Join(configDir, name), []byte("Profiles: {}\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	got, err := resolveConfigtxSource(base, "")
	if err != nil {
		t.Fatalf("resolveConfigtxSource(default) = %v", err)
	}
	if filepath.Base(got) != "configtx.yaml" {
		t.Errorf("default variant resolved to %q", filepath.Base(got))
	}

	got, err = resolveConfigtxSource(base, "complaint")
	if err != nil {
		t.Fatalf("resolveConfigtxSource(complaint) = %v", err)
	}
	if filepath.Base(got) != "configtx-complaint.yaml" {
		t.Errorf("complaint variant resolved to %q", filepath.Base(got))
	}

	if _, err := resolveConfigtxSource(base, "nonexistent"); err == nil {
		t.Error("resolveConfigtxSource(nonexistent) = nil, want an error")
	}
}

func TestLoadChannelArtifactProfiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "configtx.yaml")
	content := `
Profiles:
  TwoOrgsOrdererGenesis:
    Orderer:
      OrdererType: etcdraft
  TwoOrgsChannel:
    Application:
      Organizations:
        - Name: Org1MSP
        - Name: Org2MSP
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write configtx: %v", err)
	}

	profiles, err := loadChannelArtifactProfiles(path)
	if err != nil {
		t.Fatalf("loadChannelArtifactProfiles() = %v", err)
	}

	want := []channelArtifactProfile{{
		name:           "TwoOrgsOrdererGenesis",
		channelProfile: "TwoOrgsChannel",
		genesis:        "genesis.block",
		channel:        "mychannel",
		anchorOrgs:     []string{"Org1MSP", "Org2MSP"},
	}}
	if !reflect.DeepEqual(profiles, want) {
		t.Fatalf("profiles = %#v, want %#v", profiles, want)
	}
}

// A genesis profile with no matching *Channel profile cannot produce channel
// artifacts, so it is skipped rather than emitting a half-built profile.
func TestLoadChannelArtifactProfilesSkipsGenesisWithoutChannel(t *testing.T) {
	path := filepath.Join(t.TempDir(), "configtx.yaml")
	content := `
Profiles:
  OrphanOrdererGenesis:
    Orderer: {}
  TwoOrgsOrdererGenesis:
    Orderer: {}
  TwoOrgsChannel:
    Application:
      Organizations:
        - Name: Org1MSP
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write configtx: %v", err)
	}

	profiles, err := loadChannelArtifactProfiles(path)
	if err != nil {
		t.Fatalf("loadChannelArtifactProfiles() = %v", err)
	}
	if len(profiles) != 1 || profiles[0].name != "TwoOrgsOrdererGenesis" {
		t.Fatalf("profiles = %#v, want only TwoOrgsOrdererGenesis", profiles)
	}
}

func TestLoadChannelArtifactProfilesErrors(t *testing.T) {
	dir := t.TempDir()

	if _, err := loadChannelArtifactProfiles(filepath.Join(dir, "absent.yaml")); err == nil {
		t.Error("expected an error for a missing file")
	}

	malformed := filepath.Join(dir, "malformed.yaml")
	if err := os.WriteFile(malformed, []byte("Profiles:\n  - :\n   bad\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := loadChannelArtifactProfiles(malformed); err == nil {
		t.Error("expected an error for malformed YAML")
	}

	empty := filepath.Join(dir, "empty.yaml")
	if err := os.WriteFile(empty, []byte("Profiles: {}\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := loadChannelArtifactProfiles(empty); err == nil {
		t.Error("expected an error when no genesis/channel profile pair exists")
	}
}

func TestOrganizationNamesFallsBackToID(t *testing.T) {
	got := organizationNames([]configtxOrganization{
		{Name: "Org1MSP"},
		{ID: "Org2MSP"},             // some profiles only set ID
		{Name: "ACB", ID: "ACBMSP"}, // Name wins when both are present
		{},                          // an empty entry contributes nothing
	})

	want := []string{"Org1MSP", "Org2MSP", "ACB"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("organizationNames() = %v, want %v", got, want)
	}
}

func TestDefaultChannelID(t *testing.T) {
	cases := map[string]string{
		"TwoOrgsChannel":     "mychannel",
		"ComplaintChannel":   "complaint-channel",
		"SupplyChainChannel": "supply-chain-channel",
		"AuditChannel":       "audit-channel",
		"Channel":            "mychannel",
	}

	for profile, want := range cases {
		if got := defaultChannelID(profile); got != want {
			t.Errorf("defaultChannelID(%q) = %q, want %q", profile, got, want)
		}
	}
}

func TestCamelToKebab(t *testing.T) {
	cases := map[string]string{
		"SupplyChain": "supply-chain",
		"Audit":       "audit",
		"ABC":         "a-b-c",
		"lowerFirst":  "lower-first",
		"":            "",
	}

	for in, want := range cases {
		if got := camelToKebab(in); got != want {
			t.Errorf("camelToKebab(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestUniqueStringsPreservesOrderAndDropsEmpties(t *testing.T) {
	got := uniqueStrings([]string{"a", "b", "a", "", "c", "b", ""})

	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("uniqueStrings() = %v, want %v", got, want)
	}
}

func TestArtifactSourceLabelIsRepoRelative(t *testing.T) {
	cwd := filepath.Join("/project", "nanayam")

	// A path inside the project is recorded relative, so the metadata file stays
	// portable between machines and checkouts.
	got := artifactSourceLabel(cwd, filepath.Join(cwd, "config", "crypto-config-complaint.yaml"))
	if want := "config/crypto-config-complaint.yaml"; got != want {
		t.Errorf("artifactSourceLabel() = %q, want %q", got, want)
	}
}

func TestArtifactSourceMetadataRoundTrip(t *testing.T) {
	cwd := t.TempDir()
	cryptoConfig := filepath.Join(cwd, "config", "crypto-config-complaint.yaml")

	if err := writeArtifactSourceMetadata(cwd, cryptoConfig); err != nil {
		t.Fatalf("writeArtifactSourceMetadata() = %v", err)
	}

	got, err := readArtifactSourceMetadata(cwd)
	if err != nil {
		t.Fatalf("readArtifactSourceMetadata() = %v", err)
	}
	if want := "config/crypto-config-complaint.yaml"; got != want {
		t.Fatalf("round trip = %q, want %q", got, want)
	}
}

func TestReadArtifactSourceMetadataMissing(t *testing.T) {
	if _, err := readArtifactSourceMetadata(t.TempDir()); err == nil {
		t.Fatal("readArtifactSourceMetadata() = nil with no metadata file, want an error")
	}
}

func TestArtifactSourceMetadataPath(t *testing.T) {
	got := artifactSourceMetadataPath("/project")
	want := filepath.Join("/project", "channel-artifacts", ".nanayam-artifact-source")
	if got != want {
		t.Fatalf("artifactSourceMetadataPath() = %q, want %q", got, want)
	}
}

// configtxgen resolves the config by FABRIC_CFG_PATH and insists the file is
// named configtx.yaml, so a variant file has to be staged under that name.
func TestStageConfigtxFileRenamesToConfigtxYAML(t *testing.T) {
	cwd := t.TempDir()
	if err := os.MkdirAll(filepath.Join(cwd, "crypto-config"), 0o755); err != nil {
		t.Fatalf("mkdir crypto-config: %v", err)
	}

	source := filepath.Join(cwd, "configtx-complaint.yaml")
	if err := os.WriteFile(source, []byte("Profiles: {}\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	stagedDir, cleanup, err := stageConfigtxFile(cwd, source)
	if err != nil {
		t.Fatalf("stageConfigtxFile() = %v", err)
	}
	defer cleanup()

	staged := filepath.Join(stagedDir, "configtx.yaml")
	content, err := os.ReadFile(staged)
	if err != nil {
		t.Fatalf("staged configtx.yaml missing: %v", err)
	}
	if string(content) != "Profiles: {}\n" {
		t.Errorf("staged content = %q", content)
	}

	// The staged tree links back to the real crypto-config so configtxgen can
	// read the MSP material referenced by the profiles.
	linked := filepath.Join(filepath.Dir(stagedDir), "crypto-config")
	if _, err := os.Stat(linked); err != nil {
		t.Errorf("crypto-config not linked into the staging dir: %v", err)
	}

	cleanup()
	if _, err := os.Stat(staged); !os.IsNotExist(err) {
		t.Error("cleanup did not remove the staging directory")
	}
}

func TestStageConfigtxFileRejectsMissingSource(t *testing.T) {
	cwd := t.TempDir()
	if _, _, err := stageConfigtxFile(cwd, filepath.Join(cwd, "absent.yaml")); err == nil {
		t.Fatal("stageConfigtxFile() = nil for a missing source, want an error")
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
