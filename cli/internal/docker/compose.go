package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"
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
	args := make([]string, 0, len(r.Files)*2+1+len(extra))
	for _, f := range r.Files {
		args = append(args, "-f", f)
	}
	args = append(args, subcmd)
	args = append(args, extra...)
	return args
}

func runDockerCompose(args []string) error {
	composeBase, err := dockerComposeBaseCommand()
	if err != nil {
		return err
	}

	cmd := exec.Command(composeBase[0], append(composeBase[1:], args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func dockerComposeBaseCommand() ([]string, error) {
	if err := exec.Command("docker", "compose", "version").Run(); err == nil {
		return []string{"docker", "compose"}, nil
	}

	if _, err := exec.LookPath("docker-compose"); err == nil {
		return []string{"docker-compose"}, nil
	}

	return nil, fmt.Errorf("docker compose is not available; install Docker Compose v2 or docker-compose")
}

// FindComposeFiles returns the path to known compose files in the project
func FindComposeFiles(profile string) ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return FindComposeFilesInDir(cwd, profile)
}

// FindComposeFilesInDir returns the path to known compose files in the given project directory.
func FindComposeFilesInDir(baseDir, profile string) ([]string, error) {

	var files []string
	switch profile {
	case "basic":
		files = []string{
			filepath.Join(baseDir, "docker", "fabric-network.yaml"),
		}
	case "complaint":
		files = []string{
			filepath.Join(baseDir, "docker", "complaint-network.yaml"),
		}
	default:
		// Check if file exists directly
		path := filepath.Join(baseDir, "docker", profile+".yaml")
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

// ResolveComposeFile resolves a compose config path from an absolute path, a relative path,
// or a filename inside the project's docker/ directory.
func ResolveComposeFile(baseDir, config string) (string, error) {
	seen := make(map[string]struct{})
	for _, candidate := range composeConfigCandidates(baseDir, config) {
		if candidate == "" {
			continue
		}
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}

		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("compose config not found: %s", config)
}

func composeConfigCandidates(baseDir, config string) []string {
	if filepath.IsAbs(config) {
		return []string{config}
	}

	candidates := []string{
		filepath.Join(baseDir, config),
		filepath.Join(baseDir, "docker", config),
	}

	if filepath.Ext(config) == "" {
		candidates = append(candidates,
			filepath.Join(baseDir, config+".yaml"),
			filepath.Join(baseDir, "docker", config+".yaml"),
		)
	}

	return candidates
}

type composeDocument struct {
	Services map[string]composeService `yaml:"services"`
}

type composeService struct {
	Volumes []any `yaml:"volumes"`
}

// ValidateComposePrerequisites checks that bind-mounted files and directories required
// by a compose file exist and contain the minimum material expected by Fabric containers.
func ValidateComposePrerequisites(composeFiles []string) error {
	var issues []string
	for _, composeFile := range composeFiles {
		fileIssues, err := validateComposeFilePrerequisites(composeFile)
		if err != nil {
			return err
		}
		issues = append(issues, fileIssues...)
	}

	if len(issues) == 0 {
		return nil
	}

	sort.Strings(issues)
	return fmt.Errorf("missing or incomplete Fabric artifacts:\n  - %s", strings.Join(issues, "\n  - "))
}

func validateComposeFilePrerequisites(composeFile string) ([]string, error) {
	content, err := os.ReadFile(composeFile)
	if err != nil {
		return nil, fmt.Errorf("read compose file %s: %w", composeFile, err)
	}

	var doc composeDocument
	if err := yaml.Unmarshal(content, &doc); err != nil {
		return nil, fmt.Errorf("parse compose file %s: %w", composeFile, err)
	}

	baseDir := filepath.Dir(composeFile)
	var issues []string
	for serviceName, service := range doc.Services {
		for _, volume := range service.Volumes {
			mount, ok := volume.(string)
			if !ok {
				continue
			}

			source, target, ok := parseBindMount(mount)
			if !ok {
				continue
			}

			resolvedSource := source
			if !filepath.IsAbs(resolvedSource) {
				resolvedSource = filepath.Clean(filepath.Join(baseDir, resolvedSource))
			}

			issues = append(issues, validateBindMount(serviceName, resolvedSource, target)...)
		}
	}

	return issues, nil
}

func parseBindMount(mount string) (source string, target string, ok bool) {
	parts := strings.Split(mount, ":")
	if len(parts) < 2 {
		return "", "", false
	}

	source = strings.TrimSpace(parts[0])
	target = strings.TrimSpace(parts[1])
	if source == "" || target == "" {
		return "", "", false
	}

	if !looksLikeBindSource(source) {
		return "", "", false
	}

	return source, target, true
}

func looksLikeBindSource(source string) bool {
	return strings.HasPrefix(source, ".") || strings.HasPrefix(source, "/") || strings.HasPrefix(source, "~")
}

func validateBindMount(serviceName, source, target string) []string {
	if _, err := os.Stat(source); err != nil {
		return []string{fmt.Sprintf("%s: missing %s for mount %s", serviceName, source, target)}
	}

	switch {
	case strings.HasSuffix(target, "/msp"):
		return validateMSPDir(serviceName, source)
	case strings.HasSuffix(target, "/tls"):
		return validateTLSDir(serviceName, source)
	case strings.HasSuffix(target, ".block"):
		return validateRegularFile(serviceName, source)
	default:
		return nil
	}
}

func validateMSPDir(serviceName, source string) []string {
	signcerts := filepath.Join(source, "signcerts")
	entries, err := os.ReadDir(signcerts)
	if err != nil || len(entries) == 0 {
		return []string{fmt.Sprintf("%s: MSP directory %s is missing signcerts", serviceName, source)}
	}
	return nil
}

func validateTLSDir(serviceName, source string) []string {
	required := []string{"ca.crt", "server.crt", "server.key"}
	var issues []string
	for _, name := range required {
		if _, err := os.Stat(filepath.Join(source, name)); err != nil {
			issues = append(issues, fmt.Sprintf("%s: TLS directory %s is missing %s", serviceName, source, name))
		}
	}
	return issues
}

func validateRegularFile(serviceName, source string) []string {
	info, err := os.Stat(source)
	if err != nil {
		return []string{fmt.Sprintf("%s: missing file %s", serviceName, source)}
	}
	if info.IsDir() {
		return []string{fmt.Sprintf("%s: expected file but found directory %s", serviceName, source)}
	}
	return nil
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
