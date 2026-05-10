package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bytamilan/nanayam/cli/internal/fabric"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(channelCmd)
	channelCmd.AddCommand(channelCreateCmd)
	channelCmd.AddCommand(channelJoinCmd)
	channelCmd.AddCommand(channelUpdateAnchorCmd)
	channelCmd.AddCommand(channelListCmd)

	channelCreateCmd.Flags().String("name", "", "Channel ID (required)")
	channelCreateCmd.Flags().String("profile", "", "Configtx profile name (required)")
	channelCreateCmd.Flags().String("orderer", "orderer.nanayam.com:7050", "Orderer endpoint")
	channelCreateCmd.Flags().String("tx-file", "", "Path to channel creation tx (auto-detected if not provided)")
	channelCreateCmd.MarkFlagRequired("name")

	channelJoinCmd.Flags().String("name", "", "Channel ID (required)")
	channelJoinCmd.Flags().String("peer", "", "Peer endpoint to join (default: local peer)")
	channelJoinCmd.Flags().String("block", "", "Path to channel genesis block")
	channelJoinCmd.MarkFlagRequired("name")

	channelUpdateAnchorCmd.Flags().String("name", "", "Channel ID (required)")
	channelUpdateAnchorCmd.Flags().String("org", "", "Organization MSP ID (required)")
	channelUpdateAnchorCmd.Flags().String("peer", "", "Peer endpoint")
	channelUpdateAnchorCmd.MarkFlagRequired("name")
	channelUpdateAnchorCmd.MarkFlagRequired("org")
}

var channelCmd = &cobra.Command{
	Use:   "channel",
	Short: "Manage Fabric channels",
	Long:  `Create, join, and update channels.`,
}

var channelCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new channel",
	Example: `  nanayam channel create --name mychannel --profile TwoOrgsChannel
  nanayam channel create --name complaint-channel --profile ComplaintChannel --orderer orderer.nanayam.com:7050`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		orderer, _ := cmd.Flags().GetString("orderer")
		txFile, _ := cmd.Flags().GetString("tx-file")

		blue := color.New(color.FgBlue).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()

		if txFile == "" {
			cwd, _ := os.Getwd()
			txFile = filepath.Join(cwd, "channel-artifacts", name+".tx")
			if _, err := os.Stat(txFile); os.IsNotExist(err) {
				// Try without .tx suffix in name
				txFile = filepath.Join(cwd, "channel-artifacts", "channel.tx")
			}
		}

		fmt.Printf("%s Creating channel '%s'...\n", blue("→"), name)
		fmt.Printf("  Tx file: %s\n", txFile)
		fmt.Printf("  Orderer: %s\n", orderer)

		ordererTLS := fabric.OrdererTLSPath()

		peerCmd := fabric.DockerExec(
			"channel", "create",
			"-o", orderer,
			"-c", name,
			"-f", "/opt/gopath/src/github.com/hyperledger/fabric/peer/channel-artifacts/"+filepath.Base(txFile),
			"--tls",
			"--cafile", "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/nanayam.com/orderers/orderer.nanayam.com/msp/tlscacerts/tlsca.nanayam.com-cert.pem",
		)
		if err := peerCmd.Run(); err != nil {
			// Try fallback with local peer binary
			fmt.Printf("  Docker exec failed, trying local peer...\n")
			_ = ordererTLS
			bins := fabric.NewBinaries()
			env := fabric.DefaultPeerEnv("Org1")
			localCmd := env.Exec(bins.PeerPath(), "channel", "create",
				"-o", orderer,
				"-c", name,
				"-f", txFile,
				"--tls",
				"--cafile", ordererTLS,
			)
			if err := localCmd.Run(); err != nil {
				return fmt.Errorf("channel create failed: %w", err)
			}
		}

		fmt.Printf("%s Channel '%s' created successfully!\n", green("✓"), name)
		return nil
	},
}

var channelJoinCmd = &cobra.Command{
	Use:   "join",
	Short: "Join peer(s) to a channel",
	Example: `  nanayam channel join --name mychannel
  nanayam channel join --name mychannel --peer peer0.org2.nanayam.com:9051`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		peer, _ := cmd.Flags().GetString("peer")
		block, _ := cmd.Flags().GetString("block")

		blue := color.New(color.FgBlue).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()

		if block == "" {
			block = name + ".block"
		}

		fmt.Printf("%s Joining peer to channel '%s'...\n", blue("→"), name)
		if peer != "" {
			fmt.Printf("  Peer: %s\n", peer)
		}

		// Try docker exec first
		dockerCmd := fabric.DockerExec("channel", "join", "-b", block)
		if err := dockerCmd.Run(); err != nil {
			// Fallback to local peer with environment
			bins := fabric.NewBinaries()
			env := fabric.DefaultPeerEnv("Org1")
			if peer != "" {
				env.Address = peer
			}
			localCmd := env.Exec(bins.PeerPath(), "channel", "join", "-b", block)
			if err := localCmd.Run(); err != nil {
				return fmt.Errorf("channel join failed: %w", err)
			}
		}

		fmt.Printf("%s Peer joined channel '%s'\n", green("✓"), name)
		return nil
	},
}

var channelUpdateAnchorCmd = &cobra.Command{
	Use:   "update-anchor",
	Short: "Update anchor peers for an organization",
	Example: `  nanayam channel update-anchor --name mychannel --org Org1MSP`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		org, _ := cmd.Flags().GetString("org")
		peer, _ := cmd.Flags().GetString("peer")

		blue := color.New(color.FgBlue).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()

		cwd, _ := os.Getwd()
		anchorTx := filepath.Join(cwd, "channel-artifacts", org+"anchors.tx")

		fmt.Printf("%s Updating anchor peers for %s on channel '%s'...\n", blue("→"), org, name)

		ordererTLS := "/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/nanayam.com/orderers/orderer.nanayam.com/msp/tlscacerts/tlsca.nanayam.com-cert.pem"

		dockerCmd := fabric.DockerExec(
			"channel", "update",
			"-o", "orderer.nanayam.com:7050",
			"-c", name,
			"-f", "/opt/gopath/src/github.com/hyperledger/fabric/peer/channel-artifacts/"+org+"anchors.tx",
			"--tls",
			"--cafile", ordererTLS,
		)
		if err := dockerCmd.Run(); err != nil {
			// Fallback
			bins := fabric.NewBinaries()
			env := fabric.DefaultPeerEnv("Org1")
			if peer != "" {
				env.Address = peer
			}
			localCmd := env.Exec(bins.PeerPath(), "channel", "update",
				"-o", fabric.DefaultOrderer(),
				"-c", name,
				"-f", anchorTx,
				"--tls",
				"--cafile", fabric.OrdererTLSPath(),
			)
			if err := localCmd.Run(); err != nil {
				return fmt.Errorf("anchor peer update failed: %w", err)
			}
		}

		fmt.Printf("%s Anchor peers updated for %s\n", green("✓"), org)
		return nil
	},
}

var channelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List channels joined by a peer",
	RunE: func(cmd *cobra.Command, args []string) error {
		blue := color.New(color.FgBlue).SprintFunc()
		fmt.Printf("%s Listing channels...\n", blue("→"))

		dockerCmd := fabric.DockerExec("channel", "list")
		if err := dockerCmd.Run(); err != nil {
			bins := fabric.NewBinaries()
			env := fabric.DefaultPeerEnv("Org1")
			localCmd := env.Exec(bins.PeerPath(), "channel", "list")
			if err := localCmd.Run(); err != nil {
				return fmt.Errorf("channel list failed: %w", err)
			}
		}
		return nil
	},
}
