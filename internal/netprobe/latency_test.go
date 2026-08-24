package netprobe

import (
	"context"
	"net"
	"skiptui/internal/config"
	"testing"
	"time"
)

func TestTestProfileLatency(t *testing.T) {
	// Start a dummy local TCP listener
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen on dummy port: %v", err)
	}
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	profile := &config.Profile{
		ID:       "test-local",
		Name:     "Test-Local",
		Endpoint: listener.Addr().String(),
	}

	res := TestProfileLatency(context.Background(), profile, 1*time.Second)
	if res.Error != nil {
		t.Fatalf("expected successful latency probe, got error: %v", res.Error)
	}

	if res.LatencyMs < 0 {
		t.Errorf("expected positive latency, got %d", res.LatencyMs)
	}
}
