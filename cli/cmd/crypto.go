package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bytamilan/nanayam/cli/internal/fabric"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v3"
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

		if err := generateCryptoMaterials(bins, configFile, output); err != nil {
			return fmt.Errorf("cryptogen failed: %w", err)
		}
		fmt.Printf("%s Crypto materials generated in %s/\n", green("✓"), output)

		if genChannel {
			fmt.Println()
			fmt.Printf("%s Generating channel artifacts...\n", blue("→"))
			if err := generateChannelArtifactsForCryptoConfig(bins, cwd, configFile); err != nil {
				return err
			}
		}

		fmt.Println()
		fmt.Printf("%s Crypto generation complete!\n", green("✓"))
		return nil
	},
}

type channelArtifactProfile struct {
	name           string
	channelProfile string
	genesis        string
	channel        string
	anchorOrgs     []string
}

func resolveChannelArtifactConfig(cwd, cryptoConfigFile string) (string, []channelArtifactProfile, error) {
	variant := artifactVariantFromFile(filepath.Base(cryptoConfigFile), "crypto-config")
	configtxSource, err := resolveConfigtxSource(cwd, variant)
	if err != nil {
		return "", nil, err
	}

	profiles, err := loadChannelArtifactProfiles(configtxSource)
	if err != nil {
		return "", nil, err
	}

	return configtxSource, profiles, nil
}

func generateCryptoMaterials(bins *fabric.Binaries, configFile, output string) error {
	cryptogen := exec.Command(bins.CryptogenPath(), "generate", "--config="+configFile, "--output="+output)
	cryptogen.Stdout = os.Stdout
	cryptogen.Stderr = os.Stderr
	return cryptogen.Run()
}

func generateChannelArtifactsForCryptoConfig(bins *fabric.Binaries, cwd, cryptoConfigFile string) error {
	channelDir := filepath.Join(cwd, "channel-artifacts")
	if err := os.MkdirAll(channelDir, 0755); err != nil {
		return fmt.Errorf("create channel-artifacts dir: %w", err)
	}

	configtxSource, profiles, err := resolveChannelArtifactConfig(cwd, cryptoConfigFile)
	if err != nil {
		return err
	}

	if err := generateChannelArtifacts(bins, cwd, configtxSource, channelDir, profiles); err != nil {
		return err
	}

	if err := writeArtifactSourceMetadata(cwd, cryptoConfigFile); err != nil {
		return err
	}

	return nil
}

func generateChannelArtifacts(bins *fabric.Binaries, cwd, configtxSource, channelDir string, profiles []channelArtifactProfile) error {
	configtxDir, cleanup, err := stageConfigtxFile(cwd, configtxSource)
	if err != nil {
		return err
	}
	defer cleanup()

	env := append(os.Environ(), "FABRIC_CFG_PATH="+configtxDir)
	for _, p := range profiles {
		genesisPath := filepath.Join(channelDir, p.genesis)
		if _, err := os.Stat(genesisPath); os.IsNotExist(err) {
			configtxgen := exec.Command(bins.ConfigtxgenPath(),
				"-profile", p.name,
				"-channelID", "system-channel",
				"-outputBlock", genesisPath)
			configtxgen.Env = env
			configtxgen.Stdout = os.Stdout
			configtxgen.Stderr = os.Stderr
			if err := configtxgen.Run(); err != nil {
				return fmt.Errorf("generate genesis block with profile %s: %w", p.name, err)
			}
		}

		channelTx := filepath.Join(channelDir, p.channel+".tx")
		if _, err := os.Stat(channelTx); os.IsNotExist(err) {
			configtxgen := exec.Command(bins.ConfigtxgenPath(),
				"-profile", p.channelProfile,
				"-outputCreateChannelTx", channelTx,
				"-channelID", p.channel)
			configtxgen.Env = env
			configtxgen.Stdout = os.Stdout
			configtxgen.Stderr = os.Stderr
			if err := configtxgen.Run(); err != nil {
				return fmt.Errorf("generate channel tx with profile %s: %w", p.channelProfile, err)
			}
		}

		for _, org := range p.anchorOrgs {
			anchorTx := filepath.Join(channelDir, org+"anchors.tx")
			if _, err := os.Stat(anchorTx); os.IsNotExist(err) {
				configtxgen := exec.Command(bins.ConfigtxgenPath(),
					"-profile", p.channelProfile,
					"-outputAnchorPeersUpdate", anchorTx,
					"-channelID", p.channel,
					"-asOrg", org)
				configtxgen.Env = env
				configtxgen.Stdout = os.Stdout
				configtxgen.Stderr = os.Stderr
				if err := configtxgen.Run(); err != nil {
					return fmt.Errorf("generate anchor peers for %s: %w", org, err)
				}
			}
		}
	}

	return nil
}

func artifactVariantFromFile(baseName, prefix string) string {
	name := strings.TrimSuffix(baseName, filepath.Ext(baseName))
	if name == prefix {
		return ""
	}

	return strings.TrimPrefix(name, prefix+"-")
}

func resolveConfigtxSource(cwd, variant string) (string, error) {
	configDir := filepath.Join(cwd, "config")
	for _, candidate := range configtxCandidatesForVariant(variant) {
		configtxSource := filepath.Join(configDir, candidate)
		if _, err := os.Stat(configtxSource); err == nil {
			return configtxSource, nil
		}
	}

	return "", fmt.Errorf("no usable configtx file found for variant %q in %s", variant, configDir)
}

func configtxCandidatesForVariant(variant string) []string {
	switch variant {
	case "", "basic", "fabric":
		return uniqueStrings([]string{
			"configtx-" + variant + ".yaml",
			"configtx.yaml",
			"configtx-basic.yaml",
			"configtx-fabric.yaml",
		})
	default:
		return uniqueStrings([]string{"configtx-" + variant + ".yaml"})
	}
}

type configtxDocument struct {
	Profiles map[string]configtxProfile `yaml:"Profiles"`
}

type configtxProfile struct {
	Application configtxApplication `yaml:"Application"`
}

type configtxApplication struct {
	Organizations []configtxOrganization `yaml:"Organizations"`
}

type configtxOrganization struct {
	Name string `yaml:"Name"`
	ID   string `yaml:"ID"`
}

func loadChannelArtifactProfiles(configtxSource string) ([]channelArtifactProfile, error) {
	content, err := os.ReadFile(configtxSource)
	if err != nil {
		return nil, fmt.Errorf("read configtx source %s: %w", configtxSource, err)
	}

	var doc configtxDocument
	if err := yaml.Unmarshal(content, &doc); err != nil {
		return nil, fmt.Errorf("parse configtx source %s: %w", configtxSource, err)
	}

	var profileNames []string
	for name := range doc.Profiles {
		if strings.HasSuffix(name, "OrdererGenesis") {
			profileNames = append(profileNames, name)
		}
	}
	sort.Strings(profileNames)

	var profiles []channelArtifactProfile
	for _, genesisProfile := range profileNames {
		channelProfile := strings.TrimSuffix(genesisProfile, "OrdererGenesis") + "Channel"
		channelCfg, ok := doc.Profiles[channelProfile]
		if !ok {
			continue
		}

		profiles = append(profiles, channelArtifactProfile{
			name:           genesisProfile,
			channelProfile: channelProfile,
			genesis:        "genesis.block",
			channel:        defaultChannelID(channelProfile),
			anchorOrgs:     organizationNames(channelCfg.Application.Organizations),
		})
	}

	if len(profiles) == 0 {
		return nil, fmt.Errorf("no matching *OrdererGenesis/*Channel profiles found in %s", configtxSource)
	}

	return profiles, nil
}

func organizationNames(orgs []configtxOrganization) []string {
	result := make([]string, 0, len(orgs))
	for _, org := range orgs {
		if org.Name != "" {
			result = append(result, org.Name)
			continue
		}
		if org.ID != "" {
			result = append(result, org.ID)
		}
	}
	return result
}

func defaultChannelID(profile string) string {
	switch profile {
	case "TwoOrgsChannel":
		return "mychannel"
	case "ComplaintChannel":
		return "complaint-channel"
	default:
		base := strings.TrimSuffix(profile, "Channel")
		if base == "" {
			return "mychannel"
		}
		return camelToKebab(base) + "-channel"
	}
}

func camelToKebab(s string) string {
	var out []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			out = append(out, '-')
		}
		if r >= 'A' && r <= 'Z' {
			r = r + ('a' - 'A')
		}
		out = append(out, r)
	}
	return string(out)
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func artifactSourceLabel(cwd, cryptoConfigFile string) string {
	if rel, err := filepath.Rel(cwd, cryptoConfigFile); err == nil {
		return filepath.ToSlash(rel)
	}
	return filepath.Base(cryptoConfigFile)
}

func artifactSourceMetadataPath(cwd string) string {
	return filepath.Join(cwd, "channel-artifacts", ".nanayam-artifact-source")
}

func writeArtifactSourceMetadata(cwd, cryptoConfigFile string) error {
	path := artifactSourceMetadataPath(cwd)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create channel-artifacts dir for metadata: %w", err)
	}
	if err := os.WriteFile(path, []byte(artifactSourceLabel(cwd, cryptoConfigFile)+"\n"), 0644); err != nil {
		return fmt.Errorf("write artifact source metadata: %w", err)
	}
	return nil
}

func readArtifactSourceMetadata(cwd string) (string, error) {
	content, err := os.ReadFile(artifactSourceMetadataPath(cwd))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(content)), nil
}

func stageConfigtxFile(cwd, source string) (string, func(), error) {
	tempRoot, err := os.MkdirTemp("", "nanayam-configtx-")
	if err != nil {
		return "", nil, fmt.Errorf("create temporary configtx dir: %w", err)
	}

	configDir := filepath.Join(tempRoot, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		os.RemoveAll(tempRoot)
		return "", nil, fmt.Errorf("create staged config directory: %w", err)
	}

	configtxDest := filepath.Join(configDir, "configtx.yaml")
	content, err := os.ReadFile(source)
	if err != nil {
		os.RemoveAll(tempRoot)
		return "", nil, fmt.Errorf("read configtx source %s: %w", source, err)
	}
	if err := os.WriteFile(configtxDest, content, 0644); err != nil {
		os.RemoveAll(tempRoot)
		return "", nil, fmt.Errorf("write staged configtx file: %w", err)
	}

	cryptoTarget := filepath.Join(cwd, "crypto-config")
	cryptoLink := filepath.Join(tempRoot, "crypto-config")
	if err := os.Symlink(cryptoTarget, cryptoLink); err != nil {
		os.RemoveAll(tempRoot)
		return "", nil, fmt.Errorf("link staged crypto-config directory: %w", err)
	}

	cleanup := func() {
		_ = os.RemoveAll(tempRoot)
	}

	return configDir, cleanup, nil
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
