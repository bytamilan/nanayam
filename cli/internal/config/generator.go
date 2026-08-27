package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/bytamilan/nanayam/cli/templates"
)

// PeerNodeConfig holds parameters for peer node initialization
type PeerNodeConfig struct {
	OrgName       string
	OrgDomain     string
	Domain        string
	PeerCount     int
	UserCount     int
	PeerID        string
	PeerPort      int
	ChaincodePort int
	OpsPort       int
	MSPID         string
	CryptoPath    string
}

// OrdererNodeConfig holds parameters for orderer node initialization
type OrdererNodeConfig struct {
	OrdererID    string
	OrdererPort  int
	MSPID        string
	CryptoPath   string
	GenesisBlock string
}

// CANodeConfig holds parameters for CA node initialization
type CANodeConfig struct {
	OrgName      string
	OrgNameLower string
	CAPort       int
	OpsPort      int
}

// GenerateCryptogen writes a cryptogen config file for peer orgs
func GenerateCryptogen(cfg *PeerNodeConfig, outputPath string) error {
	tmplContent, err := templates.FS.ReadFile("config/cryptogen-peer.yaml")
	if err != nil {
		return fmt.Errorf("read template: %w", err)
	}

	return executeTemplate(string(tmplContent), cfg, outputPath)
}

// GeneratePeerCompose writes a docker-compose file for a peer
func GeneratePeerCompose(cfg *PeerNodeConfig, outputPath string) error {
	tmplContent, err := templates.FS.ReadFile("docker/peer.yaml")
	if err != nil {
		return fmt.Errorf("read template: %w", err)
	}

	return executeTemplate(string(tmplContent), cfg, outputPath)
}

// GenerateOrdererCompose writes a docker-compose file for an orderer
func GenerateOrdererCompose(cfg *OrdererNodeConfig, outputPath string) error {
	tmplContent, err := templates.FS.ReadFile("docker/orderer.yaml")
	if err != nil {
		return fmt.Errorf("read template: %w", err)
	}

	return executeTemplate(string(tmplContent), cfg, outputPath)
}

// GenerateCACompose writes a docker-compose file for a CA
func GenerateCACompose(cfg *CANodeConfig, outputPath string) error {
	cfg.OrgNameLower = strings.ToLower(cfg.OrgName)

	tmplContent, err := templates.FS.ReadFile("docker/ca.yaml")
	if err != nil {
		return fmt.Errorf("read template: %w", err)
	}

	return executeTemplate(string(tmplContent), cfg, outputPath)
}

func executeTemplate(tmplStr string, data interface{}, outputPath string) error {
	tmpl, err := template.New("config").Parse(tmplStr)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	os.MkdirAll(filepath.Dir(outputPath), 0755)
	return os.WriteFile(outputPath, buf.Bytes(), 0644)
}
