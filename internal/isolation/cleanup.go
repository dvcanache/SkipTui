package isolation

import (
	"os"
	"os/exec"
	"os/signal"
	"skiptui/internal/daemon"
	"strings"
	"syscall"
)

// SweepOrphanNamespaces scans for and cleans up any stale `skiptui-*` namespaces
// that do not belong to an active, living process.
func SweepOrphanNamespaces() []string {
	var cleaned []string

	cmd := exec.Command("ip", "netns", "list")
	out, err := cmd.Output()
	if err != nil {
		return cleaned
	}

	// Read state to find actively protected namespaces
	activeNamespaces := make(map[string]bool)
	if sessions, err := daemon.LoadSessionState(); err == nil {
		for _, s := range sessions {
			if s.Status == "running" && s.PID > 0 && daemon.IsPIDAlive(s.PID) {
				activeNamespaces[s.Namespace] = true
			}
		}
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 {
			nsName := fields[0]
			if strings.HasPrefix(nsName, "skiptui-") {
				// Don't kill active namespaces
				if activeNamespaces[nsName] {
					continue
				}
				_ = deleteNamedNetns(nsName)
				CleanupNamespaceDNS(nsName)
				cleaned = append(cleaned, nsName)
			}
		}
	}

	return cleaned
}

// RegisterSignalCleanup registers an OS signal listener to ensure sandbox teardown on termination.
func RegisterSignalCleanup(cleanupFunc func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		<-sigChan
		if cleanupFunc != nil {
			cleanupFunc()
		}
		os.Exit(0)
	}()
}
