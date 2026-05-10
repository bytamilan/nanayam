package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "nanayam",
	Short: "Nanayam — A Decentralized Private Web3 Ledger CLI",
	Long: `Nanayam is a private, permissioned Web3 ledger built on Hyperledger Fabric.

This CLI provides unified control over:
  • Node lifecycle (peer, orderer, CA, tools)
  • Network orchestration (up, down, status)
  • Channel management (create, join, update)
  • Identity management (user create, enroll)
  • Chaincode lifecycle (package, install, approve, commit)
  • Consortium connectivity (connect, join-channel)

Quick Start:
  nanayam prerequisites --auto    # Install missing dependencies
  nanayam network up              # Start the full Fabric network
  nanayam channel create --name mychannel
  nanayam user create --id alice --org Org1
`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.nanayam/config.yaml)")
	rootCmd.PersistentFlags().StringP("output", "o", "table", "Output format: table, json, yaml")
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		nanayamDir := filepath.Join(home, ".nanayam")
		os.MkdirAll(nanayamDir, 0755)

		viper.AddConfigPath(nanayamDir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix("NANAYAM")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
