package session

import (
	"context"
	"skiptui/internal/config"
	"testing"
	"time"
)

func TestSupervisorLifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	cfg := &config.Config{
		Settings: config.Settings{
			RootlessMode: true,
		},
		Profiles: []*config.Profile{
			{
				ID:         "p-test",
				Name:       "Test-Profile",
				Protocol:   config.ProtocolSOCKS5,
				Endpoint:   "127.0.0.1:1080",
				DNS:        "1.1.1.1",
				KillSwitch: true,
			},
		},
	}

	sup := NewSupervisor(cfg)

	// Launch session with echo
	info, err := sup.LaunchSession(context.Background(), cfg.Profiles[0], "echo", []string{"hello"}, false)
	if err != nil {
		t.Fatalf("LaunchSession failed: %v", err)
	}

	if info.ID == "" {
		t.Fatalf("expected non-empty session ID")
	}

	// Wait for process to complete
	time.Sleep(200 * time.Millisecond)

	sessions := sup.ListSessions()
	if len(sessions) == 0 {
		t.Fatalf("expected at least 1 session listed")
	}

	// Test clear stopped sessions
	sup.ClearStoppedSessions()
}
