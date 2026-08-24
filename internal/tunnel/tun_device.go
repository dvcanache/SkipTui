package tunnel

import (
	"fmt"
	"os/exec"
)

// SetupTunInterface creates a TUN interface inside a network namespace and assigns an IP & default route.
func SetupTunInterface(namespace, tunName, ipAddr, gwAddr string) error {
	// 1. Create TUN device in namespace: `ip -n <ns> tuntap add name <tunName> mode tun`
	cmd := exec.Command("ip", "-n", namespace, "tuntap", "add", "name", tunName, "mode", "tun")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create tun device '%s' in netns '%s': %s (%w)", tunName, namespace, string(out), err)
	}

	// 2. Assign IP address: `ip -n <ns> addr add <ipAddr> dev <tunName>`
	cmd = exec.Command("ip", "-n", namespace, "addr", "add", ipAddr, "dev", tunName)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to assign IP to '%s': %s (%w)", tunName, string(out), err)
	}

	// 3. Bring interface UP: `ip -n <ns> link set <tunName> up`
	cmd = exec.Command("ip", "-n", namespace, "link", "set", tunName, "up")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set '%s' up: %s (%w)", tunName, string(out), err)
	}

	// 4. Add default routing rule inside namespace: `ip -n <ns> route add default dev <tunName>`
	cmd = exec.Command("ip", "-n", namespace, "route", "add", "default", "dev", tunName)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add default route dev '%s': %s (%w)", tunName, string(out), err)
	}

	return nil
}

// RemoveTunInterface tears down the TUN device in the namespace.
func RemoveTunInterface(namespace, tunName string) {
	cmd := exec.Command("ip", "-n", namespace, "tuntap", "del", "name", tunName, "mode", "tun")
	_ = cmd.Run()
}
