package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print nanayam and Fabric version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("nanayam version %s (commit: %s, built: %s)\n", version, commit, date)
		fmt.Println("Hyperledger Fabric: 2.5.9")
		fmt.Println("Fabric CA: 1.5.12")
	},
}
