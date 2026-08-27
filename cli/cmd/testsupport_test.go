package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// repoRoot walks up from the test's working directory until it finds the
// repository root, identified by the checked-in config/configtx.yaml.
//
// Tests must never hardcode a developer's absolute path: doing so makes the
// suite pass on one machine and fail everywhere else, CI included.
func repoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "config", "configtx.yaml")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("repository root not found above working directory")
		}
		dir = parent
	}
}
