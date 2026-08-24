package netprobe

import (
	"context"
	"net"
	"skiptui/internal/config"
	"strings"
	"sync"
	"time"
)

// ProbeResult contains the outcome of a profile latency check.
type ProbeResult struct {
	ProfileID string
	LatencyMs int64
	Error     error
}

// TestProfileLatency tests the network latency to a profile's endpoint.
func TestProfileLatency(ctx context.Context, p *config.Profile, timeout time.Duration) ProbeResult {
	if p.Endpoint == "" {
		return ProbeResult{
			ProfileID: p.ID,
			LatencyMs: -1,
			Error:     nil,
		}
	}

	endpoint := normalizeEndpoint(p.Endpoint, p.Protocol)
	network := "tcp"

	if p.Protocol == config.ProtocolWireGuard {
		network = "udp"
	} else if p.Protocol == config.ProtocolOpenVPN && p.OpenVPN != nil && p.OpenVPN.Proto == "udp" {
		network = "udp"
	}

	start := time.Now()
	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, network, endpoint)
	if err != nil {
		p.LatencyMs = -1
		p.LastTested = time.Now()
		return ProbeResult{
			ProfileID: p.ID,
			LatencyMs: -1,
			Error:     err,
		}
	}
	defer conn.Close()

	latency := time.Since(start).Milliseconds()
	if latency == 0 {
		latency = 1
	}
	p.LatencyMs = latency
	p.LastTested = time.Now()

	return ProbeResult{
		ProfileID: p.ID,
		LatencyMs: latency,
		Error:     nil,
	}
}

func normalizeEndpoint(endpoint string, proto config.ProtocolType) string {
	if strings.Contains(endpoint, ":") {
		return endpoint
	}
	switch proto {
	case config.ProtocolSOCKS5:
		return endpoint + ":1080"
	case config.ProtocolHTTP:
		return endpoint + ":8080"
	case config.ProtocolShadowsocks:
		return endpoint + ":8388"
	case config.ProtocolWireGuard:
		return endpoint + ":51820"
	case config.ProtocolOpenVPN:
		return endpoint + ":1194"
	default:
		return endpoint + ":1080"
	}
}

// TestAllProfiles concurrently tests all profiles in the list.
func TestAllProfiles(ctx context.Context, profiles []*config.Profile, timeout time.Duration) []ProbeResult {
	results := make([]ProbeResult, len(profiles))
	var wg sync.WaitGroup

	for i, p := range profiles {
		wg.Add(1)
		go func(idx int, prof *config.Profile) {
			defer wg.Done()
			results[idx] = TestProfileLatency(ctx, prof, timeout)
		}(i, p)
	}

	wg.Wait()
	return results
}
