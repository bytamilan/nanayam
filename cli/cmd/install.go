package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(installCmd)
	addSelfUpdateFlags(installCmd, false)
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the nanayam CLI from release builds or local source",
	Long: `Install the nanayam CLI to ~/.nanayam/bin/ and add it to your PATH.

This command is primarily used by the install script:
  curl -fsSL https://raw.githubusercontent.com/bytamilan/nanayam/main/install.sh | bash

You can also run it directly to install a specific release or a local dev build.`,
	RunE: runInstallFlow,
}
