package netprobe

import (
	"context"
	"net"
	"skiptui/internal/config"
	"sync"
	"time"
)

// ProbeResult contains the outcome of a profile latency check.
type ProbeResult struct {
	ProfileID string
	LatencyMs int64
	Error     error
}

// TestProfileLatency tests the TCP connect latency to a profile's endpoint.
func TestProfileLatency(ctx context.Context, p *config.Profile, timeout time.Duration) ProbeResult {
	if p.Endpoint == "" {
		return ProbeResult{
			ProfileID: p.ID,
			LatencyMs: -1,
			Error:     nil,
		}
	}

	start := time.Now()
	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "tcp", p.Endpoint)
	if err != nil {
		return ProbeResult{
			ProfileID: p.ID,
			LatencyMs: -1,
			Error:     err,
		}
	}
	defer conn.Close()

	latency := time.Since(start).Milliseconds()
	p.LatencyMs = latency
	p.LastTested = time.Now()

	return ProbeResult{
		ProfileID: p.ID,
		LatencyMs: latency,
		Error:     nil,
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
