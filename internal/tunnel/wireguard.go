package tunnel

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"skiptui/internal/config"
	"skiptui/internal/isolation"
	"strconv"
	"strings"
	"sync/atomic"
)

// WireGuardTunnel manages a WireGuard interface moved into a network namespace.
type WireGuardTunnel struct {
	sb      *isolation.SandboxInfo
	profile *config.Profile
	iface   string
	bytesRX uint64
	bytesTX uint64
	running bool
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
		w.running = true
		return nil
	}

	// 1. Create WireGuard link in host: `ip link add <iface> type wireguard`
	cmd := exec.Command("ip", "link", "add", w.iface, "type", "wireguard")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create wireguard link '%s': %s (%w)", w.iface, string(out), err)
	}

	// 2. Move interface into the target namespace: `ip link set <iface> netns <namespace>`
	cmd = exec.Command("ip", "link", "set", w.iface, "netns", w.sb.Namespace)
	if out, err := cmd.CombinedOutput(); err != nil {
		_ = exec.Command("ip", "link", "del", w.iface).Run()
		return fmt.Errorf("failed to move wireguard link to netns '%s': %s (%w)", w.sb.Namespace, string(out), err)
	}

	// 3. Configure WireGuard keys and peer within the namespace using `wg set`
	if w.profile.WireGuard.PrivateKey != "" {
		keyFile := filepath.Join(config.GetRuntimeDir(), fmt.Sprintf("%s_priv.key", w.iface))
		if err := os.WriteFile(keyFile, []byte(w.profile.WireGuard.PrivateKey+"\n"), 0600); err == nil {
			defer os.Remove(keyFile)

			wgArgs := []string{"netns", "exec", w.sb.Namespace, "wg", "set", w.iface, "private-key", keyFile}

			if w.profile.WireGuard.PublicKey != "" {
				wgArgs = append(wgArgs, "peer", w.profile.WireGuard.PublicKey)

				endpoint := w.profile.Endpoint
				if endpoint != "" {
					wgArgs = append(wgArgs, "endpoint", endpoint)
				}

				allowedIPs := "0.0.0.0/0"
				if len(w.profile.WireGuard.AllowedIPs) > 0 {
					allowedIPs = strings.Join(w.profile.WireGuard.AllowedIPs, ",")
				}
				wgArgs = append(wgArgs, "allowed-ips", allowedIPs)

				if w.profile.WireGuard.PresharedKey != "" {
					pskFile := filepath.Join(config.GetRuntimeDir(), fmt.Sprintf("%s_psk.key", w.iface))
					if pErr := os.WriteFile(pskFile, []byte(w.profile.WireGuard.PresharedKey+"\n"), 0600); pErr == nil {
						defer os.Remove(pskFile)
						wgArgs = append(wgArgs, "preshared-key", pskFile)
					}
				}
			}

			cmd = exec.CommandContext(ctx, "ip", wgArgs...)
			_ = cmd.Run()
		}
	}

	// 4. Configure IP address inside namespace
	addr := "10.14.0.2/32"
	if w.profile.WireGuard.Address != "" {
		addr = w.profile.WireGuard.Address
	}
	cmd = exec.Command("ip", "-n", w.sb.Namespace, "addr", "add", addr, "dev", w.iface)
	_ = cmd.Run()

	// 5. Set interface UP
	cmd = exec.Command("ip", "-n", w.sb.Namespace, "link", "set", w.iface, "up")
	_ = cmd.Run()

	// 6. Add default route
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
	if w.running && !w.sb.IsRootless {
		// Read live transfer metrics from `wg show <iface> transfer`
		cmd := exec.Command("ip", "netns", "exec", w.sb.Namespace, "wg", "show", w.iface, "transfer")
		if out, err := cmd.Output(); err == nil {
			fields := strings.Fields(string(out))
			if len(fields) >= 2 {
				if rx, err := strconv.ParseUint(fields[0], 10, 64); err == nil {
					atomic.StoreUint64(&w.bytesRX, rx)
				}
				if tx, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
					atomic.StoreUint64(&w.bytesTX, tx)
				}
			}
		}
	}
	return atomic.LoadUint64(&w.bytesRX), atomic.LoadUint64(&w.bytesTX)
}

