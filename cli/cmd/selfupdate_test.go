package cmd

import "testing"

func TestReleaseAssetName(t *testing.T) {
	tests := []struct {
		name    string
		version string
		goos    string
		goarch  string
		want    string
	}{
		{name: "darwin tarball", version: "v1.2.3", goos: "darwin", goarch: "arm64", want: "nanayam_v1.2.3_darwin_arm64.tar.gz"},
		{name: "linux tarball", version: "v1.2.3", goos: "linux", goarch: "amd64", want: "nanayam_v1.2.3_linux_amd64.tar.gz"},
		{name: "windows zip", version: "v1.2.3", goos: "windows", goarch: "amd64", want: "nanayam_v1.2.3_windows_amd64.zip"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := releaseAssetName(tt.version, tt.goos, tt.goarch); got != tt.want {
				t.Fatalf("releaseAssetName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeSemver(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "already prefixed", input: "v1.2.3", want: "v1.2.3"},
		{name: "missing prefix", input: "1.2.3", want: "v1.2.3"},
		{name: "prerelease", input: "1.2.3-rc1", want: "v1.2.3-rc1"},
		{name: "invalid", input: "dev", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeSemver(tt.input); got != tt.want {
				t.Fatalf("normalizeSemver() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsUpgradeAvailable(t *testing.T) {
	tests := []struct {
		name    string
		current string
		target  string
		want    bool
	}{
		{name: "newer patch", current: "v1.2.2", target: "v1.2.3", want: true},
		{name: "same version", current: "v1.2.3", target: "v1.2.3", want: false},
		{name: "older target", current: "v1.3.0", target: "v1.2.9", want: false},
		{name: "dev build treated as different", current: "dev", target: "v1.0.0", want: true},
		{name: "prerelease to release", current: "v1.2.3-rc1", target: "v1.2.3", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isUpgradeAvailable(tt.current, tt.target); got != tt.want {
				t.Fatalf("isUpgradeAvailable(%q, %q) = %v, want %v", tt.current, tt.target, got, tt.want)
			}
		})
	}
}
