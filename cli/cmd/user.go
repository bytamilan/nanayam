package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bytamilan/nanayam/cli/internal/ca"
	"github.com/bytamilan/nanayam/cli/internal/fabric"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(userCreateCmd)
	userCmd.AddCommand(userEnrollCmd)
	userCmd.AddCommand(userListCmd)

	userCreateCmd.Flags().String("id", "", "User identity name (required)")
	userCreateCmd.Flags().String("secret", "", "Enrollment secret/password (required)")
	userCreateCmd.Flags().String("type", "client", "Identity type: admin, peer, client, orderer")
	userCreateCmd.Flags().String("org", "", "Organization name (required)")
	userCreateCmd.Flags().String("ca-url", "", "Fabric CA endpoint (e.g. https://localhost:7054)")
	userCreateCmd.Flags().String("msp-dir", "", "Output directory for enrolled MSP")
	userCreateCmd.Flags().String("affiliation", "", "CA affiliation (default: org.department)")
	userCreateCmd.Flags().String("ca-tls-cert", "", "Path to CA TLS certificate")
	userCreateCmd.MarkFlagRequired("id")
	userCreateCmd.MarkFlagRequired("secret")
	userCreateCmd.MarkFlagRequired("org")

	userEnrollCmd.Flags().String("id", "", "User identity name (required)")
	userEnrollCmd.Flags().String("secret", "", "Enrollment secret (required)")
	userEnrollCmd.Flags().String("org", "", "Organization name (required)")
	userEnrollCmd.Flags().String("ca-url", "", "Fabric CA endpoint")
	userEnrollCmd.Flags().String("msp-dir", "", "Output directory for enrolled MSP")
	userEnrollCmd.Flags().String("ca-tls-cert", "", "Path to CA TLS certificate")
	userEnrollCmd.MarkFlagRequired("id")
	userEnrollCmd.MarkFlagRequired("secret")
	userEnrollCmd.MarkFlagRequired("org")
}

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage Fabric identities and users",
	Long: `Register, enroll, and list user identities via Fabric CA.

The create command both registers and enrolls a new identity.
The enroll command only enrolls an already-registered identity.`,
}

var userCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Register and enroll a new user identity",
	Example: `  # Create a client user for Org1
  nanayam user create --id alice --secret alicepw --type client --org Org1

  # Create an admin user for ACB
  nanayam user create --id acb-admin --secret adminpw --type admin --org ACB --ca-url https://localhost:7054`,
	RunE: runUserCreate,
}

func runUserCreate(cmd *cobra.Command, args []string) error {
	id, _ := cmd.Flags().GetString("id")
	secret, _ := cmd.Flags().GetString("secret")
	userType, _ := cmd.Flags().GetString("type")
	org, _ := cmd.Flags().GetString("org")
	caURL, _ := cmd.Flags().GetString("ca-url")
	mspDir, _ := cmd.Flags().GetString("msp-dir")
	affiliation, _ := cmd.Flags().GetString("affiliation")
	caTLSCert, _ := cmd.Flags().GetString("ca-tls-cert")

	blue := color.New(color.FgBlue).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	// Defaults
	if caURL == "" {
		orgLower := strings.ToLower(org)
		caURL = fmt.Sprintf("https://localhost:7054")
		// Try to infer port based on org
		portMap := map[string]string{
			"org1":      "7054",
			"acb":       "7054",
			"org2":      "8054",
			"dept":      "8054",
			"oversight": "9054",
			"judiciary": "10054",
		}
		if p, ok := portMap[orgLower]; ok {
			caURL = fmt.Sprintf("https://localhost:%s", p)
		}
	}
	if mspDir == "" {
		cwd, _ := os.Getwd()
		domain := fmt.Sprintf("%s.nanayam.com", orgLower(org))
		mspDir = filepath.Join(cwd, "crypto-config", "peerOrganizations", domain, "users", fmt.Sprintf("%s@%s", id, domain), "msp")
	}
	if affiliation == "" {
		affiliation = fmt.Sprintf("%s.department1", orgLower(org))
	}

	fmt.Printf("%s Creating user '%s' (type: %s) for org '%s'...\n", blue("→"), id, userType, org)
	fmt.Printf("  CA URL: %s\n", caURL)
	fmt.Printf("  MSP Dir: %s\n", mspDir)

	bins := fabric.NewBinaries()
	caBinary := bins.CAClientPath()
	if _, err := os.Stat(caBinary); os.IsNotExist(err) {
		return fmt.Errorf("fabric-ca-client not found at %s\nRun 'nanayam prerequisites --install-fabric' to download it", caBinary)
	}

	homeDir := filepath.Join(os.TempDir(), "nanayam-ca")
	os.MkdirAll(homeDir, 0755)

	caClient := ca.NewClient(caBinary, homeDir, caURL)
	if caTLSCert != "" {
		caClient.TLSConfig = caTLSCert
	}

	// Register the user
	fmt.Printf("%s Registering identity with CA...\n", blue("→"))
	if err := caClient.Register("admin", "adminpw", id, secret, userType, affiliation); err != nil {
		fmt.Printf("  Registration may have failed or user already registered: %v\n", err)
	} else {
		fmt.Printf("  %s Identity registered\n", green("✓"))
	}

	// Enroll the user
	fmt.Printf("%s Enrolling identity...\n", blue("→"))
	os.MkdirAll(mspDir, 0755)
	if err := caClient.Enroll(id, secret, mspDir); err != nil {
		return fmt.Errorf("enroll failed: %w", err)
	}
	fmt.Printf("  %s Identity enrolled to %s\n", green("✓"), mspDir)

	fmt.Println()
	fmt.Printf("%s User '%s' created successfully!\n", green("✓"), id)
	return nil
}

var userEnrollCmd = &cobra.Command{
	Use:     "enroll",
	Short:   "Enroll an already-registered user identity",
	Example: `  nanayam user enroll --id alice --secret alicepw --org Org1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetString("id")
		secret, _ := cmd.Flags().GetString("secret")
		org, _ := cmd.Flags().GetString("org")
		caURL, _ := cmd.Flags().GetString("ca-url")
		mspDir, _ := cmd.Flags().GetString("msp-dir")
		caTLSCert, _ := cmd.Flags().GetString("ca-tls-cert")

		blue := color.New(color.FgBlue).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()

		if caURL == "" {
			caURL = "https://localhost:7054"
		}
		if mspDir == "" {
			cwd, _ := os.Getwd()
			domain := fmt.Sprintf("%s.nanayam.com", orgLower(org))
			mspDir = filepath.Join(cwd, "crypto-config", "peerOrganizations", domain, "users", fmt.Sprintf("%s@%s", id, domain), "msp")
		}

		fmt.Printf("%s Enrolling user '%s' for org '%s'...\n", blue("→"), id, org)

		bins := fabric.NewBinaries()
		caBinary := bins.CAClientPath()
		homeDir := filepath.Join(os.TempDir(), "nanayam-ca")
		os.MkdirAll(homeDir, 0755)

		caClient := ca.NewClient(caBinary, homeDir, caURL)
		if caTLSCert != "" {
			caClient.TLSConfig = caTLSCert
		}

		os.MkdirAll(mspDir, 0755)
		if err := caClient.Enroll(id, secret, mspDir); err != nil {
			return fmt.Errorf("enroll failed: %w", err)
		}

		fmt.Printf("%s User enrolled to %s\n", green("✓"), mspDir)
		return nil
	},
}

var userListCmd = &cobra.Command{
	Use:   "list",
	Short: "List enrolled users in an MSP directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, _ := os.Getwd()
		mspDir := filepath.Join(cwd, "crypto-config")
		fmt.Printf("Looking for enrolled identities in %s...\n", mspDir)

		// Walk crypto-config to find users
		var users []string
		filepath.Walk(mspDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() && strings.Contains(path, "/users/") {
				rel, _ := filepath.Rel(mspDir, path)
				if rel != "" && !strings.Contains(rel, "keystore") && !strings.Contains(rel, "signcerts") {
					parts := strings.Split(rel, string(filepath.Separator))
					if len(parts) >= 3 && parts[len(parts)-2] == "users" {
						users = append(users, rel)
					}
				}
			}
			return nil
		})

		if len(users) == 0 {
			fmt.Println("No enrolled users found.")
			return nil
		}

		fmt.Println("Enrolled users:")
		for _, u := range users {
			fmt.Printf("  - %s\n", u)
		}
		return nil
	},
}

func orgLower(org string) string {
	org = strings.ToLower(org)
	// Handle common abbreviations
	if org == "acb" || org == "dept" || org == "oversight" || org == "judiciary" {
		return org
	}
	return org
}
