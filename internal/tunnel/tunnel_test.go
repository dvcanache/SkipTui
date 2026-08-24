package tunnel

import (
	"context"
	"skiptui/internal/config"
	"skiptui/internal/isolation"
	"testing"
)

func TestOpenVPNRejectsHostExecution(t *testing.T) {
	sbRootless := &isolation.SandboxInfo{
		ID:         "test-sb",
		Namespace:  "rootless-test-sb",
		IsRootless: true,
	}

	profile := &config.Profile{
		ID:       "ovpn-1",
		Name:     "Test-OVPN",
		Protocol: config.ProtocolOpenVPN,
		OpenVPN:  &config.OpenVPNConfig{},
	}

	tun, err := NewOpenVPNTunnel(sbRootless, profile)
	if err != nil {
		t.Fatalf("unexpected error creating OpenVPNTunnel: %v", err)
	}

	err = tun.Start(context.Background())
	if err == nil {
		t.Fatalf("expected OpenVPNTunnel to reject execution without netns, but got nil")
	}
}

func TestWireGuardRejectsHostExecution(t *testing.T) {
	sbRootless := &isolation.SandboxInfo{
		ID:         "test-sb",
		Namespace:  "rootless-test-sb",
		IsRootless: true,
	}

	profile := &config.Profile{
		ID:        "wg-1",
		Name:      "Test-WG",
		Protocol:  config.ProtocolWireGuard,
		WireGuard: &config.WGConfig{},
	}

	tun, err := NewWireGuardTunnel(sbRootless, profile)
	if err != nil {
		t.Fatalf("unexpected error creating WireGuardTunnel: %v", err)
	}

	err = tun.Start(context.Background())
	if err == nil {
		t.Fatalf("expected WireGuardTunnel to reject execution without netns, but got nil")
	}
}
