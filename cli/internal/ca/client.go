package ca

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Client wraps fabric-ca-client commands
type Client struct {
	BinaryPath string
	HomeDir    string
	ServerURL  string
	TLSConfig  string
}

// NewClient creates a new CA client
func NewClient(binaryPath, homeDir, serverURL string) *Client {
	return &Client{
		BinaryPath: binaryPath,
		HomeDir:    homeDir,
		ServerURL:  serverURL,
	}
}

// Enroll enrolls an identity
func (c *Client) Enroll(id, secret, mspDir string, attrs ...string) error {
	args := []string{
		"enroll",
		"-u", fmt.Sprintf("%s:%s@%s", id, secret, c.ServerURL),
		"--mspdir", mspDir,
		"--home", c.HomeDir,
	}
	if c.TLSConfig != "" {
		args = append(args, "--tls.certfiles", c.TLSConfig)
	}
	for _, attr := range attrs {
		args = append(args, "--enrollment.attrs", attr)
	}

	cmd := exec.Command(c.BinaryPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Register registers a new identity (requires registrar to be enrolled first)
func (c *Client) Register(registrar, registrarSecret, id, secret, userType, affiliation string, attrs ...string) error {
	// First ensure registrar is enrolled
	registrarMSP := filepath.Join(c.HomeDir, registrar+"-msp")
	if _, err := os.Stat(registrarMSP); os.IsNotExist(err) {
		if err := c.Enroll(registrar, registrarSecret, registrarMSP); err != nil {
			return fmt.Errorf("enroll registrar: %w", err)
		}
	}

	args := []string{
		"register",
		"--id.name", id,
		"--id.secret", secret,
		"--id.type", userType,
		"--id.affiliation", affiliation,
		"--mspdir", registrarMSP,
		"--home", c.HomeDir,
	}
	if c.TLSConfig != "" {
		args = append(args, "--tls.certfiles", c.TLSConfig)
	}

	cmd := exec.Command(c.BinaryPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Reenroll re-enrolls an existing identity
func (c *Client) Reenroll(mspDir string) error {
	args := []string{
		"reenroll",
		"--mspdir", mspDir,
		"--home", c.HomeDir,
	}
	if c.TLSConfig != "" {
		args = append(args, "--tls.certfiles", c.TLSConfig)
	}

	cmd := exec.Command(c.BinaryPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// GetCAInfo gets CA information
func (c *Client) GetCAInfo() error {
	args := []string{
		"getcainfo",
		"-u", c.ServerURL,
		"--home", c.HomeDir,
	}
	cmd := exec.Command(c.BinaryPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
