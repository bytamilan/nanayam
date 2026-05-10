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
	rootCmd.AddCommand(gatewayCmd)
	gatewayCmd.Flags().String("grpc-port", ":50051", "gRPC server port")
	gatewayCmd.Flags().String("http-port", ":8080", "REST server port")
	gatewayCmd.Flags().Bool("docker", false, "Run gateway in Docker")
}

var gatewayCmd = &cobra.Command{
	Use:   "gateway",
	Short: "Start the Go distribution gateway service",
	Long: `Build and run the Go gRPC/REST gateway that connects to Fabric.

This wraps the services/gateway Go application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		grpcPort, _ := cmd.Flags().GetString("grpc-port")
		httpPort, _ := cmd.Flags().GetString("http-port")
		useDocker, _ := cmd.Flags().GetBool("docker")

		blue := color.New(color.FgBlue).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()

		fmt.Printf("%s Starting gateway (gRPC: %s, HTTP: %s)...\n", blue("→"), grpcPort, httpPort)

		cwd, _ := os.Getwd()
		gatewayDir := filepath.Join(cwd, "services", "gateway")

		if useDocker {
			fmt.Println("Building and running gateway Docker container...")
			dockerCmd := exec.Command("docker", "build", "-t", "nanayam-gateway", gatewayDir)
			dockerCmd.Stdout = os.Stdout
			dockerCmd.Stderr = os.Stderr
			if err := dockerCmd.Run(); err != nil {
				return fmt.Errorf("docker build failed: %w", err)
			}
			runCmd := exec.Command("docker", "run", "-d",
				"--name", "nanayam-gateway",
				"--network", "nanayam",
				"-p", httpPort+":"+httpPort,
				"-p", grpcPort+":"+grpcPort,
				"nanayam-gateway",
			)
			runCmd.Stdout = os.Stdout
			runCmd.Stderr = os.Stderr
			if err := runCmd.Run(); err != nil {
				return fmt.Errorf("docker run failed: %w", err)
			}
		} else {
			// Check if binary exists
			binary := filepath.Join(gatewayDir, "gateway")
			if _, err := os.Stat(binary); os.IsNotExist(err) {
				fmt.Printf("%s Building gateway binary...\n", blue("→"))
				buildCmd := exec.Command("go", "build", "-o", "gateway", ".")
				buildCmd.Dir = gatewayDir
				buildCmd.Stdout = os.Stdout
				buildCmd.Stderr = os.Stderr
				if err := buildCmd.Run(); err != nil {
					return fmt.Errorf("build failed: %w", err)
				}
			}

			fmt.Printf("%s Running gateway...\n", blue("→"))
			runCmd := exec.Command(binary,
				"-grpc-port", grpcPort,
				"-http-port", httpPort,
			)
			runCmd.Dir = cwd
			runCmd.Stdout = os.Stdout
			runCmd.Stderr = os.Stderr
			if err := runCmd.Run(); err != nil {
				return fmt.Errorf("gateway exited: %w", err)
			}
		}

		fmt.Printf("%s Gateway running\n", green("✓"))
		return nil
	},
}
