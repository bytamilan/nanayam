package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/bytamilan/nanayam/cli/internal/fabric"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cryptoCmd)
	cryptoCmd.AddCommand(cryptoGenerateCmd)
	cryptoCmd.AddCommand(cryptoRenewCmd)

	cryptoGenerateCmd.Flags().String("config", "", "Path to cryptogen config (default: config/crypto-config.yaml)")
	cryptoGenerateCmd.Flags().String("output", "crypto-config", "Output directory for crypto materials")
	cryptoGenerateCmd.Flags().Bool("channel-artifacts", true, "Also generate channel artifacts with configtxgen")
	cryptoRenewCmd.Flags().String("org", "", "Organization to renew certs for")
}

var cryptoCmd = &cobra.Command{
	Use:   "crypto",
	Short: "Manage cryptographic materials",
	Long:  `Generate and renew MSP certificates using cryptogen and Fabric CA.`,
}

var cryptoGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate crypto materials using cryptogen",
	Example: `  nanayam crypto generate
  nanayam crypto generate --config config/crypto-config-complaint.yaml --output crypto-config`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configFile, _ := cmd.Flags().GetString("config")
		output, _ := cmd.Flags().GetString("output")
		genChannel, _ := cmd.Flags().GetBool("channel-artifacts")

		bins := fabric.NewBinaries()
		if err := bins.CheckAll(); err != nil {
			return fmt.Errorf("fabric binaries not found: %w", err)
		}

		cwd, _ := os.Getwd()
		if configFile == "" {
			configFile = filepath.Join(cwd, "config", "crypto-config.yaml")
			if _, err := os.Stat(configFile); os.IsNotExist(err) {
				// Try complaint config
				configFile = filepath.Join(cwd, "config", "crypto-config-complaint.yaml")
			}
		}

		blue := color.New(color.FgBlue).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()

		fmt.Printf("%s Generating crypto materials...\n", blue("→"))
		fmt.Printf("  Config: %s\n", configFile)
		fmt.Printf("  Output: %s\n", output)

		cryptogen := exec.Command(bins.CryptogenPath(), "generate", "--config="+configFile, "--output="+output)
		cryptogen.Stdout = os.Stdout
		cryptogen.Stderr = os.Stderr
		if err := cryptogen.Run(); err != nil {
			return fmt.Errorf("cryptogen failed: %w", err)
		}
		fmt.Printf("%s Crypto materials generated in %s/\n", green("✓"), output)

		if genChannel {
			fmt.Println()
			fmt.Printf("%s Generating channel artifacts...\n", blue("→"))

			channelDir := filepath.Join(cwd, "channel-artifacts")
			os.MkdirAll(channelDir, 0755)

			// Detect which configtx to use
			configtxFile := filepath.Join(cwd, "config", "configtx.yaml")
			if _, err := os.Stat(configtxFile); os.IsNotExist(err) {
				configtxFile = filepath.Join(cwd, "config", "configtx-complaint.yaml")
			}

			// Set FABRIC_CFG_PATH for configtxgen
			env := append(os.Environ(), "FABRIC_CFG_PATH="+filepath.Dir(configtxFile))

			// Try basic profiles
			profiles := []struct {
				name      string
				genesis   string
				channel   string
				anchorOrgs []string
			}{
				{
					name:      "TwoOrgsOrdererGenesis",
					genesis:   "genesis.block",
					channel:   "mychannel",
					anchorOrgs: []string{"Org1MSP", "Org2MSP"},
				},
				{
					name:      "ComplaintOrdererGenesis",
					genesis:   "genesis.block",
					channel:   "complaint-channel",
					anchorOrgs: []string{"ACBMSP", "DeptMSP", "OversightMSP", "JudiciaryMSP"},
				},
			}

			for _, p := range profiles {
				genesisPath := filepath.Join(channelDir, p.genesis)
				if _, err := os.Stat(genesisPath); err == nil {
					continue // Already exists, skip
				}

				// Genesis block
				configtxgen := exec.Command(bins.ConfigtxgenPath(),
					"-profile", p.name,
					"-channelID", "system-channel",
					"-outputBlock", genesisPath)
				configtxgen.Env = env
				configtxgen.Stdout = os.Stdout
				configtxgen.Stderr = os.Stderr
				if err := configtxgen.Run(); err != nil {
					continue // Profile might not exist, try next
				}
				fmt.Printf("  %s Genesis block: %s\n", green("✓"), genesisPath)

				// Channel tx
				channelTx := filepath.Join(channelDir, p.channel+".tx")
				channelProfile := p.name
				if p.name == "TwoOrgsOrdererGenesis" {
					channelProfile = "TwoOrgsChannel"
				} else if p.name == "ComplaintOrdererGenesis" {
					channelProfile = "ComplaintChannel"
				}
				configtxgen = exec.Command(bins.ConfigtxgenPath(),
					"-profile", channelProfile,
					"-outputCreateChannelTx", channelTx,
					"-channelID", p.channel)
				configtxgen.Env = env
				configtxgen.Stdout = os.Stdout
				configtxgen.Stderr = os.Stderr
				if err := configtxgen.Run(); err == nil {
					fmt.Printf("  %s Channel tx: %s\n", green("✓"), channelTx)
				}

				// Anchor peer updates
				for _, org := range p.anchorOrgs {
					anchorTx := filepath.Join(channelDir, org+"anchors.tx")
					configtxgen = exec.Command(bins.ConfigtxgenPath(),
						"-profile", channelProfile,
						"-outputAnchorPeersUpdate", anchorTx,
						"-channelID", p.channel,
						"-asOrg", org)
					configtxgen.Env = env
					configtxgen.Stdout = os.Stdout
					configtxgen.Stderr = os.Stderr
					if err := configtxgen.Run(); err == nil {
						fmt.Printf("  %s Anchor peers: %s\n", green("✓"), anchorTx)
					}
				}
			}
		}

		fmt.Println()
		fmt.Printf("%s Crypto generation complete!\n", green("✓"))
		return nil
	},
}

var cryptoRenewCmd = &cobra.Command{
	Use:   "renew",
	Short: "Renew certificates for an organization",
	RunE: func(cmd *cobra.Command, args []string) error {
		org, _ := cmd.Flags().GetString("org")
		fmt.Printf("Renewing certificates for org %s...\n", org)
		fmt.Println("(Implementation coming in Phase 3)")
		return nil
	},
}
