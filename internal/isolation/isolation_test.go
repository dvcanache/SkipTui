package isolation

import (
	"context"
	"skiptui/internal/config"
	"testing"
)

func TestSanitizeProcessEnv(t *testing.T) {
	initial := []string{
		"PATH=/usr/bin:/bin",
		"USER=testuser",
		"DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/1000/bus",
		"DBUS_SYSTEM_BUS_ADDRESS=unix:path=/var/run/dbus/system_bus_socket",
		"HOME=/home/testuser",
	}

	cleaned := SanitizeProcessEnv(initial)

	for _, env := range cleaned {
		if env == "DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/1000/bus" {
			t.Errorf("DBUS_SESSION_BUS_ADDRESS was not sanitized")
		}
		if env == "DBUS_SYSTEM_BUS_ADDRESS=unix:path=/var/run/dbus/system_bus_socket" {
			t.Errorf("DBUS_SYSTEM_BUS_ADDRESS was not sanitized")
		}
	}
}

func TestRootlessEngine(t *testing.T) {
	eng := NewRootlessEngine()
	if eng.Name() != "rootless" {
		t.Errorf("expected engine name 'rootless', got '%s'", eng.Name())
	}

	sb := &SandboxInfo{
		ID:         "test-sb",
		Namespace:  "rootless-test-sb",
		DNS:        "1.1.1.1",
		IsRootless: true,
	}

	cmd, err := eng.BuildCommand(context.Background(), sb, "echo", "hello")
	if err != nil {
		t.Fatalf("failed to build command: %v", err)
	}

	if len(cmd.Args) < 4 || cmd.Args[0] != "unshare" {
		t.Errorf("expected unshare command, got %v", cmd.Args)
	}

	// Check if host supports user namespaces
	if err := eng.CheckCapabilities(); err != nil {
		t.Logf("Notice: host kernel restricts unprivileged user namespaces (%v)", err)
		return
	}

	createdSb, err := eng.CreateSandbox(context.Background(), "test-sb", &config.Profile{DNS: "1.1.1.1"})
	if err != nil {
		t.Fatalf("failed to create rootless sandbox: %v", err)
	}

	if !createdSb.IsRootless {
		t.Errorf("expected createdSb.IsRootless to be true")
	}
}
