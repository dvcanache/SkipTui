package tunnel

import (
	"context"
	"fmt"
	"os/exec"
	"skiptui/internal/config"
	"skiptui/internal/isolation"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

// OpenVPNTunnel manages an isolated OpenVPN client instance within the network namespace.
type OpenVPNTunnel struct {
	sb      *isolation.SandboxInfo
	profile *config.Profile
	cmd     *exec.Cmd
	bytesRX uint64
	bytesTX uint64
	running bool
}

func NewOpenVPNTunnel(sb *isolation.SandboxInfo, profile *config.Profile) (*OpenVPNTunnel, error) {
	if profile.OpenVPN == nil || profile.OpenVPN.ConfigPath == "" {
		return nil, fmt.Errorf("openvpn configuration path is missing for profile '%s'", profile.Name)
	}

	return &OpenVPNTunnel{
		sb:      sb,
		profile: profile,
	}, nil
}

func (o *OpenVPNTunnel) Start(ctx context.Context) error {
	var args []string
	if !o.sb.IsRootless {
		args = append(args, "netns", "exec", o.sb.Namespace, "openvpn", "--config", o.profile.OpenVPN.ConfigPath, "--dev", "tun0", "--redirect-gateway", "def1")
		if o.profile.OpenVPN.AuthUserPass != "" {
			args = append(args, "--auth-user-pass", o.profile.OpenVPN.AuthUserPass)
		}
		o.cmd = exec.CommandContext(ctx, "ip", args...)
	} else {
		args = append(args, "--config", o.profile.OpenVPN.ConfigPath, "--dev", "tun0")
		o.cmd = exec.CommandContext(ctx, "openvpn", args...)
	}

	o.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Capture output in background
	if err := o.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start openvpn process: %w", err)
	}

	o.running = true

	// Monitor OpenVPN lifecycle
	go func() {
		_ = o.cmd.Wait()
		o.running = false
	}()

	return nil
}

func (o *OpenVPNTunnel) Stop() error {
	o.running = false
	if o.cmd != nil && o.cmd.Process != nil {
		// Send SIGTERM to process group
		_ = syscall.Kill(-o.cmd.Process.Pid, syscall.SIGTERM)
		time.Sleep(200 * time.Millisecond)
		_ = syscall.Kill(-o.cmd.Process.Pid, syscall.SIGKILL)
	}
	return nil
}

func (o *OpenVPNTunnel) GetMetrics() (uint64, uint64) {
	if o.running && !o.sb.IsRootless {
		// Read RX/TX from `ip -n <ns> -s link show tun0`
		cmd := exec.Command("ip", "-n", o.sb.Namespace, "-s", "link", "show", "tun0")
		if out, err := cmd.Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			for i, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "RX:") && i+1 < len(lines) {
					fields := strings.Fields(lines[i+1])
					if len(fields) >= 1 {
						if rx, err := strconv.ParseUint(fields[0], 10, 64); err == nil {
							atomic.StoreUint64(&o.bytesRX, rx)
						}
					}
				}
				if strings.HasPrefix(line, "TX:") && i+1 < len(lines) {
					fields := strings.Fields(lines[i+1])
					if len(fields) >= 1 {
						if tx, err := strconv.ParseUint(fields[0], 10, 64); err == nil {
							atomic.StoreUint64(&o.bytesTX, tx)
						}
					}
				}
			}
		}
	}
	return atomic.LoadUint64(&o.bytesRX), atomic.LoadUint64(&o.bytesTX)
}

