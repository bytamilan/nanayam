package docker

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestComposeRunnerBuildArgsPlacesFilesBeforeSubcommand(t *testing.T) {
	runner := NewComposeRunner("docker/complaint-network.yaml", "docker/complaint-apps.yaml")

	got := runner.buildArgs("up", "-d")
	want := []string{
		"-f", "docker/complaint-network.yaml",
		"-f", "docker/complaint-apps.yaml",
		"up", "-d",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildArgs() = %v, want %v", got, want)
	}
}

func TestComposeRunnerBuildArgsWithoutFiles(t *testing.T) {
	runner := NewComposeRunner()

	got := runner.buildArgs("ps")
	want := []string{"ps"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildArgs() = %v, want %v", got, want)
	}
}

func TestResolveComposeFile(t *testing.T) {
	tempDir := t.TempDir()
	dockerDir := filepath.Join(tempDir, "docker")
	if err := os.MkdirAll(dockerDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	dockerConfig := filepath.Join(dockerDir, "fabric-network.yaml")
	if err := os.WriteFile(dockerConfig, []byte("services: {}\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	t.Run("filename under docker dir", func(t *testing.T) {
		got, err := ResolveComposeFile(tempDir, "fabric-network.yaml")
		if err != nil {
			t.Fatalf("ResolveComposeFile() error = %v", err)
		}
		if got != dockerConfig {
			t.Fatalf("ResolveComposeFile() = %q, want %q", got, dockerConfig)
		}
	})

	t.Run("relative path", func(t *testing.T) {
		got, err := ResolveComposeFile(tempDir, filepath.Join("docker", "fabric-network.yaml"))
		if err != nil {
			t.Fatalf("ResolveComposeFile() error = %v", err)
		}
		if got != dockerConfig {
			t.Fatalf("ResolveComposeFile() = %q, want %q", got, dockerConfig)
		}
	})

	t.Run("name without extension", func(t *testing.T) {
		got, err := ResolveComposeFile(tempDir, "fabric-network")
		if err != nil {
			t.Fatalf("ResolveComposeFile() error = %v", err)
		}
		if got != dockerConfig {
			t.Fatalf("ResolveComposeFile() = %q, want %q", got, dockerConfig)
		}
	})
}

func TestValidateComposePrerequisites(t *testing.T) {
	t.Run("passes for complete peer artifacts", func(t *testing.T) {
		tempDir := t.TempDir()
		composePath := filepath.Join(tempDir, "docker", "fabric-network.yaml")
		peerMSP := filepath.Join(tempDir, "crypto-config", "peerOrganizations", "org1.nanayam.com", "peers", "peer0.org1.nanayam.com", "msp")
		peerTLS := filepath.Join(tempDir, "crypto-config", "peerOrganizations", "org1.nanayam.com", "peers", "peer0.org1.nanayam.com", "tls")
		genesis := filepath.Join(tempDir, "channel-artifacts", "genesis.block")

		writeComposeFixture(t, composePath, "peer0.org1.nanayam.com", []string{
			"../crypto-config/peerOrganizations/org1.nanayam.com/peers/peer0.org1.nanayam.com/msp:/etc/hyperledger/fabric/msp",
			"../crypto-config/peerOrganizations/org1.nanayam.com/peers/peer0.org1.nanayam.com/tls:/etc/hyperledger/fabric/tls",
			"../channel-artifacts/genesis.block:/var/hyperledger/orderer/orderer.genesis.block",
		})
		writeMSPFixture(t, peerMSP)
		writeTLSFixture(t, peerTLS)
		writeFileFixture(t, genesis)

		if err := ValidateComposePrerequisites([]string{composePath}); err != nil {
			t.Fatalf("ValidateComposePrerequisites() error = %v", err)
		}
	})

	t.Run("fails for incomplete peer artifacts", func(t *testing.T) {
		tempDir := t.TempDir()
		composePath := filepath.Join(tempDir, "docker", "fabric-network.yaml")
		peerMSP := filepath.Join(tempDir, "crypto-config", "peerOrganizations", "org2.nanayam.com", "peers", "peer0.org2.nanayam.com", "msp")
		peerTLS := filepath.Join(tempDir, "crypto-config", "peerOrganizations", "org2.nanayam.com", "peers", "peer0.org2.nanayam.com", "tls")

		writeComposeFixture(t, composePath, "peer0.org2.nanayam.com", []string{
			"../crypto-config/peerOrganizations/org2.nanayam.com/peers/peer0.org2.nanayam.com/msp:/etc/hyperledger/fabric/msp",
			"../crypto-config/peerOrganizations/org2.nanayam.com/peers/peer0.org2.nanayam.com/tls:/etc/hyperledger/fabric/tls",
		})
		if err := os.MkdirAll(filepath.Join(peerMSP, "keystore"), 0755); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
		if err := os.MkdirAll(peerTLS, 0755); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}

		err := ValidateComposePrerequisites([]string{composePath})
		if err == nil {
			t.Fatal("ValidateComposePrerequisites() error = nil, want error")
		}
		if !strings.Contains(err.Error(), "missing signcerts") {
			t.Fatalf("ValidateComposePrerequisites() error = %v, want signcerts failure", err)
		}
		if !strings.Contains(err.Error(), "missing ca.crt") {
			t.Fatalf("ValidateComposePrerequisites() error = %v, want TLS failure", err)
		}
	})
}

func writeComposeFixture(t *testing.T, composePath, serviceName string, volumes []string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(composePath), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	var builder strings.Builder
	builder.WriteString("services:\n")
	builder.WriteString("  ")
	builder.WriteString(serviceName)
	builder.WriteString(":\n")
	builder.WriteString("    volumes:\n")
	for _, volume := range volumes {
		builder.WriteString("      - ")
		builder.WriteString(volume)
		builder.WriteString("\n")
	}

	if err := os.WriteFile(composePath, []byte(builder.String()), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func writeMSPFixture(t *testing.T, dir string) {
	t.Helper()
	signcerts := filepath.Join(dir, "signcerts")
	if err := os.MkdirAll(signcerts, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	writeFileFixture(t, filepath.Join(signcerts, "cert.pem"))
}

func writeTLSFixture(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	for _, name := range []string{"ca.crt", "server.crt", "server.key"} {
		writeFileFixture(t, filepath.Join(dir, name))
	}
}

func writeFileFixture(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte("fixture\n"), 0644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
