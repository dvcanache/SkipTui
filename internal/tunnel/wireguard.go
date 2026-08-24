package tunnel

import (
	"context"
	"fmt"
	"os/exec"
	"skiptui/internal/config"
	"skiptui/internal/isolation"
	"sync/atomic"
)

// WireGuardTunnel manages a WireGuard interface moved into a network namespace.
type WireGuardTunnel struct {
	sb       *isolation.SandboxInfo
	profile  *config.Profile
	iface    string
	bytesRX  uint64
	bytesTX  uint64
	running  bool
}

func NewWireGuardTunnel(sb *isolation.SandboxInfo, profile *config.Profile) (*WireGuardTunnel, error) {
	if profile.WireGuard == nil {
		return nil, fmt.Errorf("wireguard configuration missing for profile '%s'", profile.Name)
	}

	return &WireGuardTunnel{
		sb:      sb,
		profile: profile,
		iface:   "wg-" + sb.ID,
	}, nil
}

func (w *WireGuardTunnel) Start(ctx context.Context) error {
	if w.sb.IsRootless {
		// In rootless mode, WireGuard-go or user-space routing is used
		w.running = true
		return nil
	}

	// 1. Create WireGuard link in host: `ip link add <iface> type wireguard`
	cmd := exec.Command("ip", "link", "add", w.iface, "type", "wireguard")
	_ = cmd.Run()

	// 2. Move interface into the target namespace: `ip link set <iface> netns <namespace>`
	cmd = exec.Command("ip", "link", "set", w.iface, "netns", w.sb.Namespace)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to move wireguard link to netns '%s': %s (%w)", w.sb.Namespace, string(out), err)
	}

	// 3. Configure IP address inside namespace
	addr := "10.14.0.2/32"
	if w.profile.WireGuard.Address != "" {
		addr = w.profile.WireGuard.Address
	}
	cmd = exec.Command("ip", "-n", w.sb.Namespace, "addr", "add", addr, "dev", w.iface)
	_ = cmd.Run()

	// 4. Set interface UP
	cmd = exec.Command("ip", "-n", w.sb.Namespace, "link", "set", w.iface, "up")
	_ = cmd.Run()

	// 5. Add default route
	cmd = exec.Command("ip", "-n", w.sb.Namespace, "route", "add", "default", "dev", w.iface)
	_ = cmd.Run()

	w.running = true
	return nil
}

func (w *WireGuardTunnel) Stop() error {
	w.running = false
	if !w.sb.IsRootless {
		cmd := exec.Command("ip", "-n", w.sb.Namespace, "link", "del", w.iface)
		_ = cmd.Run()
	}
	return nil
}

func (w *WireGuardTunnel) GetMetrics() (uint64, uint64) {
	return atomic.LoadUint64(&w.bytesRX), atomic.LoadUint64(&w.bytesTX)
}
