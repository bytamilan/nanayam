package cmd

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestParseSemver(t *testing.T) {
	cases := []struct {
		in         string
		wantOK     bool
		major      int
		minor      int
		patch      int
		prerelease string
	}{
		{"v1.2.3", true, 1, 2, 3, ""},
		{"1.2.3", true, 1, 2, 3, ""},
		{"v1.2", true, 1, 2, 0, ""},
		{"v1", true, 1, 0, 0, ""},
		{"v1.2.3-rc1", true, 1, 2, 3, "rc1"},
		{"v1.2.3+build7", true, 1, 2, 3, ""},
		{"v1.2.3-rc1+build7", true, 1, 2, 3, "rc1"},
		{"", false, 0, 0, 0, ""},
		{"v", false, 0, 0, 0, ""},
		{"dev", false, 0, 0, 0, ""},
		{"v1.2.3.4", false, 0, 0, 0, ""},
		{"v1..3", false, 0, 0, 0, ""},
		{"v-1.2.3", false, 0, 0, 0, ""},
		{"vx.y.z", false, 0, 0, 0, ""},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got := parseSemver(tc.in)
			if got.ok != tc.wantOK {
				t.Fatalf("parseSemver(%q).ok = %v, want %v", tc.in, got.ok, tc.wantOK)
			}
			if !tc.wantOK {
				return
			}
			if got.major != tc.major || got.minor != tc.minor || got.patch != tc.patch {
				t.Errorf("parseSemver(%q) = %d.%d.%d, want %d.%d.%d",
					tc.in, got.major, got.minor, got.patch, tc.major, tc.minor, tc.patch)
			}
			if got.prerelease != tc.prerelease {
				t.Errorf("parseSemver(%q).prerelease = %q, want %q", tc.in, got.prerelease, tc.prerelease)
			}
		})
	}
}

func TestCompareParsedSemver(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"v1.0.0", "v1.0.0", 0},
		{"v1.0.0", "v1.0.1", -1},
		{"v1.0.1", "v1.0.0", 1},
		{"v1.0.0", "v1.1.0", -1},
		{"v1.0.0", "v2.0.0", -1},
		{"v2.0.0", "v10.0.0", -1}, // numeric, not lexical
		{"v1.9.0", "v1.10.0", -1},
		// A prerelease sorts before its own release.
		{"v1.0.0-rc1", "v1.0.0", -1},
		{"v1.0.0", "v1.0.0-rc1", 1},
		{"v1.0.0-rc1", "v1.0.0-rc2", -1},
	}

	for _, tc := range cases {
		got := compareParsedSemver(parseSemver(tc.a), parseSemver(tc.b))
		if got != tc.want {
			t.Errorf("compare(%s, %s) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestCompareInts(t *testing.T) {
	if got := compareInts(1, 2); got != -1 {
		t.Errorf("compareInts(1, 2) = %d, want -1", got)
	}
	if got := compareInts(2, 1); got != 1 {
		t.Errorf("compareInts(2, 1) = %d, want 1", got)
	}
	if got := compareInts(3, 3); got != 0 {
		t.Errorf("compareInts(3, 3) = %d, want 0", got)
	}
}

func TestIsUpgradeAvailableNeverDowngrades(t *testing.T) {
	// The upgrade command must not walk a user backwards onto an older release.
	if isUpgradeAvailable("v2.0.0", "v1.0.0") {
		t.Error("isUpgradeAvailable(v2.0.0 -> v1.0.0) = true, want false")
	}
	if isUpgradeAvailable("v1.0.0", "v1.0.0") {
		t.Error("isUpgradeAvailable on an identical version = true, want false")
	}
	if !isUpgradeAvailable("v1.0.0", "v1.0.1") {
		t.Error("isUpgradeAvailable(v1.0.0 -> v1.0.1) = false, want true")
	}
}

func TestIsUpgradeAvailableForDevelopmentBuilds(t *testing.T) {
	// A source build reports "dev" and has no comparable version, so any named
	// release counts as an upgrade.
	if !isUpgradeAvailable("dev", "v1.0.0") {
		t.Error("dev -> v1.0.0 should be available")
	}
	if !isUpgradeAvailable("unknown", "v1.0.0") {
		t.Error("unknown -> v1.0.0 should be available")
	}
	if !isUpgradeAvailable("", "v1.0.0") {
		t.Error("empty -> v1.0.0 should be available")
	}
	if isUpgradeAvailable("dev", "dev") {
		t.Error("dev -> dev should not be available")
	}
}

func TestIsUpgradeAvailableFallsBackToInequality(t *testing.T) {
	// Neither side parses as semver, so the only safe signal is difference.
	if !isUpgradeAvailable("nightly-2024-01-01", "nightly-2024-02-01") {
		t.Error("differing unparseable versions should be treated as an upgrade")
	}
	if isUpgradeAvailable("nightly-2024-01-01", "nightly-2024-01-01") {
		t.Error("identical unparseable versions should not be an upgrade")
	}
}

func TestIsUpgradeAvailableIgnoresSurroundingWhitespace(t *testing.T) {
	if isUpgradeAvailable("  v1.0.0  ", "v1.0.0\n") {
		t.Error("whitespace should not make an identical version look like an upgrade")
	}
}

func TestReleaseAssetURLPointsAtTheTaggedRelease(t *testing.T) {
	asset := releaseAssetName("v1.2.3", "linux", "amd64")
	url := releaseAssetURL("v1.2.3", asset)

	if !strings.HasPrefix(url, "https://") {
		t.Errorf("release URL is not HTTPS: %q", url)
	}
	if !strings.Contains(url, "v1.2.3") {
		t.Errorf("release URL %q does not carry the version tag", url)
	}
	if !strings.HasSuffix(url, asset) {
		t.Errorf("release URL %q does not end with the asset name %q", url, asset)
	}
}

func TestReleaseAssetNameUsesZipOnWindows(t *testing.T) {
	if got := releaseAssetName("v1.2.3", "windows", "amd64"); !strings.HasSuffix(got, ".zip") {
		t.Errorf("windows asset = %q, want a .zip", got)
	}
	for _, goos := range []string{"linux", "darwin"} {
		if got := releaseAssetName("v1.2.3", goos, "arm64"); !strings.HasSuffix(got, ".tar.gz") {
			t.Errorf("%s asset = %q, want a .tar.gz", goos, got)
		}
	}
}

func TestRuntimeBinaryName(t *testing.T) {
	want := "nanayam"
	if runtime.GOOS == "windows" {
		want = "nanayam.exe"
	}
	if got := runtimeBinaryName(); got != want {
		t.Errorf("runtimeBinaryName() = %q, want %q", got, want)
	}
}

func TestTitleWord(t *testing.T) {
	cases := map[string]string{
		"install": "Install",
		"upgrade": "Upgrade",
		"A":       "A",
		"":        "",
	}
	for in, want := range cases {
		if got := titleWord(in); got != want {
			t.Errorf("titleWord(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFileExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "binary")
	if err := os.WriteFile(path, []byte("x"), 0o755); err != nil {
		t.Fatalf("write: %v", err)
	}

	if !fileExists(path) {
		t.Error("fileExists() = false for an existing file")
	}
	if fileExists(filepath.Join(dir, "absent")) {
		t.Error("fileExists() = true for a missing file")
	}
}

func TestSamePath(t *testing.T) {
	if !samePath("/usr/local/bin/nanayam", "/usr/local/bin/../bin/nanayam") {
		t.Error("samePath() did not normalise a traversal")
	}
	if !samePath("/usr/local/bin/nanayam", "/usr/local/bin/nanayam/") {
		t.Error("samePath() did not normalise a trailing separator")
	}
	if samePath("/usr/local/bin/nanayam", "/opt/bin/nanayam") {
		t.Error("samePath() = true for genuinely different paths")
	}
}

func TestCopyFilePreservesContentAndMode(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dest := filepath.Join(dir, "dest")

	if err := os.WriteFile(src, []byte("binary-content"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := copyFile(src, dest, 0o755); err != nil {
		t.Fatalf("copyFile() = %v", err)
	}

	content, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	if string(content) != "binary-content" {
		t.Errorf("content = %q", content)
	}

	info, err := os.Stat(dest)
	if err != nil {
		t.Fatalf("stat dest: %v", err)
	}
	// Without the executable bit the freshly installed CLI cannot be run.
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o111 == 0 {
		t.Errorf("mode = %v, want the executable bit set", info.Mode().Perm())
	}
}

func TestCopyFileRejectsMissingSource(t *testing.T) {
	dir := t.TempDir()
	if err := copyFile(filepath.Join(dir, "absent"), filepath.Join(dir, "dest"), 0o755); err == nil {
		t.Fatal("copyFile() = nil for a missing source, want an error")
	}
}

func TestExtractTarGzBinary(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "release.tar.gz")
	writeTarGz(t, archive, map[string]string{runtimeBinaryName(): "the-binary"})

	destDir := t.TempDir()
	got, err := extractTarGzBinary(archive, destDir)
	if err != nil {
		t.Fatalf("extractTarGzBinary() = %v", err)
	}

	content, err := os.ReadFile(got)
	if err != nil {
		t.Fatalf("read extracted binary: %v", err)
	}
	if string(content) != "the-binary" {
		t.Errorf("content = %q", content)
	}
}

func TestExtractTarGzBinaryFailsWhenArchiveHasNoBinary(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "release.tar.gz")
	writeTarGz(t, archive, map[string]string{"README.md": "not a binary"})

	if _, err := extractTarGzBinary(archive, t.TempDir()); err == nil {
		t.Fatal("extractTarGzBinary() = nil for an archive with no binary, want an error")
	}
}

func TestExtractZipBinary(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "release.zip")
	// The zip path only runs on Windows releases, where the archived binary is
	// named for the current platform.
	writeZip(t, archive, map[string]string{runtimeBinaryName(): "the-binary"})

	destDir := t.TempDir()
	got, err := extractZipBinary(archive, destDir)
	if err != nil {
		t.Fatalf("extractZipBinary() = %v", err)
	}

	content, err := os.ReadFile(got)
	if err != nil {
		t.Fatalf("read extracted binary: %v", err)
	}
	if string(content) != "the-binary" {
		t.Errorf("content = %q", content)
	}
}

func TestExtractZipBinaryFailsWhenArchiveHasNoBinary(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "release.zip")
	writeZip(t, archive, map[string]string{"README.md": "not a binary"})

	if _, err := extractZipBinary(archive, t.TempDir()); err == nil {
		t.Fatal("extractZipBinary() = nil for an archive with no binary, want an error")
	}
}

func TestExtractReleaseBinaryDispatchesOnExtension(t *testing.T) {
	tarPath := filepath.Join(t.TempDir(), "release.tar.gz")
	writeTarGz(t, tarPath, map[string]string{runtimeBinaryName(): "tar-binary"})
	if _, err := extractReleaseBinary(tarPath, t.TempDir()); err != nil {
		t.Errorf("extractReleaseBinary(.tar.gz) = %v", err)
	}

	zipPath := filepath.Join(t.TempDir(), "release.zip")
	writeZip(t, zipPath, map[string]string{runtimeBinaryName(): "zip-binary"})
	if _, err := extractReleaseBinary(zipPath, t.TempDir()); err != nil {
		t.Errorf("extractReleaseBinary(.zip) = %v", err)
	}
}

func TestGetShellConfigPath(t *testing.T) {
	home := t.TempDir()

	t.Setenv("SHELL", "/bin/zsh")
	if got, want := getShellConfigPath(home), filepath.Join(home, ".zshrc"); got != want {
		t.Errorf("zsh config = %q, want %q", got, want)
	}

	t.Setenv("SHELL", "/bin/bash")
	if got := getShellConfigPath(home); !strings.Contains(got, "bash") && !strings.HasSuffix(got, ".profile") {
		t.Errorf("bash config = %q, want a bash rc or .profile", got)
	}
}

func TestResolveLocalRepoRootFindsTheCheckout(t *testing.T) {
	root := repoRoot(t)

	// Called from the repository itself, resolution must land on the checkout
	// that actually holds cli/go.mod.
	got, err := resolveLocalRepoRoot(root)
	if err != nil {
		t.Fatalf("resolveLocalRepoRoot(%q) = %v", root, err)
	}
	if !samePath(got, root) {
		t.Fatalf("resolveLocalRepoRoot() = %q, want %q", got, root)
	}
}

func TestResolveLocalRepoRootRejectsUnrelatedDirectory(t *testing.T) {
	if _, err := resolveLocalRepoRoot(t.TempDir()); err == nil {
		t.Fatal("resolveLocalRepoRoot() = nil for a directory with no checkout, want an error")
	}
}

// ---------------------------------------------------------------------------
// Archive fixtures
// ---------------------------------------------------------------------------

func writeTarGz(t *testing.T, path string, entries map[string]string) {
	t.Helper()

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create archive: %v", err)
	}
	defer file.Close()

	gz := gzip.NewWriter(file)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()

	for name, content := range entries {
		header := &tar.Header{
			Name:     name,
			Mode:     0o755,
			Size:     int64(len(content)),
			Typeflag: tar.TypeReg,
		}
		if err := tw.WriteHeader(header); err != nil {
			t.Fatalf("write tar header: %v", err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatalf("write tar body: %v", err)
		}
	}
}

func writeZip(t *testing.T, path string, entries map[string]string) {
	t.Helper()

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create archive: %v", err)
	}
	defer file.Close()

	zw := zip.NewWriter(file)
	defer zw.Close()

	for name, content := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("create zip entry: %v", err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("write zip entry: %v", err)
		}
	}
}
