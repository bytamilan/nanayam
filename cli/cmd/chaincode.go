package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/bytamilan/nanayam/cli/internal/fabric"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(chaincodeCmd)
	chaincodeCmd.AddCommand(ccPackageCmd)
	chaincodeCmd.AddCommand(ccInstallCmd)
	chaincodeCmd.AddCommand(ccApproveCmd)
	chaincodeCmd.AddCommand(ccCommitCmd)
	chaincodeCmd.AddCommand(ccInvokeCmd)
	chaincodeCmd.AddCommand(ccQueryCmd)

	ccPackageCmd.Flags().String("path", "", "Path to chaincode source (required)")
	ccPackageCmd.Flags().String("name", "", "Chaincode name (required)")
	ccPackageCmd.Flags().String("version", "1.0", "Chaincode version")
	ccPackageCmd.Flags().String("lang", "golang", "Chaincode language: golang, javascript, java")
	ccPackageCmd.Flags().String("output", "", "Output package file path")
	ccPackageCmd.MarkFlagRequired("path")
	ccPackageCmd.MarkFlagRequired("name")

	ccInstallCmd.Flags().String("package", "", "Path to chaincode package (required)")
	ccInstallCmd.Flags().String("peer", "", "Peer endpoint")
	ccInstallCmd.MarkFlagRequired("package")

	ccApproveCmd.Flags().String("name", "", "Chaincode name (required)")
	ccApproveCmd.Flags().String("version", "1.0", "Chaincode version")
	ccApproveCmd.Flags().String("sequence", "1", "Sequence number")
	ccApproveCmd.Flags().String("channel", "", "Channel ID (required)")
	ccApproveCmd.Flags().String("package-id", "", "Package ID from install (required)")
	ccApproveCmd.Flags().String("peer", "", "Peer endpoint")
	ccApproveCmd.Flags().StringSlice("init-required", []string{}, "Initialization required")
	ccApproveCmd.MarkFlagRequired("name")
	ccApproveCmd.MarkFlagRequired("channel")
	ccApproveCmd.MarkFlagRequired("package-id")

	ccCommitCmd.Flags().String("name", "", "Chaincode name (required)")
	ccCommitCmd.Flags().String("version", "1.0", "Chaincode version")
	ccCommitCmd.Flags().String("sequence", "1", "Sequence number")
	ccCommitCmd.Flags().String("channel", "", "Channel ID (required)")
	ccCommitCmd.Flags().StringSlice("peer", []string{}, "Peer endpoints to commit on")
	ccCommitCmd.Flags().StringSlice("tls-root-cert", []string{}, "TLS root cert files for peers")
	ccCommitCmd.MarkFlagRequired("name")
	ccCommitCmd.MarkFlagRequired("channel")

	ccInvokeCmd.Flags().String("channel", "", "Channel ID (required)")
	ccInvokeCmd.Flags().String("name", "", "Chaincode name (required)")
	ccInvokeCmd.Flags().String("function", "", "Function to invoke (required)")
	ccInvokeCmd.Flags().StringSlice("args", []string{}, "Function arguments")
	ccInvokeCmd.Flags().StringSlice("peer", []string{}, "Peer endpoints")
	ccInvokeCmd.Flags().Bool("wait", true, "Wait for transaction commit")
	ccInvokeCmd.MarkFlagRequired("channel")
	ccInvokeCmd.MarkFlagRequired("name")
	ccInvokeCmd.MarkFlagRequired("function")

	ccQueryCmd.Flags().String("channel", "", "Channel ID (required)")
	ccQueryCmd.Flags().String("name", "", "Chaincode name (required)")
	ccQueryCmd.Flags().String("function", "", "Function to query (required)")
	ccQueryCmd.Flags().StringSlice("args", []string{}, "Function arguments")
	ccQueryCmd.Flags().String("peer", "", "Peer endpoint")
	ccQueryCmd.MarkFlagRequired("channel")
	ccQueryCmd.MarkFlagRequired("name")
	ccQueryCmd.MarkFlagRequired("function")
}

var chaincodeCmd = &cobra.Command{
	Use:     "chaincode",
	Aliases: []string{"cc"},
	Short:   "Manage chaincode lifecycle",
	Long:    `Package, install, approve, commit, invoke, and query chaincode.`,
}

var ccPackageCmd = &cobra.Command{
	Use:   "package",
	Short: "Package chaincode",
	Example: `  nanayam chaincode package --path ./chaincode/asset-transfer-basic --name basic --version 1.0`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("path")
		name, _ := cmd.Flags().GetString("name")
		version, _ := cmd.Flags().GetString("version")
		output, _ := cmd.Flags().GetString("output")
		lang, _ := cmd.Flags().GetString("lang")

		blue := color.New(color.FgBlue).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()

		if output == "" {
			output = fmt.Sprintf("%s.tar.gz", name)
		}

		fmt.Printf("%s Packaging chaincode %s v%s...\n", blue("→"), name, version)
		fmt.Printf("  Path: %s\n", path)
		fmt.Printf("  Lang: %s\n", lang)

		bins := fabric.NewBinaries()
		peerCmd := exec.Command(bins.PeerPath(), "lifecycle", "chaincode", "package",
			output,
			"--path", path,
			"--lang", lang,
			"--label", fmt.Sprintf("%s_%s", name, version),
		)
		peerCmd.Env = append(os.Environ(), "FABRIC_CFG_PATH=./config")
		peerCmd.Stdout = os.Stdout
		peerCmd.Stderr = os.Stderr
		if err := peerCmd.Run(); err != nil {
			return fmt.Errorf("package failed: %w", err)
		}

		fmt.Printf("%s Chaincode packaged: %s\n", green("✓"), output)
		return nil
	},
}

var ccInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install chaincode package on a peer",
	RunE: func(cmd *cobra.Command, args []string) error {
		pkg, _ := cmd.Flags().GetString("package")
		peer, _ := cmd.Flags().GetString("peer")

		blue := color.New(color.FgBlue).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()

		fmt.Printf("%s Installing chaincode package %s...\n", blue("→"), pkg)

		bins := fabric.NewBinaries()
		env := fabric.DefaultPeerEnv("Org1")
		if peer != "" {
			env.Address = peer
		}

		peerCmd := env.Exec(bins.PeerPath(), "lifecycle", "chaincode", "install", pkg)
		if err := peerCmd.Run(); err != nil {
			return fmt.Errorf("install failed: %w", err)
		}

		fmt.Printf("%s Chaincode installed. Run 'nanayam chaincode queryinstalled' to get package ID.\n", green("✓"))
		return nil
	},
}

var ccApproveCmd = &cobra.Command{
	Use:   "approve",
	Short: "Approve chaincode definition for your org",
	Example: `  nanayam chaincode approve --name basic --channel mychannel --package-id basic_1.0:abc123...`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		channel, _ := cmd.Flags().GetString("channel")
		version, _ := cmd.Flags().GetString("version")
		sequence, _ := cmd.Flags().GetString("sequence")
		packageID, _ := cmd.Flags().GetString("package-id")
		peer, _ := cmd.Flags().GetString("peer")

		blue := color.New(color.FgBlue).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()

		fmt.Printf("%s Approving chaincode %s for org...\n", blue("→"), name)

		bins := fabric.NewBinaries()
		env := fabric.DefaultPeerEnv("Org1")
		if peer != "" {
			env.Address = peer
		}

		peerCmd := env.Exec(bins.PeerPath(), "lifecycle", "chaincode", "approveformyorg",
			"-o", fabric.DefaultOrderer(),
			"--channelID", channel,
			"--name", name,
			"--version", version,
			"--package-id", packageID,
			"--sequence", sequence,
			"--tls",
			"--cafile", fabric.OrdererTLSPath(),
		)
		if err := peerCmd.Run(); err != nil {
			return fmt.Errorf("approve failed: %w", err)
		}

		fmt.Printf("%s Chaincode approved for org\n", green("✓"))
		return nil
	},
}

var ccCommitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commit chaincode definition to the channel",
	Example: `  nanayam chaincode commit --name basic --channel mychannel --peer peer0.org1.nanayam.com:7051 --peer peer0.org2.nanayam.com:9051`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		channel, _ := cmd.Flags().GetString("channel")
		version, _ := cmd.Flags().GetString("version")
		sequence, _ := cmd.Flags().GetString("sequence")
		peers, _ := cmd.Flags().GetStringSlice("peer")
		tlsCerts, _ := cmd.Flags().GetStringSlice("tls-root-cert")

		blue := color.New(color.FgBlue).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()

		fmt.Printf("%s Committing chaincode %s to channel %s...\n", blue("→"), name, channel)

		bins := fabric.NewBinaries()
		env := fabric.DefaultPeerEnv("Org1")

		cmdArgs := []string{
			"lifecycle", "chaincode", "commit",
			"-o", fabric.DefaultOrderer(),
			"--channelID", channel,
			"--name", name,
			"--version", version,
			"--sequence", sequence,
			"--tls",
			"--cafile", fabric.OrdererTLSPath(),
		}

		// Add peer addresses
		for i, p := range peers {
			cmdArgs = append(cmdArgs, "--peerAddresses", p)
			if i < len(tlsCerts) {
				cmdArgs = append(cmdArgs, "--tlsRootCertFiles", tlsCerts[i])
			}
		}

		peerCmd := env.Exec(bins.PeerPath(), cmdArgs...)
		if err := peerCmd.Run(); err != nil {
			return fmt.Errorf("commit failed: %w", err)
		}

		fmt.Printf("%s Chaincode committed to channel\n", green("✓"))
		return nil
	},
}

var ccInvokeCmd = &cobra.Command{
	Use:   "invoke",
	Short: "Invoke a chaincode function",
	Example: `  nanayam chaincode invoke --channel mychannel --name basic --function InitLedger
  nanayam chaincode invoke --channel mychannel --name basic --function CreateAsset --args '{"asset1","blue","20","Alice","500"}'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		channel, _ := cmd.Flags().GetString("channel")
		name, _ := cmd.Flags().GetString("name")
		function, _ := cmd.Flags().GetString("function")
		ccArgs, _ := cmd.Flags().GetStringSlice("args")
		peers, _ := cmd.Flags().GetStringSlice("peer")
		wait, _ := cmd.Flags().GetBool("wait")

		blue := color.New(color.FgBlue).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()

		fmt.Printf("%s Invoking %s on %s/%s...\n", blue("→"), function, channel, name)

		argJSON := fmt.Sprintf(`{"function":"%s","Args":[%s]}`, function, formatArgs(ccArgs))

		bins := fabric.NewBinaries()
		env := fabric.DefaultPeerEnv("Org1")

		cmdArgs := []string{
			"chaincode", "invoke",
			"-o", fabric.DefaultOrderer(),
			"-C", channel,
			"-n", name,
			"--tls",
			"--cafile", fabric.OrdererTLSPath(),
			"-c", argJSON,
		}

		if !wait {
			cmdArgs = append(cmdArgs, "--waitForEvent", "false")
		}

		for _, p := range peers {
			cmdArgs = append(cmdArgs, "--peerAddresses", p)
		}

		peerCmd := env.Exec(bins.PeerPath(), cmdArgs...)
		if err := peerCmd.Run(); err != nil {
			return fmt.Errorf("invoke failed: %w", err)
		}

		fmt.Printf("%s Invoke successful\n", green("✓"))
		return nil
	},
}

var ccQueryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query a chaincode function",
	Example: `  nanayam chaincode query --channel mychannel --name basic --function GetAllAssets
  nanayam chaincode query --channel mychannel --name basic --function GetAsset --args '{"asset1"}'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		channel, _ := cmd.Flags().GetString("channel")
		name, _ := cmd.Flags().GetString("name")
		function, _ := cmd.Flags().GetString("function")
		ccArgs, _ := cmd.Flags().GetStringSlice("args")
		peer, _ := cmd.Flags().GetString("peer")

		blue := color.New(color.FgBlue).SprintFunc()
		fmt.Printf("%s Querying %s on %s/%s...\n", blue("→"), function, channel, name)

		argJSON := fmt.Sprintf(`{"function":"%s","Args":[%s]}`, function, formatArgs(ccArgs))

		bins := fabric.NewBinaries()
		env := fabric.DefaultPeerEnv("Org1")
		if peer != "" {
			env.Address = peer
		}

		peerCmd := env.Exec(bins.PeerPath(), "chaincode", "query",
			"-C", channel,
			"-n", name,
			"-c", argJSON,
		)
		if err := peerCmd.Run(); err != nil {
			return fmt.Errorf("query failed: %w", err)
		}

		return nil
	},
}

func formatArgs(args []string) string {
	var parts []string
	for _, a := range args {
		parts = append(parts, fmt.Sprintf("%q", a))
	}
	return strings.Join(parts, ",")
}
