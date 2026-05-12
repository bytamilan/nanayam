package fabric

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const (
	FabricVersion = "2.5.9"
	CAVersion     = "1.5.12"
)

// Binaries holds paths to Fabric CLI tools
type Binaries struct {
	BinDir string
}

// NewBinaries finds Fabric binaries in common locations
func NewBinaries() *Binaries {
	// Check project-local bin first
	cwd, _ := os.Getwd()
	localBin := filepath.Join(cwd, "bin")
	if hasFabricBinaries(localBin) {
		return &Binaries{BinDir: localBin}
	}

	// Check ~/.nanayam/fabric-bin
	home, _ := os.UserHomeDir()
	nanayamBin := filepath.Join(home, ".nanayam", "fabric-bin")
	if hasFabricBinaries(nanayamBin) {
		return &Binaries{BinDir: nanayamBin}
	}

	// Check PATH
	if path, err := exec.LookPath("peer"); err == nil {
		return &Binaries{BinDir: filepath.Dir(path)}
	}

	return &Binaries{BinDir: localBin}
}

func hasFabricBinaries(dir string) bool {
	for _, bin := range []string{"peer", "cryptogen", "configtxgen", "fabric-ca-client"} {
		if _, err := os.Stat(filepath.Join(dir, bin)); os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// PeerPath returns the path to the peer binary
func (b *Binaries) PeerPath() string {
	return filepath.Join(b.BinDir, "peer")
}

// CryptogenPath returns the path to cryptogen
func (b *Binaries) CryptogenPath() string {
	return filepath.Join(b.BinDir, "cryptogen")
}

// ConfigtxgenPath returns the path to configtxgen
func (b *Binaries) ConfigtxgenPath() string {
	return filepath.Join(b.BinDir, "configtxgen")
}

// CAClientPath returns the path to fabric-ca-client
func (b *Binaries) CAClientPath() string {
	return filepath.Join(b.BinDir, "fabric-ca-client")
}

// CheckAll verifies all required binaries exist
func (b *Binaries) CheckAll() error {
	for _, name := range []string{"peer", "cryptogen", "configtxgen", "fabric-ca-client"} {
		path := filepath.Join(b.BinDir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("missing Fabric binary: %s (looked in %s)", name, b.BinDir)
		}
	}
	return nil
}

// Download downloads Fabric binaries to the given directory
func Download(destDir string) error {
	os.MkdirAll(destDir, 0755)

	platform := runtime.GOOS
	arch := runtime.GOARCH

	fabricURL := fmt.Sprintf("https://github.com/hyperledger/fabric/releases/download/v%s/hyperledger-fabric-%s-%s-%s.tar.gz", FabricVersion, platform, arch, FabricVersion)
	if err := downloadAndExtract(fabricURL, destDir); err != nil {
		return fmt.Errorf("download fabric binaries: %w", err)
	}

	caURL := fmt.Sprintf("https://github.com/hyperledger/fabric-ca/releases/download/v%s/hyperledger-fabric-ca-%s-%s-%s.tar.gz", CAVersion, platform, arch, CAVersion)
	if err := downloadAndExtract(caURL, destDir); err != nil {
		return fmt.Errorf("download fabric-ca binaries: %w", err)
	}

	// Make executable
	entries, _ := os.ReadDir(destDir)
	for _, e := range entries {
		if !e.IsDir() {
			os.Chmod(filepath.Join(destDir, e.Name()), 0755)
		}
	}

	return nil
}

func downloadAndExtract(url, destDir string) error {
	tmpFile := filepath.Join(os.TempDir(), "nanayam-download.tar.gz")
	cmd := exec.Command("curl", "-fsSL", "-o", tmpFile, url)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("curl: %s: %w", string(out), err)
	}

	extractDir := filepath.Join(os.TempDir(), "nanayam-extract")
	os.RemoveAll(extractDir)
	os.MkdirAll(extractDir, 0755)

	cmd = exec.Command("tar", "-xzf", tmpFile, "-C", extractDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tar: %s: %w", string(out), err)
	}

	binDir := filepath.Join(extractDir, "bin")
	entries, _ := os.ReadDir(binDir)
	for _, e := range entries {
		src := filepath.Join(binDir, e.Name())
		dst := filepath.Join(destDir, e.Name())
		os.Rename(src, dst)
	}
	return nil
}
