package daemon

import (
	"os"
	"skiptui/internal/config"
	"testing"
	"time"
)

func TestStatePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", tmpDir)

	sess1 := &config.SessionInfo{
		ID:          "sb-12345678",
		Command:     "curl",
		Args:        []string{"https://api.ipify.org"},
		ProfileID:   "p-1",
		ProfileName: "Test-Profile",
		Protocol:    "socks5",
		PID:         os.Getpid(),
		Status:      "running",
		Namespace:   "skiptui-12345678",
		StartTime:   time.Now(),
	}

	// 1. Save Session
	if err := SaveSessionState([]*config.SessionInfo{sess1}); err != nil {
		t.Fatalf("SaveSessionState failed: %v", err)
	}

	// 2. Load Session
	loaded, err := LoadSessionState()
	if err != nil {
		t.Fatalf("LoadSessionState failed: %v", err)
	}
	if len(loaded) != 1 || loaded[0].ID != sess1.ID {
		t.Fatalf("expected 1 session loaded with ID %s, got %v", sess1.ID, loaded)
	}

	// 3. Get by ID
	found, err := GetSessionByID("sb-12345678")
	if err != nil || found == nil {
		t.Fatalf("GetSessionByID failed: %v", err)
	}

	// 4. Update Session
	sess1.Status = "stopped"
	if err := UpdateSession(sess1); err != nil {
		t.Fatalf("UpdateSession failed: %v", err)
	}
	updated, _ := GetSessionByID("sb-12345678")
	if updated.Status != "stopped" {
		t.Errorf("expected status 'stopped', got '%s'", updated.Status)
	}

	// 5. PID Liveness
	if !IsPIDAlive(os.Getpid()) {
		t.Errorf("expected current PID to be alive")
	}
	if IsPIDAlive(9999999) {
		t.Errorf("expected invalid PID 9999999 to not be alive")
	}

	// 6. Remove Session
	if err := RemoveSession("sb-12345678"); err != nil {
		t.Fatalf("RemoveSession failed: %v", err)
	}
	afterRemove, _ := LoadSessionState()
	if len(afterRemove) != 0 {
		t.Errorf("expected 0 sessions after removal, got %d", len(afterRemove))
	}
}
