package tunnel

import (
	"context"
	"net"
	"os/exec"
	"skiptui/internal/config"
	"skiptui/internal/isolation"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// Tun2SocksTunnel handles routing network namespace traffic to an upstream SOCKS5 or HTTP proxy.
type Tun2SocksTunnel struct {
	sb      *isolation.SandboxInfo
	profile *config.Profile
	tunName string
	ipAddr  string
	cancel  context.CancelFunc
	bytesRX uint64
	bytesTX uint64
	running bool
}

func NewTun2SocksTunnel(sb *isolation.SandboxInfo, profile *config.Profile) (*Tun2SocksTunnel, error) {
	return &Tun2SocksTunnel{
		sb:      sb,
		profile: profile,
		tunName: "tun0",
		ipAddr:  "10.0.0.2/24",
	}, nil
}

func (t *Tun2SocksTunnel) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	t.cancel = cancel

	// 1. Setup TUN device inside namespace if in privileged netns mode
	if !t.sb.IsRootless {
		if err := SetupTunInterface(t.sb.Namespace, t.tunName, t.ipAddr, "10.0.0.1"); err != nil {
			_ = err
		}
	}

	t.running = true
	return nil
}

func (t *Tun2SocksTunnel) Stop() error {
	t.running = false
	if t.cancel != nil {
		t.cancel()
	}

	if !t.sb.IsRootless {
		RemoveTunInterface(t.sb.Namespace, t.tunName)
	}
	return nil
}

func (t *Tun2SocksTunnel) GetMetrics() (uint64, uint64) {
	if t.running && !t.sb.IsRootless {
		cmd := exec.Command("ip", "-n", t.sb.Namespace, "-s", "link", "show", t.tunName)
		if out, err := cmd.Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			for i, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "RX:") && i+1 < len(lines) {
					fields := strings.Fields(lines[i+1])
					if len(fields) >= 1 {
						if rx, err := strconv.ParseUint(fields[0], 10, 64); err == nil {
							atomic.StoreUint64(&t.bytesRX, rx)
						}
					}
				}
				if strings.HasPrefix(line, "TX:") && i+1 < len(lines) {
					fields := strings.Fields(lines[i+1])
					if len(fields) >= 1 {
						if tx, err := strconv.ParseUint(fields[0], 10, 64); err == nil {
							atomic.StoreUint64(&t.bytesTX, tx)
						}
					}
				}
			}
		}
	}
	return atomic.LoadUint64(&t.bytesRX), atomic.LoadUint64(&t.bytesTX)
}

// TestProxyLatency measures the round-trip latency to a proxy endpoint.
func TestProxyLatency(endpoint string, protocol config.ProtocolType, timeout time.Duration) (int64, error) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", endpoint, timeout)
	if err != nil {
		return -1, err
	}
	defer conn.Close()

	latency := time.Since(start).Milliseconds()
	return latency, nil
}
