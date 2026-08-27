package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(consoleCmd)
	consoleCmd.Flags().String("port", "3000", "Next.js dev server port")
	consoleCmd.Flags().Bool("docker", false, "Run console in Docker instead of local dev server")
}

var consoleCmd = &cobra.Command{
	Use:   "console",
	Short: "Start the Next.js operator console",
	Long: `Run the Next.js operator console UI locally or in Docker.

Default runs 'npm run dev' in apps/org-console/.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetString("port")
		useDocker, _ := cmd.Flags().GetBool("docker")

		blue := color.New(color.FgBlue).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()

		fmt.Printf("%s Starting operator console (port: %s, docker: %v)...\n", blue("→"), port, useDocker)

		cwd, _ := os.Getwd()
		consoleDir := filepath.Join(cwd, "apps", "org-console")

		if useDocker {
			fmt.Println("Building and running console Docker container...")
			dockerCmd := exec.Command("docker", "build", "-t", "nanayam-console", consoleDir)
			dockerCmd.Stdout = os.Stdout
			dockerCmd.Stderr = os.Stderr
			if err := dockerCmd.Run(); err != nil {
				return fmt.Errorf("docker build failed: %w", err)
			}
			runCmd := exec.Command("docker", "run", "-d",
				"--name", "nanayam-console",
				"--network", "nanayam",
				"-p", port+":"+port,
				"nanayam-console",
			)
			runCmd.Stdout = os.Stdout
			runCmd.Stderr = os.Stderr
			if err := runCmd.Run(); err != nil {
				return fmt.Errorf("docker run failed: %w", err)
			}
		} else {
			// Check for node_modules
			nodeModules := filepath.Join(consoleDir, "node_modules")
			if _, err := os.Stat(nodeModules); os.IsNotExist(err) {
				fmt.Printf("%s Installing dependencies...\n", blue("→"))
				installCmd := exec.Command("npm", "install")
				installCmd.Dir = consoleDir
				installCmd.Stdout = os.Stdout
				installCmd.Stderr = os.Stderr
				if err := installCmd.Run(); err != nil {
					return fmt.Errorf("npm install failed: %w", err)
				}
			}

			fmt.Printf("%s Running Next.js dev server...\n", blue("→"))
			devCmd := exec.Command("npm", "run", "dev", "--", "-p", port)
			devCmd.Dir = consoleDir
			devCmd.Stdout = os.Stdout
			devCmd.Stderr = os.Stderr
			if err := devCmd.Run(); err != nil {
				return fmt.Errorf("dev server exited: %w", err)
			}
		}

		fmt.Printf("%s Console running at http://localhost:%s\n", green("✓"), port)
		return nil
	},
}
