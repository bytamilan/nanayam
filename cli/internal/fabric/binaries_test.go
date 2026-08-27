package fabric

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// writeFakeBinaries drops empty files named like the Fabric tools into dir.
func writeFakeBinaries(t *testing.T, dir string, names ...string) {
	t.Helper()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create bin dir: %v", err)
	}
	for _, name := range names {
		path := filepath.Join(dir, binaryName(name))
		if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatalf("write fake binary %s: %v", path, err)
		}
	}
}

func TestBinaryNameAppendsExeOnWindows(t *testing.T) {
	got := binaryName("cryptogen")

	want := "cryptogen"
	if runtime.GOOS == "windows" {
		want = "cryptogen.exe"
	}
	if got != want {
		t.Fatalf("binaryName(cryptogen) = %q, want %q", got, want)
	}
}

func TestBinaryPathsAllLiveUnderBinDir(t *testing.T) {
	b := &Binaries{BinDir: filepath.Join("/opt", "fabric", "bin")}

	paths := map[string]string{
		"peer":             b.PeerPath(),
		"cryptogen":        b.CryptogenPath(),
		"configtxgen":      b.ConfigtxgenPath(),
		"fabric-ca-client": b.CAClientPath(),
	}

	for tool, path := range paths {
		if dir := filepath.Dir(path); dir != b.BinDir {
			t.Errorf("%s path %q is not under BinDir %q", tool, path, b.BinDir)
		}
		if base := filepath.Base(path); base != binaryName(tool) {
			t.Errorf("%s path base = %q, want %q", tool, base, binaryName(tool))
		}
	}
}

func TestCheckAllSucceedsWhenEveryBinaryIsPresent(t *testing.T) {
	dir := t.TempDir()
	writeFakeBinaries(t, dir, requiredBinaries...)

	if err := (&Binaries{BinDir: dir}).CheckAll(); err != nil {
		t.Fatalf("CheckAll() = %v, want nil", err)
	}
}

func TestCheckAllNamesTheMissingBinary(t *testing.T) {
	for _, missing := range requiredBinaries {
		t.Run(missing, func(t *testing.T) {
			dir := t.TempDir()
			var present []string
			for _, name := range requiredBinaries {
				if name != missing {
					present = append(present, name)
				}
			}
			writeFakeBinaries(t, dir, present...)

			err := (&Binaries{BinDir: dir}).CheckAll()
			if err == nil {
				t.Fatalf("CheckAll() = nil, want an error naming %s", missing)
			}
			if !strings.Contains(err.Error(), missing) {
				t.Errorf("error %q does not name the missing binary %q", err, missing)
			}
			// The message must say where it looked, or the user cannot act on it.
			if !strings.Contains(err.Error(), dir) {
				t.Errorf("error %q does not report the search directory %q", err, dir)
			}
		})
	}
}

func TestCheckAllFailsOnEmptyDirectory(t *testing.T) {
	if err := (&Binaries{BinDir: t.TempDir()}).CheckAll(); err == nil {
		t.Fatal("CheckAll() = nil for an empty directory, want an error")
	}
}

func TestHasFabricBinaries(t *testing.T) {
	complete := t.TempDir()
	writeFakeBinaries(t, complete, requiredBinaries...)
	if !hasFabricBinaries(complete) {
		t.Error("hasFabricBinaries() = false for a complete directory")
	}

	partial := t.TempDir()
	writeFakeBinaries(t, partial, "peer", "cryptogen")
	if hasFabricBinaries(partial) {
		t.Error("hasFabricBinaries() = true for a partial directory")
	}

	if hasFabricBinaries(filepath.Join(t.TempDir(), "does-not-exist")) {
		t.Error("hasFabricBinaries() = true for a missing directory")
	}
}

func TestNewBinariesPrefersProjectLocalBin(t *testing.T) {
	project := t.TempDir()
	localBin := filepath.Join(project, "bin")
	writeFakeBinaries(t, localBin, requiredBinaries...)

	t.Chdir(project)

	if got := NewBinaries().BinDir; got != localBin {
		t.Fatalf("NewBinaries().BinDir = %q, want the project-local %q", got, localBin)
	}
}

func TestNewBinariesFallsBackToProjectLocalBinWhenNothingIsInstalled(t *testing.T) {
	project := t.TempDir()
	// An empty HOME and PATH means neither ~/.nanayam/fabric-bin nor PATH resolve.
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", t.TempDir())
	t.Setenv("PATH", t.TempDir())
	t.Chdir(project)

	// The fallback is where `nanayam prerequisites` installs to, so the CLI can
	// report a directory the user can actually populate.
	if got, want := NewBinaries().BinDir, filepath.Join(project, "bin"); got != want {
		t.Fatalf("NewBinaries().BinDir = %q, want %q", got, want)
	}
}

func TestNewBinariesFindsNanayamHomeInstall(t *testing.T) {
	home := t.TempDir()
	fabricBin := filepath.Join(home, ".nanayam", "fabric-bin")
	writeFakeBinaries(t, fabricBin, requiredBinaries...)

	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("PATH", t.TempDir())
	// A project without its own bin/ must fall through to the home install.
	t.Chdir(t.TempDir())

	if got := NewBinaries().BinDir; got != fabricBin {
		t.Fatalf("NewBinaries().BinDir = %q, want %q", got, fabricBin)
	}
}

func TestDownloadURLsPointAtPinnedReleases(t *testing.T) {
	// The versions are part of the install contract; a silent bump would fetch
	// binaries the compose files and configtx profiles were not tested against.
	if FabricVersion == "" || CAVersion == "" {
		t.Fatal("Fabric and CA versions must be pinned")
	}
	if strings.HasPrefix(FabricVersion, "v") || strings.HasPrefix(CAVersion, "v") {
		t.Errorf("versions must be bare semver (got %q / %q); the v prefix is added at the URL", FabricVersion, CAVersion)
	}
}
