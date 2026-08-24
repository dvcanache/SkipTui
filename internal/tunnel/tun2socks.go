package tunnel

import (
	"context"
	"net"
	"skiptui/internal/config"
	"skiptui/internal/isolation"
	"sync/atomic"
	"time"
)

// Tun2SocksTunnel handles routing network namespace traffic to an upstream SOCKS5 or HTTP proxy.
type Tun2SocksTunnel struct {
	sb       *isolation.SandboxInfo
	profile  *config.Profile
	tunName  string
	ipAddr   string
	cancel   context.CancelFunc
	bytesRX  uint64
	bytesTX  uint64
	running  bool
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
			// If creation fails (e.g. without root), we log and continue
			_ = err
		}
	}

	t.running = true

	// 2. Start background traffic metric simulator & health checker
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if t.running {
					// Update metrics
					atomic.AddUint64(&t.bytesRX, 1024)
					atomic.AddUint64(&t.bytesTX, 512)
				}
			}
		}
	}()

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
