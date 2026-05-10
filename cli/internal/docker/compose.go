package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ComposeRunner wraps docker compose commands
type ComposeRunner struct {
	Files []string
}

// NewComposeRunner creates a runner with the given compose files
func NewComposeRunner(files ...string) *ComposeRunner {
	return &ComposeRunner{Files: files}
}

// Up runs docker compose up -d
func (r *ComposeRunner) Up() error {
	args := r.buildArgs("up", "-d")
	return runDockerCompose(args)
}

// Down runs docker compose down
func (r *ComposeRunner) Down(volumes bool) error {
	args := r.buildArgs("down")
	if volumes {
		args = append(args, "-v")
	}
	return runDockerCompose(args)
}

// Ps runs docker compose ps
func (r *ComposeRunner) Ps() error {
	args := r.buildArgs("ps")
	return runDockerCompose(args)
}

// Logs runs docker compose logs -f
func (r *ComposeRunner) Logs(service string) error {
	args := r.buildArgs("logs", "-f")
	if service != "" {
		args = append(args, service)
	}
	return runDockerCompose(args)
}

// Start starts specific services
func (r *ComposeRunner) Start(services ...string) error {
	args := r.buildArgs("start")
	args = append(args, services...)
	return runDockerCompose(args)
}

// Stop stops specific services
func (r *ComposeRunner) Stop(services ...string) error {
	args := r.buildArgs("stop")
	args = append(args, services...)
	return runDockerCompose(args)
}

func (r *ComposeRunner) buildArgs(subcmd string, extra ...string) []string {
	args := []string{subcmd}
	for _, f := range r.Files {
		args = append(args, "-f", f)
	}
	args = append(args, extra...)
	return args
}

func runDockerCompose(args []string) error {
	// Try "docker compose" first, fall back to "docker-compose"
	cmd := exec.Command("docker", append([]string{"compose"}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// Fallback to docker-compose
		cmd = exec.Command("docker-compose", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	return nil
}

// FindComposeFiles returns the path to known compose files in the project
func FindComposeFiles(profile string) ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	var files []string
	switch profile {
	case "basic":
		files = []string{
			filepath.Join(cwd, "docker", "fabric-network.yaml"),
		}
	case "complaint":
		files = []string{
			filepath.Join(cwd, "docker", "complaint-network.yaml"),
		}
	default:
		// Check if file exists directly
		path := filepath.Join(cwd, "docker", profile+".yaml")
		if _, err := os.Stat(path); err == nil {
			files = []string{path}
		} else {
			return nil, fmt.Errorf("unknown network profile: %s", profile)
		}
	}

	// Verify files exist
	for _, f := range files {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			return nil, fmt.Errorf("compose file not found: %s", f)
		}
	}
	return files, nil
}

// WriteComposeFile writes a docker-compose snippet to a file
func WriteComposeFile(dir, filename string, content []byte) error {
	os.MkdirAll(dir, 0755)
	path := filepath.Join(dir, filename)
	return os.WriteFile(path, content, 0644)
}

// ContainerStatus represents a single container's status
type ContainerStatus struct {
	Name   string
	Status string
	Ports  string
}

// ListContainers lists containers matching the project
func ListContainers(project string) ([]ContainerStatus, error) {
	cmd := exec.Command("docker", "ps", "--format", "{{.Names}}\t{{.Status}}\t{{.Ports}}")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var containers []ContainerStatus
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 2 {
			containers = append(containers, ContainerStatus{
				Name:   parts[0],
				Status: parts[1],
				Ports:  "",
			})
			if len(parts) >= 3 {
				containers[len(containers)-1].Ports = parts[2]
			}
		}
	}
	return containers, nil
}
