package isolation

import (
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

// SweepOrphanNamespaces scans for and cleans up any stale `skiptui-*` namespaces.
func SweepOrphanNamespaces() []string {
	var cleaned []string

	cmd := exec.Command("ip", "netns", "list")
	out, err := cmd.Output()
	if err != nil {
		return cleaned
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 {
			nsName := fields[0]
			if strings.HasPrefix(nsName, "skiptui-") {
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
