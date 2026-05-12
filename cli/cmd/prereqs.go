package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(prereqsCmd)
	prereqsCmd.Flags().Bool("auto", false, "Automatically install missing prerequisites without prompting")
	prereqsCmd.Flags().Bool("install-fabric", false, "Also download Fabric binaries (peer, cryptogen, configtxgen, fabric-ca-client)")
}

var prereqsCmd = &cobra.Command{
	Use:   "prerequisites",
	Short: "Check and install required dependencies",
	Long: `Check if all required tools are installed and optionally auto-install them.

Checked dependencies:
  - Docker & Docker Compose
  - Go (>= 1.21)
  - Node.js (>= 18) — for operator console
  - jq
  - curl
  - git`,
	RunE: runPrereqs,
}

type dep struct {
	Name        string
	Command     string
	Args        []string
	Check       func() bool
	InstallHint string
	AutoInstall func() error
}

func runPrereqs(cmd *cobra.Command, args []string) error {
	auto, _ := cmd.Flags().GetBool("auto")
	installFabric, _ := cmd.Flags().GetBool("install-fabric")

	deps := []dep{
		{
			Name:        "Docker",
			Command:     "docker",
			Args:        []string{"--version"},
			InstallHint: "Install Docker Desktop from https://docs.docker.com/get-docker/",
			AutoInstall: nil, // Docker requires manual install or elevated permissions
		},
		{
			Name:        "Docker Compose",
			Command:     "docker",
			Args:        []string{"compose", "version"},
			InstallHint: "Included with Docker Desktop; on Linux install docker-compose-plugin",
			AutoInstall: nil,
		},
		{
			Name:        "curl",
			Command:     "curl",
			Args:        []string{"--version"},
			InstallHint: "Usually pre-installed. Use your package manager if missing.",
			AutoInstall: nil,
		},
		{
			Name:        "jq",
			Command:     "jq",
			Args:        []string{"--version"},
			InstallHint: installHintForPkgManager("jq"),
			AutoInstall: autoInstallJq,
		},
		{
			Name:        "Go",
			Command:     "go",
			Args:        []string{"version"},
			InstallHint: "Download from https://go.dev/dl/ or use your package manager",
			AutoInstall: nil,
		},
		{
			Name:        "Node.js",
			Command:     "node",
			Args:        []string{"--version"},
			InstallHint: "Download from https://nodejs.org/ or use nvm",
			AutoInstall: nil,
		},
		{
			Name:        "git",
			Command:     "git",
			Args:        []string{"--version"},
			InstallHint: "Usually pre-installed. Use your package manager if missing.",
			AutoInstall: nil,
		},
	}

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	allOK := true
	for _, d := range deps {
		if checkDep(d.Command, d.Args) {
			fmt.Printf("  %s %s\n", green("✓"), d.Name)
		} else {
			fmt.Printf("  %s %s — %s\n", red("✗"), d.Name, d.InstallHint)
			allOK = false
			if auto && d.AutoInstall != nil {
				fmt.Printf("    %s Auto-installing %s...\n", yellow("→"), d.Name)
				if err := d.AutoInstall(); err != nil {
					fmt.Printf("    %s Failed to auto-install %s: %v\n", red("✗"), d.Name, err)
				} else {
					fmt.Printf("    %s %s installed\n", green("✓"), d.Name)
				}
			}
		}
	}

	if installFabric {
		fmt.Println()
		fmt.Println("Downloading Fabric binaries...")
		if err := downloadFabricBinaries(); err != nil {
			fmt.Printf("  %s Failed: %v\n", red("✗"), err)
			allOK = false
		} else {
			fmt.Printf("  %s Fabric binaries downloaded\n", green("✓"))
		}
	}

	if !allOK {
		return fmt.Errorf("some prerequisites are missing")
	}
	fmt.Println()
	fmt.Println(green("All prerequisites satisfied!"))
	return nil
}

func checkDep(command string, args []string) bool {
	cmd := exec.Command(command, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

func installHintForPkgManager(pkg string) string {
	switch runtime.GOOS {
	case "darwin":
		return fmt.Sprintf("brew install %s", pkg)
	case "linux":
		if _, err := exec.LookPath("apt-get"); err == nil {
			return fmt.Sprintf("sudo apt-get install -y %s", pkg)
		}
		if _, err := exec.LookPath("yum"); err == nil {
			return fmt.Sprintf("sudo yum install -y %s", pkg)
		}
		return fmt.Sprintf("Install %s via your package manager", pkg)
	default:
		return fmt.Sprintf("Install %s via your package manager", pkg)
	}
}

func autoInstallJq() error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("brew", "install", "jq").Run()
	case "linux":
		if _, err := exec.LookPath("apt-get"); err == nil {
			return exec.Command("sudo", "apt-get", "install", "-y", "jq").Run()
		}
		if _, err := exec.LookPath("yum"); err == nil {
			return exec.Command("sudo", "yum", "install", "-y", "jq").Run()
		}
		return fmt.Errorf("no supported package manager found")
	default:
		return fmt.Errorf("auto-install not supported on %s", runtime.GOOS)
	}
}

func downloadFabricBinaries() error {
	fabricVersion := "2.5.9"
	caVersion := "1.5.12"

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	binDir := fmt.Sprintf("%s/.nanayam/fabric-bin", home)
	os.MkdirAll(binDir, 0755)

	platform := runtime.GOOS
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "amd64"
	} else if arch == "arm64" {
		arch = "arm64"
	}

	// Download Fabric binaries
	fabricURL := fmt.Sprintf("https://github.com/hyperledger/fabric/releases/download/v%s/hyperledger-fabric-%s-%s-%s.tar.gz", fabricVersion, platform, arch, fabricVersion)
	if err := downloadAndExtract(fabricURL, binDir); err != nil {
		return fmt.Errorf("fabric binaries: %w", err)
	}

	// Download CA binaries
	caURL := fmt.Sprintf("https://github.com/hyperledger/fabric-ca/releases/download/v%s/hyperledger-fabric-ca-%s-%s-%s.tar.gz", caVersion, platform, arch, caVersion)
	if err := downloadAndExtract(caURL, binDir); err != nil {
		return fmt.Errorf("fabric-ca binaries: %w", err)
	}

	// Make executable
	files, _ := os.ReadDir(binDir)
	for _, f := range files {
		if !f.IsDir() {
			os.Chmod(fmt.Sprintf("%s/%s", binDir, f.Name()), 0755)
		}
	}

	return nil
}

func downloadAndExtract(url, destDir string) error {
	tmpFile := "/tmp/nanayam-download.tar.gz"
	cmd := exec.Command("curl", "-fsSL", "-o", tmpFile, url)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("curl: %s: %w", strings.TrimSpace(string(out)), err)
	}
	extractDir := "/tmp/nanayam-extract"
	os.RemoveAll(extractDir)
	os.MkdirAll(extractDir, 0755)
	if out, err := exec.Command("tar", "-xzf", tmpFile, "-C", extractDir).CombinedOutput(); err != nil {
		return fmt.Errorf("tar: %s: %w", strings.TrimSpace(string(out)), err)
	}
	// Move bin contents
	entries, _ := os.ReadDir(fmt.Sprintf("%s/bin", extractDir))
	for _, e := range entries {
		src := fmt.Sprintf("%s/bin/%s", extractDir, e.Name())
		dst := fmt.Sprintf("%s/%s", destDir, e.Name())
		os.Rename(src, dst)
	}
	return nil
}
