package ca

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// newRecorder writes a stub fabric-ca-client that appends its arguments, one
// invocation per line, to a log file. Driving the real binary is not possible
// in a unit test, but the argument list is exactly what these methods build.
func newRecorder(t *testing.T) (binaryPath string, invocations func() [][]string) {
	t.Helper()

	if runtime.GOOS == "windows" {
		t.Skip("the recorder stub is a POSIX shell script")
	}

	dir := t.TempDir()
	logPath := filepath.Join(dir, "invocations.log")
	binaryPath = filepath.Join(dir, "fabric-ca-client")

	script := fmt.Sprintf("#!/bin/sh\nprintf '%%s\\n' \"$*\" >> %q\n", logPath)
	if err := os.WriteFile(binaryPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write recorder stub: %v", err)
	}

	return binaryPath, func() [][]string {
		content, err := os.ReadFile(logPath)
		if err != nil {
			return nil
		}
		var calls [][]string
		for _, line := range strings.Split(strings.TrimSpace(string(content)), "\n") {
			if line != "" {
				calls = append(calls, strings.Fields(line))
			}
		}
		return calls
	}
}

// hasFlag reports whether args contains flag immediately followed by value.
func hasFlag(args []string, flag, value string) bool {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == flag && args[i+1] == value {
			return true
		}
	}
	return false
}

func TestNewClient(t *testing.T) {
	c := NewClient("/opt/bin/fabric-ca-client", "/tmp/ca-home", "localhost:7054")

	if c.BinaryPath != "/opt/bin/fabric-ca-client" {
		t.Errorf("BinaryPath = %q", c.BinaryPath)
	}
	if c.HomeDir != "/tmp/ca-home" {
		t.Errorf("HomeDir = %q", c.HomeDir)
	}
	if c.ServerURL != "localhost:7054" {
		t.Errorf("ServerURL = %q", c.ServerURL)
	}
	// TLS is opt-in; an empty default keeps the --tls.certfiles flag off.
	if c.TLSConfig != "" {
		t.Errorf("TLSConfig = %q, want empty by default", c.TLSConfig)
	}
}

func TestEnrollBuildsCredentialURL(t *testing.T) {
	binary, calls := newRecorder(t)
	home := t.TempDir()
	c := NewClient(binary, home, "localhost:7054")

	if err := c.Enroll("alice", "alice-pw", "/msp/alice"); err != nil {
		t.Fatalf("Enroll() = %v", err)
	}

	got := calls()
	if len(got) != 1 {
		t.Fatalf("expected 1 invocation, got %d: %v", len(got), got)
	}
	args := got[0]

	if args[0] != "enroll" {
		t.Errorf("subcommand = %q, want enroll", args[0])
	}
	if !hasFlag(args, "-u", "alice:alice-pw@localhost:7054") {
		t.Errorf("enrollment URL missing or malformed: %v", args)
	}
	if !hasFlag(args, "--mspdir", "/msp/alice") {
		t.Errorf("--mspdir missing: %v", args)
	}
	if !hasFlag(args, "--home", home) {
		t.Errorf("--home missing: %v", args)
	}
}

func TestEnrollPassesAttributesAndTLS(t *testing.T) {
	binary, calls := newRecorder(t)
	c := NewClient(binary, t.TempDir(), "localhost:7054")
	c.TLSConfig = "/crypto/tls-ca.pem"

	if err := c.Enroll("alice", "pw", "/msp/alice", "role=admin:ecert", "dept=acb:ecert"); err != nil {
		t.Fatalf("Enroll() = %v", err)
	}

	args := calls()[0]
	if !hasFlag(args, "--tls.certfiles", "/crypto/tls-ca.pem") {
		t.Errorf("--tls.certfiles missing when TLSConfig is set: %v", args)
	}
	if !hasFlag(args, "--enrollment.attrs", "role=admin:ecert") {
		t.Errorf("first attribute missing: %v", args)
	}
	if !hasFlag(args, "--enrollment.attrs", "dept=acb:ecert") {
		t.Errorf("second attribute missing: %v", args)
	}
}

func TestEnrollOmitsTLSFlagWhenUnconfigured(t *testing.T) {
	binary, calls := newRecorder(t)
	c := NewClient(binary, t.TempDir(), "localhost:7054")

	if err := c.Enroll("alice", "pw", "/msp/alice"); err != nil {
		t.Fatalf("Enroll() = %v", err)
	}

	for _, arg := range calls()[0] {
		if arg == "--tls.certfiles" {
			t.Fatalf("--tls.certfiles passed with no TLSConfig: %v", calls()[0])
		}
	}
}

// Register must enroll the registrar first, otherwise the CA rejects the call
// for want of an admin identity.
func TestRegisterEnrollsTheRegistrarOnFirstUse(t *testing.T) {
	binary, calls := newRecorder(t)
	home := t.TempDir()
	c := NewClient(binary, home, "localhost:7054")

	if err := c.Register("admin", "adminpw", "alice", "alicepw", "client", "org1.department1"); err != nil {
		t.Fatalf("Register() = %v", err)
	}

	got := calls()
	if len(got) != 2 {
		t.Fatalf("expected an enroll then a register, got %d invocations: %v", len(got), got)
	}
	if got[0][0] != "enroll" {
		t.Errorf("first invocation = %q, want enroll", got[0][0])
	}

	register := got[1]
	if register[0] != "register" {
		t.Fatalf("second invocation = %q, want register", register[0])
	}
	for flag, value := range map[string]string{
		"--id.name":        "alice",
		"--id.secret":      "alicepw",
		"--id.type":        "client",
		"--id.affiliation": "org1.department1",
		"--mspdir":         filepath.Join(home, "admin-msp"),
	} {
		if !hasFlag(register, flag, value) {
			t.Errorf("%s %s missing from register: %v", flag, value, register)
		}
	}
}

func TestRegisterSkipsEnrollWhenRegistrarMSPExists(t *testing.T) {
	binary, calls := newRecorder(t)
	home := t.TempDir()

	// A pre-existing registrar MSP means the admin already enrolled.
	if err := os.MkdirAll(filepath.Join(home, "admin-msp"), 0o755); err != nil {
		t.Fatalf("seed registrar msp: %v", err)
	}

	c := NewClient(binary, home, "localhost:7054")
	if err := c.Register("admin", "adminpw", "bob", "bobpw", "client", "org1"); err != nil {
		t.Fatalf("Register() = %v", err)
	}

	got := calls()
	if len(got) != 1 {
		t.Fatalf("expected only a register call, got %d: %v", len(got), got)
	}
	if got[0][0] != "register" {
		t.Errorf("invocation = %q, want register", got[0][0])
	}
}

func TestReenroll(t *testing.T) {
	binary, calls := newRecorder(t)
	home := t.TempDir()
	c := NewClient(binary, home, "localhost:7054")

	if err := c.Reenroll("/msp/alice"); err != nil {
		t.Fatalf("Reenroll() = %v", err)
	}

	args := calls()[0]
	if args[0] != "reenroll" {
		t.Errorf("subcommand = %q, want reenroll", args[0])
	}
	if !hasFlag(args, "--mspdir", "/msp/alice") || !hasFlag(args, "--home", home) {
		t.Errorf("reenroll args incomplete: %v", args)
	}
}

func TestGetCAInfo(t *testing.T) {
	binary, calls := newRecorder(t)
	c := NewClient(binary, t.TempDir(), "localhost:7054")

	if err := c.GetCAInfo(); err != nil {
		t.Fatalf("GetCAInfo() = %v", err)
	}

	args := calls()[0]
	if args[0] != "getcainfo" {
		t.Errorf("subcommand = %q, want getcainfo", args[0])
	}
	if !hasFlag(args, "-u", "localhost:7054") {
		t.Errorf("server URL missing: %v", args)
	}
}

func TestCommandFailuresPropagate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stub is a POSIX shell script")
	}

	dir := t.TempDir()
	binary := filepath.Join(dir, "fabric-ca-client")
	if err := os.WriteFile(binary, []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatalf("write failing stub: %v", err)
	}

	c := NewClient(binary, dir, "localhost:7054")
	if err := c.Enroll("alice", "pw", "/msp/alice"); err == nil {
		t.Fatal("Enroll() = nil when the CA client exits non-zero, want an error")
	}
	if err := c.GetCAInfo(); err == nil {
		t.Fatal("GetCAInfo() = nil when the CA client exits non-zero, want an error")
	}
}
