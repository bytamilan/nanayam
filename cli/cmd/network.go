package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bytamilan/nanayam/cli/internal/docker"
	"github.com/bytamilan/nanayam/cli/internal/fabric"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var composePrerequisitesValidator = docker.ValidateComposePrerequisites
var networkArtifactGenerator = autoGenerateNetworkArtifacts

func init() {
	rootCmd.AddCommand(networkCmd)
	networkCmd.AddCommand(networkUpCmd)
	networkCmd.AddCommand(networkDownCmd)
	networkCmd.AddCommand(networkCleanCmd)
	networkCmd.AddCommand(networkStatusCmd)

	networkUpCmd.Flags().String("profile", "basic", "Network profile: basic, complaint")
	networkUpCmd.Flags().String("config", "", "Compose config file to use (path or filename under docker/)")
	networkDownCmd.Flags().Bool("volumes", false, "Also remove named volumes")
	networkCleanCmd.Flags().Bool("all", false, "Also remove Docker images")
}

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Manage the full Fabric network",
	Long:  `Bring up, tear down, and inspect the complete Fabric network stack.`,
}

var networkUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Start the full Fabric network",
	Example: `  nanayam network up
	  nanayam network up --profile complaint
	  nanayam network up --config fabric-network.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, _ := cmd.Flags().GetString("profile")
		config, _ := cmd.Flags().GetString("config")
		blue := color.New(color.FgBlue).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()

		if config != "" && cmd.Flags().Changed("profile") {
			return fmt.Errorf("--profile and --config cannot be used together")
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		composeFiles, targetLabel, err := resolveNetworkUpComposeFiles(profile, config)
		if err != nil {
			return err
		}
		recovered, setupScript, err := ensureComposePrerequisites(cwd, composeFiles)
		if err != nil {
			return err
		}
		if recovered {
			fmt.Printf("%s Generated missing Fabric certificates and artifacts using %s\n", green("✓"), setupScript)
		}

		fmt.Printf("%s Starting Fabric network (%s)...\n", blue("→"), targetLabel)

		// Ensure nanayam network exists
		exec.Command("docker", "network", "create", "nanayam").Run() // ignore error if exists

		runner := docker.NewComposeRunner(composeFiles...)
		if err := runner.Up(); err != nil {
			return fmt.Errorf("failed to start network: %w", err)
		}

		fmt.Println()
		fmt.Printf("%s Fabric network is up!\n", green("✓"))
		fmt.Println()
		fmt.Println("Next steps:")
		for _, step := range nextStepsForCompose(cwd, composeFiles) {
			fmt.Printf("  %s\n", step)
		}
		return nil
	},
}

func nextStepsForCompose(cwd string, composeFiles []string) []string {
	cryptoConfigFile, err := resolveCryptoConfigForCompose(cwd, composeFiles)
	if err != nil {
		return []string{
			"nanayam channel create --name mychannel --profile TwoOrgsChannel",
			"nanayam chaincode package --path ./chaincode/asset-transfer-basic --name basic",
		}
	}

	steps := make([]string, 0, 2)
	if _, profiles, err := resolveChannelArtifactConfig(cwd, cryptoConfigFile); err == nil && len(profiles) > 0 {
		steps = append(steps, fmt.Sprintf("nanayam channel create --name %s --profile %s", profiles[0].channel, profiles[0].channelProfile))
	}

	switch artifactVariantFromFile(filepath.Base(cryptoConfigFile), "crypto-config") {
	case "complaint":
		steps = append(steps, "nanayam chaincode package --path ./chaincode/complaint-system --name complaint")
	case "", "basic", "fabric":
		steps = append(steps, "nanayam chaincode package --path ./chaincode/asset-transfer-basic --name basic")
	}

	if len(steps) == 0 {
		return []string{
			"nanayam channel create --name mychannel --profile TwoOrgsChannel",
			"nanayam chaincode package --path ./chaincode/asset-transfer-basic --name basic",
		}
	}

	return steps
}

func resolveNetworkUpComposeFiles(profile, config string) ([]string, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, "", err
	}

	return resolveNetworkUpComposeFilesFromDir(cwd, profile, config)
}

func resolveNetworkUpComposeFilesFromDir(cwd, profile, config string) ([]string, string, error) {
	if config != "" {
		composeFile, err := docker.ResolveComposeFile(cwd, config)
		if err != nil {
			return nil, "", err
		}

		return []string{composeFile}, fmt.Sprintf("config: %s", filepath.Base(composeFile)), nil
	}

	composeFiles, err := docker.FindComposeFilesInDir(cwd, profile)
	if err != nil {
		return nil, "", err
	}

	if len(composeFiles) > 0 {
		if appsFile := resolveNetworkAppsFile(composeFiles[0]); appsFile != "" {
			composeFiles = append(composeFiles, appsFile)
		}
	}

	return composeFiles, fmt.Sprintf("profile: %s", profile), nil
}

func ensureComposePrerequisites(cwd string, composeFiles []string) (bool, string, error) {
	validationErr := composePrerequisitesValidator(composeFiles)
	needsRefresh, expectedLabel, err := networkArtifactsNeedRefresh(cwd, composeFiles)
	if err != nil {
		return false, "", err
	}

	if validationErr == nil && !needsRefresh {
		return false, "", nil
	}

	artifactLabel, err := networkArtifactGenerator(cwd, composeFiles)
	if err != nil {
		if validationErr != nil {
			return false, artifactLabel, fmt.Errorf("compose prerequisites check failed: %w\nAutomatic recovery failed: %w", validationErr, err)
		}
		return false, artifactLabel, fmt.Errorf("artifact refresh required for %s\nAutomatic recovery failed: %w", expectedLabel, err)
	}

	if err := composePrerequisitesValidator(composeFiles); err != nil {
		return false, artifactLabel, fmt.Errorf("compose prerequisites check failed after automatic recovery (%s): %w", artifactLabel, err)
	}

	return true, artifactLabel, nil
}


func autoGenerateNetworkArtifacts(cwd string, composeFiles []string) (string, error) {
	cryptoConfigFile, err := resolveCryptoConfigForCompose(cwd, composeFiles)
	if err != nil {
		return "", err
	}

	if err := generateNetworkArtifacts(cwd, cryptoConfigFile); err != nil {
		return filepath.Base(cryptoConfigFile), err
	}

	return filepath.Base(cryptoConfigFile), nil
}

func generateNetworkArtifacts(cwd, cryptoConfigFile string) error {
	bins := fabric.NewBinaries()
	if err := bins.CheckAll(); err != nil {
		return fmt.Errorf("fabric binaries not found: %w\nRun 'nanayam prerequisites --install-fabric' to download them", err)
	}

	if err := os.RemoveAll(filepath.Join(cwd, "crypto-config")); err != nil {
		return fmt.Errorf("clear crypto-config dir: %w", err)
	}
	if err := os.RemoveAll(filepath.Join(cwd, "channel-artifacts")); err != nil {
		return fmt.Errorf("clear channel-artifacts dir: %w", err)
	}

	output := filepath.Join(cwd, "crypto-config")
	if err := os.MkdirAll(output, 0755); err != nil {
		return fmt.Errorf("create crypto-config dir: %w", err)
	}

	if err := generateCryptoMaterials(bins, cryptoConfigFile, output); err != nil {
		return fmt.Errorf("generate crypto materials from %s: %w", cryptoConfigFile, err)
	}

	if err := generateChannelArtifactsForCryptoConfig(bins, cwd, cryptoConfigFile); err != nil {
		return fmt.Errorf("generate channel artifacts from %s: %w", cryptoConfigFile, err)
	}

	return nil
}

func networkArtifactsNeedRefresh(cwd string, composeFiles []string) (bool, string, error) {
	cryptoConfigFile, err := resolveCryptoConfigForCompose(cwd, composeFiles)
	if err != nil {
		return false, "", err
	}

	expected := artifactSourceLabel(cwd, cryptoConfigFile)
	actual, err := readArtifactSourceMetadata(cwd)
	if err != nil {
		if os.IsNotExist(err) {
			return true, filepath.Base(cryptoConfigFile), nil
		}
		return false, "", fmt.Errorf("read current artifact metadata: %w", err)
	}

	return actual != expected, filepath.Base(cryptoConfigFile), nil
}

func resolveCryptoConfigForCompose(cwd string, composeFiles []string) (string, error) {
	configDir := filepath.Join(cwd, "config")
	for _, variant := range composeVariantCandidates(primaryNetworkComposeFile(composeFiles)) {
		for _, candidate := range cryptoConfigCandidatesForVariant(variant) {
			path := filepath.Join(configDir, candidate)
			if _, err := os.Stat(path); err == nil {
				return path, nil
			}
		}
	}

	return "", fmt.Errorf("no matching crypto config found for %s in %s", primaryNetworkComposeFile(composeFiles), configDir)
}

func primaryNetworkComposeFile(composeFiles []string) string {
	for _, composeFile := range composeFiles {
		if strings.Contains(filepath.Base(composeFile), "network") {
			return composeFile
		}
	}
	if len(composeFiles) > 0 {
		return composeFiles[0]
	}
	return ""
}

func composeVariantCandidates(composeFile string) []string {
	base := strings.TrimSuffix(filepath.Base(composeFile), filepath.Ext(composeFile))
	base = strings.TrimSuffix(base, "-network")

	switch base {
	case "", "fabric", "basic":
		return uniqueStrings([]string{"fabric", "basic", ""})
	default:
		return uniqueStrings([]string{base})
	}
}

func cryptoConfigCandidatesForVariant(variant string) []string {
	switch variant {
	case "", "basic", "fabric":
		return uniqueStrings([]string{
			"crypto-config-" + variant + ".yaml",
			"crypto-config.yaml",
		})
	default:
		return uniqueStrings([]string{"crypto-config-" + variant + ".yaml"})
	}
}

func resolveNetworkAppsFile(composeFile string) string {
	dir := filepath.Dir(composeFile)
	base := filepath.Base(composeFile)
	var candidates []string

	switch base {
	case "fabric-network.yaml":
		candidates = append(candidates, filepath.Join(dir, "apps.yaml"))
	case "complaint-network.yaml":
		candidates = append(candidates, filepath.Join(dir, "complaint-apps.yaml"))
	default:
		if strings.HasSuffix(base, "-network.yaml") {
			prefix := strings.TrimSuffix(base, "-network.yaml")
			if prefix != "" {
				candidates = append(candidates, filepath.Join(dir, prefix+"-apps.yaml"))
			}
		}
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

var networkDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop the Fabric network",
	RunE: func(cmd *cobra.Command, args []string) error {
		volumes, _ := cmd.Flags().GetBool("volumes")
		blue := color.New(color.FgBlue).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()

		fmt.Printf("%s Stopping Fabric network...\n", blue("→"))

		// Try to find and stop all known compose files
		profiles := []string{"basic", "complaint"}
		var allFiles []string
		cwd, _ := os.Getwd()

		for _, profile := range profiles {
			files, err := docker.FindComposeFiles(profile)
			if err == nil {
				allFiles = append(allFiles, files...)
			}
		}

		// Add apps files
		for _, f := range []string{"apps.yaml", "complaint-apps.yaml"} {
			path := filepath.Join(cwd, "docker", f)
			if _, err := os.Stat(path); err == nil {
				allFiles = append(allFiles, path)
			}
		}

		if len(allFiles) > 0 {
			runner := docker.NewComposeRunner(allFiles...)
			if err := runner.Down(volumes); err != nil {
				return fmt.Errorf("failed to stop network: %w", err)
			}
		}

		fmt.Printf("%s Fabric network stopped.\n", green("✓"))
		if volumes {
			fmt.Println("Named volumes removed.")
		}
		return nil
	},
}

var networkCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Wipe all network data (crypto, channels, binaries)",
	Long: `Removes:
  - crypto-config/
  - channel-artifacts/
  - bin/ (Fabric binaries)
  - Docker volumes

Use with caution — this is destructive.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		red := color.New(color.FgRed).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()

		fmt.Printf("%s This will delete crypto-config/, channel-artifacts/, and bin/\n", yellow("⚠"))
		fmt.Printf("Type 'yes' to confirm: ")
		var confirm string
		fmt.Scanln(&confirm)
		if strings.ToLower(confirm) != "yes" {
			fmt.Println("Aborted.")
			return nil
		}

		// Stop network first
		networkDownCmd.Run(cmd, args)

		dirs := []string{"crypto-config", "channel-artifacts", "bin"}
		for _, d := range dirs {
			if err := os.RemoveAll(d); err != nil {
				fmt.Printf("  %s Failed to remove %s: %v\n", red("✗"), d, err)
			} else {
				fmt.Printf("  %s Removed %s/\n", green("✓"), d)
			}
		}

		if all {
			fmt.Println()
			fmt.Printf("%s Removing Docker images...\n", yellow("→"))
			// Remove Fabric images
			images := []string{
				"hyperledger/fabric-peer",
				"hyperledger/fabric-orderer",
				"hyperledger/fabric-tools",
				"hyperledger/fabric-ca",
				"couchdb",
			}
			for _, img := range images {
				exec.Command("docker", "rmi", "-f", img).Run()
			}
			fmt.Printf("%s Docker images removed.\n", green("✓"))
		}

		fmt.Println()
		fmt.Printf("%s Network data cleaned.\n", green("✓"))
		fmt.Println("Run 'nanayam prerequisites --install-fabric' to re-download binaries if needed.")
		return nil
	},
}

var networkStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show network container status",
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
