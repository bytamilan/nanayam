package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bytamilan/nanayam/cli/internal/config"
	"github.com/bytamilan/nanayam/cli/internal/docker"
	"github.com/bytamilan/nanayam/cli/internal/fabric"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(nodeCmd)
	nodeCmd.AddCommand(nodeInitCmd)
	nodeCmd.AddCommand(nodeStartCmd)
	nodeCmd.AddCommand(nodeStopCmd)
	nodeCmd.AddCommand(nodeStatusCmd)
	nodeCmd.AddCommand(nodeLogsCmd)

	nodeInitCmd.Flags().String("type", "peer", "Node type: peer, orderer, ca, tools")
	nodeInitCmd.Flags().String("org", "Org1", "Organization name")
	nodeInitCmd.Flags().String("domain", "", "Organization domain (default: <org>.nanayam.com)")
	nodeInitCmd.Flags().String("output", "", "Output directory for crypto materials")
	nodeInitCmd.Flags().String("profile", "", "Configtx profile name")
}

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Manage local Fabric nodes (peer, orderer, CA, tools)",
	Long:  `Initialize, start, stop, and inspect individual Fabric nodes.`,
}

var nodeInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new local node",
	Example: `  # Initialize a peer for Org1
  nanayam node init --type peer --org Org1

  # Initialize an orderer
  nanayam node init --type orderer --domain nanayam.com

  # Initialize a CA for a new org
  nanayam node init --type ca --org Dept --domain dept.nanayam.com`,
	RunE: runNodeInit,
}

func runNodeInit(cmd *cobra.Command, args []string) error {
	nodeType, _ := cmd.Flags().GetString("type")
	org, _ := cmd.Flags().GetString("org")
	domain, _ := cmd.Flags().GetString("domain")
	output, _ := cmd.Flags().GetString("output")

	if domain == "" {
		domain = fmt.Sprintf("%s.nanayam.com", strings.ToLower(org))
	}
	if output == "" {
		output = "crypto-config"
	}

	bins := fabric.NewBinaries()
	if err := bins.CheckAll(); err != nil {
		return fmt.Errorf("fabric binaries not found: %w\nRun 'nanayam prerequisites --install-fabric' to download them", err)
	}

	cwd, _ := os.Getwd()
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	switch nodeType {
	case "peer":
		fmt.Printf("%s Initializing peer node for org %s...\n", blue("→"), org)

		// Generate cryptogen config
		cryptoCfgPath := filepath.Join(cwd, "config", fmt.Sprintf("crypto-config-%s.yaml", strings.ToLower(org)))
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
		fmt.Printf("  %s Generated cryptogen config: %s\n", green("✓"), cryptoCfgPath)

		// Generate crypto
		cryptoOut := filepath.Join(cwd, output)
		cryptogen := bins.CryptogenPath()
		execCmd := []string{cryptogen, "generate", "--config=" + cryptoCfgPath, "--output=" + cryptoOut}
		fmt.Printf("  %s Running: %s\n", blue("→"), strings.Join(execCmd, " "))
		// TODO: actually run cryptogen

		// Generate peer compose
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
		fmt.Printf("  %s Generated compose: %s\n", green("✓"), composePath)
		fmt.Println()
		fmt.Printf("%s Peer node initialized. Run 'nanayam node start %s' to start it.\n", green("✓"), peerID)

	case "orderer":
		fmt.Printf("%s Initializing orderer node...\n", blue("→"))
		ordererID := fmt.Sprintf("orderer.%s", domain)
		ordererCfg := &config.OrdererNodeConfig{
			OrdererID:    ordererID,
			OrdererPort:  7050,
			MSPID:        "OrdererMSP",
			CryptoPath:   filepath.Join(cwd, output, "ordererOrganizations", domain),
			GenesisBlock: filepath.Join(cwd, "channel-artifacts", "genesis.block"),
		}
		composePath := filepath.Join(cwd, "docker", fmt.Sprintf("%s.yaml", ordererID))
		if err := config.GenerateOrdererCompose(ordererCfg, composePath); err != nil {
			return fmt.Errorf("generate orderer compose: %w", err)
		}
		fmt.Printf("  %s Generated compose: %s\n", green("✓"), composePath)
		fmt.Println()
		fmt.Printf("%s Orderer node initialized. Run 'nanayam node start %s' to start it.\n", green("✓"), ordererID)

	case "ca":
		fmt.Printf("%s Initializing CA for org %s...\n", blue("→"), org)
		caCfg := &config.CANodeConfig{
			OrgName: org,
			CAPort:  7054,
			OpsPort: 17054,
		}
		composePath := filepath.Join(cwd, "docker", fmt.Sprintf("ca-%s.yaml", strings.ToLower(org)))
		if err := config.GenerateCACompose(caCfg, composePath); err != nil {
			return fmt.Errorf("generate ca compose: %w", err)
		}
		fmt.Printf("  %s Generated compose: %s\n", green("✓"), composePath)
		fmt.Println()
		fmt.Printf("%s CA node initialized. Run 'nanayam node start ca_%s' to start it.\n", green("✓"), strings.ToLower(org))

	case "tools":
		fmt.Println("Tools container is included in the network compose files.")
		fmt.Println("Run 'nanayam network up' to start it along with the network.")

	default:
		return fmt.Errorf("unknown node type: %s", nodeType)
	}

	return nil
}

var nodeStartCmd = &cobra.Command{
	Use:   "start [service...]",
	Short: "Start local node containers",
	Args:  cobra.MinimumNArgs(1),
	Example: `  nanayam node start peer0.org1.nanayam.com
  nanayam node start -- -f docker/peer0.org1.nanayam.com.yaml up -d`,
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, service := range args {
			composeFile := filepath.Join("docker", fmt.Sprintf("%s.yaml", service))
			if _, err := os.Stat(composeFile); os.IsNotExist(err) {
				// Try network compose files
				composeFile = filepath.Join("docker", "fabric-network.yaml")
			}
			runner := docker.NewComposeRunner(composeFile)
			if err := runner.Start(service); err != nil {
				return fmt.Errorf("start %s: %w", service, err)
			}
			fmt.Printf("Started %s\n", service)
		}
		return nil
	},
}

var nodeStopCmd = &cobra.Command{
	Use:   "stop [service...]",
	Short: "Stop local node containers",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, service := range args {
			composeFile := filepath.Join("docker", fmt.Sprintf("%s.yaml", service))
			if _, err := os.Stat(composeFile); os.IsNotExist(err) {
				composeFile = filepath.Join("docker", "fabric-network.yaml")
			}
			runner := docker.NewComposeRunner(composeFile)
			if err := runner.Stop(service); err != nil {
				return fmt.Errorf("stop %s: %w", service, err)
			}
			fmt.Printf("Stopped %s\n", service)
		}
		return nil
	},
}

var nodeStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show node container status",
	RunE: func(cmd *cobra.Command, args []string) error {
		containers, err := docker.ListContainers("")
		if err != nil {
			return err
		}
		fmt.Printf("%-40s %-20s %s\n", "NAME", "STATUS", "PORTS")
		fmt.Println(strings.Repeat("-", 90))
		for _, c := range containers {
			fmt.Printf("%-40s %-20s %s\n", c.Name, c.Status, c.Ports)
		}
		return nil
	},
}

var nodeLogsCmd = &cobra.Command{
	Use:   "logs [container]",
	Short: "Tail node container logs",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		container := ""
		if len(args) > 0 {
			container = args[0]
		}
		composeFile := filepath.Join("docker", "fabric-network.yaml")
		if _, err := os.Stat(composeFile); os.IsNotExist(err) {
			// Try complaint network
			composeFile = filepath.Join("docker", "complaint-network.yaml")
		}
		runner := docker.NewComposeRunner(composeFile)
		return runner.Logs(container)
	},
}
