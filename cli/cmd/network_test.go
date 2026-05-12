package cmd

import (
	"os"
	"path/filepath"
	"reflect"
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
		filepath.Join(dockerDir, "complaint-apps.yaml"),
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

func writeTestFile(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("services: {}\n"), 0644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}