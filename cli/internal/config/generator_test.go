package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

// readGenerated runs a generator into a temp path and returns the file contents.
func readGenerated(t *testing.T, name string, generate func(outputPath string) error) string {
	t.Helper()

	outputPath := filepath.Join(t.TempDir(), name)
	if err := generate(outputPath); err != nil {
		t.Fatalf("generate %s: %v", name, err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read generated %s: %v", name, err)
	}
	return string(content)
}

// assertValidYAML fails when the generated file is not parseable, which is the
// failure mode a bad template produces: Docker Compose rejects it at run time.
func assertValidYAML(t *testing.T, label, content string) {
	t.Helper()

	var doc any
	if err := yaml.Unmarshal([]byte(content), &doc); err != nil {
		t.Fatalf("%s is not valid YAML: %v\n---\n%s", label, err, content)
	}
}

// assertNoUnresolvedPlaceholders catches fields the template referenced but the
// config never set, which text/template renders as the literal "<no value>".
func assertNoUnresolvedPlaceholders(t *testing.T, label, content string) {
	t.Helper()

	if strings.Contains(content, "<no value>") {
		t.Errorf("%s contains an unresolved template field:\n%s", label, content)
	}
	if strings.Contains(content, "{{") {
		t.Errorf("%s contains an unrendered template action:\n%s", label, content)
	}
}

func peerConfig() *PeerNodeConfig {
	return &PeerNodeConfig{
		OrgName:       "Org1",
		OrgDomain:     "org1.nanayam.com",
		Domain:        "nanayam.com",
		PeerCount:     2,
		UserCount:     1,
		PeerID:        "peer0.org1.nanayam.com",
		PeerPort:      7051,
		ChaincodePort: 7052,
		OpsPort:       9443,
		MSPID:         "Org1MSP",
		CryptoPath:    "./crypto-config",
	}
}

func TestGenerateCryptogen(t *testing.T) {
	cfg := peerConfig()
	content := readGenerated(t, "crypto-config.yaml", func(out string) error {
		return GenerateCryptogen(cfg, out)
	})

	assertValidYAML(t, "cryptogen config", content)
	assertNoUnresolvedPlaceholders(t, "cryptogen config", content)
	if !strings.Contains(content, cfg.OrgName) {
		t.Errorf("cryptogen config does not mention org %q:\n%s", cfg.OrgName, content)
	}
}

func TestGeneratePeerCompose(t *testing.T) {
	cfg := peerConfig()
	content := readGenerated(t, "peer.yaml", func(out string) error {
		return GeneratePeerCompose(cfg, out)
	})

	assertValidYAML(t, "peer compose", content)
	assertNoUnresolvedPlaceholders(t, "peer compose", content)
	if !strings.Contains(content, cfg.PeerID) {
		t.Errorf("peer compose does not mention peer id %q:\n%s", cfg.PeerID, content)
	}
}

func TestGenerateOrdererCompose(t *testing.T) {
	cfg := &OrdererNodeConfig{
		OrdererID:    "orderer.nanayam.com",
		OrdererPort:  7050,
		MSPID:        "OrdererMSP",
		CryptoPath:   "./crypto-config",
		GenesisBlock: "./channel-artifacts/genesis.block",
	}
	content := readGenerated(t, "orderer.yaml", func(out string) error {
		return GenerateOrdererCompose(cfg, out)
	})

	assertValidYAML(t, "orderer compose", content)
	assertNoUnresolvedPlaceholders(t, "orderer compose", content)
	if !strings.Contains(content, cfg.OrdererID) {
		t.Errorf("orderer compose does not mention orderer id %q:\n%s", cfg.OrdererID, content)
	}
}

func TestGenerateCAComposeDerivesLowercaseOrgName(t *testing.T) {
	cfg := &CANodeConfig{OrgName: "Org1", CAPort: 7054, OpsPort: 17054}
	content := readGenerated(t, "ca.yaml", func(out string) error {
		return GenerateCACompose(cfg, out)
	})

	assertValidYAML(t, "CA compose", content)
	assertNoUnresolvedPlaceholders(t, "CA compose", content)

	// Container and volume names must be lowercase; Docker rejects some
	// uppercase names, so the generator derives the lowercase form itself.
	if cfg.OrgNameLower != "org1" {
		t.Errorf("OrgNameLower = %q, want %q", cfg.OrgNameLower, "org1")
	}
}

func TestGeneratorsCreateMissingParentDirectories(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "nested", "deeper", "peer.yaml")

	if err := GeneratePeerCompose(peerConfig(), outputPath); err != nil {
		t.Fatalf("GeneratePeerCompose into a missing directory: %v", err)
	}
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("generated file missing: %v", err)
	}
}

func TestExecuteTemplateReportsParseErrors(t *testing.T) {
	err := executeTemplate("{{ .Unclosed", struct{}{}, filepath.Join(t.TempDir(), "out.yaml"))
	if err == nil {
		t.Fatal("executeTemplate() = nil for a malformed template, want an error")
	}
	if !strings.Contains(err.Error(), "parse template") {
		t.Errorf("error = %q, want it to mention parse template", err)
	}
}

func TestExecuteTemplateReportsExecutionErrors(t *testing.T) {
	// Calling a method that does not exist fails at execution, not parse time.
	err := executeTemplate("{{ .Missing.Field }}", struct{}{}, filepath.Join(t.TempDir(), "out.yaml"))
	if err == nil {
		t.Fatal("executeTemplate() = nil for a bad field reference, want an error")
	}
	if !strings.Contains(err.Error(), "execute template") {
		t.Errorf("error = %q, want it to mention execute template", err)
	}
}
