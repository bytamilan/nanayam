package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestResolveNetworkUpComposeFilesFromDirProfile(t *testing.T) {
	tempDir := t.TempDir()
	dockerDir := filepath.Join(tempDir, "docker")
	if err := os.MkdirAll(dockerDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	writeTestFile(t, filepath.Join(dockerDir, "fabric-network.yaml"))
	writeTestFile(t, filepath.Join(dockerDir, "apps.yaml"))

	gotFiles, gotLabel, err := resolveNetworkUpComposeFilesFromDir(tempDir, "basic", "")
	if err != nil {
		t.Fatalf("resolveNetworkUpComposeFilesFromDir() error = %v", err)
	}

	wantFiles := []string{
		filepath.Join(dockerDir, "fabric-network.yaml"),
		filepath.Join(dockerDir, "apps.yaml"),
	}
	if !reflect.DeepEqual(gotFiles, wantFiles) {
		t.Fatalf("resolveNetworkUpComposeFilesFromDir() files = %v, want %v", gotFiles, wantFiles)
	}

	if gotLabel != "profile: basic" {
		t.Fatalf("resolveNetworkUpComposeFilesFromDir() label = %q, want %q", gotLabel, "profile: basic")
	}
}

func TestResolveNetworkUpComposeFilesFromDirConfig(t *testing.T) {
	tempDir := t.TempDir()
	dockerDir := filepath.Join(tempDir, "docker")
	if err := os.MkdirAll(dockerDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	writeTestFile(t, filepath.Join(dockerDir, "complaint-network.yaml"))
	writeTestFile(t, filepath.Join(dockerDir, "complaint-apps.yaml"))

	gotFiles, gotLabel, err := resolveNetworkUpComposeFilesFromDir(tempDir, "basic", "complaint-network.yaml")
	if err != nil {
		t.Fatalf("resolveNetworkUpComposeFilesFromDir() error = %v", err)
	}

	wantFiles := []string{
		filepath.Join(dockerDir, "complaint-network.yaml"),
	}
	if !reflect.DeepEqual(gotFiles, wantFiles) {
		t.Fatalf("resolveNetworkUpComposeFilesFromDir() files = %v, want %v", gotFiles, wantFiles)
	}

	if gotLabel != "config: complaint-network.yaml" {
		t.Fatalf("resolveNetworkUpComposeFilesFromDir() label = %q, want %q", gotLabel, "config: complaint-network.yaml")
	}
}

func TestResolveNetworkAppsFileForGenericNetwork(t *testing.T) {
	tempDir := t.TempDir()
	dockerDir := filepath.Join(tempDir, "docker")
	if err := os.MkdirAll(dockerDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	composeFile := filepath.Join(dockerDir, "custom-network.yaml")
	appsFile := filepath.Join(dockerDir, "custom-apps.yaml")
	writeTestFile(t, composeFile)
	writeTestFile(t, appsFile)

	if got := resolveNetworkAppsFile(composeFile); got != appsFile {
		t.Fatalf("resolveNetworkAppsFile() = %q, want %q", got, appsFile)
	}
}

func TestComposeVariantCandidates(t *testing.T) {
	tests := []struct {
		name string
		file string
		want []string
	}{
		{name: "basic network", file: "/tmp/docker/fabric-network.yaml", want: []string{"fabric", "basic"}},
		{name: "complaint network", file: "/tmp/docker/complaint-network.yaml", want: []string{"complaint"}},
		{name: "custom network", file: "/tmp/docker/custom-network.yaml", want: []string{"custom"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := composeVariantCandidates(tt.file); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("composeVariantCandidates() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveCryptoConfigForCompose(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	writeTestFile(t, filepath.Join(configDir, "crypto-config.yaml"))
	writeTestFile(t, filepath.Join(configDir, "crypto-config-complaint.yaml"))
	writeTestFile(t, filepath.Join(configDir, "crypto-config-custom.yaml"))

	tests := []struct {
		name         string
		composeFiles []string
		wantBase     string
	}{
		{name: "basic", composeFiles: []string{filepath.Join(tempDir, "docker", "fabric-network.yaml")}, wantBase: "crypto-config.yaml"},
		{name: "complaint", composeFiles: []string{filepath.Join(tempDir, "docker", "complaint-network.yaml")}, wantBase: "crypto-config-complaint.yaml"},
		{name: "custom", composeFiles: []string{filepath.Join(tempDir, "docker", "custom-network.yaml")}, wantBase: "crypto-config-custom.yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveCryptoConfigForCompose(tempDir, tt.composeFiles)
			if err != nil {
				t.Fatalf("resolveCryptoConfigForCompose() error = %v", err)
			}
			if gotBase := filepath.Base(got); gotBase != tt.wantBase {
				t.Fatalf("resolveCryptoConfigForCompose() = %q, want %q", gotBase, tt.wantBase)
			}
		})
	}
}

func TestNextStepsForComposeComplaint(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	writeConfigtxTestFile(t, filepath.Join(configDir, "configtx-complaint.yaml"), `Profiles:
  ComplaintOrdererGenesis:
    Application: {}
  ComplaintChannel:
    Application:
      Organizations:
        - Name: ACBMSP
        - Name: DeptMSP
`)
	writeTestFile(t, filepath.Join(configDir, "crypto-config-complaint.yaml"))

	steps := nextStepsForCompose(tempDir, []string{filepath.Join(tempDir, "docker", "complaint-network.yaml")})
	if len(steps) != 2 {
		t.Fatalf("nextStepsForCompose() steps = %v, want 2 steps", steps)
	}
	if steps[0] != "nanayam channel create --name complaint-channel --profile ComplaintChannel" {
		t.Fatalf("nextStepsForCompose() channel step = %q", steps[0])
	}
	if steps[1] != "nanayam chaincode package --path ./chaincode/complaint-system --name complaint" {
		t.Fatalf("nextStepsForCompose() chaincode step = %q", steps[1])
	}
}

func TestEnsureComposePrerequisitesAutoRecovers(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	writeTestFile(t, filepath.Join(configDir, "crypto-config.yaml"))

	oldValidator := composePrerequisitesValidator
	oldGenerator := networkArtifactGenerator
	defer func() {
		composePrerequisitesValidator = oldValidator
		networkArtifactGenerator = oldGenerator
	}()

	validatorCalls := 0
	composePrerequisitesValidator = func(_ []string) error {
		validatorCalls++
		if validatorCalls == 1 {
			return fmt.Errorf("missing signcerts")
		}
		return nil
	}

	var generatedLabel string
	networkArtifactGenerator = func(_ string, _ []string) (string, error) {
		generatedLabel = "crypto-config.yaml"
		return generatedLabel, nil
	}

	recovered, artifactLabel, err := ensureComposePrerequisites(tempDir, []string{filepath.Join(tempDir, "docker", "fabric-network.yaml")})
	if err != nil {
		t.Fatalf("ensureComposePrerequisites() error = %v", err)
	}
	if !recovered {
		t.Fatal("ensureComposePrerequisites() recovered = false, want true")
	}
	if artifactLabel != "crypto-config.yaml" {
		t.Fatalf("ensureComposePrerequisites() artifactLabel = %q, want %q", artifactLabel, "crypto-config.yaml")
	}
	if generatedLabel != "crypto-config.yaml" {
		t.Fatalf("networkArtifactGenerator() label = %q, want %q", generatedLabel, "crypto-config.yaml")
	}
	if validatorCalls != 2 {
		t.Fatalf("composePrerequisitesValidator() calls = %d, want 2", validatorCalls)
	}
}

func TestEnsureComposePrerequisitesRefreshesWhenArtifactSourceMismatches(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	writeTestFile(t, filepath.Join(configDir, "crypto-config.yaml"))
	if err := os.MkdirAll(filepath.Join(tempDir, "channel-artifacts"), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(artifactSourceMetadataPath(tempDir), []byte("config/crypto-config-complaint.yaml\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	oldValidator := composePrerequisitesValidator
	oldGenerator := networkArtifactGenerator
	defer func() {
		composePrerequisitesValidator = oldValidator
		networkArtifactGenerator = oldGenerator
	}()

	composePrerequisitesValidator = func(_ []string) error { return nil }

	generatorCalled := false
	networkArtifactGenerator = func(_ string, _ []string) (string, error) {
		generatorCalled = true
		return "crypto-config.yaml", nil
	}

	recovered, artifactLabel, err := ensureComposePrerequisites(tempDir, []string{filepath.Join(tempDir, "docker", "fabric-network.yaml")})
	if err != nil {
		t.Fatalf("ensureComposePrerequisites() error = %v", err)
	}
	if !recovered {
		t.Fatal("ensureComposePrerequisites() recovered = false, want true")
	}
	if artifactLabel != "crypto-config.yaml" {
		t.Fatalf("ensureComposePrerequisites() artifactLabel = %q, want %q", artifactLabel, "crypto-config.yaml")
	}
	if !generatorCalled {
		t.Fatal("networkArtifactGenerator() not called, want true")
	}
}

func TestNetworkArtifactsNeedRefreshMatchesMetadata(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	writeTestFile(t, filepath.Join(configDir, "crypto-config.yaml"))
	if err := os.MkdirAll(filepath.Join(tempDir, "channel-artifacts"), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(artifactSourceMetadataPath(tempDir), []byte("config/crypto-config.yaml\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	refresh, label, err := networkArtifactsNeedRefresh(tempDir, []string{filepath.Join(tempDir, "docker", "fabric-network.yaml")})
	if err != nil {
		t.Fatalf("networkArtifactsNeedRefresh() error = %v", err)
	}
	if refresh {
		t.Fatal("networkArtifactsNeedRefresh() = true, want false")
	}
	if label != "crypto-config.yaml" {
		t.Fatalf("networkArtifactsNeedRefresh() label = %q, want %q", label, "crypto-config.yaml")
	}
}

func TestEnsureComposePrerequisitesRejectsUnknownCompose(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	writeTestFile(t, filepath.Join(configDir, "crypto-config-custom.yaml"))

	oldValidator := composePrerequisitesValidator
	oldGenerator := networkArtifactGenerator
	defer func() {
		composePrerequisitesValidator = oldValidator
		networkArtifactGenerator = oldGenerator
	}()

	composePrerequisitesValidator = func(_ []string) error {
		return fmt.Errorf("missing tls")
	}

	generatorCalled := false
	networkArtifactGenerator = func(_ string, _ []string) (string, error) {
		generatorCalled = true
		return "", fmt.Errorf("no matching crypto config")
	}

	recovered, artifactLabel, err := ensureComposePrerequisites(tempDir, []string{filepath.Join(tempDir, "docker", "custom-network.yaml")})
	if err == nil {
		t.Fatal("ensureComposePrerequisites() error = nil, want error")
	}
	if recovered {
		t.Fatal("ensureComposePrerequisites() recovered = true, want false")
	}
	if artifactLabel != "" {
		t.Fatalf("ensureComposePrerequisites() artifactLabel = %q, want empty", artifactLabel)
	}
	if !generatorCalled {
		t.Fatal("networkArtifactGenerator() not called, want true")
	}
	if !strings.Contains(err.Error(), "Automatic recovery failed") {
		t.Fatalf("ensureComposePrerequisites() error = %v, want automatic-recovery guidance", err)
	}
}

func writeTestFile(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("services: {}\n"), 0644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func writeConfigtxTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
