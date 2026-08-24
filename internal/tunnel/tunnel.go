package tunnel

import (
	"context"
	"skiptui/internal/config"
	"skiptui/internal/isolation"
)

// TunnelInstance represents an active running tunnel worker (tun2socks, wireguard, openvpn).
type TunnelInstance interface {
	// Start begins routing traffic through the tunnel
	Start(ctx context.Context) error

	// Stop terminates the tunnel worker and cleans up virtual interfaces
	Stop() error

	// GetMetrics returns transferred bytes (RX, TX)
	GetMetrics() (rx uint64, tx uint64)
}

// Factory creates the appropriate TunnelInstance for a given profile.
func CreateTunnel(sb *isolation.SandboxInfo, profile *config.Profile) (TunnelInstance, error) {
	switch profile.Protocol {
	case config.ProtocolWireGuard:
		return NewWireGuardTunnel(sb, profile)
	case config.ProtocolOpenVPN:
		return NewOpenVPNTunnel(sb, profile)
	case config.ProtocolSOCKS5, config.ProtocolHTTP, config.ProtocolShadowsocks:
		fallthrough
	default:
		return NewTun2SocksTunnel(sb, profile)
	}
}
