package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(upgradeCmd)
	addSelfUpdateFlags(upgradeCmd, true)
}

var upgradeCmd = &cobra.Command{
	Use:     "upgrade",
	Aliases: []string{"update"},
	Short:   "Check for and install newer Nanayam CLI releases",
	Long: `Check the latest Nanayam CLI release and install it when a newer version exists.

Useful examples:
  nanayam upgrade --check
  nanayam upgrade
  nanayam upgrade --refresh
  nanayam upgrade --dev-local --source /path/to/nanayam
  nanayam upgrade --version v1.2.3`,
	RunE: runUpgradeFlow,
}
