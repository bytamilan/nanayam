package cmd

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	nanayamReleaseAPI       = "https://api.github.com/repos/bytamilan/nanayam/releases/latest"
	nanayamReleaseDownloads = "https://github.com/bytamilan/nanayam/releases/download"
	nanayamUserAgent        = "nanayam-cli/self-update"
)

type parsedSemver struct {
	major      int
	minor      int
	patch      int
	prerelease string
	ok         bool
}

type selfUpdateOptions struct {
	Version    string
	WithFabric bool
	Setup      bool
	DevLocal   bool
	Refresh    bool
	CheckOnly  bool
	SourcePath string
}

type githubRelease struct {
	TagName string `json:"tag_name"`
}

func addSelfUpdateFlags(cmd *cobra.Command, includeCheck bool) {
	cmd.Flags().String("version", "", "Install a specific version (default: latest)")
	cmd.Flags().Bool("with-fabric", false, "Also download Fabric binaries")
	cmd.Flags().Bool("setup", false, "Run prerequisites --auto after install")
	cmd.Flags().Bool("dev-local", false, "Build from a local Nanayam checkout instead of downloading a release")
	cmd.Flags().Bool("refresh", false, "Reinstall even when the target version matches the current version")
	cmd.Flags().String("source", "", "Local Nanayam repository path to use with --dev-local")
	if includeCheck {
		cmd.Flags().Bool("check", false, "Check for a newer version without installing it")
	}
}

func selfUpdateOptionsFromFlags(cmd *cobra.Command) (selfUpdateOptions, error) {
	ver, err := cmd.Flags().GetString("version")
	if err != nil {
		return selfUpdateOptions{}, err
	}
	withFabric, err := cmd.Flags().GetBool("with-fabric")
	if err != nil {
		return selfUpdateOptions{}, err
	}
	setup, err := cmd.Flags().GetBool("setup")
	if err != nil {
		return selfUpdateOptions{}, err
	}
	devLocal, err := cmd.Flags().GetBool("dev-local")
	if err != nil {
		return selfUpdateOptions{}, err
	}
	refresh, err := cmd.Flags().GetBool("refresh")
	if err != nil {
		return selfUpdateOptions{}, err
	}
	sourcePath, err := cmd.Flags().GetString("source")
	if err != nil {
		return selfUpdateOptions{}, err
	}
	checkOnly := false
	if flag := cmd.Flags().Lookup("check"); flag != nil {
		checkOnly, err = cmd.Flags().GetBool("check")
		if err != nil {
			return selfUpdateOptions{}, err
		}
	}

	return selfUpdateOptions{
		Version:    strings.TrimSpace(ver),
		WithFabric: withFabric,
		Setup:      setup,
		DevLocal:   devLocal,
		Refresh:    refresh,
		CheckOnly:  checkOnly,
		SourcePath: strings.TrimSpace(sourcePath),
	}, nil
}

func runInstallFlow(cmd *cobra.Command, _ []string) error {
	opts, err := selfUpdateOptionsFromFlags(cmd)
	if err != nil {
		return err
	}

	if opts.CheckOnly {
		return fmt.Errorf("--check is only supported with 'nanayam upgrade'")
	}

	if opts.DevLocal {
		return installLocalBuild(opts, "install")
	}

	targetVersion, err := resolveReleaseVersion(opts.Version)
	if err != nil {
		return err
	}

	return installDownloadedRelease(targetVersion, opts, "install")
}

func runUpgradeFlow(cmd *cobra.Command, _ []string) error {
	opts, err := selfUpdateOptionsFromFlags(cmd)
	if err != nil {
		return err
	}

	if opts.DevLocal {
		if opts.CheckOnly {
			root, err := resolveLocalRepoRoot(opts.SourcePath)
			if err != nil {
				return err
			}
			fmt.Printf("Local development source detected at %s\n", root)
			fmt.Println("Run 'nanayam upgrade --dev-local --refresh' to rebuild and refresh the installed binary.")
			return nil
		}
		return installLocalBuild(opts, "upgrade")
	}

	targetVersion, err := resolveReleaseVersion(opts.Version)
	if err != nil {
		return err
	}

	available := opts.Refresh || isUpgradeAvailable(version, targetVersion)
	printUpgradeStatus(version, targetVersion, available, opts.Refresh)

	if opts.CheckOnly {
		return nil
	}

	if !available {
		fmt.Println("Already on the latest release. Use --refresh to reinstall it anyway.")
		return nil
	}

	return installDownloadedRelease(targetVersion, opts, "upgrade")
}

func printUpgradeStatus(currentVersion, targetVersion string, available, refresh bool) {
	current := strings.TrimSpace(currentVersion)
	if current == "" {
		current = "unknown"
	}
	fmt.Printf("Current version: %s\n", current)
	fmt.Printf("Target version:  %s\n", targetVersion)

	switch {
	case refresh:
		fmt.Println("Refresh requested: the target version will be reinstalled.")
	case available:
		fmt.Println("Upgrade available.")
	default:
		fmt.Println("You are already up to date.")
	}
}

func installDownloadedRelease(targetVersion string, opts selfUpdateOptions, action string) error {
	fmt.Printf("%s nanayam %s for %s/%s...\n", titleWord(action), targetVersion, runtime.GOOS, runtime.GOARCH)

	installPath, err := installReleaseBinary(targetVersion)
	if err != nil {
		return err
	}

	if err := postInstallSetup(opts); err != nil {
		return err
	}

	printInstallCompletion(targetVersion, installPath, action)
	return nil
}

func installLocalBuild(opts selfUpdateOptions, action string) error {
	repoRoot, err := resolveLocalRepoRoot(opts.SourcePath)
	if err != nil {
		return err
	}

	fmt.Printf("%s nanayam from local source at %s...\n", titleWord(action), repoRoot)

	installPath, builtVersion, err := buildAndInstallLocalBinary(repoRoot)
	if err != nil {
		return err
	}

	if err := postInstallSetup(opts); err != nil {
		return err
	}

	printInstallCompletion(builtVersion, installPath, action)
	return nil
}

func postInstallSetup(opts selfUpdateOptions) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	if err := ensurePathConfigured(home); err != nil {
		return err
	}

	if opts.WithFabric {
		fmt.Println()
		fmt.Println("Downloading Fabric binaries...")
		if err := downloadFabricBinaries(); err != nil {
			return err
		}
		fmt.Println("Fabric binaries installed.")
	}

	if opts.Setup {
		fmt.Println()
		fmt.Println("Running prerequisites setup...")
		if err := runAutoPrerequisites(); err != nil {
			return err
		}
	}

	return nil
}

func printInstallCompletion(installedVersion, installPath, action string) {
	fmt.Println()
	if action == "upgrade" {
		fmt.Printf("Upgrade complete: %s\n", installedVersion)
	} else {
		fmt.Printf("Installation complete: %s\n", installedVersion)
	}
	fmt.Printf("Binary: %s\n", installPath)
	fmt.Println("Run 'nanayam version' to verify.")
}

func runAutoPrerequisites() error {
	cmd := &cobra.Command{Use: "prerequisites"}
	cmd.Flags().Bool("auto", false, "")
	cmd.Flags().Bool("install-fabric", false, "")
	if err := cmd.Flags().Set("auto", "true"); err != nil {
		return err
	}
	return runPrereqs(cmd, nil)
}

func resolveReleaseVersion(requestedVersion string) (string, error) {
	requested := strings.TrimSpace(requestedVersion)
	if requested == "" || strings.EqualFold(requested, "latest") {
		return fetchLatestReleaseTag()
	}
	if strings.HasPrefix(requested, "v") {
		return requested, nil
	}
	return "v" + requested, nil
}

func fetchLatestReleaseTag() (string, error) {
	req, err := http.NewRequest(http.MethodGet, nanayamReleaseAPI, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", nanayamUserAgent)

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("fetch latest release: unexpected status %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("decode latest release: %w", err)
	}
	if strings.TrimSpace(release.TagName) == "" {
		return "", fmt.Errorf("latest release did not include a tag name")
	}
	return strings.TrimSpace(release.TagName), nil
}

func installReleaseBinary(targetVersion string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	binDir := filepath.Join(home, ".nanayam", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return "", err
	}

	installPath := filepath.Join(binDir, runtimeBinaryName())
	binaryPath, err := downloadReleaseBinary(targetVersion)
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(filepath.Dir(binaryPath))

	if err := replaceInstalledBinary(binaryPath, installPath); err != nil {
		return "", err
	}

	return installPath, nil
}

func downloadReleaseBinary(targetVersion string) (string, error) {
	assetName := releaseAssetName(targetVersion, runtime.GOOS, runtime.GOARCH)
	url := releaseAssetURL(targetVersion, assetName)

	tmpDir, err := os.MkdirTemp("", "nanayam-release-*")
	if err != nil {
		return "", err
	}

	archivePath := filepath.Join(tmpDir, assetName)
	if err := downloadToFile(url, archivePath); err != nil {
		os.RemoveAll(tmpDir)
		return "", err
	}

	binaryPath, err := extractReleaseBinary(archivePath, tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", err
	}

	return binaryPath, nil
}

func releaseAssetName(versionTag, goos, goarch string) string {
	ext := ".tar.gz"
	if goos == "windows" {
		ext = ".zip"
	}
	return fmt.Sprintf("nanayam_%s_%s_%s%s", versionTag, goos, goarch, ext)
}

func releaseAssetURL(versionTag, assetName string) string {
	if base := strings.TrimSpace(os.Getenv("NANAYAM_RELEASE_BASE_URL")); base != "" {
		return fmt.Sprintf("%s/%s/%s", strings.TrimRight(base, "/"), versionTag, assetName)
	}
	return fmt.Sprintf("%s/%s/%s", nanayamReleaseDownloads, versionTag, assetName)
}

func downloadToFile(url, destPath string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", nanayamUserAgent)

	client := &http.Client{Timeout: 2 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("download %s: unexpected status %s: %s", url, resp.Status, strings.TrimSpace(string(body)))
	}

	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return err
	}

	return nil
}

func extractReleaseBinary(archivePath, destDir string) (string, error) {
	if strings.HasSuffix(archivePath, ".zip") {
		return extractZipBinary(archivePath, destDir)
	}
	return extractTarGzBinary(archivePath, destDir)
}

func extractTarGzBinary(archivePath, destDir string) (string, error) {
	archive, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer archive.Close()

	gzReader, err := gzip.NewReader(archive)
	if err != nil {
		return "", err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)
	targetName := runtimeBinaryName()
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		if filepath.Base(header.Name) != targetName {
			continue
		}

		outPath := filepath.Join(destDir, targetName)
		outFile, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(outFile, tarReader); err != nil {
			outFile.Close()
			return "", err
		}
		if err := outFile.Close(); err != nil {
			return "", err
		}
		return outPath, nil
	}

	return "", fmt.Errorf("binary %s not found in %s", targetName, archivePath)
}

func extractZipBinary(archivePath, destDir string) (string, error) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	targetName := runtimeBinaryName()
	for _, file := range reader.File {
		if file.FileInfo().IsDir() || filepath.Base(file.Name) != targetName {
			continue
		}

		in, err := file.Open()
		if err != nil {
			return "", err
		}
		outPath := filepath.Join(destDir, targetName)
		out, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
		if err != nil {
			in.Close()
			return "", err
		}
		if _, err := io.Copy(out, in); err != nil {
			in.Close()
			out.Close()
			return "", err
		}
		in.Close()
		if err := out.Close(); err != nil {
			return "", err
		}
		return outPath, nil
	}

	return "", fmt.Errorf("binary %s not found in %s", targetName, archivePath)
}

func replaceInstalledBinary(binaryPath string, installPath string) error {
	if runtime.GOOS == "windows" {
		currentExe, currentErr := os.Executable()
		if currentErr == nil && samePath(currentExe, installPath) {
			return fmt.Errorf("self-upgrade on Windows must be run from install.ps1 or an external process")
		}
	}

	if err := os.MkdirAll(filepath.Dir(installPath), 0755); err != nil {
		return err
	}

	tmpPath := filepath.Join(filepath.Dir(installPath), "."+filepath.Base(installPath)+".tmp")
	if err := copyFile(binaryPath, tmpPath, 0755); err != nil {
		return err
	}

	if err := os.Chmod(tmpPath, 0755); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, installPath); err != nil {
		if removeErr := os.Remove(installPath); removeErr == nil {
			if retryErr := os.Rename(tmpPath, installPath); retryErr == nil {
				return nil
			}
		}
		return err
	}

	return nil
}

func copyFile(srcPath, destPath string, mode os.FileMode) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dest, err := os.OpenFile(destPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}

	if _, err := io.Copy(dest, src); err != nil {
		dest.Close()
		return err
	}

	return dest.Close()
}

func buildAndInstallLocalBinary(repoRoot string) (string, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}
	binDir := filepath.Join(home, ".nanayam", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return "", "", err
	}

	builtVersion := gitValue(repoRoot, "describe", "--tags", "--always", "--dirty")
	if builtVersion == "" {
		builtVersion = "dev-local"
	}
	commitValue := gitValue(repoRoot, "rev-parse", "--short", "HEAD")
	if commitValue == "" {
		commitValue = "local"
	}
	dateValue := time.Now().UTC().Format(time.RFC3339)

	tmpDir, err := os.MkdirTemp("", "nanayam-local-build-*")
	if err != nil {
		return "", "", err
	}
	defer os.RemoveAll(tmpDir)

	binaryPath := filepath.Join(tmpDir, runtimeBinaryName())
	ldflags := fmt.Sprintf("-X github.com/bytamilan/nanayam/cli/cmd.version=%s -X github.com/bytamilan/nanayam/cli/cmd.commit=%s -X github.com/bytamilan/nanayam/cli/cmd.date=%s", builtVersion, commitValue, dateValue)
	buildCmd := exec.Command("go", "build", "-ldflags", ldflags, "-o", binaryPath, ".")
	buildCmd.Dir = filepath.Join(repoRoot, "cli")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return "", "", fmt.Errorf("build local CLI: %w", err)
	}

	installPath := filepath.Join(binDir, runtimeBinaryName())
	if err := replaceInstalledBinary(binaryPath, installPath); err != nil {
		return "", "", err
	}

	return installPath, builtVersion, nil
}

func resolveLocalRepoRoot(sourcePath string) (string, error) {
	start := strings.TrimSpace(sourcePath)
	if start == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		start = cwd
	}

	absStart, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}

	if fileExists(filepath.Join(absStart, "cli", "go.mod")) {
		return absStart, nil
	}
	if fileExists(filepath.Join(absStart, "go.mod")) && filepath.Base(absStart) == "cli" {
		return filepath.Dir(absStart), nil
	}

	for dir := absStart; ; dir = filepath.Dir(dir) {
		if fileExists(filepath.Join(dir, "cli", "go.mod")) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}

	return "", fmt.Errorf("could not find a Nanayam repository from %s; pass --source with the repo path", absStart)
}

func gitValue(repoRoot string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func ensurePathConfigured(home string) error {
	shellCfg := getShellConfigPath(home)
	if shellCfg == "" {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(shellCfg), 0755); err != nil {
		return err
	}

	data, err := os.ReadFile(shellCfg)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if strings.Contains(string(data), ".nanayam/bin") {
		return nil
	}

	pathEntry := "\n# Nanayam CLI\nexport PATH=\"$HOME/.nanayam/bin:$PATH\"\n"
	f, err := os.OpenFile(shellCfg, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(pathEntry); err != nil {
		return err
	}

	fmt.Printf("Added ~/.nanayam/bin to PATH in %s\n", shellCfg)
	fmt.Println("Run 'source " + shellCfg + "' or restart your terminal to apply.")
	return nil
}

func getShellConfigPath(home string) string {
	shell := os.Getenv("SHELL")
	switch {
	case strings.Contains(shell, "zsh"):
		return filepath.Join(home, ".zshrc")
	case strings.Contains(shell, "bash"):
		return filepath.Join(home, ".bashrc")
	default:
		return filepath.Join(home, ".profile")
	}
}

func isUpgradeAvailable(currentVersion, targetVersion string) bool {
	current := strings.TrimSpace(currentVersion)
	target := strings.TrimSpace(targetVersion)
	if current == "" || current == "dev" || current == "unknown" {
		return current != target
	}
	if current == target {
		return false
	}

	currentSemver := parseSemver(current)
	targetSemver := parseSemver(target)
	if currentSemver.ok && targetSemver.ok {
		return compareParsedSemver(currentSemver, targetSemver) < 0
	}

	return current != target
}

func normalizeSemver(value string) string {
	v := strings.TrimSpace(value)
	if v == "" {
		return ""
	}
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	if !parseSemver(v).ok {
		return ""
	}
	return v
}

func parseSemver(value string) parsedSemver {
	v := strings.TrimSpace(value)
	v = strings.TrimPrefix(v, "v")
	if v == "" {
		return parsedSemver{}
	}

	core := v
	prerelease := ""
	if idx := strings.Index(core, "+"); idx >= 0 {
		core = core[:idx]
	}
	if idx := strings.Index(core, "-"); idx >= 0 {
		prerelease = core[idx+1:]
		core = core[:idx]
	}

	parts := strings.Split(core, ".")
	if len(parts) == 0 || len(parts) > 3 {
		return parsedSemver{}
	}

	nums := []int{0, 0, 0}
	for i, part := range parts {
		if part == "" {
			return parsedSemver{}
		}
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return parsedSemver{}
		}
		nums[i] = n
	}

	return parsedSemver{
		major:      nums[0],
		minor:      nums[1],
		patch:      nums[2],
		prerelease: prerelease,
		ok:         true,
	}
}

func compareParsedSemver(a, b parsedSemver) int {
	if a.major != b.major {
		return compareInts(a.major, b.major)
	}
	if a.minor != b.minor {
		return compareInts(a.minor, b.minor)
	}
	if a.patch != b.patch {
		return compareInts(a.patch, b.patch)
	}
	if a.prerelease == b.prerelease {
		return 0
	}
	if a.prerelease == "" {
		return 1
	}
	if b.prerelease == "" {
		return -1
	}
	return strings.Compare(a.prerelease, b.prerelease)
}

func compareInts(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func titleWord(value string) string {
	if value == "" {
		return ""
	}
	return strings.ToUpper(value[:1]) + value[1:]
}

func runtimeBinaryName() string {
	if runtime.GOOS == "windows" {
		return "nanayam.exe"
	}
	return "nanayam"
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func samePath(a, b string) bool {
	cleanA := filepath.Clean(a)
	cleanB := filepath.Clean(b)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(cleanA, cleanB)
	}
	return cleanA == cleanB
}
