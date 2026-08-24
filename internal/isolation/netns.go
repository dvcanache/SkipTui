package isolation

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"skiptui/internal/config"
	"time"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// NetnsEngine implements the Engine interface using Linux Network Namespaces.
type NetnsEngine struct{}

func NewNetnsEngine() *NetnsEngine {
	return &NetnsEngine{}
}

func (e *NetnsEngine) Name() string {
	return "netns"
}

// CheckCapabilities checks if the current process has CAP_NET_ADMIN or root privileges.
func (e *NetnsEngine) CheckCapabilities() error {
	if os.Geteuid() == 0 {
		return nil
	}

	// Test if we can actually manage network namespaces
	cmd := exec.Command("ip", "netns", "add", "skiptui-probe-test")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("netns capability test failed: %s (%w)", string(out), err)
	}
	// Cleanup probe namespace
	_ = exec.Command("ip", "netns", "del", "skiptui-probe-test").Run()

	return nil
}

// CreateSandbox provisions a named network namespace with loopback interface UP.
func (e *NetnsEngine) CreateSandbox(ctx context.Context, id string, profile *config.Profile) (*SandboxInfo, error) {
	nsName := "skiptui-" + id

	// 1. Create the named network namespace
	_ = deleteNamedNetns(nsName)

	if err := createNamedNetns(nsName); err != nil {
		return nil, fmt.Errorf("failed to create netns '%s': %w", nsName, err)
	}

	// 2. Bring up the loopback interface (lo) inside the new namespace
	if err := bringUpLoopback(nsName); err != nil {
		_ = deleteNamedNetns(nsName)
		return nil, fmt.Errorf("failed to bring up loopback in netns '%s': %w", nsName, err)
	}

	// 3. Setup DNS resolution inside namespace
	dns := "1.1.1.1"
	if profile != nil && profile.DNS != "" {
		dns = profile.DNS
	}
	_ = SetupNamespaceDNS(nsName, dns)

	sb := &SandboxInfo{
		ID:         id,
		Namespace:  nsName,
		DNS:        dns,
		IsRootless: false,
		CreatedAt:  time.Now().Unix(),
	}

	return sb, nil
}

// BuildCommand constructs an exec.Cmd configured to run inside the network namespace.
func (e *NetnsEngine) BuildCommand(ctx context.Context, sb *SandboxInfo, targetCmd string, args ...string) (*exec.Cmd, error) {
	var cmdArgs []string
	cmdArgs = append(cmdArgs, "netns", "exec", sb.Namespace, targetCmd)
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.CommandContext(ctx, "ip", cmdArgs...)
	cmd.Env = SanitizeProcessEnv(os.Environ())

	return cmd, nil
}

// DestroySandbox tears down the network namespace and cleans up associated resources.
func (e *NetnsEngine) DestroySandbox(ctx context.Context, sb *SandboxInfo) error {
	CleanupNamespaceDNS(sb.Namespace)
	return deleteNamedNetns(sb.Namespace)
}

// Helpers using netns / ip commands:

func createNamedNetns(name string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cmd := exec.Command("ip", "netns", "add", name)
	if out, err := cmd.CombinedOutput(); err != nil {
		_, newErr := netns.NewNamed(name)
		if newErr != nil {
			return fmt.Errorf("%s: %s (%w)", string(out), newErr.Error(), err)
		}
	}
	return nil
}

func deleteNamedNetns(name string) error {
	cmd := exec.Command("ip", "netns", "del", name)
	_ = cmd.Run()
	_ = netns.DeleteNamed(name)
	return nil
}

func bringUpLoopback(nsName string) error {
	cmd := exec.Command("ip", "-n", nsName, "link", "set", "lo", "up")
	if out, err := cmd.CombinedOutput(); err != nil {
		h, hErr := netns.GetFromName(nsName)
		if hErr != nil {
			return fmt.Errorf("ip link set lo up failed: %s (%w)", string(out), err)
		}
		defer h.Close()

		nlHandle, nlErr := netlink.NewHandleAt(h)
		if nlErr != nil {
			return fmt.Errorf("failed to get netlink handle: %w", nlErr)
		}
		defer nlHandle.Close()

		lo, lErr := nlHandle.LinkByName("lo")
		if lErr != nil {
			return fmt.Errorf("failed to get lo link: %w", lErr)
		}
		return nlHandle.LinkSetUp(lo)
	}
	return nil
}

// NamespaceExists checks if a network namespace exists on the system.
func NamespaceExists(name string) bool {
	if name == "" {
		return false
	}
	if _, err := os.Stat("/var/run/netns/" + name); err == nil {
		return true
	}
	cmd := exec.Command("ip", "netns", "list")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	for _, line := range osLines(string(out)) {
		fields := osFields(line)
		if len(fields) > 0 && fields[0] == name {
			return true
		}
	}
	return false
}

func osLines(s string) []string {
	var lines []string
	curr := ""
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, curr)
			curr = ""
		} else {
			curr += string(s[i])
		}
	}
	if curr != "" {
		lines = append(lines, curr)
	}
	return lines
}

func osFields(s string) []string {
	var fields []string
	curr := ""
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' || s[i] == '\t' || s[i] == '\r' {
			if curr != "" {
				fields = append(fields, curr)
				curr = ""
			}
		} else {
			curr += string(s[i])
		}
	}
	if curr != "" {
		fields = append(fields, curr)
	}
	return fields
}

