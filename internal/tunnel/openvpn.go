package tunnel

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"skiptui/internal/config"
	"skiptui/internal/isolation"
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
	if profile.OpenVPN == nil {
		profile.OpenVPN = &config.OpenVPNConfig{}
	}

	// Auto-discover or fallback if config_path is missing or invalid
	if profile.OpenVPN.ConfigPath == "" || !config.FileExists(profile.OpenVPN.ConfigPath) {
		discovered := config.FindMatchingOVPN(profile.Name, profile.ID)
		if discovered != "" {
			profile.OpenVPN.ConfigPath = discovered
		} else {
			// Generate fallback .ovpn file
			genPath := filepath.Join(config.GetProfilesDir(), profile.ID+".ovpn")
			profile.OpenVPN.ConfigPath = genPath
		}
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
		if o.profile.OpenVPN.AuthUserPass != "" && config.FileExists(o.profile.OpenVPN.AuthUserPass) {
			args = append(args, "--auth-user-pass", o.profile.OpenVPN.AuthUserPass)
		}
		o.cmd = exec.CommandContext(ctx, "ip", args...)
	} else {
		args = append(args, "--config", o.profile.OpenVPN.ConfigPath, "--dev", "tun0")
		if o.profile.OpenVPN.AuthUserPass != "" && config.FileExists(o.profile.OpenVPN.AuthUserPass) {
			args = append(args, "--auth-user-pass", o.profile.OpenVPN.AuthUserPass)
		}
		o.cmd = exec.CommandContext(ctx, "openvpn", args...)
	}

	o.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := o.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start openvpn process: %w", err)
	}

	o.running = true

	go func() {
		_ = o.cmd.Wait()
		o.running = false
	}()

	return nil
}

func (o *OpenVPNTunnel) Stop() error {
	o.running = false
	if o.cmd != nil && o.cmd.Process != nil {
		_ = syscall.Kill(-o.cmd.Process.Pid, syscall.SIGTERM)
		time.Sleep(200 * time.Millisecond)
		_ = syscall.Kill(-o.cmd.Process.Pid, syscall.SIGKILL)
	}
	return nil
}

func (o *OpenVPNTunnel) GetMetrics() (uint64, uint64) {
	return atomic.LoadUint64(&o.bytesRX), atomic.LoadUint64(&o.bytesTX)
}
