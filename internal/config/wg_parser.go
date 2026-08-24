package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ParseWireGuard parses a WireGuard .conf file and extracts Interface and Peer information.
func ParseWireGuard(path string) (*Profile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open wireguard conf file: %w", err)
	}
	defer file.Close()

	profile := &Profile{
		ID:         "wg-" + uuid.New().String()[:8],
		Name:       strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
		Protocol:   ProtocolWireGuard,
		KillSwitch: true,
		CreatedAt:  time.Now(),
		WireGuard: &WGConfig{
			ConfigPath: path,
		},
	}

	scanner := bufio.NewScanner(file)
	currentSection := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.ToLower(line[1 : len(line)-1])
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])

		if currentSection == "interface" {
			switch key {
			case "privatekey":
				profile.WireGuard.PrivateKey = val
			case "address":
				profile.WireGuard.Address = val
			case "dns":
				profile.DNS = val
			}
		} else if currentSection == "peer" {
			switch key {
			case "publickey":
				profile.WireGuard.PublicKey = val
			case "presharedkey":
				profile.WireGuard.PresharedKey = val
			case "endpoint":
				profile.Endpoint = val
			case "allowedips":
				ips := strings.Split(val, ",")
				for _, ip := range ips {
					profile.WireGuard.AllowedIPs = append(profile.WireGuard.AllowedIPs, strings.TrimSpace(ip))
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading wireguard conf: %w", err)
	}

	if profile.DNS == "" {
		profile.DNS = "1.1.1.1"
	}

	return profile, nil
}
