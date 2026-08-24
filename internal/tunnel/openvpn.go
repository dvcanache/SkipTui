package tunnel

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"skiptui/internal/config"
	"skiptui/internal/isolation"
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
	logDir := filepath.Join(config.GetRuntimeDir(), "logs")
	_ = os.MkdirAll(logDir, 0700)
	logFile := filepath.Join(logDir, fmt.Sprintf("openvpn-%s.log", o.sb.ID))

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

	// Pipe output to log file and buffer
	var errBuf bytes.Buffer
	if logHandle, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600); err == nil {
		o.cmd.Stdout = logHandle
		o.cmd.Stderr = ioTeeWriter(logHandle, &errBuf)
	} else {
		o.cmd.Stderr = &errBuf
	}

	if err := o.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start openvpn process: %w", err)
	}

	o.running = true

	// Check for immediate early exit (e.g. permission error on TUN creation)
	exitChan := make(chan error, 1)
	go func() {
		err := o.cmd.Wait()
		o.running = false
		exitChan <- err
	}()

	select {
	case err := <-exitChan:
		errText := strings.TrimSpace(errBuf.String())
		if strings.Contains(errText, "Operation not permitted") || strings.Contains(errText, "TUNSETIFF") {
			return fmt.Errorf("OpenVPN TUN permission denied. Run 'sudo make setcap' or run with 'sudo ./bin/skiptui'")
		}
		if strings.Contains(errText, "AUTH_FAILED") {
			return fmt.Errorf("OpenVPN authentication failed: check username/password")
		}
		if err != nil {
			return fmt.Errorf("OpenVPN failed to connect: %v (see %s)", err, logFile)
		}
	case <-time.After(1500 * time.Millisecond):
		// Still running after 1.5s -> connection is negotiating/established
	}

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

type teeWriter struct {
	w1 *os.File
	w2 *bytes.Buffer
}

func (t *teeWriter) Write(p []byte) (n int, err error) {
	if t.w1 != nil {
		_, _ = t.w1.Write(p)
	}
	if t.w2 != nil {
		_, _ = t.w2.Write(p)
	}
	return len(p), nil
}

func ioTeeWriter(w1 *os.File, w2 *bytes.Buffer) *teeWriter {
	return &teeWriter{w1: w1, w2: w2}
}
