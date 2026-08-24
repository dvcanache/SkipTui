package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseOVPN(t *testing.T) {
	tmpDir := t.TempDir()
	ovpnFile := filepath.Join(tmpDir, "test.ovpn")

	content := `
client
dev tun
proto udp
remote vpn.example.com 1194
resolv-retry infinite
nobind
persist-key
persist-tun
auth-user-pass credentials.txt
dhcp-option DNS 1.1.1.1
dhcp-option DNS 8.8.8.8
`
	if err := os.WriteFile(ovpnFile, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write test ovpn file: %v", err)
	}

	profile, err := ParseOVPN(ovpnFile)
	if err != nil {
		t.Fatalf("ParseOVPN returned error: %v", err)
	}

	if profile.Protocol != ProtocolOpenVPN {
		t.Errorf("expected ProtocolOpenVPN, got %s", profile.Protocol)
	}

	if profile.OpenVPN.RemoteHost != "vpn.example.com" {
		t.Errorf("expected remote host vpn.example.com, got %s", profile.OpenVPN.RemoteHost)
	}

	if profile.OpenVPN.RemotePort != "1194" {
		t.Errorf("expected remote port 1194, got %s", profile.OpenVPN.RemotePort)
	}

	if profile.DNS != "1.1.1.1, 8.8.8.8" {
		t.Errorf("expected DNS '1.1.1.1, 8.8.8.8', got '%s'", profile.DNS)
	}
}

func TestParseWireGuard(t *testing.T) {
	tmpDir := t.TempDir()
	wgFile := filepath.Join(tmpDir, "wg0.conf")

	content := `
[Interface]
PrivateKey = a2V5X3Rlc3RfZGF0YQ==
Address = 10.14.0.2/32
DNS = 1.1.1.1

[Peer]
PublicKey = c2VydmVyX3B1YmxpY19rZXk=
Endpoint = 185.220.101.5:51820
AllowedIPs = 0.0.0.0/0
`
	if err := os.WriteFile(wgFile, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write test wg conf: %v", err)
	}

	profile, err := ParseWireGuard(wgFile)
	if err != nil {
		t.Fatalf("ParseWireGuard returned error: %v", err)
	}

	if profile.Protocol != ProtocolWireGuard {
		t.Errorf("expected ProtocolWireGuard, got %s", profile.Protocol)
	}

	if profile.WireGuard.Address != "10.14.0.2/32" {
		t.Errorf("expected address 10.14.0.2/32, got %s", profile.WireGuard.Address)
	}

	if profile.Endpoint != "185.220.101.5:51820" {
		t.Errorf("expected endpoint 185.220.101.5:51820, got %s", profile.Endpoint)
	}
}

func TestConfigProfileOperations(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	_ = InitDirs()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Add profile
	p := &Profile{
		ID:       "test-id",
		Name:     "Test-Profile",
		Protocol: ProtocolSOCKS5,
		Endpoint: "127.0.0.1:1080",
	}

	if err := cfg.AddProfile(p); err != nil {
		t.Fatalf("failed to add profile: %v", err)
	}

	found := cfg.GetProfile("Test-Profile")
	if found == nil || found.Endpoint != "127.0.0.1:1080" {
		t.Errorf("expected profile found with endpoint 127.0.0.1:1080")
	}

	// Delete profile
	if err := cfg.DeleteProfile("test-id"); err != nil {
		t.Fatalf("failed to delete profile: %v", err)
	}

	if cfg.GetProfile("test-id") != nil {
		t.Errorf("expected profile to be deleted")
	}
}
