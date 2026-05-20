package cmd

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestResolveChannelArtifactConfigBasic(t *testing.T) {
	cwd := "/Users/ajithberlin/Documents/GitHub/nanayam"
	configtxSource, profiles, err := resolveChannelArtifactConfig(cwd, filepath.Join(cwd, "config", "crypto-config.yaml"))
	if err != nil {
		t.Fatalf("resolveChannelArtifactConfig() error = %v", err)
	}
	if got, want := filepath.Base(configtxSource), "configtx.yaml"; got != want {
		t.Fatalf("resolveChannelArtifactConfig() source = %q, want %q", got, want)
	}
	wantProfiles := []channelArtifactProfile{{
		name:           "TwoOrgsOrdererGenesis",
		channelProfile: "TwoOrgsChannel",
		genesis:        "genesis.block",
		channel:        "mychannel",
		anchorOrgs:     []string{"Org1MSP", "Org2MSP"},
	}}
	if !reflect.DeepEqual(profiles, wantProfiles) {
		t.Fatalf("resolveChannelArtifactConfig() profiles = %#v, want %#v", profiles, wantProfiles)
	}
}

func TestResolveChannelArtifactConfigComplaint(t *testing.T) {
	cwd := "/Users/ajithberlin/Documents/GitHub/nanayam"
	configtxSource, profiles, err := resolveChannelArtifactConfig(cwd, filepath.Join(cwd, "config", "crypto-config-complaint.yaml"))
	if err != nil {
		t.Fatalf("resolveChannelArtifactConfig() error = %v", err)
	}
	if got, want := filepath.Base(configtxSource), "configtx-complaint.yaml"; got != want {
		t.Fatalf("resolveChannelArtifactConfig() source = %q, want %q", got, want)
	}
	wantProfiles := []channelArtifactProfile{{
		name:           "ComplaintOrdererGenesis",
		channelProfile: "ComplaintChannel",
		genesis:        "genesis.block",
		channel:        "complaint-channel",
		anchorOrgs:     []string{"ACBMSP", "DeptMSP", "OversightMSP", "JudiciaryMSP"},
	}}
	if !reflect.DeepEqual(profiles, wantProfiles) {
		t.Fatalf("resolveChannelArtifactConfig() profiles = %#v, want %#v", profiles, wantProfiles)
	}
}
