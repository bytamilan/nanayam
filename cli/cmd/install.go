package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().String("version", "", "Install a specific version (default: latest)")
	installCmd.Flags().Bool("with-fabric", false, "Also download Fabric binaries")
	installCmd.Flags().Bool("setup", false, "Run prerequisites --auto after install")
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install or update the nanayam CLI",
	Long: `Install the nanayam CLI to ~/.nanayam/bin/ and add it to your PATH.

This command is primarily used by the install script:
  curl -fsSL https://nanayam.io/install.sh | bash

You can also run it directly to self-update.`,
	RunE: runInstall,
}

func runInstall(cmd *cobra.Command, args []string) error {
	ver, _ := cmd.Flags().GetString("version")
	withFabric, _ := cmd.Flags().GetBool("with-fabric")
	setup, _ := cmd.Flags().GetBool("setup")

	if ver == "" {
		ver = "latest"
	}

	fmt.Printf("Installing nanayam %s for %s/%s...\n", ver, runtime.GOOS, runtime.GOARCH)

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	binDir := fmt.Sprintf("%s/.nanayam/bin", home)
	os.MkdirAll(binDir, 0755)

	// In dev mode, just symlink the current binary
	exe, _ := os.Executable()
	if exe != "" && version == "dev" {
		target := fmt.Sprintf("%s/nanayam", binDir)
		os.Remove(target)
		if err := os.Symlink(exe, target); err != nil {
			return fmt.Errorf("symlink: %w", err)
		}
		fmt.Println("Linked dev binary to", target)
	} else {
		fmt.Println("Download from GitHub Releases would happen here.")
		fmt.Printf("  OS: %s, Arch: %s, Version: %s\n", runtime.GOOS, runtime.GOARCH, ver)
	}

	// Add to PATH if not present
	shellCfg := getShellConfigPath(home)
	if shellCfg != "" {
		pathEntry := fmt.Sprintf("export PATH=\"$HOME/.nanayam/bin:$PATH\"\n")
		data, _ := os.ReadFile(shellCfg)
		if data != nil && !contains(string(data), ".nanayam/bin") {
			f, _ := os.OpenFile(shellCfg, os.O_APPEND|os.O_WRONLY, 0644)
			f.WriteString("\n# Nanayam CLI\n" + pathEntry)
			f.Close()
			fmt.Printf("Added ~/.nanayam/bin to PATH in %s\n", shellCfg)
			fmt.Println("Run 'source " + shellCfg + "' or restart your terminal to apply.")
		}
	}

	if withFabric {
		fmt.Println()
		fmt.Println("Downloading Fabric binaries...")
		if err := downloadFabricBinaries(); err != nil {
			return err
		}
		fmt.Println("Fabric binaries installed.")
	}

	if setup {
		fmt.Println()
		fmt.Println("Running prerequisites setup...")
		prereqsCmd.Run(cmd, args)
	}

	fmt.Println()
	fmt.Println("Installation complete!")
	fmt.Println("Run 'nanayam version' to verify.")
	return nil
}

func getShellConfigPath(home string) string {
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "zsh") {
		return fmt.Sprintf("%s/.zshrc", home)
	}
	return fmt.Sprintf("%s/.bashrc", home)
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
