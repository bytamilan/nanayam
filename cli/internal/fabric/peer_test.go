package fabric

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultPeerEnvPortsPerOrg(t *testing.T) {
	// Each org in the complaint network listens on its own peer port; a wrong
	// port here silently sends every peer command to the wrong organization.
	cases := map[string]string{
		"Org1":      "7051",
		"Org2":      "9051",
		"acb":       "7051",
		"dept":      "9051",
		"oversight": "10051",
		"judiciary": "11051",
	}

	for org, wantPort := range cases {
		t.Run(org, func(t *testing.T) {
			env := DefaultPeerEnv(org)

			host, port, found := strings.Cut(env.Address, ":")
			if !found {
				t.Fatalf("Address = %q, want host:port", env.Address)
			}
			if port != wantPort {
				t.Errorf("port = %q, want %q", port, wantPort)
			}
			if want := "peer0." + org + ".nanayam.com"; host != want {
				t.Errorf("host = %q, want %q", host, want)
			}
		})
	}
}

func TestDefaultPeerEnvMSPID(t *testing.T) {
	if got, want := DefaultPeerEnv("Org1").LocalMSPID, "Org1MSP"; got != want {
		t.Errorf("LocalMSPID = %q, want %q", got, want)
	}
	if got, want := DefaultPeerEnv("acb").LocalMSPID, "acbMSP"; got != want {
		t.Errorf("LocalMSPID = %q, want %q", got, want)
	}
}

func TestDefaultPeerEnvCryptoPaths(t *testing.T) {
	project := t.TempDir()
	t.Chdir(project)

	env := DefaultPeerEnv("Org1")

	wantTLS := filepath.Join(project, "crypto-config", "peerOrganizations",
		"Org1.nanayam.com", "peers", "peer0.Org1.nanayam.com", "tls", "ca.crt")
	if env.TLSCertFile != wantTLS {
		t.Errorf("TLSCertFile = %q, want %q", env.TLSCertFile, wantTLS)
	}

	wantMSP := filepath.Join(project, "crypto-config", "peerOrganizations",
		"Org1.nanayam.com", "users", "Admin@Org1.nanayam.com", "msp")
	if env.MSPConfigPath != wantMSP {
		t.Errorf("MSPConfigPath = %q, want %q", env.MSPConfigPath, wantMSP)
	}
}

func TestPeerEnvExecSetsFabricEnvironment(t *testing.T) {
	env := &PeerEnv{
		Address:       "peer0.org1.nanayam.com:7051",
		LocalMSPID:    "Org1MSP",
		TLSCertFile:   "/crypto/tls/ca.crt",
		MSPConfigPath: "/crypto/users/Admin/msp",
	}

	cmd := env.Exec("/opt/fabric/bin/peer", "channel", "list")

	if got, want := filepath.Base(cmd.Path), "peer"; got != want {
		t.Errorf("command = %q, want %q", got, want)
	}
	if got := cmd.Args[1:]; len(got) != 2 || got[0] != "channel" || got[1] != "list" {
		t.Errorf("args = %v, want [channel list]", got)
	}

	want := map[string]string{
		"CORE_PEER_ADDRESS":           env.Address,
		"CORE_PEER_LOCALMSPID":        env.LocalMSPID,
		"CORE_PEER_TLS_ROOTCERT_FILE": env.TLSCertFile,
		"CORE_PEER_MSPCONFIGPATH":     env.MSPConfigPath,
		"CORE_PEER_TLS_ENABLED":       "true",
	}
	got := envMap(cmd.Env)
	for key, wantValue := range want {
		if got[key] != wantValue {
			t.Errorf("%s = %q, want %q", key, got[key], wantValue)
		}
	}
}

func TestDockerExecRunsPeerInsideTheCLIContainer(t *testing.T) {
	cmd := DockerExec("channel", "list")

	if got, want := filepath.Base(cmd.Path), "docker"; got != want {
		t.Fatalf("command = %q, want %q", got, want)
	}
	want := []string{"docker", "exec", "cli", "peer", "channel", "list"}
	if len(cmd.Args) != len(want) {
		t.Fatalf("args = %v, want %v", cmd.Args, want)
	}
	for i := range want {
		if cmd.Args[i] != want[i] {
			t.Fatalf("args = %v, want %v", cmd.Args, want)
		}
	}
}

// DockerExec appends to a fresh slice each call; a shared backing array would
// let one invocation's arguments leak into the next.
func TestDockerExecDoesNotShareStateBetweenCalls(t *testing.T) {
	first := DockerExec("channel", "list")
	second := DockerExec("chaincode", "query")

	if first.Args[4] != "channel" || first.Args[5] != "list" {
		t.Fatalf("first call args were mutated: %v", first.Args)
	}
	if second.Args[4] != "chaincode" || second.Args[5] != "query" {
		t.Fatalf("second call args = %v, want chaincode query", second.Args)
	}
}

func TestOrdererDefaults(t *testing.T) {
	if got, want := DefaultOrderer(), "orderer.nanayam.com:7050"; got != want {
		t.Errorf("DefaultOrderer() = %q, want %q", got, want)
	}

	project := t.TempDir()
	t.Chdir(project)

	want := filepath.Join(project, "crypto-config", "ordererOrganizations", "nanayam.com",
		"orderers", "orderer.nanayam.com", "msp", "tlscacerts", "tlsca.nanayam.com-cert.pem")
	if got := OrdererTLSPath(); got != want {
		t.Errorf("OrdererTLSPath() = %q, want %q", got, want)
	}
}

func envMap(entries []string) map[string]string {
	out := make(map[string]string, len(entries))
	for _, entry := range entries {
		key, value, found := strings.Cut(entry, "=")
		if found {
			out[key] = value
		}
	}
	return out
}
