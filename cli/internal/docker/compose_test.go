package docker

import (
	"os"
	"path/filepath"
	"reflect"
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