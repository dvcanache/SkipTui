package isolation

import (
	"fmt"
	"hash/crc32"
	"os/exec"
	"strings"
)

// VethUplinkInfo holds information about a created veth pair between host and netns.
type VethUplinkInfo struct {
	HostVeth  string
	PeerVeth  string
	SubnetIdx int
	HostIP    string
	PeerIP    string
	Namespace string
}

// SetupVethUplink provisions a virtual ethernet pair connecting the host to a network namespace,
// enabling outbound internet access for VPN daemons running strictly inside the namespace without
// modifying host routes.
func SetupVethUplink(namespace, sessionID string) (*VethUplinkInfo, error) {
	// Generate unique short interface names (<= 15 chars for Linux interface limits)
	cleanID := strings.TrimPrefix(sessionID, "sb-")
	if len(cleanID) > 8 {
		cleanID = cleanID[:8]
	}
	hostVeth := fmt.Sprintf("vh-%s", cleanID)
	peerVeth := fmt.Sprintf("vc-%s", cleanID)

	// Determine unique subnet index (between 1 and 240) based on IEEE CRC32 of sessionID
	checksum := crc32.ChecksumIEEE([]byte(sessionID))
	subnetIdx := int((checksum % 240) + 1)

	hostIP := fmt.Sprintf("10.215.%d.1/30", subnetIdx)
	peerIP := fmt.Sprintf("10.215.%d.2/30", subnetIdx)
	hostGW := fmt.Sprintf("10.215.%d.1", subnetIdx)
	subnetCIDR := fmt.Sprintf("10.215.%d.0/30", subnetIdx)

	// 1. Clean up any existing stale interfaces with this name
	_ = exec.Command("ip", "link", "del", hostVeth).Run()

	// 2. Create the veth pair
	cmd := exec.Command("ip", "link", "add", hostVeth, "type", "veth", "peer", "name", peerVeth)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to create veth pair (%s <-> %s): %s (%w)", hostVeth, peerVeth, string(out), err)
	}

	// 3. Move peer end into namespace
	cmd = exec.Command("ip", "link", "set", peerVeth, "netns", namespace)
	if out, err := cmd.CombinedOutput(); err != nil {
		_ = exec.Command("ip", "link", "del", hostVeth).Run()
		return nil, fmt.Errorf("failed to move %s to netns %s: %s (%w)", peerVeth, namespace, string(out), err)
	}

	// 4. Configure host IP and bring host interface UP
	cmd = exec.Command("ip", "addr", "add", hostIP, "dev", hostVeth)
	if out, err := cmd.CombinedOutput(); err != nil {
		_ = exec.Command("ip", "link", "del", hostVeth).Run()
		return nil, fmt.Errorf("failed to assign IP to %s: %s (%w)", hostVeth, string(out), err)
	}
	_ = exec.Command("ip", "link", "set", hostVeth, "up").Run()

	// 5. Configure namespace peer IP and bring peer interface UP
	cmd = exec.Command("ip", "-n", namespace, "addr", "add", peerIP, "dev", peerVeth)
	if out, err := cmd.CombinedOutput(); err != nil {
		_ = exec.Command("ip", "link", "del", hostVeth).Run()
		return nil, fmt.Errorf("failed to assign IP to %s in netns %s: %s (%w)", peerVeth, namespace, string(out), err)
	}
	_ = exec.Command("ip", "-n", namespace, "link", "set", peerVeth, "up").Run()

	// 6. Add default gateway inside namespace pointing to host interface
	cmd = exec.Command("ip", "-n", namespace, "route", "add", "default", "via", hostGW, "dev", peerVeth)
	_ = cmd.Run()

	// 7. Enable IP forwarding on host
	_ = exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1").Run()

	// 8. Add NAT masquerade rule for the namespace subnet so VPN daemon can reach remote server
	_ = exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", subnetCIDR, "-j", "MASQUERADE").Run()

	return &VethUplinkInfo{
		HostVeth:  hostVeth,
		PeerVeth:  peerVeth,
		SubnetIdx: subnetIdx,
		HostIP:    hostIP,
		PeerIP:    peerIP,
		Namespace: namespace,
	}, nil
}

// CleanupVethUplink removes the veth pair and deletes iptables NAT rules.
func CleanupVethUplink(info *VethUplinkInfo) {
	if info == nil {
		return
	}
	subnetCIDR := fmt.Sprintf("10.215.%d.0/30", info.SubnetIdx)
	_ = exec.Command("iptables", "-t", "nat", "-D", "POSTROUTING", "-s", subnetCIDR, "-j", "MASQUERADE").Run()
	if info.HostVeth != "" {
		_ = exec.Command("ip", "link", "del", info.HostVeth).Run()
	}
}
