package fabric

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// PeerEnv holds environment variables for peer commands
type PeerEnv struct {
	Address       string
	LocalMSPID    string
	TLSCertFile   string
	MSPConfigPath string
}

// DefaultPeerEnv returns the default peer environment for a given org
func DefaultPeerEnv(org string) *PeerEnv {
	cwd, _ := os.Getwd()
	domain := fmt.Sprintf("%s.nanayam.com", org)
	if org == "acb" || org == "dept" || org == "oversight" || org == "judiciary" {
		domain = fmt.Sprintf("%s.nanayam.com", org)
	}

	peerID := fmt.Sprintf("peer0.%s", domain)
	peerPort := "7051"
	if org == "Org2" || org == "dept" {
		peerPort = "9051"
	} else if org == "oversight" {
		peerPort = "10051"
	} else if org == "judiciary" {
		peerPort = "11051"
	}

	cryptoBase := filepath.Join(cwd, "crypto-config", "peerOrganizations", domain)
	return &PeerEnv{
		Address:       fmt.Sprintf("%s:%s", peerID, peerPort),
		LocalMSPID:    fmt.Sprintf("%sMSP", org),
		TLSCertFile:   filepath.Join(cryptoBase, "peers", peerID, "tls", "ca.crt"),
		MSPConfigPath: filepath.Join(cryptoBase, "users", fmt.Sprintf("Admin@%s", domain), "msp"),
	}
}

// Exec runs a peer command with the correct environment
func (e *PeerEnv) Exec(peerBinary string, args ...string) *exec.Cmd {
	cmd := exec.Command(peerBinary, args...)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("CORE_PEER_ADDRESS=%s", e.Address),
		fmt.Sprintf("CORE_PEER_LOCALMSPID=%s", e.LocalMSPID),
		fmt.Sprintf("CORE_PEER_TLS_ROOTCERT_FILE=%s", e.TLSCertFile),
		fmt.Sprintf("CORE_PEER_MSPCONFIGPATH=%s", e.MSPConfigPath),
		"CORE_PEER_TLS_ENABLED=true",
		"FABRIC_CFG_PATH=./config",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

// DockerExec runs a peer command inside the cli container
func DockerExec(args ...string) *exec.Cmd {
	dockerArgs := append([]string{
		"exec", "cli", "peer",
	}, args...)
	cmd := exec.Command("docker", dockerArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

// OrdererTLSPath returns the path to the orderer TLS CA cert
func OrdererTLSPath() string {
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, "crypto-config", "ordererOrganizations", "nanayam.com", "orderers", "orderer.nanayam.com", "msp", "tlscacerts", "tlsca.nanayam.com-cert.pem")
}

// DefaultOrderer returns the default orderer endpoint
func DefaultOrderer() string {
	return "orderer.nanayam.com:7050"
}
