package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/bytamilan/nanayam/cli/internal/config"
	"github.com/bytamilan/nanayam/cli/internal/fabric"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(consortiumCmd)
	consortiumCmd.AddCommand(consortiumConnectCmd)
	consortiumCmd.AddCommand(consortiumJoinChannelCmd)

	consortiumConnectCmd.Flags().String("orderer", "", "Orderer endpoint (required)")
	consortiumConnectCmd.Flags().String("tls-cert", "", "Path to orderer TLS CA cert (required)")
	consortiumConnectCmd.Flags().String("genesis", "", "Path to genesis block")
	consortiumConnectCmd.Flags().String("org", "", "Local organization name (required)")
	consortiumConnectCmd.Flags().String("domain", "", "Local organization domain (required)")
	consortiumConnectCmd.Flags().String("crypto-output", "crypto-config", "Output directory for crypto materials")
	consortiumConnectCmd.MarkFlagRequired("orderer")
	consortiumConnectCmd.MarkFlagRequired("tls-cert")
	consortiumConnectCmd.MarkFlagRequired("org")
	consortiumConnectCmd.MarkFlagRequired("domain")

	consortiumJoinChannelCmd.Flags().String("name", "", "Channel ID (required)")
	consortiumJoinChannelCmd.Flags().String("block", "", "Path to channel genesis block")
	consortiumJoinChannelCmd.Flags().String("from-peer", "", "Peer endpoint to fetch block from")
	consortiumJoinChannelCmd.Flags().String("peer", "", "Local peer endpoint to join")
	consortiumJoinChannelCmd.MarkFlagRequired("name")
}

var consortiumCmd = &cobra.Command{
	Use:   "consortium",
	Short: "Connect to an existing consortium",
	Long: `Join an existing Fabric consortium by connecting to its orderer,
generating local org crypto, and updating the channel config to add your org.`,
}

var consortiumConnectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect local org to an existing consortium orderer",
	Example: `  nanayam consortium connect \
    --orderer orderer.example.com:7050 \
    --tls-cert ./orderer-tls-ca.crt \
    --org NewOrg \
    --domain neworg.example.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		orderer, _ := cmd.Flags().GetString("orderer")
		tlsCert, _ := cmd.Flags().GetString("tls-cert")
		genesis, _ := cmd.Flags().GetString("genesis")
		org, _ := cmd.Flags().GetString("org")
		domain, _ := cmd.Flags().GetString("domain")
		cryptoOutput, _ := cmd.Flags().GetString("crypto-output")

		blue := color.New(color.FgBlue).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()

		fmt.Printf("%s Connecting org '%s' to consortium orderer %s...\n", blue("→"), org, orderer)
		fmt.Printf("  Domain: %s\n", domain)
		fmt.Printf("  TLS Cert: %s\n", tlsCert)

		bins := fabric.NewBinaries()
		if err := bins.CheckAll(); err != nil {
			return fmt.Errorf("fabric binaries not found: %w", err)
		}

		// Step 1: Generate crypto for the new org
		fmt.Println()
		fmt.Printf("%s Step 1: Generating crypto materials for %s...\n", blue("→"), org)
		cwd, _ := os.Getwd()
		cryptoCfgPath := filepath.Join(cwd, "config", fmt.Sprintf("crypto-config-%s.yaml", org))
		peerCfg := &config.PeerNodeConfig{
			OrgName:   org,
			OrgDomain: domain,
			Domain:    "nanayam.com",
			PeerCount: 1,
			UserCount: 1,
		}
		if err := config.GenerateCryptogen(peerCfg, cryptoCfgPath); err != nil {
			return fmt.Errorf("generate cryptogen config: %w", err)
		}
		fmt.Printf("  %s Cryptogen config: %s\n", green("✓"), cryptoCfgPath)

		cryptoOut := filepath.Join(cwd, cryptoOutput)
		// Actually run cryptogen
		fmt.Printf("  %s Running cryptogen...\n", blue("→"))
		cryptogenCmd := exec.Command(bins.CryptogenPath(), "generate", "--config="+cryptoCfgPath, "--output="+cryptoOut)
		cryptogenCmd.Stdout = os.Stdout
		cryptogenCmd.Stderr = os.Stderr
		if err := cryptogenCmd.Run(); err != nil {
			return fmt.Errorf("cryptogen failed: %w", err)
		}
		fmt.Printf("  %s Crypto materials generated\n", green("✓"))

		// Step 2: Fetch channel config (if genesis not provided)
		if genesis == "" {
			fmt.Println()
			fmt.Printf("%s Step 2: Fetching system channel config from orderer...\n", blue("→"))
			fmt.Printf("  %s Use --genesis flag to provide a genesis block directly\n", yellow("⚠"))
		}

		// Step 3: Generate peer compose
		fmt.Println()
		fmt.Printf("%s Step 3: Generating peer docker-compose...\n", blue("→"))
		peerID := fmt.Sprintf("peer0.%s", domain)
		peerCompose := &config.PeerNodeConfig{
			PeerID:        peerID,
			PeerPort:      7051,
			ChaincodePort: 7052,
			OpsPort:       9444,
			MSPID:         fmt.Sprintf("%sMSP", org),
			CryptoPath:    filepath.Join(cryptoOut, "peerOrganizations", domain),
		}
		composePath := filepath.Join(cwd, "docker", fmt.Sprintf("%s.yaml", peerID))
		if err := config.GeneratePeerCompose(peerCompose, composePath); err != nil {
			return fmt.Errorf("generate peer compose: %w", err)
		}
		fmt.Printf("  %s Peer compose: %s\n", green("✓"), composePath)

		fmt.Println()
		fmt.Printf("%s Consortium connection prepared!\n", green("✓"))
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("  1. Share your org's MSP with existing consortium members")
		fmt.Println("  2. Ask them to update the channel config to add your org")
		fmt.Println("  3. Start your peer: nanayam node start", peerID)
		fmt.Println("  4. Join the channel: nanayam consortium join-channel --name <channel>")
		return nil
	},
}

var consortiumJoinChannelCmd = &cobra.Command{
	Use:     "join-channel",
	Short:   "Join an existing channel from a consortium",
	Example: `  nanayam consortium join-channel --name mychannel --block ./mychannel.block`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		block, _ := cmd.Flags().GetString("block")
		fromPeer, _ := cmd.Flags().GetString("from-peer")
		peer, _ := cmd.Flags().GetString("peer")

		blue := color.New(color.FgBlue).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()

		if block == "" {
			block = name + ".block"
		}

		fmt.Printf("%s Joining channel '%s'...\n", blue("→"), name)
		if fromPeer != "" {
			fmt.Printf("  Fetching block from peer: %s\n", fromPeer)
		}

		bins := fabric.NewBinaries()
		env := fabric.DefaultPeerEnv("Org1")
		if peer != "" {
			env.Address = peer
		}

		peerCmd := env.Exec(bins.PeerPath(), "channel", "join", "-b", block)
		if err := peerCmd.Run(); err != nil {
			return fmt.Errorf("join channel failed: %w", err)
		}

		fmt.Printf("%s Joined channel '%s'\n", green("✓"), name)
		return nil
	},
}
